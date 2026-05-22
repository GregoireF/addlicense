// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package integration_test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
)

const (
	testAuthor       = "Grégoire"
	testLicense      = "MIT"
	testApache       = "Apache-2.0"
	testMainGo       = "main.go"
	testBody         = "package main\n"
	testLicensedBody = "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\n\npackage main\n"
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
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
}

// ── Core behaviour ────────────────────────────────────────────────────────────

func TestRun_AddsMITHeader(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo:   testBody,
		"app.ts":     "export default {}\n",
		"script.sh":  "#!/bin/bash\necho hello\n",
		"config.yml": "key: value\n",
	})

	if err := runner.Run(defaultOpts(testLicense, root)); err != nil {
		t.Fatalf("Run: %v", err)
	}

	cases := map[string][]string{
		testMainGo:   {"// SPDX-License-Identifier: MIT", "// Copyright 2026 Grégoire"},
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
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := defaultOpts(testLicense, root)

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, filepath.Join(root, testMainGo))

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	second := readFile(t, filepath.Join(root, testMainGo))

	if first != second {
		t.Errorf("second run modified the file:\nbefore:\n%s\nafter:\n%s", first, second)
	}
}

func TestRun_PreservesShebang(t *testing.T) {
	root := makeProject(t, map[string]string{
		"deploy.sh": "#!/usr/bin/env bash\nset -euo pipefail\n",
	})

	if err := runner.Run(defaultOpts(testApache, root)); err != nil {
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
		testMainGo: "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\npackage main\n",
	})
	opts := config.Options{
		License:   testLicense,
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
	}
	if err := runner.Run(opts); err != nil {
		t.Errorf("check mode on licensed file should pass, got: %v", err)
	}
}

func TestRun_CheckMode_FailsOnMissing(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License:   testLicense,
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
		License:   testLicense,
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
		testMainGo:            testBody,
		"vendor/lib.go":       "package lib\n",
		"node_modules/pkg.js": "module.exports = {}\n",
	})

	if err := runner.Run(defaultOpts(testLicense, root)); err != nil {
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
		testMainGo:  testBody,
		"data.json": "{}\n",
		"image.png": "\x89PNG\r\n",
	})

	if err := runner.Run(defaultOpts(testLicense, root)); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"data.json", "image.png"} {
		content := readFile(t, filepath.Join(root, rel))
		if strings.Contains(content, "SPDX-License-Identifier") {
			t.Errorf("%s should not have been touched", rel)
		}
	}
}

// ── Template ──────────────────────────────────────────────────────────────────

func TestRun_CustomTemplate(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	tmplPath := filepath.Join(t.TempDir(), "header.txt")
	if err := os.WriteFile(tmplPath, []byte("Copyright {{.Year}} {{.Author}} — All rights reserved."), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := config.Options{
		License:  testLicense,
		Author:   "ACME Corp",
		Year:     2026,
		Template: tmplPath,
		Ignore:   config.DefaultIgnore,
		Paths:    []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "Copyright 2026 ACME Corp — All rights reserved.") {
		t.Errorf("custom template not applied:\n%s", content)
	}
}

func TestRun_TemplateFileNotFound(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License:  testLicense,
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
		testMainGo:           testBody,
		".addlicenserc.yaml": "license: Apache-2.0\nauthor: ConfigAuthor\nyear: 2025\n",
	})
	opts := config.Options{Ignore: config.DefaultIgnore, Paths: []string{root}}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
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
	if err := runner.Run(defaultOpts(testLicense, root)); err != nil {
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
	if err := runner.Run(defaultOpts(testLicense, root)); err != nil {
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
	if err := runner.Run(defaultOpts(testLicense, root)); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, "schema.sql"))
	if !strings.Contains(content, "-- SPDX-License-Identifier: MIT") {
		t.Errorf("missing SQL comment prefix:\n%s", content)
	}
}

// ── EU licences ───────────────────────────────────────────────────────────────

func TestRun_EULicense_EUPL(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	if err := runner.Run(defaultOpts("EUPL-1.2", root)); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "SPDX-License-Identifier: EUPL-1.2") {
		t.Errorf("missing EUPL-1.2 identifier:\n%s", content)
	}
}

