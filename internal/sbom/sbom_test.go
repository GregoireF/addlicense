// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package sbom_test

import (
	"strings"
	"testing"
	"time"

	"github.com/GregoireF/addlicense/internal/sbom"
)

func fixedTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2026-05-22T12:00:00Z")
	return t
}

func baseDoc(entries []sbom.FileEntry) sbom.Document {
	return sbom.Document{
		Name:    "my-project",
		Tool:    "addlicense-1.0.0",
		Created: fixedTime(),
		Entries: entries,
	}
}

func TestBuild_HeaderSection(t *testing.T) {
	got := sbom.Build(baseDoc(nil))

	for _, want := range []string{
		"SPDXVersion: SPDX-2.3",
		"DataLicense: CC0-1.0",
		"SPDX-ID: SPDXRef-DOCUMENT",
		"DocumentName: my-project",
		"Creator: Tool: addlicense-1.0.0",
		"Created: 2026-05-22T12:00:00Z",
		"PackageName: my-project",
		"SPDX-ID: SPDXRef-Package",
		"Relationship: SPDXRef-DOCUMENT DESCRIBES SPDXRef-Package",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestBuild_FileWithHeader(t *testing.T) {
	entries := []sbom.FileEntry{
		{Path: "main.go", LicenseID: "MIT", CopyrightText: "2026 Acme Corp"},
	}
	got := sbom.Build(baseDoc(entries))

	for _, want := range []string{
		"FileName: ./main.go",
		"LicenseConcluded: MIT",
		"LicenseInfoInFile: MIT",
		"FileCopyrightText: 2026 Acme Corp",
		"Relationship: SPDXRef-Package CONTAINS SPDXRef-File-0",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

func TestBuild_FileWithoutHeader_NOASSERTION(t *testing.T) {
	entries := []sbom.FileEntry{
		{Path: "unlicensed.go"},
	}
	got := sbom.Build(baseDoc(entries))

	if !strings.Contains(got, "LicenseConcluded: NOASSERTION") {
		t.Errorf("expected NOASSERTION for file with no header:\n%s", got)
	}
	if !strings.Contains(got, "LicenseInfoInFile: NOASSERTION") {
		t.Errorf("expected LicenseInfoInFile NOASSERTION:\n%s", got)
	}
	if !strings.Contains(got, "FileCopyrightText: NOASSERTION") {
		t.Errorf("expected FileCopyrightText NOASSERTION:\n%s", got)
	}
}

func TestBuild_FallbackLicense(t *testing.T) {
	// No LicenseID in file but FallbackLicense set (from --license flag).
	entries := []sbom.FileEntry{
		{Path: "script.sh", FallbackLicense: "Apache-2.0"},
	}
	got := sbom.Build(baseDoc(entries))

	if !strings.Contains(got, "LicenseConcluded: Apache-2.0") {
		t.Errorf("expected fallback license Apache-2.0 in LicenseConcluded:\n%s", got)
	}
	// LicenseInfoInFile must still be NOASSERTION (not in the file itself).
	if !strings.Contains(got, "LicenseInfoInFile: NOASSERTION") {
		t.Errorf("expected LicenseInfoInFile NOASSERTION when header absent:\n%s", got)
	}
}

func TestBuild_MultipleFiles(t *testing.T) {
	entries := []sbom.FileEntry{
		{Path: "a.go", LicenseID: "MIT", CopyrightText: "2026 Alice"},
		{Path: "b.go", LicenseID: "Apache-2.0", CopyrightText: "2026 Bob"},
	}
	got := sbom.Build(baseDoc(entries))

	if !strings.Contains(got, "SPDXRef-File-0") {
		t.Errorf("expected SPDXRef-File-0:\n%s", got)
	}
	if !strings.Contains(got, "SPDXRef-File-1") {
		t.Errorf("expected SPDXRef-File-1:\n%s", got)
	}
	if !strings.Contains(got, "LicenseConcluded: MIT") {
		t.Errorf("expected MIT:\n%s", got)
	}
	if !strings.Contains(got, "LicenseConcluded: Apache-2.0") {
		t.Errorf("expected Apache-2.0:\n%s", got)
	}
}

func TestBuild_DocumentNamespace_ContainsName(t *testing.T) {
	got := sbom.Build(baseDoc(nil))
	if !strings.Contains(got, "DocumentNamespace: https://spdx.org/spdxdocs/my-project-") {
		t.Errorf("unexpected namespace:\n%s", got)
	}
}

func TestBuild_DefaultsWhenEmpty(t *testing.T) {
	// Empty document with zero values — should not panic, should produce valid header.
	doc := sbom.Document{}
	got := sbom.Build(doc)
	if !strings.Contains(got, "SPDXVersion: SPDX-2.3") {
		t.Errorf("expected SPDXVersion even with zero Document:\n%s", got)
	}
}

func TestBuild_PathNormalisation(t *testing.T) {
	// Windows-style backslash paths must be normalised to forward slashes.
	entries := []sbom.FileEntry{
		{Path: `sub\dir\file.go`, LicenseID: "MIT"},
	}
	got := sbom.Build(baseDoc(entries))
	if strings.Contains(got, `\`) {
		t.Errorf("backslash found in SPDX output — paths must use forward slashes:\n%s", got)
	}
}
