// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
)

// makeProject creates a temporary directory tree from a map of relative path → content.
func makeProject(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func defaultOpts(license, root string) config.Options {
	return config.Options{
		License: license,
		Author:  "Grégoire",
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
}

// ── Core behaviour ────────────────────────────────────────────────────────────

func TestRun_AddsMITHeader(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":    "package main\n",
		"app.ts":     "export default {}\n",
		"script.sh":  "#!/bin/bash\necho hello\n",
		"config.yml": "key: value\n",
	})

	if err := runner.Run(defaultOpts("MIT", root)); err != nil {
		t.Fatalf("Run: %v", err)
	}

	cases := map[string][]string{
		"main.go":    {"// SPDX-License-Identifier: MIT", "// Copyright 2026 Grégoire"},
		"app.ts":     {"// SPDX-License-Identifier: MIT"},
		"script.sh":  {"# SPDX-License-Identifier: MIT"},
		"config.yml": {"# SPDX-License-Identifier: MIT"},
	}
	for rel, want := range cases {
		content := readFile(t, filepath.Join(root, rel))
		for _, line := range want {
			if !strings.Contains(content, line) {
				t.Errorf("%s: missing %q\ngot:\n%s", rel, line, content)
			}
		}
	}
}

func TestRun_IdempotentOnSecondCall(t *testing.T) {
	root := makeProject(t, map[string]string{"main.go": "package main\n"})
	opts := defaultOpts("MIT", root)

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, filepath.Join(root, "main.go"))

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	second := readFile(t, filepath.Join(root, "main.go"))

	if first != second {
		t.Errorf("second run modified the file:\nbefore:\n%s\nafter:\n%s", first, second)
	}
}

func TestRun_PreservesShebang(t *testing.T) {
	root := makeProject(t, map[string]string{
		"deploy.sh": "#!/usr/bin/env bash\nset -euo pipefail\n",
	})

	if err := runner.Run(defaultOpts("Apache-2.0", root)); err != nil {
		t.Fatal(err)
	}

	content := readFile(t, filepath.Join(root, "deploy.sh"))
	if !strings.HasPrefix(content, "#!/usr/bin/env bash") {
		t.Errorf("shebang not on first line:\n%s", content)
	}
	if !strings.Contains(content, "SPDX-License-Identifier: Apache-2.0") {
		t.Errorf("missing SPDX header:\n%s", content)
	}
}

// ── Check mode ────────────────────────────────────────────────────────────────

func TestRun_CheckMode_DetectsOK(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go": "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\npackage main\n",
	})
	opts := config.Options{
		License:   "MIT",
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
	}
	if err := runner.Run(opts); err != nil {
		t.Errorf("check mode on licensed file should pass, got: %v", err)
	}
}

func TestRun_CheckMode_FailsOnMissing(t *testing.T) {
	root := makeProject(t, map[string]string{"main.go": "package main\n"})
	opts := config.Options{
		License:   "MIT",
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("check mode should fail when headers are missing")
	}
}

func TestRun_CheckMode_ListsAllMissing(t *testing.T) {
	root := makeProject(t, map[string]string{
		"a.go": "package a\n",
		"b.go": "package b\n",
		"c.go": "package c\n",
	})
	opts := config.Options{
		License:   "MIT",
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
	}
	err := runner.Run(opts)
	if err == nil {
		t.Fatal("expected error listing missing files")
	}
	if !strings.Contains(err.Error(), "3 file(s)") {
		t.Errorf("expected 3 missing files in error, got: %v", err)
	}
}

// ── Ignore patterns ──────────────────────────────────────────────────────────

func TestRun_IgnoresVendorAndNodeModules(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":             "package main\n",
		"vendor/lib.go":       "package lib\n",
		"node_modules/pkg.js": "module.exports = {}\n",
	})

	if err := runner.Run(defaultOpts("MIT", root)); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"vendor/lib.go", "node_modules/pkg.js"} {
		content := readFile(t, filepath.Join(root, rel))
		if strings.Contains(content, "SPDX-License-Identifier") {
			t.Errorf("%s should not have been modified", rel)
		}
	}
}

func TestRun_SkipsUnsupportedExtensions(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":   "package main\n",
		"README.md": "# hello\n",
		"data.json": "{}\n",
	})

	if err := runner.Run(defaultOpts("MIT", root)); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"README.md", "data.json"} {
		content := readFile(t, filepath.Join(root, rel))
		if strings.Contains(content, "SPDX-License-Identifier") {
			t.Errorf("%s should not have been touched", rel)
		}
	}
}

// ── Template ──────────────────────────────────────────────────────────────────

