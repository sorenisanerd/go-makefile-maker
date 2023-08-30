// Copyright 2021 SAP SE
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

package ghworkflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sapcc/go-bits/must"

	"github.com/sapcc/go-makefile-maker/internal/core"
)

const workflowDir = ".github/workflows"

// Render renders GitHub workflows.
func Render(cfg *core.Configuration) {
	ghwCfg := cfg.GitHubWorkflow

	must.Succeed(os.MkdirAll(workflowDir, 0o755))

	// remove renamed files
	must.Succeed(os.RemoveAll(filepath.Join(workflowDir, "dependency-review.yaml")))
	must.Succeed(os.RemoveAll(filepath.Join(workflowDir, "license.yaml")))
	must.Succeed(os.RemoveAll(filepath.Join(workflowDir, "spell.yaml")))

	checksWorkflow(ghwCfg, cfg.SpellCheck.IgnoreWords)

	if ghwCfg.CI.Enabled {
		ciWorkflow(ghwCfg, cfg.Golang.EnableVendoring, len(cfg.Binaries) > 0)
	}
	if ghwCfg.SecurityChecks.Enabled {
		codeQLWorkflow(ghwCfg)
	}
	if ghwCfg.PushContainerToGhcr.Enabled {
		ghcrWorkflow(ghwCfg)
	}
}

func writeWorkflowToFile(w *workflow) {
	name := strings.ToLower(strings.ReplaceAll(w.Name, " ", "-"))
	path := filepath.Join(workflowDir, name+".yaml")
	f := must.Return(os.Create(path))
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	defer encoder.Close()
	encoder.SetIndent(2)

	fmt.Fprintln(f, core.AutogeneratedHeader)
	fmt.Fprintln(f, "")
	must.Succeed(encoder.Encode(w))
}
