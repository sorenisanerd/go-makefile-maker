// Copyright 2022 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dockerfile

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/must"

	"github.com/sapcc/go-makefile-maker/internal/core"
)

func RenderConfig(cfg core.Configuration) {
	if cfg.Dockerfile.User != "" {
		if cfg.Dockerfile.User == "root" {
			logg.Fatal("the `dockerfile.user` config option has been removed; set `dockerfile.runAsRoot` if you need to run as root")
		} else {
			logg.Fatal("the `dockerfile.user` config option has been removed; commands now run as user `appuser` (ID 4200) in group `appgroup` (ID 4200)")
		}
	}

	var goBuildflags, packages, userCommand, entrypoint, workingDir, addUserGroup, extraCommands string

	if cfg.Golang.EnableVendoring {
		goBuildflags = ` GO_BUILDFLAGS='-mod vendor'`
	}

	for _, v := range append([]string{"ca-certificates"}, cfg.Dockerfile.ExtraPackages...) {
		packages += fmt.Sprintf(" %s", v)
	}

	if cfg.Dockerfile.RunAsRoot {
		userCommand = ""
		workingDir = "/"
		addUserGroup = ""
	} else {
		// this is the same as `USER appuser:appgroup`, but using numeric IDs
		// should allow Kubernetes to validate the `runAsNonRoot` rule without
		// requiring an explicit `runAsUser: 4200` setting in the container spec
		userCommand = "USER 4200:4200\n"
		workingDir = "/home/appuser"
		addUserGroup = `RUN addgroup -g 4200 appgroup \
  && adduser -h /home/appuser -s /sbin/nologin -G appgroup -D -u 4200 appuser

`
	}

	// if there is an entrypoint configured use that otherwise fallback to the binary name
	if len(cfg.Dockerfile.Entrypoint) > 0 {
		entrypoint = fmt.Sprintf(`"%s"`, strings.Join(cfg.Dockerfile.Entrypoint, `", "`))
	} else {
		entrypoint = fmt.Sprintf(`"/usr/bin/%s"`, cfg.Binaries[0].Name)
	}

	if cfg.Dockerfile.WithLinkerdAwait {
		extraCommands = fmt.Sprintf(`
RUN wget -qO /usr/bin/linkerd-await https://github.com/linkerd/linkerd-await/releases/download/release%%2Fv%[1]s/linkerd-await-v%[1]s-amd64 \
  && chmod 755 /usr/bin/linkerd-await
`, core.DefaultLinkerdAwaitVersion)

		// add linkrd-await after the fallback for entrypoint has been set
		entrypoint = `"/usr/bin/linkerd-await", "--shutdown", "--", ` + entrypoint
	}

	extraDirectives := strings.Join(cfg.Dockerfile.ExtraDirectives, "\n")
	if extraDirectives != "" {
		extraDirectives += "\n"
	}

	dockerfile := fmt.Sprintf(
		`FROM golang:%[1]s%[2]s as builder

RUN apk add --no-cache --no-progress gcc git make musl-dev

COPY . /src
ARG BININFO_BUILD_DATE BININFO_COMMIT_HASH BININFO_VERSION # provided to 'make install'
RUN make -C /src install PREFIX=/pkg GOTOOLCHAIN=local%[3]s

################################################################################

FROM alpine:%[2]s

%[4]s# upgrade all installed packages to fix potential CVEs in advance
# also remove apk package manager to hopefully remove dependecy on openssl 🤞
RUN apk upgrade --no-cache --no-progress \
  && apk add --no-cache --no-progress%[5]s \
  && apk del --no-cache --no-progress apk-tools alpine-keys
%[6]s
COPY --from=builder /pkg/ /usr/

ARG BININFO_BUILD_DATE BININFO_COMMIT_HASH BININFO_VERSION
LABEL source_repository="%[7]s" \
  org.opencontainers.image.url="%[7]s" \
  org.opencontainers.image.created=${BININFO_BUILD_DATE} \
  org.opencontainers.image.revision=${BININFO_COMMIT_HASH} \
  org.opencontainers.image.version=${BININFO_VERSION}

%[8]s%[9]sWORKDIR %[10]s
ENTRYPOINT [ %[11]s ]
`, core.DefaultGolangImagePrefix, core.DefaultAlpineImage, goBuildflags, addUserGroup, packages, extraCommands, cfg.Metadata.URL, extraDirectives, userCommand, workingDir, entrypoint)

	must.Succeed(os.WriteFile("Dockerfile", []byte(dockerfile), 0666))

	dockerignoreLines := append([]string{
		`.dockerignore`,
		`# TODO: uncomment when applications no longer use git to get version information`,
		`#.git/`,
		`.github/`,
		`.gitignore`,
		`.goreleaser.yml`,
		`/*.env*`,
		`.golangci.yaml`,
		`build/`,
		`CONTRIBUTING.md`,
		`Dockerfile`,
		`docs/`,
		`LICENSE*`,
		`Makefile.maker.yaml`,
		`README.md`,
		`report.html`,
		`shell.nix`,
		`/testing/`,
	}, cfg.Dockerfile.ExtraIgnores...)
	dockerignore := strings.Join(dockerignoreLines, "\n") + "\n"

	must.Succeed(os.WriteFile(".dockerignore", []byte(dockerignore), 0666))
}
