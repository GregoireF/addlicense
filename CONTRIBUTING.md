# Contributing to addlicense

Thank you for your interest. This document covers the development setup, conventions, and process for contributing to this project.

---

## Prerequisites

- **Go 1.23+** — the module requires Go 1.23 (`go.mod`).
- **golangci-lint v2** — for local linting. Install via the [official script](https://golangci-lint.run/welcome/install/).

No other toolchain dependency is required — the binary has no CGO and zero non-standard runtime dependencies.

---

## Local setup

```bash
git clone https://github.com/GregoireF/addlicense.git
cd addlicense
go mod tidy
```

---

## Running tests

```bash
# All tests (unit + integration) with race detector and cross-package coverage
go test -race -coverprofile=coverage.out -coverpkg=./... ./...

# Coverage report in browser
go tool cover -html=coverage.out
```

**Test organisation:**

| Location | Package | Purpose |
| :-- | :-- | :-- |
| `internal/*/..._test.go` | `package foo_test` | Unit tests — fast, no I/O beyond temp dirs |
| `tests/integration/` | `package integration_test` | Integration tests — run the full pipeline on real filesystems |

The split follows Go conventions: unit tests live alongside source (standard), integration tests get a dedicated top-level directory so they can be run or excluded independently. The `-coverpkg=./...` flag ensures integration tests attribute coverage to the `internal/` packages they exercise, rather than appearing as uncovered.

---

## Running the linter

```bash
golangci-lint run
```

The configuration is in `.golangci.yml` (golangci-lint v2 format). Key decisions:

- **goimports** is in `formatters.enable`, not `linters.enable` — this is a v2 requirement; formatters and linters are now separate sections.
- **goconst.min-occurrences: 5** — raised from the default (3) to avoid false positives in test files that legitimately repeat short string constants like `"MIT"`.
- **godot.scope: declarations** — checks periods only on exported declaration comments, not inline comments.
- **revive** rules are explicit (`exported`, `var-naming`, `error-return`, etc.) rather than using the default set, which gives stable behaviour across golangci-lint updates.

---

## Commit convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/), enforced by CI (`commitlint`).

```
<type>(<scope>): <short description>
```

Common types: `feat`, `fix`, `docs`, `chore`, `ci`, `refactor`, `test`.

Examples:

```
feat(header): add EUPL-1.2 built-in template
fix(injector): preserve shebang on files with Windows line endings
docs(readme): add CI integration example for Docker
chore(deps): bump actions/setup-go from v5 to v6
```

Scope is optional but recommended for larger codebases. Breaking changes use `!` after the type: `feat!: rename --check to --check-only`.

---

## Pull requests

1. Fork and create a branch from `main`.
2. Write or update tests. The 90% coverage threshold is enforced on both project and patch by Codecov.
3. Run `go test -race ./...` and `golangci-lint run` locally before pushing.
4. Open a PR — CI runs automatically. A PR that drops coverage below 90% or introduces lint violations will not be merged.
5. Commits are squash-merged to keep `main` clean.

---

## Adding a language

Language support lives in `internal/header/header.go`, in the `langs` map. Each entry maps a file extension to a `Lang` struct describing the comment style:

```go
type Lang struct {
    LineComment string // prefix for single-line comments, e.g. "//"
    BlockOpen   string // opening delimiter for block comments, e.g. "/*"
    BlockClose  string // closing delimiter, e.g. " */"
    BlockPrefix string // per-line prefix inside a block, e.g. " * "
}
```

If `LineComment` is set, line comments are used. Otherwise, block comment style. HTML comment style (`<!-- ... -->`) is a special case handled by the `htmlComment` variable.

To add a language:
1. Add the extension → `Lang` mapping to the `langs` map (keep within its comment group, sorted alphabetically within the group).
2. Add a test case in `tests/integration/runner_test.go` — at minimum one test that verifies the correct comment delimiters appear in the output.
3. Update the **Supported languages** table in `README.md`.

---

## Adding a licence template

Built-in templates live in `internal/header/header.go`, in the `BuiltinTemplate` function. Each `case` maps one or more SPDX identifiers (upper-cased) to a Go template string.

Template data fields:

| Field | Type | Description |
| :-- | :-- | :-- |
| `.Year` | `int` | Copyright year |
| `.Author` | `string` | Copyright holder (empty string if not set) |
| `.License` | `string` | Raw licence identifier as passed by the user |
| `.SPDX` | `string` | Upper-cased SPDX identifier |

To add a template:
1. Add a `case` in `BuiltinTemplate` — use the canonical SPDX identifier (e.g. `"EUPL-1.2"`) and common aliases.
2. Add a test in `tests/integration/runner_test.go` that runs with that identifier and asserts the SPDX identifier appears in the output.
3. Add the identifier to the **Supported SPDX identifiers** list in `README.md`.
4. Document the addition in `CHANGELOG.md` under `[Unreleased]`.

---

## Release process

Releases are fully automated via GoReleaser triggered by a `v*` tag.

```bash
git tag v0.3.0
git push origin v0.3.0
```

This builds multi-platform binaries, publishes a GitHub Release, pushes a Docker image to GHCR, and updates the Homebrew tap. No manual steps.

**Homebrew note:** as of GoReleaser v2, the tap formula is written to `Casks/addlicense.rb` (via `homebrew_casks:`) instead of `Formula/addlicense.rb`. The tap repository must have a `Casks/` directory present.

---

## Design principles

**Minimal and composable.** addlicense does one thing: manage licence headers. It does not lint, generate SBOMs, or validate dependency licences. Those concerns belong to other tools.

**Zero-config for the common case.** `addlicense --license MIT .` should work with no config file, no environment variables, no setup. The config file is opt-in for teams that want shared defaults.

**Idempotence is non-negotiable.** Running addlicense twice must produce the same result as running it once. The detection window (top 20 lines, `SPDX-License-Identifier:` or `copyright`) is intentionally broad to handle manually-edited headers.

**No parallel processing (yet).** The scanner + injector pipeline is I/O-bound, not CPU-bound. For typical repos (< 10 000 files), sequential processing completes in under a second. Goroutine overhead (mutex on stdout, error channels, race testing complexity) is not worth it until benchmarks justify it. Deferred to v0.4.0.
