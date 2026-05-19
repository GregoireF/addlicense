// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package injector_test

import (
	"os"
	"path/filepath"
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
