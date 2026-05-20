# addlicense

> Fast, minimal license header manager for monorepos and CI pipelines.

[![CI](https://github.com/GregoireF/addlicense/actions/workflows/ci.yml/badge.svg)](https://github.com/GregoireF/addlicense/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/GregoireF/addlicense/graph/badge.svg)](https://codecov.io/gh/GregoireF/addlicense)
[![Go version](https://img.shields.io/github/go-mod/go-version/GregoireF/addlicense)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/GregoireF/addlicense)](https://github.com/GregoireF/addlicense/releases)

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

# REUSE/FSFE compliance — emit SPDX-FileCopyrightText: instead of Copyright
addlicense --license MIT --reuse .

# Custom template
addlicense --template ./header.txt .

# Ignore paths
addlicense --ignore dist,vendor,*.gen.go .
```

---

## Flags

| Flag | Short | Default | Description |
| :-- | :-- | :-- | :-- |
| `--license` | `-l` | `MIT` | SPDX license identifier |
| `--author` | `-a` | — | Copyright holder |
| `--year` | `-y` | current year | Copyright year |
| `--template` | `-t` | — | Path to a custom header template file |
| `--ignore` | `-i` | see below | Comma-separated glob patterns to skip |
| `--check` | `-c` | false | Check mode — no writes, exit 1 if any file is missing a header |
| `--reuse` | `-r` | false | [REUSE/FSFE](https://reuse.software/) mode — emit `SPDX-FileCopyrightText:` instead of `Copyright` |
| `--version` | | | Print version and build info |

**Default ignore list:** `vendor`, `node_modules`, `.git`, `dist`, `build`, `*.pb.go`, `*.gen.go`

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
ignore:
  - vendor
  - node_modules
  - "*.gen.go"
```

---

## CI integration

**GitHub Actions**

```yaml
- name: Install addlicense
  run: |
    curl -sSL https://github.com/GregoireF/addlicense/releases/latest/download/addlicense_linux_amd64.tar.gz \
      | tar -xz -C /usr/local/bin addlicense

- name: Check license headers
  run: addlicense --check .
```

**Docker (no install)**

```yaml
- name: Check license headers
  run: |
    docker run --rm -v "$PWD:/src" -w /src \
      ghcr.io/gregoiref/addlicense:latest --check .
```

More examples (GitLab CI, pre-commit, Makefile) → [Wiki: CI Integration](../../wiki/CI-Integration).

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