func TestRun_EULicense_AGPL(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	if err := runner.Run(defaultOpts("AGPL-3.0", root)); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "SPDX-License-Identifier: AGPL-3.0-only") {
		t.Errorf("missing AGPL-3.0-only identifier:\n%s", content)
	}
}

// ── REUSE compliance ──────────────────────────────────────────────────────────

func TestRun_Reuse_EmitsSPDXFileCopyrightText(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo:   testBody,
		"script.sh":  "#!/bin/bash\necho hello\n",
		"config.yml": "key: value\n",
	})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Reuse:   true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}
	cases := map[string][]string{
		testMainGo:   {"// SPDX-FileCopyrightText: 2026 Grégoire", "// SPDX-License-Identifier: MIT"},
		"script.sh":  {"# SPDX-FileCopyrightText: 2026 Grégoire", "# SPDX-License-Identifier: MIT"},
		"config.yml": {"# SPDX-FileCopyrightText: 2026 Grégoire", "# SPDX-License-Identifier: MIT"},
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

func TestRun_Reuse_NoAuthor(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testApache,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Reuse:   true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "SPDX-FileCopyrightText: 2026") {
		t.Errorf("missing SPDX-FileCopyrightText: 2026:\n%s", content)
	}
	// plain "Copyright 2026" (the standalone notice) must not appear
	if strings.Contains(content, "Copyright 2026") {
		t.Errorf("should not contain plain 'Copyright 2026' in reuse mode:\n%s", content)
	}
}

func TestRun_Reuse_Idempotent(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Reuse:   true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, filepath.Join(root, testMainGo))

	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	second := readFile(t, filepath.Join(root, testMainGo))

	if first != second {
		t.Errorf("second reuse run modified the file:\nbefore:\n%s\nafter:\n%s", first, second)
	}
}

// ── Remove mode ───────────────────────────────────────────────────────────────

func TestRun_Remove_LineComment(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: testLicensedBody,
	})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if strings.Contains(content, "SPDX") || strings.Contains(content, "Copyright") {
		t.Errorf("header should be removed:\n%s", content)
	}
	if !strings.Contains(content, "package main") {
		t.Errorf("body should be preserved:\n%s", content)
	}
}

func TestRun_Remove_BlockComment(t *testing.T) {
	root := makeProject(t, map[string]string{})
	htmlPath := filepath.Join(root, "index.html")
	if err := os.WriteFile(htmlPath, []byte("<!--\nSPDX-License-Identifier: MIT\nCopyright 2026 Grégoire\n-->\n<html></html>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, htmlPath)
	if strings.Contains(content, "SPDX") {
		t.Errorf("header should be removed:\n%s", content)
	}
	if !strings.Contains(content, "<html>") {
		t.Errorf("body should be preserved:\n%s", content)
	}
}

func TestRun_Remove_NoHeader_NoChange(t *testing.T) {
	original := "package main\n\nfunc main() {}\n"
	root := makeProject(t, map[string]string{testMainGo: original})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if content != original {
		t.Errorf("file without header should be unchanged:\ngot:\n%s", content)
	}
}

func TestRun_Remove_NonLicenseComment_NoChange(t *testing.T) {
	original := "// Package main does something interesting.\n// It is purely documentary.\npackage main\n"
	root := makeProject(t, map[string]string{testMainGo: original})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if content != original {
		t.Errorf("non-license comment block should be unchanged:\ngot:\n%s", content)
	}
}

// ── Update mode ───────────────────────────────────────────────────────────────

func TestRun_Update_ReplaceHeader(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: testLicensedBody,
	})
	opts := config.Options{
		License: testApache,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Update:  true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if strings.Contains(content, "SPDX-License-Identifier: MIT") {
		t.Errorf("old MIT header should be replaced:\n%s", content)
	}
	if !strings.Contains(content, "SPDX-License-Identifier: Apache-2.0") {
		t.Errorf("new Apache-2.0 header should be present:\n%s", content)
	}
	if !strings.Contains(content, "package main") {
		t.Errorf("body should be preserved:\n%s", content)
	}
	if strings.Count(content, "SPDX-License-Identifier") > 1 {
		t.Errorf("duplicate SPDX line detected:\n%s", content)
	}
}

func TestRun_Update_FileWithoutHeader(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Update:  true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "SPDX-License-Identifier: MIT") {
		t.Errorf("update on unlicensed file should add header:\n%s", content)
	}
	if !strings.Contains(content, "package main") {
		t.Errorf("body should be preserved:\n%s", content)
	}
}

