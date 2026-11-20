# Release Process

Releases are fully automated via GoReleaser triggered by a `v*` tag. No manual steps are required beyond tagging.

## Prerequisites

- All changes committed and pushed to `main`
- CI green (Build & Test ✅ + Lint ✅)
- `CHANGELOG.md` `[Unreleased]` section promoted to `[X.Y.Z]`
- `ROADMAP.md` released-versions section updated

## Cutting a release

```bash
git tag v0.5.0
git push origin v0.5.0
```

This triggers the GoReleaser GitHub Actions workflow (`.github/workflows/release.yml`), which:

1. **Builds** multi-platform binaries: Linux, macOS, Windows × amd64 + arm64
2. **Creates** a GitHub Release with the binaries and generated release notes
3. **Pushes** a Docker image to GHCR (`ghcr.io/gregoiref/addlicense:v0.4.0` + `:latest`)
4. **Attests** every archive (SLSA build provenance, `actions/attest-build-provenance`)

Within 30 minutes, the tap `GregoireF/homebrew-tap` picks the release up on its own (see below).

## Versioning rules

This project follows [Semantic Versioning](https://semver.org/):

| Change | Version bump |
|:--|:--|
| New flag, new language, new SPDX template | Minor (`0.X.0`) |
| Bug fix | Patch (`0.0.X`) |
| Breaking change (flag rename, default behaviour change) | Major (`X.0.0`) |

During the pre-1.0 phase (`0.x.y`), minor versions may contain breaking changes if documented in the CHANGELOG.

## GoReleaser configuration

Key decisions in [`.goreleaser.yaml`](https://github.com/GregoireF/addlicense/blob/main/.goreleaser.yaml):

- **No Homebrew section** — the tap pulls releases itself; addlicense holds no token for it.
- **Docker `FROM scratch`** — the image is ~3 MB, contains only the binary. No shell, no libc. Use `docker run --rm -v "$PWD:/src" -w /src ghcr.io/gregoiref/addlicense:latest`.
- **CGO disabled** — `CGO_ENABLED=0` ensures the binary runs on any Linux distribution.

## Homebrew tap

The tap repository is [`GregoireF/homebrew-tap`](https://github.com/GregoireF/homebrew-tap). Every 30 minutes it reads the latest stable release, checks each archive against `checksums.txt` and its build provenance, writes `Formula/addlicense.rb` and the Scoop manifest `bucket/addlicense.json`, tests the installs on macOS, Linux and Windows, then publishes. The archive names are declared in the tap's `tools.yml`: rename an archive in `.goreleaser.yaml` only together with it.

```bash
brew install gregoiref/tap/addlicense
scoop bucket add gregoiref https://github.com/GregoireF/homebrew-tap; scoop install gregoiref/addlicense
```

**Note:** since 2026-11, addlicense is a formula again (`Formula/addlicense.rb`), not a cask: macOS quarantines what a cask installs, and the binaries are not notarized. Users of the cask run `brew uninstall --cask addlicense && brew install gregoiref/tap/addlicense`.

## Release checklist

- [ ] `CHANGELOG.md` `[Unreleased]` → `[X.Y.Z] — YYYY-MM-DD`
- [ ] `CHANGELOG.md` new empty `[Unreleased]` section added
- [ ] `ROADMAP.md` released-versions summary updated
- [ ] CI green on `main`
- [ ] Tag pushed: `git tag vX.Y.Z && git push origin vX.Y.Z`
- [ ] GoReleaser workflow completes successfully
- [ ] GitHub Release published with correct binaries
- [ ] Docker image pushed to GHCR
- [ ] Within 30 minutes: `Formula/addlicense.rb` updated in the tap

## Hotfix process

For critical bug fixes on a released version:

```bash
git checkout -b hotfix/v0.4.1 v0.4.0
# apply fix
git commit -m "fix: ..."
git push origin hotfix/v0.4.1
# open PR targeting main
# after merge, tag from main
git tag v0.4.1
git push origin v0.4.1
```
