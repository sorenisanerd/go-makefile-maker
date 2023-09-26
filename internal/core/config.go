/******************************************************************************
*
*  Copyright 2020 SAP SE
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
*
******************************************************************************/

package core

import (
	"os/exec"
	"strings"

	"github.com/sapcc/go-bits/logg"

	"github.com/sapcc/go-makefile-maker/internal/renovate"
)

var AutogeneratedHeader = strings.TrimSpace(`
################################################################################
# This file is AUTOGENERATED with <https://github.com/sapcc/go-makefile-maker> #
# Edit Makefile.maker.yaml instead.                                            #
################################################################################
`)

///////////////////////////////////////////////////////////////////////////////
// Core configuration

// Configuration is the data structure that we read from the input file.
type Configuration struct {
	Verbatim       string                       `yaml:"verbatim"`
	VariableValues map[string]string            `yaml:"variables"`
	Binaries       []BinaryConfiguration        `yaml:"binaries"`
	Test           TestConfiguration            `yaml:"testPackages"`
	Coverage       CoverageConfiguration        `yaml:"coverageTest"`
	Golang         GolangConfiguration          `yaml:"golang"`
	GolangciLint   GolangciLintConfiguration    `yaml:"golangciLint"`
	Goreleaser     GoreleaserConfiguration      `yaml:"goreleaser"`
	SpellCheck     SpellCheckConfiguration      `yaml:"spellCheck"`
	GitHubWorkflow *GithubWorkflowConfiguration `yaml:"githubWorkflow"`
	Makefile       MakefileConfig               `yaml:"makefile"`
	Renovate       RenovateConfig               `yaml:"renovate"`
	Dockerfile     DockerfileConfig             `yaml:"dockerfile"`
	Metadata       Metadata                     `yaml:"metadata"`
}

// Variable returns the value of this variable if it's overridden in the config,
// or the default value otherwise.
func (c Configuration) Variable(name, defaultValue string) string {
	value, exists := c.VariableValues[name]
	if exists {
		return " " + value
	}
	if defaultValue == "" {
		return ""
	}
	return " " + defaultValue
}

// BinaryConfiguration appears in type Configuration.
type BinaryConfiguration struct {
	Name        string `yaml:"name"`
	FromPackage string `yaml:"fromPackage"`
	InstallTo   string `yaml:"installTo"`
}

// TestConfiguration appears in type Configuration.
type TestConfiguration struct {
	Only   string `yaml:"only"`
	Except string `yaml:"except"`
}

// CoverageConfiguration appears in type Configuration.
type CoverageConfiguration struct {
	Only   string `yaml:"only"`
	Except string `yaml:"except"`
}

// GolangConfiguration appears in type Configuration.
type GolangConfiguration struct {
	EnableVendoring bool `yaml:"enableVendoring"`
	SetGoModVersion bool `yaml:"setGoModVersion"`
}

// GolangciLintConfiguration appears in type Configuration.
type GolangciLintConfiguration struct {
	CreateConfig     bool     `yaml:"createConfig"`
	ErrcheckExcludes []string `yaml:"errcheckExcludes"`
	SkipDirs         []string `yaml:"skipDirs"`
}

type GoreleaserConfiguration struct {
	Enabled bool `yaml:"enabled"`
}

// SpellCheckConfiguration appears in type Configuration.
type SpellCheckConfiguration struct {
	IgnoreWords []string `yaml:"ignoreWords"`
}

///////////////////////////////////////////////////////////////////////////////
// GitHub workflow configuration

// GithubWorkflowConfiguration appears in type Configuration.
type GithubWorkflowConfiguration struct {
	// These global-level settings are applicable for all workflows. They are
	// superseded by their workflow-level counterpart(s).
	Global struct {
		DefaultBranch string `yaml:"defaultBranch"`
		GoVersion     string `yaml:"goVersion"`
	} `yaml:"global"`

	CI                  CIWorkflowConfig             `yaml:"ci"`
	IsSelfHostedRunner  bool                         `yaml:"omit"`
	License             LicenseWorkflowConfig        `yaml:"license"`
	PushContainerToGhcr PushContainerToGhcrConfig    `yaml:"pushContainerToGhcr"`
	Release             ReleaseWorkflowConfig        `yaml:"release"`
	SpellCheck          SpellCheckWorkflowConfig     `yaml:"spellCheck"`
	SecurityChecks      SecurityChecksWorkflowConfig `yaml:"securityChecks"`
}

