// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

// Package cli_test exercises the full CLI binary built from ./cmd/addlicense.
// These tests cover the Cobra layer (flag parsing, version, exit codes, error messages)
// which the integration tests in tests/integration/ bypass by calling runner.Run directly.
package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// binaryPath is set by TestMain after building the binary once.
var binaryPath string

func TestMain(m *testing.M) {
	bin, cleanup, err := buildBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	binaryPath = bin
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// run executes the binary with args. Returns stdout, stderr, and the exit code.
func run(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("unexpected exec error: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// writeMainGo creates a main.go file in dir with the given content.
func writeMainGo(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ── version / help ────────────────────────────────────────────────────────────

func TestCLI_Version(t *testing.T) {
	stdout, _, code := run(t, "--version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "addlicense") {
		t.Errorf("version output missing 'addlicense':\n%s", stdout)
	}
}

func TestCLI_Help(t *testing.T) {
	stdout, _, code := run(t, "--help")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	for _, flag := range []string{"--license", "--check", "--remove", "--update", "--year-range", "--dep5"} {
		if !strings.Contains(stdout, flag) {
			t.Errorf("help output missing flag %s", flag)
		}
	}
}

func TestCLI_NoArgs_ExitsNonZero(t *testing.T) {
	_, _, code := run(t)
	if code == 0 {
		t.Error("expected non-zero exit when called with no arguments")
	}
}

// ── add / check ───────────────────────────────────────────────────────────────

func TestCLI_Add_Then_Check(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	_, _, code := run(t, "--license", "MIT", "--author", "Test", dir)
	if code != 0 {
		t.Fatalf("add: expected exit 0, got %d", code)
	}

	_, _, code = run(t, "--check", dir)
	if code != 0 {
		t.Fatalf("check after add: expected exit 0, got %d", code)
	}
}

func TestCLI_Check_Missing_ExitsOne(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	_, stderr, code := run(t, "--check", dir)
	if code != 1 {
		t.Fatalf("expected exit 1 for missing header, got %d", code)
	}
	if !strings.Contains(stderr, "missing") {
		t.Errorf("expected 'missing' in stderr:\n%s", stderr)
	}
}

// ── remove ────────────────────────────────────────────────────────────────────

func TestCLI_Remove(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")

	_, _, code := run(t, "--remove", dir)
	if code != 0 {
		t.Fatalf("remove: expected exit 0, got %d", code)
	}

	_, _, code = run(t, "--check", dir)
	if code != 1 {
		t.Error("after remove, check should fail (header gone)")
	}
}

// ── dry-run ───────────────────────────────────────────────────────────────────

func TestCLI_DryRun_NoWrite(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	stdout, _, code := run(t, "--license", "MIT", "--author", "Test", "--dry-run", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "dry-run") {
		t.Errorf("expected '[dry-run]' in stdout:\n%s", stdout)
	}

	// File must remain unmodified.
	got, _ := os.ReadFile(filepath.Join(dir, "main.go"))
	if strings.Contains(string(got), "SPDX") {
		t.Error("dry-run must not write to file")
	}
}

// ── format json ───────────────────────────────────────────────────────────────

func TestCLI_Format_JSON(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")

	stdout, _, code := run(t, "--check", "--format", "json", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	var rec struct {
		File   string `json:"file"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &rec); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, stdout)
	}
	if rec.Status != "ok" {
		t.Errorf("expected status=ok, got %q", rec.Status)
	}
}

func TestCLI_Format_Invalid(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	_, stderr, code := run(t, "--format", "xml", dir)
	if code == 0 {
		t.Error("expected non-zero exit for invalid --format")
	}
	if !strings.Contains(stderr, "xml") {
		t.Errorf("expected 'xml' in error message:\n%s", stderr)
	}
}

// ── mutual exclusion ─────────────────────────────────────────────────────────

func TestCLI_MutualExclusion_VerboseQuiet(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")
	_, stderr, code := run(t, "--verbose", "--quiet", dir)
	if code == 0 {
		t.Error("expected non-zero exit for --verbose --quiet")
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' in error:\n%s", stderr)
	}
}

func TestCLI_MutualExclusion_CheckRemove(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")
	_, _, code := run(t, "--check", "--remove", dir)
	if code == 0 {
		t.Error("expected non-zero exit for --check --remove")
	}
}

func TestCLI_MutualExclusion_RemoveUpdate(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")
	_, _, code := run(t, "--remove", "--update", dir)
	if code == 0 {
		t.Error("expected non-zero exit for --remove --update")
	}
}

// ── year-range ────────────────────────────────────────────────────────────────

func TestCLI_YearRange(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "// Copyright 2020 Test\n// SPDX-License-Identifier: MIT\n\npackage main\n")

	_, _, code := run(t, "--license", "MIT", "--author", "Test", "--update", "--year-range", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	got, _ := os.ReadFile(filepath.Join(dir, "main.go"))
	if !strings.Contains(string(got), "2020-") {
		t.Errorf("expected year range '2020-' in header:\n%s", got)
	}
}

// ── dep5 ─────────────────────────────────────────────────────────────────────

func TestCLI_Dep5(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")
	if err := os.WriteFile(filepath.Join(dir, "logo.png"), []byte("\x89PNG"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, code := run(t, "--license", "MIT", "--author", "Test", "--dep5", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	dep5, err := os.ReadFile(filepath.Join(dir, ".reuse", "dep5"))
	if err != nil {
		t.Fatalf(".reuse/dep5 not created: %v", err)
	}
	if !strings.Contains(string(dep5), "logo.png") {
		t.Errorf("logo.png missing from dep5:\n%s", dep5)
	}
}

// ── multi-author ──────────────────────────────────────────────────────────────

func TestCLI_MultiAuthor(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	_, _, code := run(t, "--license", "MIT", "--author", "Alice, Bob", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	got, _ := os.ReadFile(filepath.Join(dir, "main.go"))
	if !strings.Contains(string(got), "Copyright") || !strings.Contains(string(got), "Alice") {
		t.Errorf("expected Alice copyright in header:\n%s", got)
	}
	if !strings.Contains(string(got), "Bob") {
		t.Errorf("expected Bob copyright in header:\n%s", got)
	}
}

// ── diff ──────────────────────────────────────────────────────────────────────

func TestCLI_Diff_NoWrite_ExitsOne(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	stdout, _, code := run(t, "--license", "MIT", "--author", "Test", "--diff", dir)
	if code != 1 {
		t.Fatalf("expected exit 1 (file would be modified), got %d", code)
	}
	// JSON with header field must be present.
	if !strings.Contains(stdout, "diff-add") {
		t.Errorf("expected 'diff-add' in stdout:\n%s", stdout)
	}
	if !strings.Contains(stdout, "header") {
		t.Errorf("expected 'header' field in JSON:\n%s", stdout)
	}
	// File must be untouched.
	got, _ := os.ReadFile(filepath.Join(dir, "main.go"))
	if strings.Contains(string(got), "SPDX") {
		t.Error("--diff must not write to file")
	}
}

func TestCLI_Diff_AlreadyLicensed_ExitsZero(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")

	_, _, code := run(t, "--diff", dir)
	if code != 0 {
		t.Fatalf("already-licensed file: expected exit 0, got %d", code)
	}
}

func TestCLI_Diff_CheckConflict(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")
	_, _, code := run(t, "--diff", "--check", dir)
	if code == 0 {
		t.Error("expected non-zero exit for --diff --check")
	}
}

// v1.0.0: --sbom

func TestCLI_Sbom_WritesFile(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")
	out := filepath.Join(t.TempDir(), "sbom.spdx")

	_, _, code := run(t, "--sbom", out, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading sbom output: %v", err)
	}
	if !strings.Contains(string(content), "SPDXVersion: SPDX-2.3") {
		t.Errorf("expected SPDX 2.3 output:\n%s", string(content))
	}
	if !strings.Contains(string(content), "LicenseConcluded: MIT") {
		t.Errorf("expected MIT in output:\n%s", string(content))
	}
}

func TestCLI_Sbom_Stdout(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "// SPDX-License-Identifier: MIT\n// Copyright 2026 Test\n\npackage main\n")

	stdout, _, code := run(t, "--sbom", "-", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "SPDXVersion: SPDX-2.3") {
		t.Errorf("expected SPDX output on stdout:\n%s", stdout)
	}
}

func TestCLI_Sbom_MutualExclusion_Check(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")
	out := filepath.Join(t.TempDir(), "sbom.spdx")

	_, _, code := run(t, "--sbom", out, "--check", dir)
	if code == 0 {
		t.Error("expected non-zero exit for --sbom --check")
	}
}

// v0.9.0: --author-file

func TestCLI_AuthorFile_InjectsAllAuthors(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	authorsFile := filepath.Join(t.TempDir(), "AUTHORS")
	if err := os.WriteFile(authorsFile, []byte("Alice Dupont\n# comment\nBob Martin\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, code := run(t, "--license", "MIT", "--author-file", authorsFile, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	got, _ := os.ReadFile(filepath.Join(dir, "main.go"))
	content := string(got)
	if !strings.Contains(content, "Alice Dupont") {
		t.Errorf("Alice not in header:\n%s", content)
	}
	if !strings.Contains(content, "Bob Martin") {
		t.Errorf("Bob not in header:\n%s", content)
	}
}

func TestCLI_AuthorFile_MutualExclusionWithAuthor(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	authorsFile := filepath.Join(t.TempDir(), "AUTHORS")
	if err := os.WriteFile(authorsFile, []byte("Alice\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, code := run(t, "--license", "MIT", "--author", "Bob", "--author-file", authorsFile, dir)
	if code == 0 {
		t.Error("expected non-zero exit for --author + --author-file")
	}
}

func TestCLI_AuthorFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	_, _, code := run(t, "--license", "MIT", "--author-file", "/nonexistent/AUTHORS", dir)
	if code == 0 {
		t.Error("expected non-zero exit for missing author file")
	}
}

func TestCLI_AuthorFile_Empty_Error(t *testing.T) {
	dir := t.TempDir()
	writeMainGo(t, dir, "package main\n")

	authorsFile := filepath.Join(t.TempDir(), "AUTHORS")
	// Only comments and blank lines — no actual authors.
	if err := os.WriteFile(authorsFile, []byte("# just a comment\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, code := run(t, "--license", "MIT", "--author-file", authorsFile, dir)
	if code == 0 {
		t.Error("expected non-zero exit for empty author file")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildBinary compiles cmd/addlicense to a temp binary and returns its path
// along with a cleanup function.
func buildBinary() (string, func(), error) {
	// Find module root: go up from this file until we find go.mod.
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", func() {}, fmt.Errorf("go.mod not found")
		}
		dir = parent
	}

	tmp, err := os.MkdirTemp("", "addlicense-cli-test-*")
	if err != nil {
		return "", func() {}, err
	}

	bin := filepath.Join(tmp, "addlicense")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, "./cmd/addlicense")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(tmp)
		return "", func() {}, fmt.Errorf("%w\n%s", err, out)
	}

	return bin, func() { _ = os.RemoveAll(tmp) }, nil
}
