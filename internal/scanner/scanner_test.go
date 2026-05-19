// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GregoireF/addlicense/internal/scanner"
)

func makeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	files := map[string]string{
		"main.go":             "package main",
		"cmd/cli.go":          "package cmd",
		"node_modules/lib.js": "module.exports = {}",
		"vendor/pkg/pkg.go":   "package pkg",
		"script.sh":           "#!/bin/bash",
		"README.md":           "# hello",
		"config.yaml":         "key: value",
	}

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

func TestWalk_IgnoresNodeModulesAndVendor(t *testing.T) {
	root := makeTree(t)
	files, err := scanner.Walk([]string{root}, []string{"node_modules", "vendor"})
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if contains(f.Path, "node_modules") || contains(f.Path, "vendor") {
			t.Errorf("should have ignored %s", f.Path)
		}
	}
}

func TestWalk_FindsGoAndYAML(t *testing.T) {
	root := makeTree(t)
	files, err := scanner.Walk([]string{root}, []string{"node_modules", "vendor"})
	if err != nil {
		t.Fatal(err)
	}

	exts := make(map[string]bool)
	for _, f := range files {
		exts[f.Ext] = true
	}

	for _, want := range []string{".go", ".sh", ".yaml"} {
		if !exts[want] {
			t.Errorf("expected to find %s files, got: %v", want, exts)
		}
	}
}

func TestWalk_GlobPattern(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"main.go":      "package main",
		"model.pb.go":  "// generated",
		"types.gen.go": "// generated",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	found, err := scanner.Walk([]string{root}, []string{"*.pb.go", "*.gen.go"})
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range found {
		base := filepath.Base(f.Path)
		if base == "model.pb.go" || base == "types.gen.go" {
			t.Errorf("glob-ignored file should be skipped: %s", f.Path)
		}
	}
	if len(found) != 1 {
		t.Errorf("expected 1 file (main.go), got %d", len(found))
	}
}

func TestWalk_InvalidRoot(t *testing.T) {
	_, err := scanner.Walk([]string{filepath.Join(t.TempDir(), "nonexistent")}, nil)
	if err == nil {
		t.Error("expected error for nonexistent root, got nil")
	}
}

func TestWalk_SkipsNoExtension(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Makefile", "Dockerfile", "main.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	found, err := scanner.Walk([]string{root}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range found {
		base := filepath.Base(f.Path)
		if base == "Makefile" || base == "Dockerfile" {
			t.Errorf("file with no extension should be skipped: %s", f.Path)
		}
	}
}

func TestWalk_CollectsMarkdownAndJSON(t *testing.T) {
	// The scanner is language-agnostic: it collects ALL files with extensions.
	// Filtering by supported language is the runner's responsibility, not the scanner's.
	root := makeTree(t)
	files, err := scanner.Walk([]string{root}, []string{"node_modules", "vendor"})
	if err != nil {
		t.Fatal(err)
	}

	exts := make(map[string]bool)
	for _, f := range files {
		exts[f.Ext] = true
	}

	// .md is in the tree (README.md) — scanner should return it.
	if !exts[".md"] {
		t.Error("scanner should collect .md files (language filtering happens in runner)")
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
