# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [0.5.0] — 2026-05-20

### Added

- **`--year-range` flag** — when combined with `--update`, reads the original copyright year from the existing header and emits `YYYY-YYYY` (e.g. `Copyright 2023-2026 Author`) instead of overwriting with the current year alone. Falls back silently to a single year when the original year cannot be extracted or equals the current year. Works with `--reuse` (emits `SPDX-FileCopyrightText: 2023-2026 Author`).
- **`--dep5` flag** — generates a REUSE-compliant `.reuse/dep5` bulk-licence declaration for files that cannot carry inline SPDX headers (images, fonts, binaries, and any other extension not in the language map). The dep5 root defaults to the first path argument. Supports `--dry-run` (prints `[dry-run] would-write: .reuse/dep5`).
- **Native GitHub Action (`action.yml`)** — composite action using `runner.os` and `runner.arch` detection to download the correct binary from the GitHub Release tag, add it to `$GITHUB_PATH`, and run addlicense. Supports `ubuntu-*`, `macos-*`, and `windows-*` runners on amd64 + arm64. Usage: `uses: GregoireF/addlicense@v0.5.0` with optional `args` input (default: `--check .`).
- **`internal/dep5` package** — pure Go dep5 document generator; `Build(paths, year, author, license)` returns the formatted dep5 content.
- **Extended test suite**:
  - `internal/injector`: 5 new unit tests for `ExtractYear` — copyright, SPDX-FileCopyrightText, no year, file not found, beyond scan window.
  - `internal/dep5`: 5 unit tests — with paths, empty paths (header-only), no author, multiple paths (continuation lines), forward-slash path normalisation.
  - `tests/integration/`: 7 new tests — `--year-range` (original preserved, same year, no header, with REUSE); `--dep5` (creates file, no unhandled files, dry-run).

---

## [0.4.0] — 2026-05-20

### Added

