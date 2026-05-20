# REUSE Compliance

## What is REUSE?

The [REUSE specification](https://reuse.software/) from the Free Software Foundation Europe (FSFE) defines a machine-readable standard for per-file copyright and licensing information. It is gaining adoption in:

- **German federal agencies** — the German Federal Agency for IT Security (BSI) and others require REUSE compliance on publicly-funded software.
- **Netherlands open government** — REUSE is a criterion in OSS procurement and public sector software publication.
- **European Commission** — recommended alongside EUPL-1.2 for open-source public sector projects.
- **Academic and research institutions** — REUSE is endorsed by several European university consortia.

## Standard vs. REUSE format

The key difference is in the copyright line:

| Format | Header |
|:--|:--|
| Standard (default) | `Copyright 2026 Grégoire Favreau` |
| REUSE (`--reuse`) | `SPDX-FileCopyrightText: 2026 Grégoire Favreau` |

The `SPDX-License-Identifier:` line is identical in both cases.

A REUSE-compliant Go file looks like:

```go
// SPDX-FileCopyrightText: 2026 Grégoire Favreau
// SPDX-License-Identifier: MIT

package main
```

## Using `--reuse`

```bash
# Add REUSE-compliant headers
addlicense --license MIT --author "Grégoire Favreau" --reuse .

# Check that all files are REUSE-compliant
addlicense --check --reuse .
```

In a config file (`.addlicenserc.yaml`), `--reuse` is not available — it is a CLI-only flag. Add it to your CI command explicitly.

## Idempotence

The idempotence detection scan looks (case-insensitively) for the string `copyright`. Since `FileCopyrightText` contains `copyright` as a substring, files with REUSE-formatted headers are correctly detected as already licensed and are not re-processed.

This means you can safely run `addlicense --reuse .` multiple times — it will not double-inject headers.

## Migration from standard to REUSE

If your project already has standard `Copyright` headers and you want to migrate to REUSE format:

1. Wait for the `--remove` flag (planned for v0.4.0) or strip headers manually.
2. Re-run with `--reuse`:

```bash
# Currently (before --remove is available):
# 1. Strip Copyright lines manually or via sed
# 2. Re-run addlicense
addlicense --license MIT --author "You" --reuse .
```

The upcoming `--update` flag (v0.4.0) will combine strip + inject in a single pass.

## Validating with fsfe/reuse-action

Once all headers are in REUSE format, you can validate with the official FSFE action:

```yaml
- name: REUSE compliance check
  uses: fsfe/reuse-action@v4
```

This validates not just the header format but the full REUSE specification (LICENSES/ directory, `.reuse/dep5` for unlicensable files, etc.).

**Note:** addlicense's `--reuse` flag handles the per-file header part of the spec. Full REUSE compliance also requires:
- A `LICENSES/` directory with the full licence texts (one file per SPDX identifier)
- A `.reuse/dep5` file for files that cannot carry headers (images, binaries)

The `--dep5` flag for generating `.reuse/dep5` is planned for v0.5.0.

## Why opt-in, not default?

Making `SPDX-FileCopyrightText:` the default would break existing tooling that relies on `grep -r "Copyright"` to find copyright notices. It would also silently change the format for all existing users on upgrade.

The `--reuse` flag is opt-in so existing projects are unaffected. If REUSE becomes the dominant standard, making it the default would require a major version bump (v2.0.0).