func TestRun_DryRun_Update_WithHeader(t *testing.T) {
	original := testLicensedBody
	root := makeProject(t, map[string]string{testMainGo: original})
	opts := config.Options{
		License: testApache,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Update:  true,
		DryRun:  true,
	}

	out := captureStdout(t, func() {
		if err := runner.Run(opts); err != nil {
			t.Error(err)
		}
	})

	content := readFile(t, filepath.Join(root, testMainGo))
	if content != original {
		t.Errorf("dry-run --update should not modify file:\n%s", content)
	}
	if !strings.Contains(out, "would-update") {
		t.Errorf("expected would-update in output, got:\n%s", out)
	}
}

func TestRun_DryRun_Update_WithoutHeader(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Update:  true,
		DryRun:  true,
	}

	out := captureStdout(t, func() {
		if err := runner.Run(opts); err != nil {
			t.Error(err)
		}
	})

	content := readFile(t, filepath.Join(root, testMainGo))
	if content != testBody {
		t.Errorf("dry-run --update should not modify file:\n%s", content)
	}
	if !strings.Contains(out, "would-add") {
		t.Errorf("expected would-add in output, got:\n%s", out)
	}
}

// ── Dry-run mode ──────────────────────────────────────────────────────────────

func TestRun_DryRun_NoWrite(t *testing.T) {
	original := testBody
	root := makeProject(t, map[string]string{testMainGo: original})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		DryRun:  true,
		Quiet:   true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if content != original {
		t.Errorf("dry-run should not modify files:\ngot:\n%s", content)
	}
}

func TestRun_DryRun_Remove(t *testing.T) {
	original := testLicensedBody
	root := makeProject(t, map[string]string{testMainGo: original})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
		DryRun: true,
		Quiet:  true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if content != original {
		t.Errorf("dry-run --remove should not modify files:\ngot:\n%s", content)
	}
}

func TestRun_DryRun_Remove_NoHeader(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
		DryRun: true,
		Quiet:  true,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if content != testBody {
		t.Errorf("dry-run --remove on file without header should be a no-op:\n%s", content)
	}
}

func TestRun_ErrorOnReadOnlyFile(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	path := filepath.Join(root, testMainGo)
	if err := os.Chmod(path, 0o444); err != nil {
		t.Skip("cannot set read-only attribute: " + err.Error())
	}
	defer os.Chmod(path, 0o644) //nolint:errcheck

	// Verify read-only is actually enforced (administrators can bypass it).
	if err := os.WriteFile(path, []byte("test"), 0o644); err == nil {
		t.Skip("read-only not enforced (running as administrator?)")
	}

	opts := defaultOpts(testLicense, root)
	if err := runner.Run(opts); err == nil {
		t.Error("expected error when writing to a read-only file")
	}
}

// ── Mutually exclusive flags ───────────────────────────────────────────────────

func TestRun_CheckAndRemoveConflict(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
		Remove:    true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("--check --remove should return an error")
	}
}

func TestRun_CheckAndUpdateConflict(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License:   testLicense,
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		CheckOnly: true,
		Update:    true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("--check --update should return an error")
	}
}

func TestRun_VerboseAndQuietConflict(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Verbose: true,
		Quiet:   true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("--verbose --quiet should return an error")
	}
}

func TestRun_RemoveAndUpdateConflict(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
		Update: true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("--remove --update should return an error")
	}
}

// ── Output format ─────────────────────────────────────────────────────────────

