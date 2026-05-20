# Configuration Reference

## CLI flags

| Flag | Short | Type | Default | Description |
| :-- | :-- | :-- | :-- | :-- |
| `--license` | `-l` | string | `MIT` | SPDX identifier. Any value is accepted; built-in templates exist for common ones. |
| `--author` | `-a` | string | — | Copyright holder name. Empty string omits the author from the header. |
| `--year` | `-y` | int | current year | Copyright year. Defaults to `time.Now().Year()` if 0. |
| `--template` | `-t` | string | — | Path to a custom Go template file. Overrides the built-in template for the chosen licence. |
| `--ignore` | `-i` | string | see below | Comma-separated glob patterns to skip. Merged with the default list. |
| `--check` | `-c` | bool | false | Check mode. No files are modified. Exits 1 with a list of missing files if any are unlicensed. |
| `--remove` | `-R` | bool | false | Strip existing license headers. Only removes blocks containing `SPDX-License-Identifier` or `copyright`. |
| `--update` | `-u` | bool | false | Replace existing headers (`--remove` + inject). Canonical migration workflow. |
| `--dry-run` | `-n` | bool | false | Preview changes without writing to disk. Prints `[dry-run] would-add/remove/update: <path>`. |
| `--verbose` | `-v` | bool | false | Print every file, including already-licensed ones (`skipped` / `ok`). |
| `--quiet` | `-q` | bool | false | Suppress all stdout. Errors still go to stderr. |
| `--format` | `-f` | string | `text` | Output format: `text` or `json` (JSON Lines: `{"file":"…","status":"…","error":"…"}`). |
| `--workers` | | int | `runtime.NumCPU()` | Number of parallel goroutines for file processing. |
| `--reuse` | `-r` | bool | false | REUSE/FSFE mode. Emits `SPDX-FileCopyrightText:` instead of `Copyright`. Also available in config file. |
| `--version` | | | | Prints version, commit, and build date. |

### Mutually exclusive flag combinations

| Combination | Error |
| :-- | :-- |
| `--verbose` + `--quiet` | `--verbose and --quiet are mutually exclusive` |
| `--check` + `--remove` | `--check cannot be combined with --remove or --update` |
| `--check` + `--update` | `--check cannot be combined with --remove or --update` |
| `--remove` + `--update` | `--remove and --update are mutually exclusive; --update implies removal` |

## Default ignore list

```
vendor
node_modules
.git
dist
build
*.pb.go
*.gen.go
```

Additional patterns are added via `--ignore` or the `ignore` key in the config file.

## Config file

addlicense auto-detects a config file in priority order (first match wins):

1. `.addlicenserc.yaml`
2. `.addlicenserc.yml`
3. `.addlicenserc.json`
4. `addlicense.json`

The file is resolved relative to the first path argument, then the working directory.

**CLI flags always override the config file.**

### YAML schema

```yaml
# .addlicenserc.yaml
license: MIT           # string — SPDX identifier
author: Acme Corp      # string — copyright holder
year: 2026             # int — copyright year (defaults to current year)
reuse: false           # bool — emit SPDX-FileCopyrightText: instead of Copyright
ignore:                # list of glob patterns (merged with defaults)
  - vendor
  - node_modules
  - "*.gen.go"
  - "third_party/**"
```

### JSON schema

```json
{
  "license": "MIT",
  "author": "Acme Corp",
  "year": 2026,
  "ignore": ["vendor", "node_modules", "*.gen.go"]
}
```

### Priority rules

| Setting | Resolution |
| :-- | :-- |
| `license` | CLI flag wins, then config file, then default `MIT` |
| `author` | CLI flag wins, then config file, then empty string |
| `year` | CLI flag wins, then config file, then `time.Now().Year()` |
| `ignore` | CLI `--ignore` replaces the config list entirely (not merged) |
| `--check` | CLI only — not available in config file |
| `reuse` | CLI `--reuse` wins; config `reuse: true` sets it when flag is not passed |

## Custom template

The `--template` flag accepts a path to a file containing a Go `text/template` string. The template is rendered with the `header.Data` struct:

```go
type Data struct {
    Year          int    // copyright year
    Author        string // copyright holder (empty if not set)
    License       string // raw identifier as passed by the user
    SPDX          string // upper-cased identifier
    CopyrightLine string // pre-formatted: "Copyright 2026 Author"
                         // or "SPDX-FileCopyrightText: 2026 Author" with --reuse
}
```

**Example custom template:**

```
{{.CopyrightLine}}
SPDX-License-Identifier: {{.SPDX}}
This file is part of the Acme project.
```

The template is wrapped in the appropriate comment style for the file type. You do not need to add comment prefixes (`//`, `#`, etc.) — addlicense handles that.

## Dogfooding

The repo itself uses `.addlicenserc.yaml`:

```yaml
license: MIT
author: Grégoire Favreau
ignore:
  - vendor
  - node_modules
  - .git
  - "*.pb.go"
  - "*.gen.go"
```

The `license-check` CI workflow runs `addlicense --check .` on every push to verify all source files are licensed.