// CIWorkflowConfig appears in type Configuration.
type CIWorkflowConfig struct {
	Enabled     bool     `yaml:"enabled"`
	IgnorePaths []string `yaml:"ignorePaths"`
	RunnerType  []string `yaml:"runOn"`
	Coveralls   bool     `yaml:"coveralls"`
	Postgres    struct {
		Enabled bool   `yaml:"enabled"`
		Version string `yaml:"version"`
	} `yaml:"postgres"`
	KubernetesEnvtest struct {
		Enabled bool   `yaml:"enabled"`
		Version string `yaml:"version"`
	} `yaml:"kubernetesEnvtest"`
}

// LicenseWorkflowConfig appears in type Configuration.
type LicenseWorkflowConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Patterns       []string `yaml:"patterns"`
	IgnorePatterns []string `yaml:"ignorePatterns"`
}

type PushContainerToGhcrConfig struct {
	Enabled bool `yaml:"enabled"`
}

type ReleaseWorkflowConfig struct {
	Enabled bool `yaml:"enabled"`
}

// SpellCheckWorkflowConfig appears in type Configuration.
type SpellCheckWorkflowConfig struct {
	Enabled bool `yaml:"enabled"`
}

// SecurityChecksWorkflowConfig appears in type Configuration.
type SecurityChecksWorkflowConfig struct {
	Enabled bool `yaml:"enabled"`
}

// RenovateConfig appears in type Configuration.
type RenovateConfig struct {
	Enabled      bool                   `yaml:"enabled"`
	Assignees    []string               `yaml:"assignees"`
	GoVersion    string                 `yaml:"goVersion"`
	PackageRules []renovate.PackageRule `yaml:"packageRules"`
}

// DockerfileConfig appears in type Configuration.
type DockerfileConfig struct {
	Enabled          bool     `yaml:"enabled"`
	Entrypoint       []string `yaml:"entrypoint"`
	ExtraDirectives  []string `yaml:"extraDirectives"`
	ExtraIgnores     []string `yaml:"extraIgnores"`
	ExtraPackages    []string `yaml:"extraPackages"`
	RunAsRoot        bool     `yaml:"runAsRoot"`
	User             string   `yaml:"user"` //obsolete; will produce an error when used
	WithLinkerdAwait bool     `yaml:"withLinkerdAwait"`
}

type MakefileConfig struct {
	Enabled *bool `yaml:"enabled"` // this is a pointer to bool to treat an absence as true for backwards compatibility
}

type Metadata struct {
	URL string `yaml:"url"`
}

///////////////////////////////////////////////////////////////////////////////
// Helper functions

func (c *Configuration) Validate() {
	if c.Dockerfile.Enabled {
		if c.Metadata.URL == "" {
			logg.Fatal("metadata.url must be set when docker.enabled is true")
		}
	}

	// Validate GolangciLintConfiguration.
	if len(c.GolangciLint.ErrcheckExcludes) > 0 && !c.GolangciLint.CreateConfig {
		logg.Fatal("golangciLint.createConfig must be set to 'true' if golangciLint.errcheckExcludes is defined")
	}

	// Validate GithubWorkflowConfiguration.
	ghwCfg := c.GitHubWorkflow
	if ghwCfg != nil {
		if c.Metadata.URL == "" {
			logg.Fatal("metadata.url must be set when any github workflow is configured otherwise it cannot be determined which github runner type should be used")
		}

		// Validate global options.
		if ghwCfg.Global.DefaultBranch == "" {
			errMsg := "could not find default branch using git, you can define manually be setting 'githubWorkflow.global.defaultBranch' in config"
			b, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").CombinedOutput()
			if err != nil {
				logg.Fatal("%s: %s", errMsg, err.Error())
			}

			branch := strings.TrimPrefix(string(b), "refs/remotes/origin/")
			if branch == string(b) {
				logg.Fatal(errMsg)
			} else {
				c.GitHubWorkflow.Global.DefaultBranch = strings.TrimSpace(branch)
			}
		}

		// Validate CI workflow configuration.
		if ghwCfg.CI.Postgres.Enabled || ghwCfg.CI.KubernetesEnvtest.Enabled {
			if !ghwCfg.CI.Enabled {
				logg.Fatal("githubWorkflow.ci.enabled must be set to 'true' when githubWorkflow.ci.postgres or githubWorkflow.ci.kubernetesEnvtest is enabled")
			}
			if len(ghwCfg.CI.RunnerType) > 0 {
				if len(ghwCfg.CI.RunnerType) > 1 || !strings.HasPrefix(ghwCfg.CI.RunnerType[0], "ubuntu") {
					logg.Fatal("githubWorkflow.ci.runOn must only define a single Ubuntu based runner when githubWorkflow.ci.postgres or githubWorkflow.ci.kubernetesEnvtest is enabled")
				}
			}
		}
	}
}
