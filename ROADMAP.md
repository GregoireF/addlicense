# Roadmap

Tracks the direction of addlicense. Nothing here is a commitment — issues and real-world usage drive priorities.

See the [GitHub Wiki](../../wiki) for architecture decisions, design principles, and detailed contributor guides.

---

## Context — French and European norms

License headers are not just developer hygiene: they are the foundation of open source compliance and, increasingly, a regulatory requirement in the European ecosystem.

### EUPL-1.2 — European Union Public Licence

Created by the European Commission, EUPL-1.2 is the reference licence for public sector software across EU member states. France's [DINUM](https://www.numerique.gouv.fr/) recommends it for state-produced software; the [socle interministériel de logiciels libres (SILL)](https://sill.etalab.gouv.fr/) lists it as preferred for contributions to the commons. Unlike most OSS licences, EUPL-1.2 is legally binding in all 23 EU official languages, which resolves jurisdiction ambiguity for cross-border public bodies.

It is copyleft but explicitly compatible with GPL-2.0, GPL-3.0, AGPL-3.0, EUPL-1.1, MPL-2.0, and others — making it safe to use alongside most major open source stacks.

**Status:** built-in template since v0.2.0 (shipped as v0.3.0).

### REUSE Specification — FSFE

The [REUSE specification](https://reuse.software/) from the Free Software Foundation Europe defines a machine-readable standard for per-file copyright and licensing. It is gaining adoption in European public sector projects (German federal agencies, Netherlands open government initiative) and is increasingly a criterion in OSS procurement.

| Field | addlicense default | REUSE (`--reuse`) |
| --- | --- | --- |
| Copyright | `Copyright 2026 Author` | `SPDX-FileCopyrightText: 2026 Author` |
| Licence | `SPDX-License-Identifier: MIT` | `SPDX-License-Identifier: MIT` |

**Status:** `--reuse` flag shipped in v0.3.0.

### CRA — Cyber Resilience Act

The EU Cyber Resilience Act (enforcement from 2027) requires software manufacturers to maintain a Software Bill of Materials (SBOM). SPDX identifiers — already emitted by addlicense — are the standard format for SBOM licence data. SPDX headers at the file level compose naturally into project-level SBOM documents.

**Positioning:** addlicense is already CRA-adjacent infrastructure. The path to full CRA tooling is a `--sbom` flag that aggregates scanned headers into an SPDX document — planned for v1.0.0.

### AGPL-3.0 — Closing the SaaS loophole

Widely used in European cloud-native and SaaS projects (Nextcloud, Framasoft ecosystem) to close the "application service provider loophole". Increasingly relevant as French and EU public procurement favours OSS with strong copyleft guarantees.

**Status:** built-in template since v0.2.0 (shipped as v0.3.0).

---

## Released

### v0.1.0 — 2026-05-18
Core CLI: unified command, SPDX templates (MIT, Apache-2.0, GPL-3.0, MPL-2.0, BSD-2/3-Clause), custom templates, config auto-detection, idempotence, shebang preservation, check mode. GoReleaser multi-platform + Homebrew tap + Docker.

### v0.2.0 — 2026-05-19
Extended language support (HTML/Vue/Svelte, CSS/SCSS, proto, SQL), Docker GHCR, Codecov integration, `.addlicenserc.yaml` dogfooding, Dependabot.

### v0.3.0 — 2026-05-20
EU licence templates (EUPL-1.2, AGPL-3.0-only, LGPL-2.1/3.0-only). `--reuse` flag for FSFE/REUSE compliance. 90 %+ coverage, golangci-lint v2 config, full gofmt compliance. Contribution infrastructure: issue templates, PR template, CODEOWNERS, GitHub Wiki, Discussions.

---

## Planned versions

### v0.4.0 — Header removal and power-user flags

- **`--remove` flag**: strip existing headers — needed when migrating between licences (MIT → EUPL-1.2, Apache-2.0 → AGPL-3.0). Detect the header block (already identified by the idempotence scan window), remove it, leave the rest of the file untouched.
- **`--update` flag**: combine remove + inject in a single pass. The common migration workflow: `addlicense --update --license EUPL-1.2 .`
- **`--verbose` / `--quiet` flags**: `--verbose` prints every file processed (not just injected ones); `--quiet` suppresses all stdout for pure CI usage (non-zero exit on error is still emitted to stderr).
- **`--format json` output**: machine-readable report (file path, status: added/skipped/missing/error) for integration with other tools (dashboards, pre-commit hooks with structured output).
- **Parallel file processing**: `filepath.WalkDir` with a worker pool (bounded goroutines) for large monorepos. Deferred until benchmarks justify it — sequential is under 1s for < 10 000 files; goroutine overhead is real.

### v0.5.0 — Ecosystem and DX

- **Pre-commit hook support**: publish an official `.pre-commit-hooks.yaml` so `addlicense` can be added to [pre-commit](https://pre-commit.com/) without manual configuration.
- **Native GitHub Action** (`action.yml`): no Docker pull, no binary download, instant cold start. Cross-compiles the binary and ships it inside the action repository on each release. Removes the 10 s Docker image pull from CI.
- **Editor integration hints**: VS Code extension or devcontainer feature that runs `addlicense --check` in the background and highlights unlicensed files (out of scope for the binary itself, but documentation and integration guide).
- **`--year-range` flag**: emit `Copyright 2024–2026 Author` for files that span multiple years. Useful for long-lived projects or when re-running addlicense years after initial injection.
- **`.reuse/dep5` stub generation**: for assets that cannot carry inline headers (images, binaries, generated files), emit a REUSE-compliant `dep5` bulk-licence declaration.

### v1.0.0 — Stable API and SBOM

- **Stable public Go API** — `github.com/GregoireF/addlicense/pkg` importable as a library with a stable, documented interface. Semantic versioning guarantee from this point forward.
- **SBOM generation (`--sbom`)**: aggregate scanned SPDX headers into an [SPDX 2.3](https://spdx.github.io/spdx-spec/v2.3/) document (JSON or tag-value). CRA readiness: one command produces a file-level SBOM suitable for submission to procurement or auditors.
- **`--check --format json` exit contracts**: stable exit codes and JSON schema for downstream tooling (lint runners, CI dashboards).
- **Multi-licence support**: some files legitimately carry two identifiers (`SPDX-License-Identifier: MIT OR Apache-2.0`). v1.0.0 preserves but does not inject multi-licence expressions.

---

## Open questions

**REUSE as default vs. opt-in** — `SPDX-FileCopyrightText:` is more future-proof and better aligned with the FSFE ecosystem, but breaks `grep -r "Copyright"` conventions. The `--reuse` flag preserves backward compatibility; making it the default would require a major version bump. Lean toward flag for now.

**`--remove` heuristic** — the current scan window (top 20 lines, `SPDX-License-Identifier:` or `copyright`) reliably detects headers, but removing them requires knowing where the header ends. Heuristic: remove all consecutive comment lines from the top of the file (after shebang if present). Edge case: files where the first non-header line is also a comment. Needs a test matrix before shipping.

**GitHub Action packaging** — a native action requires the binary to ship inside the action repository (no Docker). This adds complexity to the release pipeline (cross-compile + commit to action repo on each tag). The Docker action path is simpler but adds ~10 s of pull time per CI run. Decision deferred to v0.5.0.

**Year-range detection** — to emit `2024–2026` instead of `2026`, addlicense would need to read the existing header (if present) to extract the original year, then compare with the current year. This adds a read pass before the inject pass. The complexity may not be worth it for most users.

**EUPL multi-language legal binding** — EUPL-1.2 exists in 23 languages, all legally binding. Should addlicense expose `--lang fr` to write `EUPL-1.2` with a French-language note? Almost certainly out of scope — the SPDX identifier is language-agnostic and sufficient for all tooling downstream.

**Scope boundary** — addlicense is deliberately minimal. Features that belong in a full SBOM tool (CycloneDX, SPDX-tools) should not be replicated here. The `--sbom` flag in v1.0.0 is a convenience wrapper, not a replacement for dedicated tooling.
