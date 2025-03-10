/******************************************************************************
*
*  Copyright 2022 SAP SE
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
	"os"

	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/must"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// ScanResult contains data obtained through a scan of the configuration files
// in the repository. At the moment, only `go.mod` is scanned.
type ScanResult struct {
	ModulePath           string           // from "module" directive in go.mod, e.g. "github.com/foo/bar"
	GoVersion            string           // from "go" directive in go.mod, e.g. "1.17"
	GoDirectDependencies []module.Version // from "require" directive(s) in go.mod without the "// indirect" comment
	HasBinInfo           bool             // whether we can produce linker instructions for "github.com/sapcc/go-api-declarations/bininfo"
	UsesPostgres         bool             // wether postgres is used
}

const ModFilename = "go.mod"

func Scan() ScanResult {
	modFileBytes := must.Return(os.ReadFile(ModFilename))
	modFile := must.Return(modfile.Parse(ModFilename, modFileBytes, nil))

	var goDeps []module.Version
	hasBinInfo := false
	usesPostgres := false

	for _, v := range modFile.Require {
		if !v.Indirect {
			goDeps = append(goDeps, v.Mod)
		}
		if v.Mod.Path == "github.com/sapcc/go-api-declarations" {
			if semver.Compare(v.Mod.Version, "v1.2.0") >= 0 {
				hasBinInfo = true
			}
		}
		if v.Mod.Path == "github.com/lib/pq" {
			usesPostgres = true
		}
	}

	return ScanResult{
		GoVersion:            modFile.Go.Version,
		ModulePath:           modFile.Module.Mod.Path,
		GoDirectDependencies: goDeps,
		HasBinInfo:           hasBinInfo,
		UsesPostgres:         usesPostgres,
	}
}

// MustModulePath reads the ModulePath field, but fails if it is empty.
func (sr ScanResult) MustModulePath() string {
	if sr.ModulePath == "" {
		logg.Fatal("could not find module path from go.mod file, make sure it is defined")
	}
	return sr.ModulePath
}
