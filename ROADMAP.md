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

---

## Planned versions

---

### v0.5.0 — Ecosystem integrations

**Target:** next minor release. Focuses on reducing friction at the developer workstation and CI layer.

#### Native GitHub Action

**What:** Ship `action.yml` so users can reference `uses: GregoireF/addlicense@v0.5.0` directly in their workflows without a separate binary download or Docker pull.

**Justification:**
- GitHub's [official guidance](https://docs.github.com/en/actions/sharing-automations/creating-actions/about-custom-actions) distinguishes three action types: JavaScript, Docker container, and composite. A composite action that calls the pre-built binary combines the reliability of a compiled binary with zero external downloads.
- Reference implementations: [golangci-lint-action](https://github.com/golangci/golangci-lint-action) ships the linter binary inside the action and achieves ~3 s cold start vs ~10–15 s for a `curl | tar` install. [shfmt-action](https://github.com/mfinelli/setup-shfmt) follows the same pattern.
- The binary already targets linux/amd64 and linux/arm64 (the two GitHub-hosted runner architectures). GoReleaser can attach the binaries to the `action.yml` repository via a post-release workflow.

**Pros:**
- No external HTTP request in CI — binary shipped inside the action, fully offline-capable.
- Single line of YAML for the end user: `uses: GregoireF/addlicense@v0.5.0`.
- Version pinning is inherited from the `@v0.5.0` action tag, not a separate download URL.

**Cons:**
- Requires a post-release step to commit the binary into the action repo (or the main repo). Adds ~30 s to the release pipeline.
- Action repo and release repo must stay in sync; a failed release step leaves them divergent.
- macOS and Windows runners would need a different binary — the action would need a `runs.using: composite` with OS-conditional steps.

**Decision:** Implement as a composite action using `runs.using: composite` with `${{ runner.os }}` detection. Binaries are downloaded from the GitHub Release (already published) rather than committed — avoids repository bloat while still using the action tag for version pinning.

---

#### `--year-range` flag

**What:** When updating a file that already has a header from a previous year, emit `Copyright 2023–2026 Author` instead of overwriting with `Copyright 2026 Author`.

**Justification:**
- The [SPDX specification §11.3](https://spdx.github.io/spdx-spec/v2.3/file-information/) allows year ranges in `FileCopyrightText` fields. The Linux kernel, LLVM, GCC, and most long-lived FOSS projects use year ranges in headers.
- Without this, `--update` on a multi-year project silently discards the original year — a legal regression for projects in jurisdictions where publication date affects copyright duration (e.g. EU author's rights under [Directive 2001/29/EC](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=celex%3A32001L0029), Art. 5).
- The [REUSE FAQ](https://reuse.software/faq/) explicitly recommends year ranges: "For a work that spans several years, you can list them all".

**Pros:**
- Preserves the full copyright timeline, which matters legally and historically.
- Composable with `--update`: `addlicense --update --year-range .`
- No breaking change: opt-in flag, existing behaviour unchanged.

**Cons:**
- Requires parsing the existing header to extract the original year (a read pass before the inject pass). Increases the per-file syscall count from 1 to 2 for files that already have a header.
- Edge cases: year ranges with gaps (`2021, 2023–2026`), ranges already in the header, corrupted dates. Parsing heuristics can fail on non-standard headers.
- Does not integrate cleanly with REUSE `dep5` bulk declarations, which have their own date syntax.

**Decision:** Implement with a conservative regex (`Copyright (20\d{2})`) that extracts the first year. Emit `YYYY–<current>` only when original year < current year. If parsing fails, fall back to single year with a warning.

---

#### `.reuse/dep5` stub generation

**What:** For files that cannot carry inline headers (images, fonts, compiled binaries, generated files), emit a [REUSE-compliant `dep5`](https://reuse.software/spec/#dep5) bulk-licence declaration at `.reuse/dep5`.

**Justification:**
- The [REUSE specification 3.3](https://reuse.software/spec/) §4.3 requires that _every_ file in a repository be covered by a licence — including non-source assets. Inline headers are impossible for binary assets. The only REUSE-compliant solution is a `.reuse/dep5` file.
- The [FSFE's `reuse` tool](https://github.com/fsfe/reuse-tool) generates `dep5` stubs but is a separate Python dependency. addlicense users who want REUSE compliance without adding a Python toolchain need a Go-native solution.
- European public sector projects increasingly require REUSE compliance as a condition of OSS procurement ([OpenChain ISO/IEC 5230](https://openchainproject.org/featured/2020/12/17/iso-5230)), and `dep5` coverage is a blocking check for `reuse lint`.

**Pros:**
- Closes the last gap between addlicense and full REUSE compliance.
- Pure Go — no additional runtime dependencies.
- Integrates naturally with the existing `scanner.Walk` output: files without an inline header and without a `dep5` entry are the `--check` failures.

**Cons:**
- `dep5` format is a subset of Debian's [`Machine-readable debian/copyright`](https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/) specification. Parsing and generating it correctly requires careful handling of glob patterns, paragraph delimiters, and multi-paragraph files.
- The feature adds a new output artefact (`.reuse/dep5`) that is outside the header-per-file model addlicense is built around. This expands the scope boundary.
- Most users wanting `dep5` already use the dedicated `reuse` CLI. Duplicating its functionality risks incomplete coverage and maintenance burden.

**Decision:** Implement as a separate `--dep5` flag (not the default). The flag emits a minimal `dep5` covering only unhandled file types (images, fonts, binaries). The `reuse lint` exit code is the acceptance test.

---

### v0.6.0 — Multi-author and organisation workflows

**Target:** second minor release after v0.5.0. Focuses on team and enterprise use cases.

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
