# addlicense

> Fast, minimal license header manager for monorepos and CI pipelines.

[![CI](https://github.com/GregoireF/addlicense/actions/workflows/ci.yml/badge.svg)](https://github.com/GregoireF/addlicense/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/GregoireF/addlicense/graph/badge.svg)](https://codecov.io/gh/GregoireF/addlicense)
[![Go Report Card](https://goreportcard.com/badge/github.com/GregoireF/addlicense)](https://goreportcard.com/report/github.com/GregoireF/addlicense)
[![Go version](https://img.shields.io/github/go-mod/go-version/GregoireF/addlicense)](go.mod)
[![Release](https://img.shields.io/github/v/release/GregoireF/addlicense)](https://github.com/GregoireF/addlicense/releases)
[![Docker](https://img.shields.io/badge/docker-ghcr.io-2496ED?logo=docker&logoColor=white)](https://github.com/GregoireF/addlicense/pkgs/container/addlicense)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`addlicense` scans a directory tree, detects which files are missing a license header, and injects one. On the second run, it does nothing — already-licensed files are left untouched.

```
addlicense --license MIT .
```

No config required for the common case. See the **[Wiki](../../wiki)** for architecture decisions, advanced usage, and contributor guides.

---

## Installation

**Homebrew**

```bash
brew install GregoireF/tap/addlicense
```

**Binary** — download from [Releases](https://github.com/GregoireF/addlicense/releases) (Linux, macOS, Windows — amd64 / arm64).

**Docker** — no installation needed:

```bash
docker run --rm -v "$PWD:/src" -w /src ghcr.io/gregoiref/addlicense --license MIT .
```

**Go install**

```bash
go install github.com/GregoireF/addlicense/cmd/addlicense@latest
```

---

## Quick start

```bash
# Add MIT headers to the whole project
addlicense --license MIT --author "Acme Corp" .

# Check only — exit 1 with list of unlicensed files (useful in CI)
addlicense --check .

# Remove existing headers
addlicense --remove .

# Replace all headers with a new licence in one pass
addlicense --update --license Apache-2.0 --author "Acme Corp" .

# Preview changes without writing to disk
addlicense --dry-run .

# REUSE/FSFE compliance — emit SPDX-FileCopyrightText: instead of Copyright
addlicense --license MIT --reuse .

# Multiple copyright holders — comma-separated
addlicense --license MIT --author "Alice, Bob" .

# Show what would change without writing (JSON Lines output, exit 1 if changes needed)
addlicense --diff .

# Custom template
addlicense --template ./header.txt .

# Ignore paths
addlicense --ignore "dist,vendor,*.gen.go" .
```

---

## Flags

| Flag | Short | Default | Description |
| :-- | :-- | :-- | :-- |
| `--license` | `-l` | `MIT` | SPDX license identifier |
| `--author` | `-a` | — | Copyright holder (comma-separated for multiple: `"Alice, Bob"`) |
| `--year` | `-y` | current year | Copyright year |
| `--template` | `-t` | — | Path to a custom header template file |
| `--ignore` | `-i` | see below | Comma-separated glob patterns to skip |
| `--check` | `-c` | `false` | Check mode — no writes, exit 1 if any file is missing a header |
| `--remove` | `-R` | `false` | Strip existing license headers |
| `--update` | `-u` | `false` | Replace headers in one pass (`--remove` + inject) |
| `--dry-run` | `-n` | `false` | Preview changes without writing to disk |
| `--diff` | | `false` | Emit JSON Lines with rendered header for each file that would change; no writes, exit 1 if any file would be modified |
| `--verbose` | `-v` | `false` | Print every processed file, including already-licensed ones |
| `--quiet` | `-q` | `false` | Suppress all stdout; errors still go to stderr |
| `--format` | `-f` | `text` | Output format: `text` or `json` (JSON Lines) |
| `--workers` | | `NumCPU` | Parallel goroutines for file processing |
| `--reuse` | `-r` | `false` | [REUSE/FSFE](https://reuse.software/) mode — emit `SPDX-FileCopyrightText:` |
| `--year-range` | | `false` | Preserve original copyright year on `--update` — emits `YYYY-YYYY` range |
| `--dep5` | | `false` | Generate `.reuse/dep5` for files that cannot carry inline headers (images, binaries…) |
| `--version` | | | Print version and build info |

**Mutually exclusive:** `--verbose`/`--quiet` · `--check`/`--remove` · `--check`/`--update` · `--remove`/`--update` · `--diff`/`--check`

**Default ignore list:** `vendor`, `node_modules`, `.git`, `dist`, `build`, `*.pb.go`, `*.gen.go`

**`--ignore` matching rules:** Each pattern is tested as a glob against the file name and each path segment (e.g. `vendor` ignores any file under a `vendor/` directory at any depth). Patterns without wildcards also match as substrings of the full path (e.g. `generated` matches `auto_generated.go`). Glob patterns like `*.pb.go` use standard shell globbing.

**Multi-author with `--update`:** `--update` replaces the entire existing header with the new one. If `--author "Alice, Bob"` is set, the result is two copyright lines for Alice and Bob — the previous author is not preserved. To keep an existing author, include them in the `--author` list.

---

## Supported languages

| Extensions | Comment style |
| :-- | :-- |
| `.go` `.ts` `.tsx` `.js` `.jsx` `.rs` `.swift` `.kt` `.scala` `.php` `.cs` `.proto` | `// line` |
| `.java` `.c` `.cpp` `.h` `.css` `.scss` | `/* block */` |
| `.html` `.vue` `.svelte` | `<!-- block -->` |
| `.py` `.sh` `.bash` `.yaml` `.yml` `.tf` `.toml` `.rb` | `# line` |
| `.sql` | `-- line` |

[Request a language](https://github.com/GregoireF/addlicense/issues/new?template=feature_request.md) — adding one takes < 5 minutes (see [Wiki: Adding a Language](../../wiki/Adding-a-Language)).

---

## Supported SPDX identifiers

Any SPDX identifier works. Built-in templates:

`MIT` · `Apache-2.0` · `GPL-3.0-only` · `AGPL-3.0-only` · `LGPL-2.1-only` · `LGPL-3.0-only` · `EUPL-1.2` · `MPL-2.0` · `BSD-2-Clause` · `BSD-3-Clause`

Any other identifier uses a generic template: `Copyright <year> <author> / SPDX-License-Identifier: <id>`.

---

## Configuration file

Auto-detected in priority order: `.addlicenserc.yaml` → `.addlicenserc.yml` → `.addlicenserc.json` → `addlicense.json`. CLI flags always take precedence.

```yaml
# .addlicenserc.yaml
license: Apache-2.0
author: Acme Corp
year: 2026
reuse: false
ignore:
  - vendor
  - node_modules
  - "*.gen.go"
```

---

## CI integration

**GitHub Action** (since v0.5.0 — recommended)

```yaml
- uses: GregoireF/addlicense@v0.5.0
  with:
    args: --check .
```

Works on Linux, macOS, and Windows runners (amd64 + arm64). See [Wiki: CI Integration](../../wiki/CI-Integration) for advanced examples.

**pre-commit** (since v0.4.0 — binary installed automatically)

```yaml
repos:
  - repo: https://github.com/GregoireF/addlicense
    rev: v0.4.0
    hooks:
      - id: addlicense-check
```

**Docker (no install)**

```yaml
- name: Check license headers
  run: |
    docker run --rm -v "$PWD:/src" -w /src \
      ghcr.io/gregoiref/addlicense:latest --check .
```

More examples (GitLab CI, curl install, Makefile, JSON output) → [Wiki: CI Integration](../../wiki/CI-Integration).

---

## Performance

`addlicense` is built for speed. The worker pool scales linearly with CPU count — the default (`--workers 0`) uses all available cores.

| Scenario | Files | Time |
| :-- | --: | --: |
| Add headers (1 worker) | 200 | ~159 ms |
| Add headers (4 workers) | 200 | ~50 ms |
| Add headers (NumCPU = 20) | 200 | ~38 ms |
| Add headers | 1 000 | ~184 ms |
| Check only (all licensed) | 1 000 | ~55 ms |
| `HasHeader` per file | 1 | ~27 µs |

_Measured on an i7-12700F (20 logical cores), NVMe SSD, Go 1.23, Linux._

Run the benchmarks yourself:

```bash
go test -bench=. -benchtime=3s ./tests/bench/       # end-to-end pipeline
go test -bench=. -benchtime=3s ./internal/injector/ # injector micro-benchmarks
```

---

## Contributing

Issues and PRs are welcome. Quick start:

```bash
git clone https://github.com/GregoireF/addlicense.git
cd addlicense
go test -race -coverprofile=coverage.out -coverpkg=./... ./...
golangci-lint run
```

Please follow [Conventional Commits](https://www.conventionalcommits.org/) — enforced by CI. Full guide → [CONTRIBUTING.md](CONTRIBUTING.md) · [Wiki](../../wiki).

---

## License

MIT — see [LICENSE](LICENSE).
