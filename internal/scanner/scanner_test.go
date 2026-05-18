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
		"main.go":              "package main",
		"cmd/cli.go":           "package cmd",
		"node_modules/lib.js":  "module.exports = {}",
		"vendor/pkg/pkg.go":    "package pkg",
		"script.sh":            "#!/bin/bash",
		"README.md":            "# hello",
		"config.yaml":          "key: value",
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

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
