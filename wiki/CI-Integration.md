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

Until an official `.pre-commit-hooks.yaml` is published (planned for v0.5.0), use the `local` hook type:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: addlicense-check
        name: Check license headers
        entry: addlicense --check
        language: system
        pass_filenames: false
        types_or: [go, python, javascript, typescript, java, rust]
```

This requires `addlicense` to be installed on the developer's machine (e.g. via Homebrew).

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