func TestRun_Format_JSON(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Format:  "json",
	}

	out := captureStdout(t, func() {
		if err := runner.Run(opts); err != nil {
			t.Error(err)
		}
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("expected at least one JSON line on stdout")
	}
	for _, line := range lines {
		if line == "" {
			continue
		}
		var rec struct {
			File   string `json:"file"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Errorf("invalid JSON line %q: %v", line, err)
			continue
		}
		if rec.File == "" {
			t.Errorf("JSON record missing 'file' field: %s", line)
		}
		if rec.Status == "" {
			t.Errorf("JSON record missing 'status' field: %s", line)
		}
	}
}

func TestRun_Format_Invalid(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Format:  "xml",
	}
	if err := runner.Run(opts); err == nil {
		t.Error("--format xml should return an error")
	}
}

func TestRun_Quiet_SuppressesStdout(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Quiet:   true,
	}

	out := captureStdout(t, func() {
		if err := runner.Run(opts); err != nil {
			t.Error(err)
		}
	})

	if strings.TrimSpace(out) != "" {
		t.Errorf("--quiet should suppress stdout, got:\n%s", out)
	}
}

func TestRun_Verbose_ShowsSkippedFiles(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\npackage main\n",
	})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Verbose: true,
	}

	out := captureStdout(t, func() {
		if err := runner.Run(opts); err != nil {
			t.Error(err)
		}
	})

	if !strings.Contains(out, "skipped") && !strings.Contains(out, "ok") {
		t.Errorf("--verbose should report already-licensed files, got:\n%s", out)
	}
}

// ── Parallel processing ───────────────────────────────────────────────────────

func TestRun_Parallel_MultipleFiles(t *testing.T) {
	files := make(map[string]string, 20)
	for i := 0; i < 20; i++ {
		files[strings.Repeat("a", i+1)+".go"] = testBody
	}
	root := makeProject(t, files)
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Workers: 4,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	for rel := range files {
		content := readFile(t, filepath.Join(root, rel))
		if !strings.Contains(content, "SPDX-License-Identifier: MIT") {
			t.Errorf("%s: missing header", rel)
		}
	}
}

func TestRun_Workers_SingleWorker(t *testing.T) {
	root := makeProject(t, map[string]string{
		"a.go": "package a\n",
		"b.go": "package b\n",
	})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Workers: 1,
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"a.go", "b.go"} {
		if !strings.Contains(readFile(t, filepath.Join(root, rel)), "SPDX") {
			t.Errorf("%s: missing header", rel)
		}
	}
}

// ── Config reuse from file ────────────────────────────────────────────────────

func TestRun_Config_ReuseFromFile(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo:           testBody,
		".addlicenserc.yaml": "license: MIT\nauthor: Grégoire\nyear: 2026\nreuse: true\n",
	})
	opts := config.Options{Ignore: config.DefaultIgnore, Paths: []string{root}}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "SPDX-FileCopyrightText:") {
		t.Errorf("reuse: true in config should emit SPDX-FileCopyrightText:\n%s", content)
	}
	if strings.Contains(content, "Copyright 2026") {
		t.Errorf("should not contain plain 'Copyright 2026' in reuse mode:\n%s", content)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// captureStdout runs fn while redirecting os.Stdout to a pipe, then returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// ── Error paths ───────────────────────────────────────────────────────────────

func TestRun_InvalidRoot_Error(t *testing.T) {
	opts := config.Options{
		License: testLicense,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{filepath.Join(t.TempDir(), "nonexistent")},
	}
	if err := runner.Run(opts); err == nil {
		t.Error("expected error for nonexistent root path")
	}
}

func TestRun_ErrorOnReadOnlyFile_Update(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: testLicensedBody,
	})
	path := filepath.Join(root, testMainGo)
	if err := os.Chmod(path, 0o444); err != nil {
		t.Skip("cannot set read-only attribute: " + err.Error())
	}
	defer os.Chmod(path, 0o644) //nolint:errcheck

	if err := os.WriteFile(path, []byte("test"), 0o644); err == nil {
		t.Skip("read-only not enforced (running as administrator?)")
	}

	opts := config.Options{
		License: testApache,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Update:  true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("expected error when updating a read-only file")
	}
}

func TestRun_ErrorOnReadOnlyFile_Remove(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: testLicensedBody,
	})
	path := filepath.Join(root, testMainGo)
	if err := os.Chmod(path, 0o444); err != nil {
		t.Skip("cannot set read-only attribute: " + err.Error())
	}
	defer os.Chmod(path, 0o644) //nolint:errcheck

	if err := os.WriteFile(path, []byte("test"), 0o644); err == nil {
		t.Skip("read-only not enforced (running as administrator?)")
	}

	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Remove: true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("expected error when removing header from read-only file")
	}
}

func TestRun_Format_JSON_WithError(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	path := filepath.Join(root, testMainGo)
	if err := os.Chmod(path, 0o444); err != nil {
		t.Skip("cannot set read-only attribute: " + err.Error())
	}
	defer os.Chmod(path, 0o644) //nolint:errcheck

	if err := os.WriteFile(path, []byte("test"), 0o644); err == nil {
		t.Skip("read-only not enforced (running as administrator?)")
	}

	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Format:  "json",
	}

	out := captureStdout(t, func() {
		runner.Run(opts) //nolint:errcheck
	})

	// The JSON record for the error should include the "error" field.
	if !strings.Contains(out, `"error"`) {
		t.Errorf("JSON error record should include error field, got:\n%s", out)
	}
}

func TestRun_NoFilesToProcess(t *testing.T) {
	// A project with only extension-less files (Makefile, Dockerfile) produces an
	// empty scanner.Walk result, which exercises the n==0 guard in runParallel.
	root := t.TempDir()
	for _, name := range []string{"Makefile", "Dockerfile"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("all:"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	opts := defaultOpts(testLicense, root)
	if err := runner.Run(opts); err != nil {
		t.Fatalf("empty project should not error: %v", err)
	}
}

// ── Multiple roots ────────────────────────────────────────────────────────────

// ── --year-range ──────────────────────────────────────────────────────────────

func TestRun_YearRange_PreservesOriginalYear(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: "// SPDX-License-Identifier: MIT\n// Copyright 2023 " + testAuthor + "\n\n" + testBody,
	})
	opts := config.Options{
		License:   testLicense,
		Author:    testAuthor,
		Year:      2026,
		Update:    true,
		YearRange: true,
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "2023-2026") {
		t.Errorf("expected year range 2023-2026 in header:\n%s", content)
	}
}

func TestRun_YearRange_SameYear_NoRange(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: "// SPDX-License-Identifier: MIT\n// Copyright 2026 " + testAuthor + "\n\n" + testBody,
	})
	opts := config.Options{
		License:   testLicense,
		Author:    testAuthor,
		Year:      2026,
		Update:    true,
		YearRange: true,
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if strings.Contains(content, "2026-2026") {
		t.Errorf("should not emit range when years are equal:\n%s", content)
	}
	if !strings.Contains(content, "Copyright 2026") {
		t.Errorf("expected single year copyright:\n%s", content)
	}
}

func TestRun_YearRange_NoExistingHeader_NoRange(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: testBody,
	})
	opts := config.Options{
		License:   testLicense,
		Author:    testAuthor,
		Year:      2026,
		Update:    true,
		YearRange: true,
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if strings.Contains(content, "Copyright 2026-") {
		t.Errorf("should not emit year range when no original header:\n%s", content)
	}
	if !strings.Contains(content, "Copyright 2026") {
		t.Errorf("expected single year copyright:\n%s", content)
	}
}

func TestRun_YearRange_Reuse_PreservesOriginalYear(t *testing.T) {
	root := makeProject(t, map[string]string{
		testMainGo: "// SPDX-FileCopyrightText: 2022 " + testAuthor + "\n// SPDX-License-Identifier: MIT\n\n" + testBody,
	})
	opts := config.Options{
		License:   testLicense,
		Author:    testAuthor,
		Year:      2026,
		Update:    true,
		YearRange: true,
		Reuse:     true,
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "2022-2026") {
		t.Errorf("expected REUSE year range 2022-2026 in header:\n%s", content)
	}
	if !strings.Contains(content, "SPDX-FileCopyrightText:") {
		t.Errorf("expected REUSE prefix in header:\n%s", content)
	}
}

// ── --dep5 ────────────────────────────────────────────────────────────────────

func TestRun_Dep5_CreatesFile(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":     testBody,
		"logo.png":    "\x89PNG\r\n",
		"diagram.svg": "<svg></svg>",
	})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Dep5:    true,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}

	dep5Path := filepath.Join(root, ".reuse", "dep5")
	content := readFile(t, dep5Path)

	if !strings.Contains(content, "Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/") {
		t.Errorf("missing dep5 Format header:\n%s", content)
	}
	if !strings.Contains(content, "logo.png") {
		t.Errorf("logo.png missing from dep5:\n%s", content)
	}
	if !strings.Contains(content, "diagram.svg") {
		t.Errorf("diagram.svg missing from dep5:\n%s", content)
	}
	if strings.Contains(content, "main.go") {
		t.Errorf("main.go (handled by inline header) must not appear in dep5:\n%s", content)
	}
}

func TestRun_Dep5_NoUnhandledFiles(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go": testBody,
		"app.ts":  "const x = 1;\n",
	})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Dep5:    true,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}

	dep5Path := filepath.Join(root, ".reuse", "dep5")
	content := readFile(t, dep5Path)
	if strings.Contains(content, "Files:") {
		t.Errorf("Files paragraph should be absent when no unhandled files:\n%s", content)
	}
}

func TestRun_Dep5_DryRun(t *testing.T) {
	root := makeProject(t, map[string]string{
		"main.go":  testBody,
		"logo.png": "\x89PNG\r\n",
	})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Dep5:    true,
		DryRun:  true,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}

	dep5Path := filepath.Join(root, ".reuse", "dep5")
	if _, err := os.Stat(dep5Path); err == nil {
		t.Error("--dry-run must not create .reuse/dep5")
	}
}

func TestRun_MultipleRoots(t *testing.T) {
	rootA := makeProject(t, map[string]string{"a.go": "package a\n"})
	rootB := makeProject(t, map[string]string{"b.go": "package b\n"})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
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

// ── Multi-author ──────────────────────────────────────────────────────────────

func TestRun_MultiAuthor_TwoAuthors(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  "Alice, Bob",
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "Copyright 2026 Alice") {
		t.Errorf("expected 'Copyright 2026 Alice' in header:\n%s", content)
	}
	if !strings.Contains(content, "Copyright 2026 Bob") {
		t.Errorf("expected 'Copyright 2026 Bob' in header:\n%s", content)
	}
	if !strings.Contains(content, "SPDX-License-Identifier: MIT") {
		t.Errorf("expected SPDX identifier in header:\n%s", content)
	}
}

func TestRun_MultiAuthor_Reuse(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  "Alice,Bob",
		Year:    2026,
		Reuse:   true,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
	}
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if !strings.Contains(content, "SPDX-FileCopyrightText: 2026 Alice") {
		t.Errorf("expected SPDX-FileCopyrightText for Alice:\n%s", content)
	}
	if !strings.Contains(content, "SPDX-FileCopyrightText: 2026 Bob") {
		t.Errorf("expected SPDX-FileCopyrightText for Bob:\n%s", content)
	}
}

func TestRun_MultiAuthor_SingleAuthor_Unchanged(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := defaultOpts(testLicense, root)
	if err := runner.Run(opts); err != nil {
		t.Fatal(err)
	}
	content := readFile(t, filepath.Join(root, testMainGo))
	if strings.Contains(content, "\nCopyright") {
		t.Errorf("single author should produce exactly one copyright line:\n%s", content)
	}
}

// ── --diff ────────────────────────────────────────────────────────────────────

func TestRun_Diff_NoWrite_ReturnsChangedCount(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Diff:    true,
	}
	err := runner.Run(opts)
	if err == nil {
		t.Fatal("expected non-nil error (file would be modified), got nil")
	}
	if !strings.Contains(err.Error(), "would be modified") {
		t.Errorf("expected 'would be modified' in error, got: %v", err)
	}
	// File must not have been touched.
	content := readFile(t, filepath.Join(root, testMainGo))
	if content != testBody {
		t.Errorf("--diff must not write to file:\n%s", content)
	}
}

func TestRun_Diff_AlreadyLicensed_NoChange(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testLicensedBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Diff:    true,
	}
	if err := runner.Run(opts); err != nil {
		t.Errorf("all files already licensed, expected nil error, got: %v", err)
	}
}

func TestRun_Diff_EmitsJSONWithHeader(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Diff:    true,
	}

	out := captureStdout(t, func() {
		runner.Run(opts) //nolint:errcheck
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("expected at least one JSON line on stdout")
	}
	var rec struct {
		File   string `json:"file"`
		Status string `json:"status"`
		Header string `json:"header"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &rec); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, lines[0])
	}
	if rec.Status != "diff-add" {
		t.Errorf("expected status=diff-add, got %q", rec.Status)
	}
	if !strings.Contains(rec.Header, "SPDX-License-Identifier") {
		t.Errorf("header field should contain rendered header:\n%s", rec.Header)
	}
}

