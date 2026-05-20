# Adding an SPDX Template

Built-in templates produce well-known SPDX headers. Adding a new one takes under 10 minutes.

## Step 1 — Add a `case` in `BuiltinTemplate`

Open [`internal/header/header.go`](https://github.com/GregoireF/addlicense/blob/main/internal/header/header.go) and find `BuiltinTemplate`. Add a `case` for the new identifier:

```go
func BuiltinTemplate(license string) string {
    switch strings.ToUpper(license) {
    // ...existing cases...

    case "OSL-3.0", "OSL":
        return "{{.CopyrightLine}}\nSPDX-License-Identifier: OSL-3.0"
    }
}
```

**Rules:**
- Use the canonical SPDX identifier as the primary `case` value (e.g. `"OSL-3.0"`).
- Add common short aliases as additional `case` values (e.g. `"OSL"`).
- All cases must be **upper-cased** — the switch receives `strings.ToUpper(license)`.
- The template body uses `{{.CopyrightLine}}` for the copyright line. Do not use `Copyright {{.Year}}...` directly — that bypasses `--reuse` mode.
- The second line is always `SPDX-License-Identifier: <canonical-id>`.

**Available template fields:**

| Field | Value |
| :-- | :-- |
| `{{.CopyrightLine}}` | `"Copyright 2026 Author"` or `"SPDX-FileCopyrightText: 2026 Author"` |
| `{{.Year}}` | `2026` |
| `{{.Author}}` | `"Grégoire"` (empty string if not set) |
| `{{.License}}` | Raw identifier as passed by the user |
| `{{.SPDX}}` | Upper-cased identifier |

## Step 2 — Add an integration test

Open [`tests/integration/runner_test.go`](https://github.com/GregoireF/addlicense/blob/main/tests/integration/runner_test.go) and add a test that verifies the canonical SPDX identifier appears in the output:

```go
func TestRun_OSL3License(t *testing.T) {
    root := makeProject(t, map[string]string{"main.go": "package main\n"})
    if err := runner.Run(defaultOpts("OSL-3.0", root)); err != nil {
        t.Fatal(err)
    }
    content := readFile(t, filepath.Join(root, "main.go"))
    if !strings.Contains(content, "SPDX-License-Identifier: OSL-3.0") {
        t.Errorf("missing OSL-3.0 identifier:\n%s", content)
    }
}
```

## Step 3 — Update README and CHANGELOG

Add the identifier to the **Supported SPDX identifiers** list in `README.md` and document the addition under `[Unreleased]` in `CHANGELOG.md`.

## Checklist

- [ ] `case` added in `BuiltinTemplate` with canonical SPDX identifier + aliases
- [ ] Template body uses `{{.CopyrightLine}}` (not inline `Copyright {{.Year}}`)
- [ ] Integration test added and passes
- [ ] `go test -race ./...` passes
- [ ] README identifiers list updated
- [ ] CHANGELOG `[Unreleased]` updated
