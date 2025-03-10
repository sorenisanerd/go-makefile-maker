// Copyright 2023 SAP SE
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

package core

const (
	DefaultAlpineImage       = "3.18"
	DefaultGolangImagePrefix = "1.21.4-alpine"

	DefaultGoVersion           = "1.21"
	DefaultPostgresVersion     = "12"
	DefaultLinkerdAwaitVersion = "0.2.7"
	DefaultK8sEnvtestVersion   = "1.26.x!"
	DefaultGitHubComRunnerType = "ubuntu-latest"
)

var DefaultGitHubEnterpriseRunnerType = [...]string{"self-hosted", "Linux", "X64"}

const (
	CacheAction            = "actions/cache@v3"
	CheckoutAction         = "actions/checkout@v4"
	SetupGoAction          = "actions/setup-go@v4"
	DependencyReviewAction = "actions/dependency-review-action@v3"

	DockerLoginAction     = "docker/login-action@v3"
	DockerMetadataAction  = "docker/metadata-action@v5"
	DockerBuildPushAction = "docker/build-push-action@v5"

	CodeqlInitAction      = "github/codeql-action/init@v2"
	CodeqlAnalyzeAction   = "github/codeql-action/analyze@v2"
	CodeqlAutobuildAction = "github/codeql-action/autobuild@v2"

	GolangciLintAction = "golangci/golangci-lint-action@v3"
	GoreleaserAction   = "goreleaser/goreleaser-action@v5"
	GovulncheckAction  = "golang/govulncheck-action@v1"
	MisspellAction     = "reviewdog/action-misspell@v1"
)