func TestRun_Diff_Update_EmitsDiffUpdate(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testLicensedBody})
	opts := config.Options{
		License: testApache,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Diff:    true,
		Update:  true,
	}

	out := captureStdout(t, func() {
		runner.Run(opts) //nolint:errcheck
	})

	if !strings.Contains(out, "diff-update") {
		t.Errorf("expected diff-update in output:\n%s", out)
	}
	// File must be unchanged.
	if content := readFile(t, filepath.Join(root, testMainGo)); content != testLicensedBody {
		t.Errorf("--diff --update must not modify file:\n%s", content)
	}
}

func TestRun_Diff_CheckConflict(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License:   testLicense,
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		Diff:      true,
		CheckOnly: true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("--diff --check should return an error")
	}
}

func TestRun_Diff_Remove_WithHeader(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testLicensedBody})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Diff:   true,
		Remove: true,
	}
	out := captureStdout(t, func() {
		if err := runner.Run(opts); err == nil {
			t.Error("expected non-nil error (file would be modified)")
		}
	})
	if !strings.Contains(out, "diff-remove") {
		t.Errorf("expected diff-remove in JSON output:\n%s", out)
	}
	// File must be unchanged.
	if content := readFile(t, filepath.Join(root, testMainGo)); content != testLicensedBody {
		t.Errorf("--diff --remove must not modify file:\n%s", content)
	}
}

