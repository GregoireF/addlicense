# Roadmap

Tracks the direction of addlicense. Nothing here is a commitment — issues and real-world usage drive priorities.

See the [GitHub Wiki](../../wiki) for architecture decisions, design principles, and detailed contributor guides.

---

## Context — French and European norms

License headers are not just developer hygiene: they are the foundation of open source compliance and, increasingly, a regulatory requirement in the European ecosystem.

### EUPL-1.2 — European Union Public Licence

Created by the European Commission, EUPL-1.2 is the reference licence for public sector software across EU member states. France's [DINUM](https://www.numerique.gouv.fr/) recommends it for state-produced software; the [socle interministériel de logiciels libres (SILL)](https://sill.etalab.gouv.fr/) lists it as preferred for contributions to the commons. Unlike most OSS licences, EUPL-1.2 is legally binding in all 23 EU official languages, which resolves jurisdiction ambiguity for cross-border public bodies.

It is copyleft but explicitly compatible with GPL-2.0, GPL-3.0, AGPL-3.0, EUPL-1.1, MPL-2.0, and others — making it safe to use alongside most major open source stacks.

**Status:** built-in template since v0.2.0.

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

**Status:** built-in template since v0.2.0.

---

## Released

### v0.1.0 — 2026-05-18
Core CLI: unified command, SPDX templates (MIT, Apache-2.0, GPL-3.0, MPL-2.0, BSD-2/3-Clause), custom templates, config auto-detection, idempotence, shebang preservation, check mode. GoReleaser multi-platform + Homebrew tap + Docker.

### v0.2.0 — 2026-05-19
Extended language support (HTML/Vue/Svelte, CSS/SCSS, proto, SQL), Docker GHCR, Codecov integration, `.addlicenserc.yaml` dogfooding, Dependabot.

### v0.3.0 — 2026-05-20
EU licence templates (EUPL-1.2, AGPL-3.0-only, LGPL-2.1/3.0-only). `--reuse` flag for FSFE/REUSE compliance. 90%+ coverage, golangci-lint v2 config, full gofmt compliance. Contribution infrastructure: issue templates, PR template, CODEOWNERS, GitHub Wiki, Discussions.

### v0.4.0 — 2026-05-20
Header removal and power-user flags. `--remove` / `-R` strips existing headers; `--update` / `-u` replaces them in one pass. `--dry-run` / `-n` previews without writing. `--verbose` / `-v` and `--quiet` / `-q` control output verbosity. `--format json` emits JSON Lines for machine consumption. Parallel worker pool (`--workers`). `reuse:` field in `.addlicenserc.yaml`. Official `.pre-commit-hooks.yaml` (`addlicense-check` + `addlicense-add`). Mutual-exclusion validation for conflicting flags. Coverage raised to 90.2%.

### v0.5.0 — 2026-05-20
Ecosystem integrations. `--year-range` preserves the original copyright year during `--update`, emitting `YYYY-YYYY` ranges (opt-in, backward-compatible, composable with `--reuse`). `--dep5` generates a REUSE-compliant `.reuse/dep5` bulk-licence declaration for files that cannot carry inline headers (images, fonts, binaries). Native GitHub Action (`action.yml`) — composite action with OS/arch detection, binary downloaded from the GitHub Release tag, zero external services, `ubuntu-*` / `macos-*` / `windows-*` × amd64 + arm64.

### v0.6.0 — 2026-05-22
Multi-author and diff. `--author "Alice, Bob"` emits one copyright line per author, composable with `--reuse` and `--year-range`. `--diff` emits JSON Lines with the rendered header for each file that would change (no writes; exit 1 if any file would be modified) — designed for PR validation and compliance review without destructive writes. Quality: injector and pipeline benchmark suites, `SilenceErrors` fix (error no longer printed twice). 9 new integration tests, 4 new CLI tests.

### v0.7.0 — 2026-05-22
Security and correctness. `injector.Inject` and `injector.Remove` now preserve original file permissions — executable scripts (`chmod +x`) no longer silently lose their execute bit after a run. Coverage raised from 89.1% to 90.4% by covering `diffFile` error paths and removing dead code in `emit`. `--ignore` and multi-author `--update` semantics documented.

### v0.8.0 — 2026-05-22
Public Go API. `github.com/GregoireF/addlicense/pkg/addlicense` exposes `Options` and `Run` as a stable, documented library interface. Distinct struct (not type alias) keeps the `internal/` packages free to evolve without breaking callers. `DefaultIgnore`, `FormatText`, `FormatJSON` constants re-exported. 12 public API tests added. README updated with library usage examples.

### v0.9.0 — 2026-05-22
Language expansion and `--author-file`. Added Lua, Nix, Zig, `.dockerfile`, `.mk`, Markdown (`.md`/`.mdx`) — 8 new file types, zero new dependencies, fully backward-compatible. `--author-file` reads copyright holders from a text file (one per line, `#` comments ignored), composable with `--update` and `--reuse`. 9 new tests added across header, integration, and CLI packages.

