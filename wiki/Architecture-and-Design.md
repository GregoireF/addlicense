# Architecture & Design

## Pipeline

```
CLI flags + config file
        │
        ▼
  config.Load()        resolves Options (file → flag merge)
  opts.Normalize()     fills Year from time.Now() if 0
        │
        ▼
  header.BuiltinTemplate(license)   selects template string
  runner.buildCopyrightLine(...)    formats copyright prefix
        │
        ▼
  scanner.Walk(paths, ignore)       returns []File{Path, Ext}
        │
        ▼
  for each file:
    header.LangFor(ext)      → Lang struct (comment delimiters)
    injector.HasHeader(path) → bool (idempotence check)
    header.Render(tmpl, data, lang) → rendered comment block
    injector.Inject(path, block)    → write to disk
```

## Package responsibilities

| Package | Role |
| :-- | :-- |
| `internal/config` | Loads `.addlicenserc.*` config file and merges with CLI flags. Priority: config file < CLI flags. |
| `internal/header` | SPDX template strings, comment-wrapping logic (`wrapComment`, `writeLineComment`, `writeBlockComment`), `LangFor` lookup. |
| `internal/scanner` | `filepath.WalkDir` wrapper — collects files, applies ignore globs, extracts extensions. |
| `internal/injector` | `HasHeader` (idempotence scan), `Inject` (prepend header, preserve shebang). |
| `internal/runner` | Orchestrates the pipeline: config → template → scan → inject. |
| `internal/cmd` | Cobra command definition, flag wiring. |

## Why Go

A license tool runs in CI on every push. Two things matter: startup time and zero-dependency installation.

- **Zero runtime**: Go produces a self-contained binary — no Node.js, no Python, no JVM warmup.
- **Instant startup**: < 10 ms cold start vs. 100–500 ms for interpreted language tools.
- **Single file install**: `curl | tar` extracts one binary. No `npm install`, no `pip install`, no brew shims.
- **Cross-compilation**: GoReleaser produces Linux/macOS/Windows × amd64/arm64 in one `goreleaser release` command.

## Idempotence

`injector.HasHeader` reads the first 20 lines of each file and checks (case-insensitively) for:
- `spdx-license-identifier` — matches SPDX-format headers
- `copyright` — matches both `Copyright 2026 Author` and `SPDX-FileCopyrightText: 2026 Author` (the string `copyright` is a substring of `FileCopyrightText`)

If either is found, the file is skipped. The 20-line window is intentionally broad — it handles manual edits that add blank lines above the header.

**Why not hash the header?** Hashing would break if the year changes or the author is updated. The pattern-based approach is robust to manual edits and survives `--reuse` migrations.

## Template system

Built-in templates use Go's `text/template`. All built-in templates follow the same two-line structure:

```
{{.CopyrightLine}}
SPDX-License-Identifier: <id>
```

`CopyrightLine` is pre-computed by `runner.buildCopyrightLine(year, author, reuse)`:
- `reuse=false` → `"Copyright 2026 Author"`
- `reuse=true` → `"SPDX-FileCopyrightText: 2026 Author"`

This design keeps all REUSE/non-REUSE logic in one place (`runner.go`) rather than duplicating `{{if .Reuse}}` conditionals across every template.

Custom templates (via `--template`) can use any `header.Data` field: `.Year`, `.Author`, `.License`, `.SPDX`, `.CopyrightLine`.

## Comment wrapping

`header.wrapComment` dispatches on `Lang`:
- `LineComment != ""` → `writeLineComment` — prefixes each line with `// ` (or `# `, `-- `, etc.)
- `BlockOpen != ""` → `writeBlockComment` — wraps with `/* ... */` or `<!-- ... -->`

HTML comment style (`<!-- -->`) is a special-cased `Lang` with `BlockOpen: "<!--"` and `BlockClose: "-->"` and no `BlockPrefix` — the entire block is wrapped as a single unit.

## Config auto-detection

`config.Load` walks up from the first scanned path looking for:
1. `.addlicenserc.yaml`
2. `.addlicenserc.yml`
3. `.addlicenserc.json`
4. `addlicense.json`

CLI flags always override the file. Boolean flags (like `--check`, `--reuse`) cannot be set to `false` via config — they can only be set to `true`. This is a deliberate limitation: boolean-false is the zero value in Go and is indistinguishable from "not set".

## Design principles

**Minimal and composable.** addlicense does one thing: manage licence headers. It does not lint, generate SBOMs, or validate dependency licences. Those concerns belong to other tools.

**Zero-config for the common case.** `addlicense --license MIT .` must work with no config file, no environment variables, no setup.

**Idempotence is non-negotiable.** Running addlicense twice must produce the same result as running it once.

**No parallel processing (yet).** The scanner + injector pipeline is I/O-bound, not CPU-bound. For typical repos (< 10 000 files), sequential processing completes in under a second. Deferred to v0.4.0 pending benchmarks.
