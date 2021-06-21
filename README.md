# go-makefile-maker

[![CI](https://github.com/sapcc/go-makefile-maker/actions/workflows/ci.yaml/badge.svg)](https://github.com/sapcc/go-makefile-maker/actions/workflows/ci.yaml)

Generates a Makefile and GitHub workflows for your Go application:

* Makefile follows established Unix conventions for installing and packaging,
  and includes targets for vendoring, running tests and checking code quality.
* GitHub Workflows use [GitHub Actions](https://github.com/features/actions) to
  lint, build, and test your code. Additionally, they can also check your
  codebase for spelling errors and missing license headers.

## Installation

The easiest way is `go get github.com/sapcc/go-makefile-maker`.

We also support the usual Makefile invocations: `make`, `make check` and `make install`. The latter understands the conventional environment variables for choosing install locations, `DESTDIR` and `PREFIX`.
You usually want something like `make && sudo make install PREFIX=/usr/local`.

## Usage

Put a `Makefile.maker.yaml` in your Git repository's root directory, then run `go-makefile-maker` to render the Makefile and GitHub workflows from it.
Commit both the `Makefile.maker.yaml` and the Makefile, so that your users don't need to have `go-makefile-maker` installed.

The `Makefile.maker.yaml` is a YAML file with the following sections:

* [binaries](#binaries)
* [coverageTest](#coveragetest)
* [variables](#variables)
* [vendoring](#vendoring)
* [staticCheck](#staticcheck)
* [verbatim](#verbatim)
* [githubWorkflows](#githubworkflows)
  * [githubWorkflows\.global](#githubworkflowsglobal)
  * [githubWorkflows\.ci](#githubworkflowsci)
  * [githubWorkflows\.license](#githubworkflowslicense)
  * [githubWorkflows\.spellCheck](#githubworkflowsspellcheck)

### `binaries`

```yaml
binaries:
  - name:        example
    fromPackage: ./cmd/example
    installTo:   bin/
  - name:        test-helper
    fromPackage: ./cmd/test-helper
```

For each binary specified here, a target will be generated that builds it with `go build` and puts it in `build/$NAME`.
The `fromPackage` is a Go module path relative to the directory containing the Makefile.

If `installTo` is set for at least one binary, the `install` target is added to the Makefile, and all binaries with `installTo` are installed by it.
In this case, `example` would be installed as `/usr/bin/example` by default, and `test-helper` would not be installed.

### `coverageTest`

```yaml
coverageTest:
  only: '/internal'
  except: '/test/util|/test/mock'
```

When `make check` runs `go test`, it produces a test coverage report.
By default, all packages inside the repository are subject to coverage testing, but this section can be used to restrict this.
The values in `only` and `except` are regexes for `grep -E`.
Since only entire packages (not single source files) can be selected for coverage testing, the regexes have to match package names, not on file names.

### `variables`

```yaml
variables:
  GO_BUILDFLAGS: '-mod vendor'
  GO_LDFLAGS: ''
  GO_TESTENV: ''
```

Allows to override the default values of Makefile variables used by the autogenerated recipes.
This mechanism cannot be used to define new variables to use in your own rules; use `verbatim` for that.
By default, all accepted variables are empty.
The only exception is that `GO_BUILDFLAGS` defaults to `-mod vendor` when vendoring is enabled (see below).

A typical usage of this is to give compile-time values to the Go compiler with the `-X` linker flag:

```yaml
variables:
  GO_LDFLAGS: '-X github.com/foo/bar.Version = $(shell git describe --abbrev=7)'
```

`GO_TESTENV` can contain environment variables to pass to `go test`:

```yaml
variables:
  GO_TESTENV: 'POSTGRES_HOST=localhost POSTGRES_DATABASE=unittestdb'
```

### `vendoring`

```yaml
vendoring:
  enabled: false
```

Set `vendoring.enabled` to `true` if you vendor all dependencies in your repository. With vendoring enabled:

1. The default for `GO_BUILDFLAGS` is set to `-mod vendor`, so that build targets default to using vendored dependencies.
   This means that building binaries does not require a network connection.
2. The `make tidy-deps` target is replaced by a `make vendor` target that runs `go mod tidy && go mod verify` just like `make tidy-deps`, but also runs `go
   mod vendor`.
   This target can be used to get the vendor directory up-to-date before commits.

### `staticCheck`

```yaml
staticCheck:
  golangciLint: false
```

Set `staticCheck.golangciLint` to `true`, if you want to use [`golangci-lint`](https://golangci-lint.run/) for static checking instead of `gofmt`, `golint`, and `go vet`.

### `verbatim`

```yaml
verbatim: |
  run-example: build/example
    ./build/example example-config.txt
```

This field can be used to add your own definitions and rules to the Makefile.
The text in this field is copied into the Makefile mostly verbatim, with one exception:
Since YAML does not like tabs for indentation, we allow rule recipes to be indented with spaces.
This indentation will be replaced with tabs before writing it into the actual Makefile.

### `githubWorkflows`

This is how a minimal and complete workflow configuration would look like:

```yaml
githubWorkflows:
  global:
    ignorePaths:
      - "**.md" # all Markdown files
  ci:
    enabled: true
    coveralls: true
  license:
    enabled: true
    patterns: [ "**.go" ] # all Go files
  spellCheck:
    enabled: true
    ignorePaths: [] # override global setting so that nothing is ignored
```

#### `githubWorkflows.global`

This section defines global settings that apply to all workflows. If the same
setting is also defined for a specific workflow then that will override the
global value.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `githubWorkflows.global.defaultBranch` | string | Value of `git symbolic-ref refs/remotes/origin/HEAD \| sed 's@^refs/remotes/origin/@@'` | Git branch on which `push` actions will trigger the workflows. Pull requests will automatically trigger all workflows. |
| `githubWorkflows.global.ignorePaths` | list | *(optional)* | A list of filename patterns. Workflows will not trigger if a path name matches pattern in this list. [More info][ref-onpushpull] and [filter pattern cheat sheet][ref-pattern-cheat-sheet]. |

[ref-onpushpull]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onpushpull_requestpaths
[ref-pattern-cheat-sheet]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet

#### `githubWorkflows.ci`

This workflow:

* checks your code using `gofmt`, `golint`, and `go vet` (or `golangci-lint` if `staticCheck.golangciLint` is `true`)
* ensures that your code compiles successfully
* runs tests and generates test coverage report
* uploads the test coverage report to [Coveralls](https://coveralls.io) (you will need to enable Coveralls for your repo).

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `githubWorkflows.ci.enabled` | boolean | `false` | Enables generation of this workflow. |
| `githubWorkflows.ci.goVersion` | string | Go version in `go.mod` file | Specify the Go version to use for CI jobs (`lint`, `build`, `test`). |
| `githubWorkflows.ci.runOn` | list | `ubuntu-latest` | The type of machine(s) to run the `build` and `test` job on ([more info][ref-runs-on]). Use this to ensure that your build compilation and tests are successful on multiple operating systems. |
| `githubWorkflows.ci.coveralls` | boolean | `false` | Enables sending the test coverage report to Coveralls. |
| `githubWorkflows.ci.ignorePaths` | list | *(optional)* | Refer to the description for `githubWorkflows.global.ignorePaths`. |

You can disable this workflow for a specific commit by including `[ci skip]` in
the commit message.

[ref-runs-on]: https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idruns-on

#### `githubWorkflows.license`

This workflow ensures that all your source code files have a license header. It
uses [`addlicense`](https://github.com/google/addlicense) for this.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `githubWorkflows.license.enabled` | boolean | `false` | Enables generation of this workflow. |
| `githubWorkflows.license.patterns` | list | *(required)* | A list of filename patterns to check. Directory patterns are scanned recursively. |
| `githubWorkflows.license.ignorePaths` | list | *(optional)* | Refer to the description for `githubWorkflows.global.ignorePaths`. |

#### `githubWorkflows.spellCheck`

This workflow checks for spelling (American english) errors. It uses
[`misspell`](https://github.com/client9/misspell) for this.

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `githubWorkflows.spellCheck.enabled` | boolean | `false` | Enables generation of this workflow. |
| `githubWorkflows.spellCheck.ignorePaths` | list | *(optional)* | Refer to the description for `githubWorkflows.global.ignorePaths`. |
