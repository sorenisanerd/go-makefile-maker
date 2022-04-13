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
// limitations under the License.

package ghworkflow

import "github.com/sapcc/go-makefile-maker/internal/core"

func dependencyReviewWorkflow(cfg *core.GithubWorkflowConfiguration) error {
	w := newWorkflow("Dependency", cfg.Global.DefaultBranch, nil)
	w.On.Push.Branches = []string{}

	j := baseJob("")
	j.addStep(jobStep{
		Name: "Dependency Review",
		Uses: "actions/dependency-review-action@v1",
	})
	w.Jobs = map[string]job{"Review": j}
	return writeWorkflowToFile(w)
}