func TestRun_CustomTemplate(t *testing.T) {
	root := makeProject(t, map[string]string{"main.go": "package main\n"})
	tmplPath := filepath.Join(t.TempDir(), "header.txt")
	if err := os.WriteFile(tmplPath, []byte("Copyright {{.Year}} {{.Author}} — All rights reserved."), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := config.Options{
		License:  "MIT",
		Author:   "ACME Corp",
		Year:     2026,
		Template: tmplPath,
		Ignore:   config.DefaultIgnore,
		Paths:    []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, "main.go"))
	if !strings.Contains(content, "Copyright 2026 ACME Corp — All rights reserved.") {
		t.Errorf("custom template not applied:\n%s", content)
	}
}

func TestRun_TemplateFileNotFound(t *testing.T) {
	root := makeProject(t, map[string]string{"main.go": "package main\n"})
	opts := config.Options{
		License:  "MIT",
		Template: "/nonexistent/header.txt",
		Ignore:   config.DefaultIgnore,
		Paths:    []string{root},
	}
	if err := runner.Run(opts); err == nil {
		t.Error("expected error for missing template file")
	}
}

// ── Config file ───────────────────────────────────────────────────────────────

func TestRun_ConfigFileOverride(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":            "package main\n",
		".addlicenserc.yaml": "license: Apache-2.0\nauthor: ConfigAuthor\nyear: 2025\n",
	})
	opts := config.Options{Ignore: config.DefaultIgnore, Paths: []string{root}}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, "main.go"))
	if !strings.Contains(content, "SPDX-License-Identifier: Apache-2.0") {
		t.Errorf("expected Apache-2.0 from config:\n%s", content)
	}
	if !strings.Contains(content, "ConfigAuthor") {
		t.Errorf("expected ConfigAuthor from config:\n%s", content)
	}
}

// ── Language comment styles ───────────────────────────────────────────────────

func TestRun_BlockCommentLanguages(t *testing.T) {
	root := makeProject(t, map[string]string{
		"Main.java": "public class Main {}\n",
		"main.c":    "#include <stdio.h>\n",
		"style.css": "body { margin: 0; }\n",
	})
	if err := runner.Run(defaultOpts("MIT", root)); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"Main.java", "main.c", "style.css"} {
		content := readFile(t, filepath.Join(root, rel))
		if !strings.Contains(content, "/*") || !strings.Contains(content, "*/") {
			t.Errorf("%s: missing block comment markers:\n%s", rel, content)
		}
	}
}

func TestRun_HTMLComment(t *testing.T) {
	root := makeProject(t, map[string]string{
		"index.html": "<html></html>\n",
		"App.vue":    "<template></template>\n",
	})
	if err := runner.Run(defaultOpts("MIT", root)); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"index.html", "App.vue"} {
		content := readFile(t, filepath.Join(root, rel))
		if !strings.Contains(content, "<!--") || !strings.Contains(content, "-->") {
			t.Errorf("%s: missing HTML comment markers:\n%s", rel, content)
		}
	}
}

func TestRun_SQLComment(t *testing.T) {
	root := makeProject(t, map[string]string{
		"schema.sql": "CREATE TABLE foo (id INT);\n",
	})
	if err := runner.Run(defaultOpts("MIT", root)); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, "schema.sql"))
	if !strings.Contains(content, "-- SPDX-License-Identifier: MIT") {
		t.Errorf("missing SQL comment prefix:\n%s", content)
	}
}

// ── EU licences ───────────────────────────────────────────────────────────────

func TestRun_EULicense_EUPL(t *testing.T) {
	root := makeProject(t, map[string]string{"main.go": "package main\n"})
	if err := runner.Run(defaultOpts("EUPL-1.2", root)); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, "main.go"))
	if !strings.Contains(content, "SPDX-License-Identifier: EUPL-1.2") {
		t.Errorf("missing EUPL-1.2 identifier:\n%s", content)
	}
}

func TestRun_EULicense_AGPL(t *testing.T) {
	root := makeProject(t, map[string]string{"main.go": "package main\n"})
	if err := runner.Run(defaultOpts("AGPL-3.0", root)); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, "main.go"))
	if !strings.Contains(content, "SPDX-License-Identifier: AGPL-3.0-only") {
		t.Errorf("missing AGPL-3.0-only identifier:\n%s", content)
	}
}

// ── Multiple roots ────────────────────────────────────────────────────────────

func TestRun_MultipleRoots(t *testing.T) {
	rootA := makeProject(t, map[string]string{"a.go": "package a\n"})
	rootB := makeProject(t, map[string]string{"b.go": "package b\n"})
	opts := config.Options{
		License: "MIT",
		Author:  "Grégoire",
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{rootA, rootB},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		filepath.Join(rootA, "a.go"),
		filepath.Join(rootB, "b.go"),
	} {
		content := readFile(t, p)
		if !strings.Contains(content, "SPDX-License-Identifier: MIT") {
			t.Errorf("%s: missing header", p)
		}
	}
}
