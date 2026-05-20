# Removing and Updating Headers

addlicense can strip existing license headers (`--remove`) or replace them with a new one (`--update`). Both modes target only comment blocks that contain a license marker (`SPDX-License-Identifier` or `copyright`, case-insensitive) — non-license comments are never touched.

---

## `--remove` — strip existing headers

```bash
addlicense --remove .
```

Removes the leading license header from every supported source file under `.`. Files without a header are skipped silently. Returns exit 0 even when no file is modified.

**What counts as a license header?**

A contiguous block of comment lines at the top of the file (after the shebang if present) that contains at least one line with `SPDX-License-Identifier` or `copyright` (case-insensitive). Pure documentation comments are left untouched.

**Supported comment styles:**

| Style | Example |
| :-- | :-- |
| Line comment | `// Copyright …` (Go, JS, Java, …) |
| Line comment | `# Copyright …` (Python, Shell, YAML, …) |
| Line comment | `-- Copyright …` (SQL) |
| Block comment | `/* … */` (CSS, C, Java) |
| Block comment | `<!-- … -->` (HTML, Vue, Svelte) |

---

## `--update` — replace headers

```bash
addlicense --update --license EUPL-1.2 --author "Acme Corp" .
```

`--update` is equivalent to `--remove` followed by inject. It removes the existing header (whatever licence it was) and injects a new one with the current flags or config. This is the canonical command for licence migration.

**Example: migrate from MIT to EUPL-1.2**

```bash
addlicense --update --license EUPL-1.2 --author "Acme Corp" .
```

Before:

```go
// SPDX-License-Identifier: MIT
// Copyright 2024 Acme Corp

package main
```

After:

```go
// SPDX-License-Identifier: EUPL-1.2
// Copyright 2026 Acme Corp

package main
```

---

## `--dry-run` — preview without writing

Both `--remove` and `--update` respect `--dry-run`. No files are modified; instead, the planned action is printed:

```bash
addlicense --remove --dry-run .
# [dry-run] would-remove: internal/config/config.go
# [dry-run] would-remove: internal/runner/runner.go
# …

addlicense --update --license Apache-2.0 --dry-run .
# [dry-run] would-update: internal/config/config.go
# [dry-run] would-add:    README.md  (if unlicensed)
```

---

## Combine with `--format json`

For scripting or dashboards, pipe the output as JSON Lines:

```bash
addlicense --remove --dry-run --format json . | jq 'select(.status == "would-remove") | .file'
```

---

## Idempotence

Running `--remove` twice on the same file is safe — the second run detects no header and skips the file (`changed=false`).

Running `--update` twice is also safe — the first run injects the new header; the second run detects it and skips.

---

## Mutually exclusive combinations

| Combination | Behaviour |
| :-- | :-- |
| `--remove` + `--update` | Error: mutually exclusive |
| `--check` + `--remove` | Error: `--check` cannot be combined with `--remove` |
| `--check` + `--update` | Error: `--check` cannot be combined with `--update` |

---

## Usage in CI

To enforce that headers are present and correct after a migration:

```bash
addlicense --update --license EUPL-1.2 --author "Acme Corp" .
addlicense --check .
```

Or in a single step using `--dry-run` to validate first:

```bash
addlicense --update --license EUPL-1.2 --dry-run .   # preview
addlicense --update --license EUPL-1.2 .              # apply
```
