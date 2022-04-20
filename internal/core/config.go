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
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var AutogeneratedHeader = strings.TrimSpace(`
################################################################################
# This file is AUTOGENERATED with <https://github.com/sapcc/go-makefile-maker> #
# Edit Makefile.maker.yaml instead.                                            #
################################################################################
`)

///////////////////////////////////////////////////////////////////////////////
// Core configuration

//Configuration is the data structure that we read from the input file.
type Configuration struct {
	Verbatim       string                       `yaml:"verbatim"`
	VariableValues map[string]string            `yaml:"variables"`
	Binaries       []BinaryConfiguration        `yaml:"binaries"`
	Coverage       CoverageConfiguration        `yaml:"coverageTest"`
	Vendoring      VendoringConfiguration       `yaml:"vendoring"`
	GolangciLint   GolangciLintConfiguration    `yaml:"golangciLint"`
	SpellCheck     SpellCheckConfiguration      `yaml:"spellCheck"`
	GitHubWorkflow *GithubWorkflowConfiguration `yaml:"githubWorkflow"`
	Renovate       RenovateConfig               `yaml:"renovate"`
}

//Variable returns the value of this variable if it's overridden in the config,
//or the default value otherwise.
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

//BinaryConfiguration appears in type Configuration.
type BinaryConfiguration struct {
	Name        string `yaml:"name"`
	FromPackage string `yaml:"fromPackage"`
	InstallTo   string `yaml:"installTo"`
}

//CoverageConfiguration appears in type Configuration.
type CoverageConfiguration struct {
	Only   string `yaml:"only"`
	Except string `yaml:"except"`
}

//VendoringConfiguration appears in type Configuration.
type VendoringConfiguration struct {
	Enabled bool `yaml:"enabled"`
}

//GolangciLintConfiguration appears in type Configuration.
type GolangciLintConfiguration struct {
	CreateConfig     bool     `yaml:"createConfig"`
	ErrcheckExcludes []string `yaml:"errcheckExcludes"`
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
		commonWorkflowConfigOpts `yaml:",inline"`

		Assignees     []string `yaml:"assignees"`
		DefaultBranch string   `yaml:"defaultBranch"`
		GoVersion     string   `yaml:"goVersion"`
	} `yaml:"global"`

	CI             CIWorkflowConfig             `yaml:"ci"`
	License        LicenseWorkflowConfig        `yaml:"license"`
	SpellCheck     SpellCheckWorkflowConfig     `yaml:"spellCheck"`
	SecurityChecks SecurityChecksWorkflowConfig `yaml:"securityChecks"`
}

// CIWorkflowConfig appears in type Configuration.
type CIWorkflowConfig struct {
	commonWorkflowConfigOpts `yaml:",inline"`

	Enabled      bool     `yaml:"enabled"`
	RunnerOSList []string `yaml:"runOn"`
	Coveralls    bool     `yaml:"coveralls"`
	Postgres     struct {
		Enabled bool   `yaml:"enabled"`
		Version string `yaml:"version"`
	} `yaml:"postgres"`
}

// LicenseWorkflowConfig appears in type Configuration.
type LicenseWorkflowConfig struct {
	commonWorkflowConfigOpts `yaml:",inline"`

	Enabled  bool     `yaml:"enabled"`
	Patterns []string `yaml:"patterns"`
}

// SpellCheckWorkflowConfig appears in type Configuration.
type SpellCheckWorkflowConfig struct {
	commonWorkflowConfigOpts `yaml:",inline"`

	Enabled bool `yaml:"enabled"`
}

// SecurityChecksWorkflowConfig appears in type Configuration.
type SecurityChecksWorkflowConfig struct {
	Enabled bool `yaml:"enabled"`
}

// commonWorkflowConfigOpts holds common configuration options that are applicable for all
// workflows.
type commonWorkflowConfigOpts struct {
	IgnorePaths []string `yaml:"ignorePaths"`
}

// RenovateConfig appears in type Configuration.
type RenovateConfig struct {
	Enabled   bool   `yaml:"enabled"`
	GoVersion string `yaml:"goVersion"`
}

///////////////////////////////////////////////////////////////////////////////
// Helper functions

func (c *Configuration) Validate() error {
	// Validate GolangciLintConfiguration.
	if len(c.GolangciLint.ErrcheckExcludes) > 0 && !c.GolangciLint.CreateConfig {
		return errors.New("golangciLint.createConfig needs to be set to 'true' if golangciLint.errcheckExcludes is defined")
	}

	// Validate GithubWorkflowConfiguration.
	ghwCfg := c.GitHubWorkflow
	if ghwCfg != nil {
		// Validate global options.
		if ghwCfg.Global.DefaultBranch == "" {
			errMsg := "could not find default branch using git, you can define manually be setting 'githubWorkflow.global.defaultBranch' in config"
			b, err := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD").CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s: %s", errMsg, err.Error())
			}

			branch := strings.TrimPrefix(string(b), "refs/remotes/origin/")
			if branch == string(b) {
				return errors.New(errMsg)
			} else {
				c.GitHubWorkflow.Global.DefaultBranch = strings.TrimSpace(branch)
			}
		}

		// Validate CI workflow configuration.
		if ghwCfg.CI.Postgres.Enabled {
			if !ghwCfg.CI.Enabled {
				return errors.New("githubWorkflow.ci.enabled needs to be set to 'true' when githubWorkflow.ci.postgres.enabled is 'true'")
			}
			if len(ghwCfg.CI.RunnerOSList) > 0 {
				if len(ghwCfg.CI.RunnerOSList) > 1 || !strings.HasPrefix(ghwCfg.CI.RunnerOSList[0], "ubuntu") {
					return errors.New("githubWorkflow.ci.runOn must only define a single Ubuntu based runner when githubWorkflow.ci.postgres.enabled is 'true'")
				}
			}
		}
	}

	return nil
}
