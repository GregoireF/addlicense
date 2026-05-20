// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package dep5_test

import (
	"strings"
	"testing"

	"github.com/GregoireF/addlicense/internal/dep5"
)

func TestBuild_WithPaths(t *testing.T) {
	got := dep5.Build([]string{"assets/logo.png", "docs/diagram.svg"}, 2026, "Acme Corp", "MIT")

	if !strings.Contains(got, "Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/") {
		t.Error("missing Format header")
	}
	if !strings.Contains(got, "Upstream-Contact: Acme Corp") {
		t.Error("missing Upstream-Contact")
	}
	if !strings.Contains(got, "Files: assets/logo.png") {
		t.Error("missing first file path")
	}
	if !strings.Contains(got, "docs/diagram.svg") {
		t.Error("missing second file path")
	}
	if !strings.Contains(got, "Copyright: 2026 Acme Corp") {
		t.Error("missing Copyright line")
	}
	if !strings.Contains(got, "License: MIT") {
		t.Error("missing License line")
	}
}

func TestBuild_NoPaths_HeaderOnly(t *testing.T) {
	got := dep5.Build(nil, 2026, "Acme Corp", "MIT")

	if !strings.Contains(got, "Format:") {
		t.Error("missing Format header")
	}
	if strings.Contains(got, "Files:") {
		t.Error("Files paragraph should be absent when paths is empty")
	}
	if strings.Contains(got, "Copyright:") {
		t.Error("Copyright line should be absent when paths is empty")
	}
}

func TestBuild_NoAuthor(t *testing.T) {
	got := dep5.Build([]string{"logo.png"}, 2026, "", "MIT")

	if strings.Contains(got, "Upstream-Contact:") {
		t.Error("Upstream-Contact should be omitted when author is empty")
	}
	if !strings.Contains(got, "Copyright: 2026\n") {
		t.Errorf("expected bare year copyright line, got:\n%s", got)
	}
}

func TestBuild_MultiplePaths_ContinuationLines(t *testing.T) {
	paths := []string{"a.png", "b.jpg", "c.gif"}
	got := dep5.Build(paths, 2026, "Author", "MIT")

	lines := strings.Split(got, "\n")
	var filesLines []string
	inFiles := false
	for _, l := range lines {
		if strings.HasPrefix(l, "Files:") {
			inFiles = true
		}
		if inFiles && (strings.HasPrefix(l, "Files:") || strings.HasPrefix(l, " ")) {
			filesLines = append(filesLines, l)
		} else if inFiles {
			break
		}
	}
	if len(filesLines) != 3 {
		t.Errorf("expected 3 Files lines (1 main + 2 continuation), got %d: %v", len(filesLines), filesLines)
	}
	if !strings.HasPrefix(filesLines[1], " ") {
		t.Errorf("continuation line must start with space: %q", filesLines[1])
	}
}

func TestBuild_ForwardSlashPaths(t *testing.T) {
	got := dep5.Build([]string{"assets/images/logo.png"}, 2026, "A", "MIT")
	if strings.Contains(got, `\`) {
		t.Error("dep5 paths must use forward slashes")
	}
}
