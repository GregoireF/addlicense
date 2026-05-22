// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package addlicense_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	addlicense "github.com/GregoireF/addlicense/pkg/addlicense"
)

const (
	testLicense = "MIT"
	testAuthor  = "Acme Corp"
)

// writeGoFile writes a Go source file into dir and returns its path.
func writeGoFile(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "main.go")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeGoFile: %v", err)
	}
	return p
}

func TestDefaultIgnore_NonEmpty(t *testing.T) {
	if len(addlicense.DefaultIgnore) == 0 {
		t.Fatal("DefaultIgnore must not be empty")
	}
}

func TestFormatConstants(t *testing.T) {
	if addlicense.FormatText != "text" {
		t.Errorf("FormatText = %q, want %q", addlicense.FormatText, "text")
	}
	if addlicense.FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", addlicense.FormatJSON, "json")
	}
}

func TestRun_CheckOnly_Licensed_NoError(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")

	err := addlicense.Run(addlicense.Options{
		CheckOnly: true,
		Paths:     []string{dir},
	})
	if err != nil {
		t.Fatalf("expected nil error for licensed file, got: %v", err)
	}
}

func TestRun_CheckOnly_Unlicensed_Error(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "package main\n")

	err := addlicense.Run(addlicense.Options{
		CheckOnly: true,
		Paths:     []string{dir},
	})
	if err == nil {
		t.Fatal("expected non-nil error for unlicensed file, got nil")
	}
}

func TestRun_Add_InjectsHeader(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "package main\n")

	err := addlicense.Run(addlicense.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Paths:   []string{dir},
		Quiet:   true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(got)
	if !strings.Contains(content, "SPDX-License-Identifier: MIT") {
		t.Errorf("header not injected; got:\n%s", content)
	}
	if !strings.Contains(content, "Acme Corp") {
		t.Errorf("author not in header; got:\n%s", content)
	}
}

func TestRun_Add_Idempotent(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "package main\n")

	opts := addlicense.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Paths:   []string{dir},
		Quiet:   true,
	}

	if err := addlicense.Run(opts); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	first, _ := os.ReadFile(p)

	if err := addlicense.Run(opts); err != nil {
		t.Fatalf("second Run: %v", err)
	}
	second, _ := os.ReadFile(p)

	if string(first) != string(second) {
		t.Error("Run is not idempotent — file changed on second invocation")
	}
}

func TestRun_Diff_Unlicensed_Error_NoWrite(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "package main\n")
	before, _ := os.ReadFile(p)

	err := addlicense.Run(addlicense.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Diff:    true,
		Paths:   []string{dir},
	})
	if err == nil {
		t.Fatal("expected non-nil error from Diff on unlicensed file, got nil")
	}

	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Error("Diff mode must not modify files")
	}
}

func TestRun_Diff_Licensed_NoError(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")

	err := addlicense.Run(addlicense.Options{
		License: testLicense,
		Diff:    true,
		Paths:   []string{dir},
	})
	if err != nil {
		t.Fatalf("Diff on already-licensed tree should return nil, got: %v", err)
	}
}

func TestRun_MultiAuthor_BothLinesPresent(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "package main\n")

	err := addlicense.Run(addlicense.Options{
		License: testLicense,
		Author:  "Alice Dupont, Bob Martin",
		Year:    2026,
		Paths:   []string{dir},
		Quiet:   true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := os.ReadFile(p)
	content := string(got)
	if !strings.Contains(content, "Alice Dupont") {
		t.Errorf("Alice not in header; got:\n%s", content)
	}
	if !strings.Contains(content, "Bob Martin") {
		t.Errorf("Bob not in header; got:\n%s", content)
	}
}

func TestRun_DryRun_NoWrite(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "package main\n")
	before, _ := os.ReadFile(p)

	err := addlicense.Run(addlicense.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		DryRun:  true,
		Paths:   []string{dir},
	})
	if err != nil {
		t.Fatalf("DryRun: %v", err)
	}

	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Error("DryRun must not modify files")
	}
}

func TestRun_Remove_StripsHeader(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")

	err := addlicense.Run(addlicense.Options{
		Remove: true,
		Paths:  []string{dir},
		Quiet:  true,
	})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	got, _ := os.ReadFile(p)
	if strings.Contains(string(got), "SPDX-License-Identifier") {
		t.Errorf("header not removed; got:\n%s", string(got))
	}
}

func TestRun_Reuse_EmitsSPDXFileCopyrightText(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "package main\n")

	err := addlicense.Run(addlicense.Options{
		License: testLicense,
		Author:  testAuthor,
		Year:    2026,
		Reuse:   true,
		Paths:   []string{dir},
		Quiet:   true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := os.ReadFile(p)
	if !strings.Contains(string(got), "SPDX-FileCopyrightText:") {
		t.Errorf("REUSE header not found; got:\n%s", string(got))
	}
}

func TestRun_Sbom_WritesDocument(t *testing.T) {
	dir := t.TempDir()
	p := writeGoFile(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")
	_ = p
	out := dir + "/sbom.spdx"

	err := addlicense.Run(addlicense.Options{
		Sbom:  out,
		Paths: []string{dir},
	})
	if err != nil {
		t.Fatalf("Run --sbom: %v", err)
	}

	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading sbom: %v", err)
	}
	if !strings.Contains(string(content), "SPDXVersion: SPDX-2.3") {
		t.Errorf("expected SPDX 2.3 output:\n%s", string(content))
	}
}

func TestRun_InvalidMutualExclusion_CheckAndDiff(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "package main\n")

	err := addlicense.Run(addlicense.Options{
		CheckOnly: true,
		Diff:      true,
		Paths:     []string{dir},
	})
	if err == nil {
		t.Fatal("expected error for CheckOnly+Diff mutual exclusion, got nil")
	}
}
