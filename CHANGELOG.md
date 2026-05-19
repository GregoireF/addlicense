# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `--reuse` / `-r` flag: emits `SPDX-FileCopyrightText: <year> <author>` instead of `Copyright <year> <author>` — [REUSE/FSFE specification](https://reuse.software/) compliance. Idempotence detection already handled (`FileCopyrightText` contains `copyright` as a substring). Choice: opt-in flag rather than default to avoid breaking `grep -r "Copyright"` tooling; making it default requires a major version bump (see [ROADMAP.md](ROADMAP.md)).
- `CopyrightLine` field added to `header.Data` — pre-formatted copyright string passed to templates; all built-in templates now use `{{.CopyrightLine}}` instead of inline `Copyright {{.Year}} {{.Author}}` conditionals, which removes template duplication and enables REUSE mode without template changes.
- Built-in templates for EUPL-1.2, AGPL-3.0, LGPL-2.1, LGPL-3.0 — covers European public sector and copyleft licences (see [ROADMAP.md](ROADMAP.md) for EU compliance context)
- CHANGELOG.md (this file) and ROADMAP.md — EU/French norms context + version planning
- golangci-lint v2 configuration (`.golangci.yml`): 15 linters beyond the default set — `bodyclose`, `exhaustive`, `goconst`, `gocritic`, `gocyclo`, `godot`, `misspell`, `nestif`, `nilerr`, `prealloc`, `revive`, `unconvert`, `unparam`; `goimports` in `formatters` section (golangci-lint v2 split formatters from linters)
- Codecov configuration (`.codecov.yml`): 90% coverage threshold enforced on both project and patch; structured comment layout
- Extended test suite — coverage raised to 90%+:
  - `internal/scanner`: 4 new unit tests — glob ignore patterns, invalid root, no-extension files, markdown/JSON collection
  - `internal/injector`: 5 new unit tests — file-not-found paths, empty file, scan window boundary, case-insensitive detection
  - `internal/config`: 11 new unit tests — `Load` priority rules, YAML/YML/JSON auto-detection, CLI flag override, invalid YAML error path
  - `tests/integration/`: integration tests moved from `internal/runner/` to a dedicated package; 8 additional scenarios covering block/HTML/SQL comment styles, EU licences, multiple roots, custom templates
- `CONTRIBUTING.md` — development setup, test commands, commit and PR conventions

### Changed
- GoReleaser v2: `brews:` key (removed in v2.10) replaced by `homebrew_casks:` — Homebrew tap formula moves from `Formula/` to `Casks/` on next release
- CI test command now uses `-coverpkg=./...` for cross-package coverage attribution — integration tests in `tests/integration/` correctly attribute coverage to `internal/` packages
- Test organisation: runner integration tests promoted to `tests/integration/` package (`package integration_test`) following Go conventions — unit tests (`_test.go`) stay alongside source, integration tests get their own top-level directory

### Fixed
- `internal/header/header.go`: `Lang` struct alignment and `langs` map formatting brought into strict gofmt compliance — gofmt aligns map values per comment-separated group (comments reset the tabwriter), not globally

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
