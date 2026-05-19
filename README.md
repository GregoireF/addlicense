# addlicense

> Fast, minimal license header manager for monorepos and CI pipelines.

[![CI](https://github.com/GregoireF/addlicense/actions/workflows/ci.yml/badge.svg)](https://github.com/GregoireF/addlicense/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/GregoireF/addlicense/graph/badge.svg)](https://codecov.io/gh/GregoireF/addlicense)
[![Go version](https://img.shields.io/github/go-mod/go-version/GregoireF/addlicense)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/GregoireF/addlicense)](https://github.com/GregoireF/addlicense/releases)

---

## What it does

`addlicense` scans a directory tree, detects which files are missing a license header, and injects one. On the second run, it does nothing — already-licensed files are left untouched.

```
addlicense --license MIT .
```

That's the common case. No config required.

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

**Add MIT headers**

```bash
addlicense --license MIT .
```

**Specify an author**

```bash
addlicense --license MIT --author "Acme Corp" .
```

**Check only — useful in CI**

```bash
addlicense --check .
```

Exit 0 if all files have headers. Exit 1 with a list of missing files otherwise.

**Custom template**

```bash
addlicense --template ./header.txt .
```

**REUSE/FSFE compliance**

```bash
addlicense --license MIT --reuse .
```

Emits `SPDX-FileCopyrightText: 2026 Author` instead of `Copyright 2026 Author` — required by the [REUSE specification](https://reuse.software/).

**Ignore paths**

```bash
addlicense --ignore dist,vendor,*.gen.go .
```

---

## Flags

| Flag | Short | Default | Description |
| :-- | :-- | :-- | :-- |
| `--license` | `-l` | `MIT` | SPDX license identifier |
| `--author` | `-a` | — | Copyright holder |
| `--year` | `-y` | current year | Copyright year |
| `--template` | `-t` | — | Path to a custom header template |
| `--ignore` | `-i` | see below | Patterns to skip |
| `--check` | `-c` | false | Check mode — no writes, exit 1 if missing |
| `--reuse` | `-r` | false | REUSE/FSFE mode — emit `SPDX-FileCopyrightText:` instead of `Copyright` |
| `--version` | | | Print version and build info |

**Default ignore list:** `vendor`, `node_modules`, `.git`, `dist`, `build`, `*.pb.go`, `*.gen.go`

---

## Supported languages

| Extension | Comment style |
| :-- | :-- |
| `.go` `.ts` `.tsx` `.js` `.jsx` `.rs` `.swift` `.kt` `.scala` `.php` `.cs` `.proto` | `//` |
| `.java` `.c` `.cpp` `.h` `.css` `.scss` | `/* */` |
| `.html` `.vue` `.svelte` | `<!-- -->` |
| `.py` `.sh` `.bash` `.yaml` `.yml` `.tf` `.toml` `.rb` | `#` |
| `.sql` | `--` |

More languages are added on request — open an issue.

---

## Configuration file

`addlicense` auto-detects a config file in the current directory or the scanned path, in this order:

1. `.addlicenserc.yaml`
2. `.addlicenserc.yml`
3. `.addlicenserc.json`
4. `addlicense.json`

CLI flags always override the config file.

**Example `.addlicenserc.yaml`:**

```yaml
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

**GitHub Actions — check on every push**

```yaml
name: License check

on: [push, pull_request]

jobs:
  license:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install addlicense
        run: |
          curl -sSL https://github.com/GregoireF/addlicense/releases/latest/download/addlicense_linux_amd64.tar.gz \
            | tar -xz -C /usr/local/bin addlicense

      - name: Check license headers
        run: addlicense --check .
```

**With Docker (no install needed):**

```yaml
      - name: Check license headers
        run: |
          docker run --rm -v "$PWD:/src" -w /src \
            ghcr.io/gregoiref/addlicense:latest --check .
```

---

## Supported SPDX identifiers

Any SPDX identifier works with `--license`. Built-in templates exist for the most common ones:

- `MIT`
- `Apache-2.0`
- `GPL-3.0-only`
- `AGPL-3.0-only`
- `LGPL-2.1-only`
- `LGPL-3.0-only`
- `EUPL-1.2` — European Union Public Licence (recommended for EU public sector software)
- `MPL-2.0`
- `BSD-2-Clause`
- `BSD-3-Clause`

For anything else, a generic template is used: `Copyright <year> <author>\nSPDX-License-Identifier: <id>`.

See [ROADMAP.md](ROADMAP.md) for the EU/French compliance context (REUSE spec, CRA) and planned versions.

---

## Design

**Why Go**

A license tool runs in CI on every push. Runtime startup matters, dependency installation is noise. Go produces a self-contained binary with no runtime requirement — it installs in one step and runs in milliseconds.

**Idempotence via SPDX detection**

The first pass scans the top 20 lines of each file for `SPDX-License-Identifier:` or any line containing `copyright`. If found, the file is skipped — no re-injection, no drift. This is the approach used by tooling in the SPDX ecosystem and is robust to manual edits.

**Config auto-detection**

No mandatory config file. Zero-config for the common case (`addlicense --license MIT .`), opt-in config for teams that want consistent defaults checked into the repo. Same pattern as Biome, ESLint, and Prettier.

**Template system**

The built-in templates emit SPDX identifiers rather than full license text. This keeps headers minimal and machine-readable — compatible with SBOM tools and license scanners downstream.

**Single binary, Docker, Homebrew**

Three distribution channels cover the three common contexts: local dev (Homebrew), CI without installation (Docker `FROM scratch`), and projects that manage dependencies with Go tooling (`go install`).

---

## Contributing

Issues and PRs are welcome. Please follow [Conventional Commits](https://www.conventionalcommits.org/) — enforced by CI.

---

## License

MIT — see [LICENSE](LICENSE).
