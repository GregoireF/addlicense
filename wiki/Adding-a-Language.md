# Adding a Language

Adding support for a new file extension takes less than 5 minutes and requires changes to two files.

## Step 1 — Add the extension to `langs` in `header.go`

Open [`internal/header/header.go`](https://github.com/GregoireF/addlicense/blob/main/internal/header/header.go) and find the `langs` map. Add your entry in the appropriate comment group, sorted alphabetically within the group.

```go
var langs = map[string]Lang{
    // Go / systems
    ".go":  {LineComment: "//"},
    ".rs":  {LineComment: "//"},
    // ...

    // Your new group or entry:
    ".zig": {LineComment: "//"},
}
```

**Comment style reference:**

| Pattern | Example | Use when |
| :-- | :-- | :-- |
| `{LineComment: "//"}` | `// Copyright...` | Most C-family languages |
| `{LineComment: "#"}` | `# Copyright...` | Python, Shell, YAML, Ruby |
| `{LineComment: "--"}` | `-- Copyright...` | SQL, Lua, Haskell |
| `cBlock` (pre-defined) | `/* ... */` | C, Java, CSS |
| `htmlComment` (pre-defined) | `<!-- ... -->` | HTML, Vue, Svelte |
| Custom block | see below | Any language with block-only comments |

**Custom block comment example:**

```go
// Nix uses /* */ but without the leading " * " per line
".nix": {BlockOpen: "/*", BlockClose: "*/", BlockPrefix: "  "},
```

**gofmt alignment rule:** the `langs` map uses per-comment-group alignment (each comment line resets gofmt's tabwriter). Align values to the longest key in the same group + 1 space. Run `gofmt -d internal/header/header.go` locally to confirm.

## Step 2 — Add an integration test

Open [`tests/integration/runner_test.go`](https://github.com/GregoireF/addlicense/blob/main/tests/integration/runner_test.go) and add a test case. At minimum, verify the correct comment delimiters appear in the output:

```go
func TestRun_ZigLineComment(t *testing.T) {
    root := makeProject(t, map[string]string{
        "main.zig": "const std = @import(\"std\");\n",
    })
    if err := runner.Run(defaultOpts("MIT", root)); err != nil {
        t.Fatal(err)
    }
    content := readFile(t, filepath.Join(root, "main.zig"))
    if !strings.Contains(content, "// SPDX-License-Identifier: MIT") {
        t.Errorf("missing line comment prefix:\n%s", content)
    }
}
```

For block comment languages, assert both opening and closing delimiters:

```go
if !strings.Contains(content, "/*") || !strings.Contains(content, "*/") {
    t.Errorf("missing block comment markers:\n%s", content)
}
```

## Step 3 — Update the README

Add the extension to the **Supported languages** table in [`README.md`](https://github.com/GregoireF/addlicense/blob/main/README.md).

## Step 4 — Update CHANGELOG

Add an entry under `[Unreleased]` in [`CHANGELOG.md`](https://github.com/GregoireF/addlicense/blob/main/CHANGELOG.md):

```markdown
- Extended language support: `.zig` (`//`)
```

## Checklist

- [ ] Extension added to `langs` map in `internal/header/header.go`
- [ ] `gofmt -d internal/header/header.go` produces no diff
- [ ] Integration test added in `tests/integration/runner_test.go`
- [ ] `go test -race ./...` passes
- [ ] README supported-languages table updated
- [ ] CHANGELOG `[Unreleased]` updated