---

## Planned versions

---

### v1.0.0 — Stable API and SBOM

**Target:** after v0.9.x stabilisation, API freeze, and community validation of the `pkg/` surface.

#### API stability guarantee

With `pkg/addlicense` shipped in v0.8.0, v1.0.0 freezes the public API under semantic versioning: any rename, signature change, or removal of an exported symbol is a breaking change requiring v2.0.0. The internal packages remain free to evolve.

**Pre-1.0.0 checklist:**
- Survey `pkg/addlicense` API for any fields or names likely to be regretted at v2.0.0.
- Decide whether `AuthorFile string` should be added to `pkg.Options` (currently CLI-only — library callers read files themselves).
- Confirm `DefaultIgnore` slice behaviour (currently a shared reference — should it return a copy?).

#### SBOM generation (`--sbom`)

**What:** Aggregate scanned SPDX headers into an [SPDX 2.3](https://spdx.github.io/spdx-spec/v2.3/) document (JSON or tag-value format).

**Justification:** The EU Cyber Resilience Act (enforcement 2027) mandates an SBOM for CE-marked software. SPDX identifiers — already emitted by addlicense — are the standard format for SBOM licence data at the file level.

**Decision:** Use [`spdx/tools-golang`](https://github.com/spdx/tools-golang) for serialisation. Scope to file-level SPDX documents only. Document clearly that this satisfies CRA file-level requirements but does not replace a full dependency SBOM tool (syft, cdxgen).

**Pros:** Direct CRA readiness for existing users — one additional flag. Differentiator vs. competitors (none produce SPDX output).

**Cons:** SPDX 2.3 document structure is complex; adds a dependency; minimal `--sbom` omits package-level metadata auditors may still require separately.

---

### v0.6.0 content (now released — see Released section above)

#### Multi-author copyright lines

**What:** `--author "Alice Dupont,Bob Martin"` emits two `Copyright` lines (or `SPDX-FileCopyrightText` lines in `--reuse` mode):

```
// Copyright 2026 Alice Dupont
// Copyright 2026 Bob Martin
// SPDX-License-Identifier: MIT
```

**Justification:**
- The [SPDX specification §3.8](https://spdx.github.io/spdx-spec/v2.3/package-information/) and the REUSE specification both explicitly support multiple `FileCopyrightText` entries per file. `reuse lint` validates multi-author headers correctly.
- In European employment law ([Directive 2009/24/EC](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=celex%3A32009L0024) on computer programs), copyright co-authorship by multiple contributors is the norm for team software. A single corporate author line is legally weaker than named contributors.
- Most enterprise monorepos have team ownership per module/package. `--author` accepting a comma-separated list is the minimal addition needed.

**Pros:**
- Directly supported by SPDX and REUSE — no spec deviation.
- Backward-compatible: single author continues to work unchanged.
- Useful for OSS foundations (Apache, Eclipse) where the header must name the foundation AND the original contributor.

**Cons:**
- Interacts non-trivially with `--update`: when updating, existing author lines must be preserved or replaced. Deciding the merge semantics (append / replace / error) requires a flag or a convention.
- Template rendering complexity increases: `{{.CopyrightLine}}` becomes a multi-line string or a slice, breaking the current single-string contract.

**Decision:** Extend `header.Data.CopyrightLine` to a slice internally; `Render` joins them with `\n` before template execution. `--update` in multi-author mode appends new authors to existing lines rather than replacing.

---

#### Licence migration report (`--diff`)

**What:** `addlicense --diff --from MIT --to Apache-2.0 .` performs a dry run and emits a structured diff showing exactly which files would change and the before/after header content, without writing to disk.

**Justification:**
- Large-scale licence migrations (MIT → Apache-2.0, GPL-2.0+ → GPL-3.0-or-later) are high-risk changes in regulated software. The [FSFE's licence compatibility matrix](https://reuse.software/faq/#license-compatibility) and [OSI's compatibility guide](https://opensource.org/licenses) both advise reviewing the scope of a migration before committing.
- Currently, `--update --dry-run --format json` gives a file list but not the content diff. Legal reviewers need to see the before/after header, not just the filename.
- `git diff` after a dry-run-less `--update` is destructive. A `--diff` mode is safer for review workflows.

**Pros:**
- Zero-risk preview for compliance officers and legal teams.
- JSON output from `--diff --format json` can feed into audit dashboards.
- Complements the existing `--dry-run` flag without replacing it.

**Cons:**
- Requires reading every file twice (current + rendered), doubling I/O on large repos.
- The output format must be specified carefully to avoid ambiguity with `git diff` or `unified diff` conventions. A custom JSON schema is safer but requires documentation.
- Edge case: files where the header is spread across non-contiguous comment blocks — the diff would show partial matches.

**Decision:** Implement `--diff` as a superset of `--dry-run`. The JSON output adds `"before"` and `"after"` fields to the existing record schema, preserving backward compatibility.

---

### v1.0.0 — Stable API and SBOM

#### Stable public Go API

**What:** Expose `github.com/GregoireF/addlicense/pkg` as an importable Go library with a stable, documented interface. Semantic versioning guarantee applies from this point forward.

**Justification:**
- The internal packages (`internal/header`, `internal/injector`, `internal/scanner`) are mature and stable. Making them public allows embedding addlicense in CI runners, code generators, and IDEs without invoking a subprocess.
- Go module conventions ([pkg.go.dev standards](https://go.dev/blog/package-names)) recommend exposing a `pkg/` tree for library consumers, keeping `internal/` for implementation details.
- Reference: [golangci-lint exposes `pkg/`](https://github.com/golangci/golangci-lint/tree/master/pkg) for use in editor integrations; [goreleaser/goreleaser](https://github.com/goreleaser/goreleaser) exposes its config parser as a library.

**Pros:**
- Enables IDE plugins and editor extensions to call the injector natively.
- Reduces subprocess overhead for tools that integrate addlicense programmatically.

**Cons:**
- Once public, the API is a compatibility commitment. Any rename or signature change is a breaking change requiring a major version bump.
- Requires a public API design review before v1.0.0 — the internal structs were not designed with external consumers in mind.

---

#### SBOM generation (`--sbom`)

**What:** Aggregate scanned SPDX headers into an [SPDX 2.3](https://spdx.github.io/spdx-spec/v2.3/) document (JSON or tag-value format), suitable for submission under the EU Cyber Resilience Act.

**Justification:**
- The CRA (enforcement 2027) mandates an SBOM for CE-marked software products. [ENISA's CRA guidance](https://www.enisa.europa.eu/publications/cyber-resilience-act-requirements-standards-mapping) identifies SPDX 2.3 as the preferred format for file-level licence data.
- addlicense already has all the data needed: for each file, it knows the SPDX identifier, the copyright year, and the author. The `--sbom` flag just aggregates this into the standard document structure.
- The [Linux Foundation's OpenChain ISO 5230](https://openchainproject.org/featured/2020/12/17/iso-5230) compliance programme requires a "written representation of the SBOM". addlicense-generated SPDX documents would satisfy this requirement at the file-licence level.

**Pros:**
- Direct CRA readiness for existing addlicense users — one additional flag, no new toolchain.
- Differentiator vs. competitors (google/addlicense, skywalking-eyes): none produce SPDX output.
- SPDX is also the input format for `syft`, `grype`, and most vulnerability scanners — composability is built-in.

**Cons:**
- SPDX 2.3 document structure is complex (Document, Package, File, Snippet, Relationship sections). A complete implementation requires careful handling of SPDX-ID uniqueness, verification codes, and relationship graphs.
- A minimal `--sbom` that only covers file-licence data omits package-level metadata (dependency graph, purl identifiers), which auditors may still require separately.
- The [spdx/tools-golang](https://github.com/spdx/tools-golang) library exists and should be used rather than re-implementing the SPDX serializer — adds a dependency.

**Decision:** Use `spdx/tools-golang` for serialisation. Scope the v1.0.0 implementation to file-level SPDX documents only (no package graph). Clearly document that this satisfies CRA file-level requirements but does not replace a full dependency SBOM tool (syft, cdxgen).

---

## Open questions

**REUSE as default vs. opt-in** — `SPDX-FileCopyrightText:` is more future-proof and better aligned with the FSFE ecosystem, but breaks `grep -r "Copyright"` conventions. The `--reuse` flag preserves backward compatibility; making it the default would require a major version bump. Lean toward flag for now.

**GitHub Action packaging** — a native action requires the binary to ship inside the action repository (no Docker). This adds complexity to the release pipeline (cross-compile + commit to action repo on each tag). The Docker action path is simpler but adds ~10 s of pull time per CI run. Decision deferred to v0.5.0.

**Year-range detection** — to emit `2024–2026` instead of `2026`, addlicense would need to read the existing header (if present) to extract the original year, then compare with the current year. This adds a read pass before the inject pass. The complexity may not be worth it for most users.

**EUPL multi-language legal binding** — EUPL-1.2 exists in 23 languages, all legally binding. Should addlicense expose `--lang fr` to write `EUPL-1.2` with a French-language note? Almost certainly out of scope — the SPDX identifier is language-agnostic and sufficient for all tooling downstream.

**Scope boundary** — addlicense is deliberately minimal. Features that belong in a full SBOM tool (CycloneDX, SPDX-tools) should not be replicated here. The `--sbom` flag in v1.0.0 is a convenience wrapper, not a replacement for dedicated tooling.

**Multi-author merge semantics** — when `--update` is combined with multiple authors, the safest default is append (keep old authors, add new ones). But in practice users may want replacement (new org name replaces old). A `--update-strategy append|replace` sub-flag may be needed.
