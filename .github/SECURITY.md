# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| latest (`v1.x`) | ✅ |
| older tags | ❌ — upgrade to the latest release |

## Reporting a vulnerability

Prefer **GitHub private advisories** so the issue stays confidential until a fix is available:

1. Open [Security → Report a vulnerability](https://github.com/GregoireF/addlicense/security/advisories/new)
2. Include: affected version, steps to reproduce, and potential impact

Alternatively, email **gfavreau.wrprojects@gmail.com** with the subject `[addlicense] Security`.

**Response SLA**

| Milestone | Target |
|---|---|
| Acknowledgement | 48 h |
| Triage + severity | 7 days |
| Fix — critical / high | 14 days |
| Fix — moderate | 30 days |

## Supply-chain controls

| Control | Detail |
|---|---|
| SHA-pinned Actions | All `uses:` steps reference a commit SHA, not a mutable tag |
| Dependabot | Weekly updates for Go modules and GitHub Actions |
| CODEOWNERS | All changes require review by @GregoireF |
| GoReleaser checksums | `checksums.txt` published alongside every release |
| Isolated builds | Binaries are built in ephemeral CI runners — no local build artifacts shipped |
| SBOM | `--sbom` flag generates an SPDX 2.3 document for dependency auditing (EU CRA compliant) |
| Minimal permissions | Each workflow declares only the permissions it needs |
