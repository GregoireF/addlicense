# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Built-in templates for EUPL-1.2, AGPL-3.0, LGPL-2.1, LGPL-3.0 — covers European public sector licences and common copyleft variants
- CHANGELOG.md (this file)
- ROADMAP.md — EU/French norms context + version planning

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
