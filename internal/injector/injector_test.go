// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package injector_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/GregoireF/addlicense/internal/injector"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.go")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestHasHeader_NoHeader(t *testing.T) {
	path := writeTemp(t, "package main\n\nfunc main() {}\n")
	has, err := injector.HasHeader(path)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("expected no header, got true")
	}
}

func TestHasHeader_WithSPDX(t *testing.T) {
	content := "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\npackage main\n"
	path := writeTemp(t, content)
	has, err := injector.HasHeader(path)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("expected header detected, got false")
	}
}

func TestHasHeader_WithCopyright(t *testing.T) {
	content := "// Copyright 2026 Grégoire Favreau\npackage main\n"
	path := writeTemp(t, content)
	has, err := injector.HasHeader(path)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("expected header detected, got false")
	}
}

func TestInject_AddsHeader(t *testing.T) {
	path := writeTemp(t, "package main\n\nfunc main() {}\n")
	header := "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\n"

	if err := injector.Inject(path, header); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(got)
	if content[:len(header)] != header {
		t.Errorf("header not at start of file:\n%s", content)
	}
}

func TestInject_PreservesShebang(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.sh")
	original := "#!/usr/bin/env bash\necho hello\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	header := "# SPDX-License-Identifier: MIT\n"
	if err := injector.Inject(path, header); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(got)
	if content[:len("#!/usr/bin/env bash\n")] != "#!/usr/bin/env bash\n" {
		t.Errorf("shebang not preserved:\n%s", content)
	}
	if !contains(content, header) {
		t.Errorf("header not found after shebang:\n%s", content)
	}
}

func TestInject_ShebangNoNewline(t *testing.T) {
	// Shebang that is the entire file content (no trailing newline) exercises the
	// else branch in Inject's shebang handling.
	path := writeTemp(t, "#!/usr/bin/env bash")
	h := "# SPDX-License-Identifier: MIT\n"
	if err := injector.Inject(path, h); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	content := string(got)
	if !strings.HasPrefix(content, "#!/usr/bin/env bash") {
		t.Errorf("shebang should be first:\n%s", content)
	}
	if !contains(content, h) {
		t.Errorf("header not found:\n%s", content)
	}
}

func TestHasHeader_FileNotFound(t *testing.T) {
	_, err := injector.HasHeader(nonexistentPath(t))
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestInject_FileNotFound(t *testing.T) {
	err := injector.Inject(nonexistentPath(t), "// header\n")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestHasHeader_EmptyFile(t *testing.T) {
	path := writeTemp(t, "")
	has, err := injector.HasHeader(path)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("empty file should have no header")
	}
}

func TestHasHeader_HeaderBeyondScanWindow(t *testing.T) {
	// Copyright notice past line 20 must NOT be detected — we only scan the top 20 lines.
	var b strings.Builder
	for i := 0; i < 21; i++ {
		b.WriteString("// line\n")
	}
	b.WriteString("// Copyright 2026 Someone\n")

	path := writeTemp(t, b.String())
	has, err := injector.HasHeader(path)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("copyright beyond scan window should not be detected")
	}
}

func TestHasHeader_CaseInsensitive(t *testing.T) {
	path := writeTemp(t, "// COPYRIGHT 2026 SOMEONE\npackage main\n")
	has, err := injector.HasHeader(path)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("detection should be case-insensitive")
	}
}

// ── Remove ────────────────────────────────────────────────────────────────────

func TestRemove_LineComment_RemovesHeader(t *testing.T) {
	content := "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\n\npackage main\n\nfunc main() {}\n"
	path := writeTemp(t, content)

	changed, err := injector.Remove(path, "//", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	got, _ := os.ReadFile(path)
	result := string(got)
	if strings.Contains(result, "SPDX") || strings.Contains(result, "Copyright") {
		t.Errorf("header still present:\n%s", result)
	}
	if !strings.Contains(result, "package main") {
		t.Errorf("body was removed:\n%s", result)
	}
}

func TestRemove_LineComment_PreservesShebang(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.sh")
	content := "#!/usr/bin/env bash\n# SPDX-License-Identifier: MIT\n# Copyright 2026 Grégoire\n\necho hello\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := injector.Remove(path, "#", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	got, _ := os.ReadFile(path)
	result := string(got)
	if !strings.HasPrefix(result, "#!/usr/bin/env bash\n") {
		t.Errorf("shebang not preserved:\n%s", result)
	}
	if strings.Contains(result, "SPDX") {
		t.Errorf("SPDX line still present:\n%s", result)
	}
}

func TestRemove_LineComment_NoHeader_NoChange(t *testing.T) {
	content := "package main\n\nfunc main() {}\n"
	path := writeTemp(t, content)

	changed, err := injector.Remove(path, "//", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("expected changed=false for file without header")
	}

	got, _ := os.ReadFile(path)
	if string(got) != content {
		t.Error("file should be unmodified")
	}
}

func TestRemove_LineComment_NonLicenseComment_NoChange(t *testing.T) {
	content := "// Package main does something.\n// It is not a license.\npackage main\n"
	path := writeTemp(t, content)

	changed, err := injector.Remove(path, "//", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("non-license comment block must not be removed")
	}
}

func TestRemove_BlockComment_HTML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.html")
	content := "<!--\nSPDX-License-Identifier: MIT\nCopyright 2026 Grégoire\n-->\n<html></html>\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := injector.Remove(path, "", "<!--", "-->")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	got, _ := os.ReadFile(path)
	result := string(got)
	if strings.Contains(result, "SPDX") {
		t.Errorf("header still present:\n%s", result)
	}
	if !strings.Contains(result, "<html>") {
		t.Errorf("body was removed:\n%s", result)
	}
}

