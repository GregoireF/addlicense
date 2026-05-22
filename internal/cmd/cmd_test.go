// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package cmd

import (
	"testing"
)

func TestParseAuthorFile_Empty(t *testing.T) {
	if got := parseAuthorFile(""); len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestParseAuthorFile_SingleAuthor(t *testing.T) {
	got := parseAuthorFile("Alice Dupont\n")
	if len(got) != 1 || got[0] != "Alice Dupont" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestParseAuthorFile_CommentSkipped(t *testing.T) {
	got := parseAuthorFile("# This is a comment\nAlice\n")
	if len(got) != 1 || got[0] != "Alice" {
		t.Errorf("expected [Alice], got %v", got)
	}
}

func TestParseAuthorFile_BlankLinesSkipped(t *testing.T) {
	got := parseAuthorFile("Alice\n\n\nBob\n")
	if len(got) != 2 || got[0] != "Alice" || got[1] != "Bob" {
		t.Errorf("expected [Alice Bob], got %v", got)
	}
}

func TestParseAuthorFile_WhitespaceTrimmed(t *testing.T) {
	got := parseAuthorFile("  Alice Dupont  \n  Bob Martin  \n")
	if len(got) != 2 || got[0] != "Alice Dupont" || got[1] != "Bob Martin" {
		t.Errorf("expected trimmed authors, got %v", got)
	}
}

func TestParseAuthorFile_CommentsAndBlanks(t *testing.T) {
	input := "# Copyright holders\n\nAlice\n# another comment\nBob\n"
	got := parseAuthorFile(input)
	if len(got) != 2 || got[0] != "Alice" || got[1] != "Bob" {
		t.Errorf("expected [Alice Bob], got %v", got)
	}
}

func TestParseAuthorFile_NoNewlineAtEnd(t *testing.T) {
	got := parseAuthorFile("Alice")
	if len(got) != 1 || got[0] != "Alice" {
		t.Errorf("expected [Alice], got %v", got)
	}
}
