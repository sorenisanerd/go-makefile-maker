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

package renovate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sapcc/go-bits/must"

	"github.com/sapcc/go-makefile-maker/internal/core"
)

type constraints struct {
	Go string `json:"go"`
}

type config struct {
	Extends             []string           `json:"extends"`
	Assignees           []string           `json:"assignees,omitempty"`
	CommitMessageAction string             `json:"commitMessageAction"`
	Constraints         constraints        `json:"constraints"`
	PostUpdateOptions   []string           `json:"postUpdateOptions"`
	PackageRules        []core.PackageRule `json:"packageRules,omitempty"`
	PrHourlyLimit       int                `json:"prHourlyLimit"`
	Schedule            []string           `json:"schedule"`
	SemanticCommits     string             `json:"semanticCommits,omitempty"`
}

func (c *config) addPackageRule(rule core.PackageRule) {
	c.PackageRules = append(c.PackageRules, rule)
}

func RenderConfig(cfgRenovate core.RenovateConfig, scanResult core.ScanResult, url string, isApplicationRepo bool) {
	isGoMakefileMakerRepo := scanResult.MustModulePath() == "github.com/sapcc/go-makefile-maker"
	isInternalRenovate := strings.HasPrefix(url, "https://github.wdf.sap.corp")

	// Our default rule is to have Renovate send us PRs once a week. (More
	// frequent PRs become too overwhelming with the sheer amount of repos that
	// we manage.) Friday works well for this because when we merge, we can let
	// the changes simmer in QA over the weekend, and then have high confidence
	// when we deploy these updates on Monday.
	schedule := "before 8am on Friday"
	if isInternalRenovate {
		schedule = "on Friday"
	}
	// However, for pure library repos, we do the PRs on Thursday instead, so
	// that the dependency updates in these library repos trickle down into the
	// application repos without an extra week of delay.
	if !isApplicationRepo {
		schedule = "before 8am on Thursday"
		if isInternalRenovate {
			schedule = "on Thursday"
		}
	}

	cfg := config{
		Extends: []string{
			"config:base",
			"default:pinDigestsDisabled",
			"github>whitesource/merge-confidence:beta",
		},
		Assignees: cfgRenovate.Assignees,
		// CommitMessageAction is the verb that appears at the start of Renovate's
		// commit messages (and therefore, PR titles). The default value is "Update".
		// We choose something more specific because some of us have filter rules
		// in their mail client to separate Renovate PRs from other PRs.
		CommitMessageAction: "Renovate: Update",
		Constraints: constraints{
			Go: cfgRenovate.GoVersion,
		},
		PostUpdateOptions: []string{
			"gomodUpdateImportPaths",
		},
		PrHourlyLimit:   0,
		Schedule:        []string{schedule},
		SemanticCommits: "disabled",
	}
	if cfgRenovate.GoVersion == "1.17" {
		cfg.PostUpdateOptions = append([]string{"gomodTidy1.17"}, cfg.PostUpdateOptions...)
	} else {
		cfg.PostUpdateOptions = append([]string{"gomodTidy"}, cfg.PostUpdateOptions...)
	}

	// Default package rules.
	//
	// NOTE: When changing this list, please also adjust the documentation for
	// default package rules in the README.
	cfg.addPackageRule(core.PackageRule{
		MatchPackageNames: []string{"golang"},
		AllowedVersions:   fmt.Sprintf("%s.x", cfgRenovate.GoVersion),
	})

	// combine and automerge all dependencies under github.com/sapcc/
	cfg.addPackageRule(core.PackageRule{
		MatchPackagePatterns: []string{`^github\.com\/sapcc\/.*`},
		GroupName:            "github.com/sapcc",
		AutoMerge:            true,
	})

	// combine all dependencies not under github.com/sapcc/
	cfg.addPackageRule(core.PackageRule{
		MatchPackagePatterns:   []string{`.*`},
		ExcludePackagePatterns: []string{`^github\.com\/sapcc\/.*`},
		GroupName:              "External dependencies",
		AutoMerge:              false,
	})

	// Only enable Dockerfile and github-actions updates for go-makefile-maker itself.
	if isGoMakefileMakerRepo {
		cfg.Extends = append(cfg.Extends, "docker:enableMajor", "regexManagers:dockerfileVersions")
	} else {
		cfg.Extends = append(cfg.Extends, "docker:disable")
	}
	hasK8sIOPkgs := false
	for _, v := range scanResult.GoDirectDependencies {
		switch dep := v.Path; {
		case strings.HasPrefix(dep, "k8s.io/"):
			hasK8sIOPkgs = true
		}
	}
	if hasK8sIOPkgs {
		cfg.addPackageRule(core.PackageRule{
			MatchPackagePrefixes: []string{"k8s.io/"},
			// Since our clusters use k8s v1.26 and k8s has a support policy of -/+ 1 minor version we set the allowedVersions to `0.27.x`.
			// k8s.io/* deps use v0.x.y instead of v1.x.y therefore we use 0.x instead of 1.x.
			// Ref: https://docs.renovatebot.com/configuration-options/#allowedversions
			AllowedVersions: "0.27.x",
		})
	}

	// Custom package rules specified in config.
	//
	// Renovate will evaluate all packageRules and not stop once it gets a first match
	// therefore the packageRules should be in the order of importance so that user
	// defined rules can override settings from earlier rules.
	for _, rule := range cfgRenovate.PackageRules {
		cfg.addPackageRule(rule)
	}

	must.Succeed(os.MkdirAll(".github", 0750))
	f := must.Return(os.Create(".github/renovate.json"))

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // in order to preserve `<` in allowedVersions field
	must.Succeed(encoder.Encode(cfg))

	must.Succeed(f.Close())
}