func TestRemove_BlockComment_CSS(t *testing.T) {
	path := filepath.Join(t.TempDir(), "style.css")
	content := "/* SPDX-License-Identifier: MIT\n   Copyright 2026 Grégoire */\nbody { margin: 0; }\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := injector.Remove(path, "", "/*", "*/")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	got, _ := os.ReadFile(path)
	result := string(got)
	if strings.Contains(result, "SPDX") {
		t.Errorf("header still present:\n%s", result)
	}
	if !strings.Contains(result, "body") {
		t.Errorf("body was removed:\n%s", result)
	}
}

func TestRemove_Idempotent(t *testing.T) {
	content := "// SPDX-License-Identifier: MIT\n// Copyright 2026 Grégoire\n\npackage main\n"
	path := writeTemp(t, content)

	changed1, err := injector.Remove(path, "//", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !changed1 {
		t.Fatal("first Remove should change the file")
	}

	changed2, err := injector.Remove(path, "//", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if changed2 {
		t.Error("second Remove should be a no-op (idempotent)")
	}
}

func TestRemove_NonExistentFile_Error(t *testing.T) {
	_, err := injector.Remove(nonexistentPath(t), "//", "", "")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestRemove_BlockComment_NonLicenseContent_NoChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.html")
	content := "<!--\nThis is a regular HTML comment.\nNo license info here.\n-->\n<html></html>\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := injector.Remove(path, "", "<!--", "-->")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("non-license block comment must not be removed")
	}
	got, _ := os.ReadFile(path)
	if string(got) != content {
		t.Error("file should be unmodified")
	}
}

func TestRemove_BlockComment_UnclosedBlock_NoChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.html")
	content := "<!-- SPDX-License-Identifier: MIT\n<html></html>\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := injector.Remove(path, "", "<!--", "-->")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("unclosed block comment should not be removed")
	}
}

func TestRemove_BlockComment_NoOpenAtStart_NoChange(t *testing.T) {
	// Content that doesn't start with the block-open marker; findBlockCommentHeader
	// must return (0, false) immediately without scanning further.
	path := filepath.Join(t.TempDir(), "main.c")
	content := "int main() { return 0; }\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := injector.Remove(path, "", "/*", "*/")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("file not starting with block-open must not be modified")
	}
}

// ── ExtractYear ───────────────────────────────────────────────────────────────

func TestExtractYear_Copyright(t *testing.T) {
	path := writeTemp(t, "// Copyright 2023 Author\n// SPDX-License-Identifier: MIT\npackage main\n")
	if got := injector.ExtractYear(path); got != 2023 {
		t.Errorf("expected 2023, got %d", got)
	}
}

func TestExtractYear_SPDXFileCopyrightText(t *testing.T) {
	path := writeTemp(t, "// SPDX-FileCopyrightText: 2021 Author\npackage main\n")
	if got := injector.ExtractYear(path); got != 2021 {
		t.Errorf("expected 2021, got %d", got)
	}
}

func TestExtractYear_NoYear(t *testing.T) {
	path := writeTemp(t, "package main\nfunc main() {}\n")
	if got := injector.ExtractYear(path); got != 0 {
		t.Errorf("expected 0 for file without year, got %d", got)
	}
}

func TestExtractYear_FileNotFound(t *testing.T) {
	if got := injector.ExtractYear(nonexistentPath(t)); got != 0 {
		t.Errorf("expected 0 for missing file, got %d", got)
	}
}

func TestExtractYear_BeyondScanWindow(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 21; i++ {
		b.WriteString("// line\n")
	}
	b.WriteString("// Copyright 2019 Someone\n")
	path := writeTemp(t, b.String())
	if got := injector.ExtractYear(path); got != 0 {
		t.Errorf("expected 0 (year past scan window), got %d", got)
	}
}

// ── Permission preservation ───────────────────────────────────────────────────

func TestInject_PreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits not supported on Windows")
	}
	path := filepath.Join(t.TempDir(), "script.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho hello\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := injector.Inject(path, "# SPDX-License-Identifier: MIT\n# Copyright 2026 Test"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("expected permissions 0755, got %04o", info.Mode().Perm())
	}
}

func TestRemove_PreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits not supported on Windows")
	}
	path := filepath.Join(t.TempDir(), "script.sh")
	content := "#!/bin/sh\n# SPDX-License-Identifier: MIT\n# Copyright 2026 Test\n\necho hello\n"
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := injector.Remove(path, "#", "", ""); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("expected permissions 0755, got %04o", info.Mode().Perm())
	}
}

// nonexistentPath returns a path guaranteed not to exist.
func nonexistentPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "does_not_exist_12345.go")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
