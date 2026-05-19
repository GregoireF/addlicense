# Roadmap

Tracks the direction of addlicense. Nothing here is a commitment — issues and real-world usage drive priorities.

---

## Context — French and European norms

License headers are not just developer hygiene: they are the foundation of open source compliance and, increasingly, a regulatory requirement in the European ecosystem.

### EUPL-1.2 — European Union Public Licence

Created by the European Commission, EUPL-1.2 is the reference licence for public sector software across EU member states. France's [DINUM](https://www.numerique.gouv.fr/) recommends it for state-produced software; the [socle interministériel de logiciels libres (SILL)](https://sill.etalab.gouv.fr/) lists it as preferred for contributions to the commons. Unlike most OSS licences, EUPL-1.2 is legally binding in all 23 EU official languages, which resolves jurisdiction ambiguity for cross-border public bodies.

It is copyleft but explicitly compatible with GPL-2.0, GPL-3.0, AGPL-3.0, EUPL-1.1, MPL-2.0, and others — making it safe to use alongside most major open source stacks.

**Status in addlicense:** built-in template added in v0.2.1.

### REUSE Specification — FSFE

The [REUSE specification](https://reuse.software/) from the Free Software Foundation Europe defines a machine-readable standard for per-file copyright and licensing. It is gaining adoption in European public sector projects (German federal agencies, Netherlands open government initiative) and is increasingly a criterion in OSS procurement. Key difference from the current addlicense format:

| Field | addlicense default | REUSE |
| --- | --- | --- |
| Copyright | `Copyright 2026 Author` | `SPDX-FileCopyrightText: 2026 Author` |
| Licence | `SPDX-License-Identifier: MIT` | `SPDX-License-Identifier: MIT` |

The `copyright` keyword is caught by addlicense's idempotence check regardless of format, so REUSE-formatted files are already correctly skipped. The gap is in *writing* REUSE-compliant headers.

**Planned:** `--reuse` flag in v0.3.0.

### CRA — Cyber Resilience Act

The EU Cyber Resilience Act (enforcement from 2027) requires software manufacturers to maintain a Software Bill of Materials (SBOM). SPDX identifiers — already emitted by addlicense — are the standard format for SBOM licence data. SPDX headers at the file level compose naturally into project-level SBOM documents.

**Positioning:** addlicense is already CRA-adjacent infrastructure. The path to full CRA tooling is a `--sbom` flag that aggregates scanned headers into an SPDX document — planned for v1.0.0.

### AGPL-3.0 — Closing the SaaS loophole

The GNU Affero GPL is widely used in European cloud-native and SaaS projects (Nextcloud, Framasoft ecosystem, many OSS-funded startups) specifically to close the "application service provider loophole" — companies that run modified OSS internally without distributing binaries are still required to publish source under AGPL. Increasingly relevant as French and EU public procurement favours OSS with strong copyleft guarantees.

**Status in addlicense:** built-in template added in v0.2.1.

---

## Planned versions

### v0.3.0 — REUSE compliance

- `--reuse` flag: emit `SPDX-FileCopyrightText:` instead of `Copyright`
- Validate output against [fsfe/reuse-action](https://github.com/fsfe/reuse-action) in CI (dogfooding)
- Optionally: `.reuse/dep5` stub generation for assets that cannot carry inline headers (images, binaries)

### v0.4.0 — Power user features

- `--remove` flag: strip existing headers — needed when migrating between licences (MIT → EUPL-1.2, Apache-2.0 → AGPL-3.0)
- Parallel file processing: `filepath.WalkDir` with goroutines for large monorepos
- `--verbose` / `--quiet` flags for CI tuning
- JSON output (`--format json`) for integration with other tools

### v1.0.0 — Ecosystem integration

- Native GitHub Action (`action.yml`) — no binary download, no Docker pull, instant cold start
- SBOM generation: aggregate scanned headers into an SPDX 2.3 document (CRA readiness)
- Stable public Go API — `github.com/GregoireF/addlicense/pkg` importable as a library
- Semantic versioning guarantee on the public API from this point forward

---

## Open questions

**REUSE as default vs. opt-in** — The `SPDX-FileCopyrightText:` format is more future-proof and better aligned with the FSFE ecosystem, but it breaks the convention assumed by most existing tooling (`grep -r "Copyright"`). The `--reuse` flag preserves backward compatibility; making it the default would require a major version bump. Lean toward flag for now.

**GitHub Action packaging** — A native action requires the binary to ship inside the action repository (no Docker). This works but adds complexity to the release pipeline (cross-compile + commit to action repo on each tag). The Docker action path is simpler but adds ~10 seconds of pull time per CI run. Decision deferred to v1.0.0.

**EUPL multi-language legal binding** — EUPL-1.2 exists in 23 languages, all legally binding. Should addlicense expose `--lang fr` to write `EUPL-1.2` with a French-language note? Almost certainly out of scope — the SPDX identifier is language-agnostic and sufficient for all tooling downstream.

**Scope boundary** — addlicense is deliberately minimal. Features that belong in a full SBOM tool (CycloneDX, SPDX-tools) should not be replicated here. The `--sbom` flag in v1.0.0 is a convenience wrapper, not a replacement for dedicated tooling.
