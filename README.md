# addlicense

> Fast, minimal and developer-friendly license header manager.

`addlicense` helps you automatically add, update and maintain license headers across your projects.

Designed for:
- OSS maintainers
- monorepos
- CI pipelines
- developer tooling ecosystems
- compliance automation

---

# Features

- blazing fast
- recursive scanning
- multi-language support
- customizable templates
- CI-friendly
- idempotent operations
- monorepo aware
- zero-config defaults

---

# Why

Most license tools are:
- outdated
- hard to configure
- language-limited
- not CI-native

`addlicense` focuses on:
- simplicity
- modern DX
- automation
- developer workflows

---

# Installation

```bash
npm install -g addlicense
```

---

# Quick Start

## Add MIT headers

```bash
addlicense --license MIT .
```

---

## Custom author

```bash
addlicense --license MIT --author "Grégoire"
```

---

## Custom template

```bash
addlicense --template ./header.txt
```

---

# Supported Languages

- TypeScript
- JavaScript
- Go
- Rust
- Python
- Shell
- Terraform
- YAML
- Dockerfile
- Java
- C/C++
- more coming soon

---

# Examples

## Add headers to a monorepo

```bash
addlicense packages/** src/**
```

---

## Ignore files

```bash
addlicense . --ignore dist,node_modules
```

---

# CI Integration

Example GitHub Actions workflow:

```yaml
name: License Check

on: [push]

jobs:
  license:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - run: npm install -g addlicense

      - run: addlicense --check .
```

---

# Philosophy

Licensing should be:
- automatic
- invisible
- reproducible
- enforced by tooling

---

# Roadmap

- SPDX support
- SBOM integration
- provenance metadata
- autofix mode
- workspace presets
- git hooks
- license compatibility checks

---

# License

MIT