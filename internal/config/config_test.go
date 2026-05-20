// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GregoireF/addlicense/internal/config"
)

// ── Load ──────────────────────────────────────────────────────────────────────

func TestLoad_NoConfigFile(t *testing.T) {
	dir := t.TempDir()
	opts := config.Options{License: "MIT", Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load with no config file: %v", err)
	}
	if opts.License != "MIT" {
		t.Errorf("license should stay MIT, got %q", opts.License)
	}
}

func TestLoad_ReadsYAML(t *testing.T) {
	dir := t.TempDir()
	content := "license: Apache-2.0\nauthor: TestAuthor\nyear: 2025\n"
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), content)

	opts := config.Options{Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if opts.License != "Apache-2.0" {
		t.Errorf("license: want Apache-2.0, got %q", opts.License)
	}
	if opts.Author != "TestAuthor" {
		t.Errorf("author: want TestAuthor, got %q", opts.Author)
	}
	if opts.Year != 2025 {
		t.Errorf("year: want 2025, got %d", opts.Year)
	}
}

func TestLoad_ReadsYML(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yml"), "license: GPL-3.0-only\n")

	opts := config.Options{Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if opts.License != "GPL-3.0-only" {
		t.Errorf("want GPL-3.0-only, got %q", opts.License)
	}
}

func TestLoad_ReadsJSON(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "addlicense.json"), `{"license":"BSD-2-Clause","author":"JSONAuthor"}`)

	opts := config.Options{Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if opts.License != "BSD-2-Clause" {
		t.Errorf("want BSD-2-Clause, got %q", opts.License)
	}
	if opts.Author != "JSONAuthor" {
		t.Errorf("want JSONAuthor, got %q", opts.Author)
	}
}

func TestLoad_CLIFlagTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), "license: Apache-2.0\nauthor: FileAuthor\n")

	opts := config.Options{License: "MIT", Author: "CLIAuthor", Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if opts.License != "MIT" {
		t.Errorf("CLI license should win, got %q", opts.License)
	}
	if opts.Author != "CLIAuthor" {
		t.Errorf("CLI author should win, got %q", opts.Author)
	}
}

func TestLoad_CLIIgnoreTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), "ignore:\n  - generated\n")

	opts := config.Options{Ignore: []string{"custom"}, Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(opts.Ignore) != 1 || opts.Ignore[0] != "custom" {
		t.Errorf("CLI ignore should win, got %v", opts.Ignore)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), "{{invalid yaml}}")

	opts := config.Options{Paths: []string{dir}}
	if err := config.Load(&opts); err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoad_YAMLPriorityOverYML(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), "license: MIT\n")
	write(t, filepath.Join(dir, ".addlicenserc.yml"), "license: Apache-2.0\n")

	opts := config.Options{Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if opts.License != "MIT" {
		t.Errorf(".yaml should take priority over .yml, got %q", opts.License)
	}
}

// ── Normalize ─────────────────────────────────────────────────────────────────

func TestNormalize_Defaults(t *testing.T) {
	opts := config.Options{}
	opts.Normalize()

	if opts.License != "MIT" {
		t.Errorf("default license: want MIT, got %q", opts.License)
	}
	if opts.Year == 0 {
		t.Error("year should be set to current year, got 0")
	}
	if len(opts.Ignore) == 0 {
		t.Error("default ignore list should be non-empty")
	}
}

func TestNormalize_DoesNotOverrideSetValues(t *testing.T) {
	opts := config.Options{License: "EUPL-1.2", Year: 2020}
	opts.Normalize()

	if opts.License != "EUPL-1.2" {
		t.Errorf("should not override license, got %q", opts.License)
	}
	if opts.Year != 2020 {
		t.Errorf("should not override year, got %d", opts.Year)
	}
}

func TestNormalize_CustomIgnorePreserved(t *testing.T) {
	opts := config.Options{Ignore: []string{"mydir"}}
	opts.Normalize()

	if len(opts.Ignore) != 1 || opts.Ignore[0] != "mydir" {
		t.Errorf("custom ignore should be preserved, got %v", opts.Ignore)
	}
}

// ── Reuse ─────────────────────────────────────────────────────────────────────

func TestLoad_ReuseFromConfig(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), "license: MIT\nreuse: true\n")

	opts := config.Options{Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !opts.Reuse {
		t.Error("expected Reuse=true from config file")
	}
}

func TestLoad_CLIReuseTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), "reuse: true\n")

	opts := config.Options{Reuse: true, Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !opts.Reuse {
		t.Error("Reuse should remain true")
	}
}

func TestLoad_ReuseDefaultsFalse(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".addlicenserc.yaml"), "license: MIT\n")

	opts := config.Options{Paths: []string{dir}}
	if err := config.Load(&opts); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if opts.Reuse {
		t.Error("expected Reuse=false when not set in config")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