func TestRun_Diff_Remove_WithoutHeader_Skipped(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		Ignore: config.DefaultIgnore,
		Paths:  []string{root},
		Diff:   true,
		Remove: true,
	}
	if err := runner.Run(opts); err != nil {
		t.Errorf("no-header file in diff-remove should be skipped (exit 0), got: %v", err)
	}
}

func TestRun_Diff_Update_NoHeader_EmitsDiffAdd(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	opts := config.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Diff:    true,
		Update:  true,
	}
	out := captureStdout(t, func() {
		runner.Run(opts) //nolint:errcheck
	})
	if !strings.Contains(out, "diff-add") {
		t.Errorf("expected diff-add for unlicensed file in --diff --update:\n%s", out)
	}
}

func TestRun_Diff_BadTemplate_Error(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	tmplPath := filepath.Join(t.TempDir(), "bad.txt")
	if err := os.WriteFile(tmplPath, []byte("{{."), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := config.Options{
		License:  testLicense,
		Author:   testAuthor,
		Year:     2026,
		Template: tmplPath,
		Ignore:   config.DefaultIgnore,
		Paths:    []string{root},
		Diff:     true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("expected error for malformed template in --diff mode")
	}
}

func TestRun_Diff_BadTemplate_Update_Error(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testLicensedBody})
	tmplPath := filepath.Join(t.TempDir(), "bad.txt")
	if err := os.WriteFile(tmplPath, []byte("{{."), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := config.Options{
		License:  testLicense,
		Author:   testAuthor,
		Year:     2026,
		Template: tmplPath,
		Ignore:   config.DefaultIgnore,
		Paths:    []string{root},
		Diff:     true,
		Update:   true,
	}
	if err := runner.Run(opts); err == nil {
		t.Error("expected error for malformed template in --diff --update mode")
	}
}