- **`--remove` / `-R` flag** — strips existing license headers from files. Detection uses the same heuristic as the idempotence check (top-20-line scan for `spdx-license-identifier` or `copyright`); only comment blocks that contain a license marker are removed — non-license comment blocks (package docs, etc.) are left untouched. Handles both line-comment (`//`, `#`, `--`) and block-comment (`<!-- -->`, `/* */`) styles.
- **`--update` / `-u` flag** — replaces the existing header with a new one in a single pass (`--remove` + inject). The canonical migration workflow: `addlicense --update --license EUPL-1.2 .`
- **`--dry-run` / `-n` flag** — previews changes without writing to disk. Emits `[dry-run] would-add: <path>`, `[dry-run] would-remove: <path>`, `[dry-run] would-update: <path>` for every file that would be affected.
- **`--verbose` / `-v` flag** — prints every processed file, including already-licensed ones (`skipped` / `ok`), in addition to the default modified-file output.
- **`--quiet` / `-q` flag** — suppresses all stdout; errors still go to stderr. Designed for pure CI pipelines where only the exit code matters.
- **`--format` / `-f` flag** — selects output format: `text` (default, human-readable) or `json` (JSON Lines: `{"file":"…","status":"…","error":"…"}`). The string flag is extensible to future formats without breaking the CLI contract.
- **`--workers` flag** — controls the number of parallel goroutines (default: `runtime.NumCPU()`). Files are processed concurrently via a buffered jobs channel and a `sync.WaitGroup`-managed worker pool, bounded to `min(workers, len(files))`.
- **`reuse:` field in `.addlicenserc.yaml`** — `reuse: true` in the config file now activates REUSE/FSFE mode without requiring the `--reuse` flag on every invocation. CLI flag takes precedence per the usual merge order.
- **`.pre-commit-hooks.yaml`** — official [pre-commit](https://pre-commit.com/) hook definitions shipped in the repository root. Two hooks: `addlicense-check` (exit 1 on missing headers, read-only) and `addlicense-add` (inject missing headers). Language is `golang`; pre-commit installs the binary automatically via `go install`. Both hooks set `pass_filenames: false` and `always_run: true`.
- **Extended test suite**:
  - `internal/injector`: 8 new unit tests covering `Remove` — line-comment removal, shebang preservation, idempotence, non-license-comment safety, block-comment HTML/CSS styles, nonexistent-file error.
  - `internal/config`: 3 new unit tests covering `reuse:` merge — from config file, CLI precedence, default false.
  - `tests/integration/`: 20 new integration tests covering `--remove`, `--update`, `--dry-run`, `--quiet`, `--verbose`, `--format json`, `--workers`, `--format xml` (invalid), all mutual-exclusion flag combinations, `reuse:` from config file, and parallel processing across 20 files.

### Changed

- **Mutual exclusion validation** (`validateOpts`): `--verbose` ⊕ `--quiet`, `--check` ⊕ (`--remove` ∨ `--update`), `--remove` ⊕ `--update` — all return descriptive errors before any file I/O.
- **Runner refactored** into `runParallel` + `processFile` + per-mode functions (`checkFile`, `addFile`, `removeFile`, `updateFile`) to keep cyclomatic complexity under the gocyclo-15 limit.

---

## [0.3.0] — 2026-05-20

### Added

- **`--reuse` / `-r` flag** — emits `SPDX-FileCopyrightText: <year> <author>` instead of `Copyright <year> <author>`, complying with the [REUSE/FSFE specification](https://reuse.software/). Idempotence works transparently (`FileCopyrightText` contains `copyright` as a substring). Opt-in flag rather than default to preserve `grep -r "Copyright"` workflows; making it the default would require a major version bump.
- **`header.Data.CopyrightLine`** — pre-formatted string computed once by `runner.buildCopyrightLine(year, author, reuse)` and consumed by all built-in templates via `{{.CopyrightLine}}`; removes per-template `{{if .Author}}` conditionals and centralises REUSE/non-REUSE logic in one place.
- **Built-in SPDX templates**: EUPL-1.2, AGPL-3.0-only, LGPL-2.1-only, LGPL-3.0-only — EU public sector and strong-copyleft licences (see [ROADMAP.md](ROADMAP.md) for compliance context).
- **golangci-lint v2 configuration** (`.golangci.yml`): 15 linters beyond the default — `bodyclose`, `exhaustive`, `goconst`, `gocritic`, `gocyclo`, `godot`, `misspell`, `nestif`, `nilerr`, `prealloc`, `revive`, `unconvert`, `unparam`; `goimports` in `formatters` section (v2 separates formatters from linters).
- **Codecov** (`.codecov.yml`): 90 % coverage threshold enforced on both project and patch.
- **Extended test suite** — coverage ≥ 90 %:
  - `internal/scanner`: 4 unit tests — glob ignore patterns, invalid root, no-extension files, markdown/JSON collection
  - `internal/injector`: 5 unit tests — file-not-found, empty file, scan-window boundary, case-insensitive detection
  - `internal/config`: 11 unit tests — `Load` priority rules, YAML/YML/JSON auto-detection, CLI flag override, invalid YAML error
  - `tests/integration/`: 3 REUSE tests + 8 additional scenarios (block/HTML/SQL comment styles, EU licences, multiple roots, custom templates)
- **`CONTRIBUTING.md`** — development setup, test commands, lint config decisions, commit/PR conventions, adding-a-language and adding-a-template guides, release process, design principles.
- **`CHANGELOG.md`** and **`ROADMAP.md`** — version planning and EU/French compliance context.
- **GitHub issue templates** (`.github/ISSUE_TEMPLATE/`): bug report and feature request with structured fields.
- **GitHub PR template** (`.github/PULL_REQUEST_TEMPLATE.md`): checklist covering tests, CHANGELOG, lint, coverage.
- **`CODEOWNERS`** — GregoireF as default reviewer on all PRs.

### Changed

- **GoReleaser v2**: `brews:` key replaced by `homebrew_casks:` (removed in v2.10) — Homebrew tap formula now lives at `Casks/addlicense.rb`; old `Formula/addlicense.rb` removed from tap.
- **CI coverage**: `-coverpkg=./...` flag ensures integration tests in `tests/integration/` correctly attribute coverage to `internal/` packages.
- **Test organisation**: integration tests promoted from `internal/runner/` to dedicated `tests/integration/` package (`package integration_test`), following Go conventions.

### Fixed

- **`internal/header/header.go`**: `Lang` struct and `langs` map brought into strict gofmt compliance — gofmt aligns map-literal values per comment-separated group (comments reset the tabwriter), not globally.

---

## [0.2.0] — 2026-05-19

### Added
- Extended language support: `.html`, `.vue`, `.svelte` (`<!-- -->`), `.css`, `.scss` (`/* */`), `.proto` (`//`), `.sql` (`--`)
- Docker image published to GHCR (`ghcr.io/gregoiref/addlicense`) on every release tag
- Codecov integration — coverage badge in README
- `.addlicenserc.yaml` dogfooding — repo now uses its own config file; CI no longer hardcodes `--license` and `--author` flags
- Dependabot for Go modules and GitHub Actions (weekly, auto-merge on passing checks)
- Auto-merge workflow for Dependabot PRs

### Changed
- GitHub Actions bumped to Node.js 24 compatible versions: `actions/checkout` v6, `actions/setup-go` v6, `goreleaser/goreleaser-action` v7, `golangci-lint-action` v9
- CI caching re-enabled after `go.sum` committed
- `goreleaser-action` version locked to `~> v2` (was `latest`)

### Fixed
- `errcheck` lint violations: `HasHeader` refactored from `bufio.Scanner` + `defer f.Close()` to `os.ReadFile` — no unchecked close error
- `f.Close()` in injector test helper now checked and fatal on error

---

## [0.1.0] — 2026-05-18

### Added
- Initial release
- Unified CLI — single command, no subcommands: `addlicense [flags] [path...]`
- Flags: `--license` / `-l`, `--author` / `-a`, `--year` / `-y` (defaults to current year), `--template` / `-t`, `--ignore` / `-i`, `--check` / `-c`, `--version`
- Built-in SPDX templates: MIT, Apache-2.0, GPL-3.0-only, MPL-2.0, BSD-2-Clause, BSD-3-Clause
- Generic fallback template for any other SPDX identifier
- Custom header template support via `--template ./header.txt`
- Config file auto-detection in priority order: `.addlicenserc.yaml`, `.addlicenserc.yml`, `.addlicenserc.json`, `addlicense.json`
- Idempotent header detection — scans first 20 lines for `SPDX-License-Identifier:` or `copyright`; already-licensed files are skipped
- Shebang line preservation (`#!/usr/bin/env bash` stays on line 1)
- Supported languages: Go, TypeScript/TSX, JavaScript/JSX, Java, C/C++/H, Rust, Python, Shell/Bash, YAML, Terraform, TOML, Ruby, Swift, Kotlin, Scala, PHP, C#
- Check mode: exit 0 if all files are licensed, exit 1 with list of missing files
- Multi-platform binaries: Linux, macOS, Windows × amd64, arm64 — via GoReleaser
- Homebrew tap: `brew install GregoireF/tap/addlicense`
- Docker image (`FROM scratch`, ~3 MB): `ghcr.io/gregoiref/addlicense`
- Dogfooding: the `license-check` CI workflow runs addlicense on its own source on every push
