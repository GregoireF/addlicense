# CI Integration

## GitHub Actions

### Check on every push (recommended)

```yaml
name: License check

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  license:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install addlicense
        run: |
          curl -sSL https://github.com/GregoireF/addlicense/releases/latest/download/addlicense_linux_amd64.tar.gz \
            | tar -xz -C /usr/local/bin addlicense

      - name: Check license headers
        run: addlicense --check .
```

### Using Docker (no binary download)

```yaml
      - name: Check license headers
        run: |
          docker run --rm -v "$PWD:/src" -w /src \
            ghcr.io/gregoiref/addlicense:latest --check .
```

### Add headers automatically on PRs

```yaml
name: Add license headers

on:
  pull_request:
    types: [opened, synchronize]

permissions:
  contents: write

jobs:
  add-headers:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Install addlicense
        run: |
          curl -sSL https://github.com/GregoireF/addlicense/releases/latest/download/addlicense_linux_amd64.tar.gz \
            | tar -xz -C /usr/local/bin addlicense

      - name: Add missing headers
        run: addlicense --license MIT --author "Your Name" .

      - name: Commit if changed
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git diff --quiet || git commit -am "chore: add missing license headers"
          git push
```

### With a config file (`.addlicenserc.yaml`)

When the repo has a config file, the check command needs no extra flags:

```yaml
      - name: Check license headers
        run: addlicense --check .
```

---

## GitLab CI

```yaml
license-check:
  image: alpine:latest
  stage: test
  before_script:
    - apk add --no-cache curl tar
    - curl -sSL https://github.com/GregoireF/addlicense/releases/latest/download/addlicense_linux_amd64.tar.gz
        | tar -xz -C /usr/local/bin addlicense
  script:
    - addlicense --check .
```

---

## pre-commit

addlicense ships an official `.pre-commit-hooks.yaml` since v0.4.0. pre-commit installs the binary automatically via `go install` — no manual setup required.

### Check mode (recommended for most teams)

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/GregoireF/addlicense
    rev: v0.4.0
    hooks:
      - id: addlicense-check
        # optionally pass flags; or rely on .addlicenserc.yaml
        # args: [--license, MIT, --author, "Your Name"]
```

This hook exits non-zero if any tracked source file is missing a header. It does **not** modify files, making it safe for all CI runners.

### Auto-add mode

```yaml
repos:
  - repo: https://github.com/GregoireF/addlicense
    rev: v0.4.0
    hooks:
      - id: addlicense-add
        args: [--license, MIT, --author, "Your Name"]
```

This hook injects missing headers before the commit is finalised. If you use a config file (`.addlicenserc.yaml`), you can omit the `args`.

### Using both hooks

```yaml
repos:
  - repo: https://github.com/GregoireF/addlicense
    rev: v0.4.0
    hooks:
      - id: addlicense-add   # inject first
      - id: addlicense-check # then verify
```

---

## Makefile

```makefile
.PHONY: license-check license-add

license-check:
	addlicense --check .

license-add:
	addlicense --license MIT --author "Your Name" .
```

---

## Pin a specific version

To avoid breaking CI when a new release ships, pin to a specific version:

```yaml
      - name: Install addlicense v0.3.0
        run: |
          curl -sSL https://github.com/GregoireF/addlicense/releases/download/v0.3.0/addlicense_linux_amd64.tar.gz \
            | tar -xz -C /usr/local/bin addlicense
```

Or with Docker:

```yaml
        run: docker run --rm -v "$PWD:/src" -w /src ghcr.io/gregoiref/addlicense:v0.3.0 --check .
```

---

## Ignore generated or vendored files

addlicense's default ignore list covers `vendor`, `node_modules`, `*.pb.go`, `*.gen.go`. For additional paths, use `--ignore` or a config file:

```yaml
      - name: Check license headers
        run: addlicense --check --ignore "third_party,testdata,*.mock.go" .
```