// v0.9.0: new language extensions

func TestRun_NewLang_Lua(t *testing.T) {
	root := makeProject(t, map[string]string{"init.lua": "return {}\n"})
	if err := runner.Run(config.Options{
		License: testLicense, Author: testAuthor, Year: 2026,
		Ignore: config.DefaultIgnore, Paths: []string{root}, Quiet: true,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := readFile(t, filepath.Join(root, "init.lua"))
	if !strings.Contains(got, "-- Copyright") {
		t.Errorf("expected Lua line comment, got:\n%s", got)
	}
}

func TestRun_NewLang_Nix(t *testing.T) {
	root := makeProject(t, map[string]string{"default.nix": "{ pkgs }: {}\n"})
	if err := runner.Run(config.Options{
		License: testLicense, Author: testAuthor, Year: 2026,
		Ignore: config.DefaultIgnore, Paths: []string{root}, Quiet: true,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := readFile(t, filepath.Join(root, "default.nix"))
	if !strings.Contains(got, "# Copyright") {
		t.Errorf("expected Nix hash comment, got:\n%s", got)
	}
}

func TestRun_NewLang_Zig(t *testing.T) {
	root := makeProject(t, map[string]string{"main.zig": "const std = @import(\"std\");\n"})
	if err := runner.Run(config.Options{
		License: testLicense, Author: testAuthor, Year: 2026,
		Ignore: config.DefaultIgnore, Paths: []string{root}, Quiet: true,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := readFile(t, filepath.Join(root, "main.zig"))
	if !strings.Contains(got, "// Copyright") {
		t.Errorf("expected Zig line comment, got:\n%s", got)
	}
}

func TestRun_NewLang_Dockerfile(t *testing.T) {
	root := makeProject(t, map[string]string{"app.dockerfile": "FROM scratch\n"})
	if err := runner.Run(config.Options{
		License: testLicense, Author: testAuthor, Year: 2026,
		Ignore: config.DefaultIgnore, Paths: []string{root}, Quiet: true,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := readFile(t, filepath.Join(root, "app.dockerfile"))
	if !strings.Contains(got, "# Copyright") {
		t.Errorf("expected Dockerfile hash comment, got:\n%s", got)
	}
}

func TestRun_NewLang_Markdown(t *testing.T) {
	root := makeProject(t, map[string]string{"README.md": "# Hello\n"})
	if err := runner.Run(config.Options{
		License: testLicense, Author: testAuthor, Year: 2026,
		Ignore: config.DefaultIgnore, Paths: []string{root}, Quiet: true,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := readFile(t, filepath.Join(root, "README.md"))
	if !strings.Contains(got, "<!--") || !strings.Contains(got, "-->") {
		t.Errorf("expected HTML comment in Markdown, got:\n%s", got)
	}
}

// v1.0.0: --sbom mode

func TestRun_Sbom_WritesFile(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testLicensedBody})
	out := filepath.Join(t.TempDir(), "sbom.spdx")
	if err := runner.Run(config.Options{
		License: testLicense,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Sbom:    out,
	}); err != nil {
		t.Fatalf("Run --sbom: %v", err)
	}
	content := readFile(t, out)
	if !strings.Contains(content, "SPDXVersion: SPDX-2.3") {
		t.Errorf("expected SPDX header:\n%s", content)
	}
	if !strings.Contains(content, "LicenseConcluded: MIT") {
		t.Errorf("expected LicenseConcluded MIT from file header:\n%s", content)
	}
}

func TestRun_Sbom_UnlicensedFile_FallbackLicense(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testBody})
	out := filepath.Join(t.TempDir(), "sbom.spdx")
	if err := runner.Run(config.Options{
		License: testLicense,
		Ignore:  config.DefaultIgnore,
		Paths:   []string{root},
		Sbom:    out,
	}); err != nil {
		t.Fatalf("Run --sbom: %v", err)
	}
	content := readFile(t, out)
	// No SPDX-License-Identifier in file → fallback to opts.License.
	if !strings.Contains(content, "LicenseConcluded: MIT") {
		t.Errorf("expected fallback MIT in LicenseConcluded:\n%s", content)
	}
	// LicenseInfoInFile must be NOASSERTION (nothing in the file itself).
	if !strings.Contains(content, "LicenseInfoInFile: NOASSERTION") {
		t.Errorf("expected LicenseInfoInFile NOASSERTION:\n%s", content)
	}
}

func TestRun_Sbom_MutualExclusion_Check(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testLicensedBody})
	out := filepath.Join(t.TempDir(), "sbom.spdx")
	err := runner.Run(config.Options{
		Ignore:    config.DefaultIgnore,
		Paths:     []string{root},
		Sbom:      out,
		CheckOnly: true,
	})
	if err == nil {
		t.Error("expected error for --sbom --check, got nil")
	}
}

func TestRun_Sbom_Stdout(t *testing.T) {
	root := makeProject(t, map[string]string{testMainGo: testLicensedBody})
	out := captureStdout(t, func() {
		if err := runner.Run(config.Options{
			License: testLicense,
			Ignore:  config.DefaultIgnore,
			Paths:   []string{root},
			Sbom:    "-",
		}); err != nil {
			t.Errorf("Run --sbom -: %v", err)
		}
	})
	if !strings.Contains(out, "SPDXVersion: SPDX-2.3") {
		t.Errorf("expected SPDX output on stdout:\n%s", out)
	}
}
