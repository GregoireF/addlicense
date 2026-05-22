// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

// Package sbom generates minimal SPDX 2.3 tag-value documents for file-level
// licence data. It is intentionally minimal: it covers file-level
// LicenseConcluded/LicenseInfoInFile/FileCopyrightText fields and does not
// attempt to build a dependency graph (that belongs in a full SBOM tool such
// as syft or cdxgen).
//
// The output satisfies the EU Cyber Resilience Act requirement for file-level
// SPDX identification as described in ENISA's CRA guidance (SPDX 2.3,
// file-element section).
package sbom

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const noAssertion = "NOASSERTION"

// FileEntry holds the licence information for one file.
type FileEntry struct {
	// Path is the file path as it will appear in the SPDX document (e.g. "./main.go").
	Path string
	// LicenseID is the SPDX identifier extracted from the file's SPDX-License-Identifier
	// header line (e.g. "MIT"). Empty when the file has no header.
	LicenseID string
	// CopyrightText is the copyright line extracted from the file header
	// (e.g. "2026 Acme Corp"). Empty when none found.
	CopyrightText string
	// FallbackLicense is used as LicenseConcluded when LicenseID is empty
	// (typically the value of --license from the CLI). Empty means NOASSERTION.
	FallbackLicense string
}

// Document describes the SPDX document to generate.
type Document struct {
	// Name is the project / package name (used in DocumentName and SPDX-IDs).
	Name string
	// Version is written as PackageVersion. Defaults to NOASSERTION.
	Version string
	// Tool is the creator string (e.g. "addlicense-1.0.0").
	Tool string
	// Created is the document creation timestamp.
	Created time.Time
	// Entries lists all files to include in the document.
	Entries []FileEntry
}

var nonSafe = regexp.MustCompile(`[^a-zA-Z0-9.\-]`)

// Build renders a valid SPDX 2.3 tag-value document and returns it as a string.
func Build(doc Document) string {
	if doc.Version == "" {
		doc.Version = noAssertion
	}
	if doc.Tool == "" {
		doc.Tool = "addlicense"
	}
	if doc.Created.IsZero() {
		doc.Created = time.Now().UTC()
	}
	safeName := nonSafe.ReplaceAllString(doc.Name, "-")
	if safeName == "" {
		safeName = "project"
	}
	ns := fmt.Sprintf("https://spdx.org/spdxdocs/%s-%d", safeName, doc.Created.Unix())
	created := doc.Created.UTC().Format("2006-01-02T15:04:05Z")

	var b strings.Builder

	// ── Document creation info ────────────────────────────────────────────────
	fmt.Fprintf(&b, "SPDXVersion: SPDX-2.3\n")
	fmt.Fprintf(&b, "DataLicense: CC0-1.0\n")
	fmt.Fprintf(&b, "SPDX-ID: SPDXRef-DOCUMENT\n")
	fmt.Fprintf(&b, "DocumentName: %s\n", doc.Name)
	fmt.Fprintf(&b, "DocumentNamespace: %s\n", ns)
	fmt.Fprintf(&b, "Creator: Tool: %s\n", doc.Tool)
	fmt.Fprintf(&b, "Created: %s\n", created)

	// ── Package (root) ────────────────────────────────────────────────────────
	fmt.Fprintf(&b, "\nPackageName: %s\n", doc.Name)
	fmt.Fprintf(&b, "SPDX-ID: SPDXRef-Package\n")
	fmt.Fprintf(&b, "PackageVersion: %s\n", doc.Version)
	fmt.Fprintf(&b, "PackageDownloadLocation: %s\n", noAssertion)
	fmt.Fprintf(&b, "FilesAnalyzed: true\n")
	fmt.Fprintf(&b, "PackageLicenseConcluded: %s\n", noAssertion)
	fmt.Fprintf(&b, "PackageLicenseDeclared: %s\n", noAssertion)
	fmt.Fprintf(&b, "PackageCopyrightText: %s\n", noAssertion)
	fmt.Fprintf(&b, "\nRelationship: SPDXRef-DOCUMENT DESCRIBES SPDXRef-Package\n")

	// ── File elements ─────────────────────────────────────────────────────────
	for i, e := range doc.Entries {
		spdxFileRef := fileRef(i, e.Path)
		licenseConcluded := licenseConcluded(e)
		licenseInFile := e.LicenseID
		if licenseInFile == "" {
			licenseInFile = noAssertion
		}
		copyrightText := e.CopyrightText
		if copyrightText == "" {
			copyrightText = noAssertion
		}

		// Normalise path separator and prefix with "./"
		rel := filepath.ToSlash(e.Path)
		if !strings.HasPrefix(rel, "./") && !strings.HasPrefix(rel, "/") {
			rel = "./" + rel
		}

		fmt.Fprintf(&b, "\nFileName: %s\n", rel)
		fmt.Fprintf(&b, "SPDX-ID: %s\n", spdxFileRef)
		fmt.Fprintf(&b, "LicenseConcluded: %s\n", licenseConcluded)
		fmt.Fprintf(&b, "LicenseInfoInFile: %s\n", licenseInFile)
		fmt.Fprintf(&b, "FileCopyrightText: %s\n", copyrightText)
		fmt.Fprintf(&b, "Relationship: SPDXRef-Package CONTAINS %s\n", spdxFileRef)
	}

	return b.String()
}

func licenseConcluded(e FileEntry) string {
	if e.LicenseID != "" {
		return e.LicenseID
	}
	if e.FallbackLicense != "" {
		return e.FallbackLicense
	}
	return noAssertion
}

func fileRef(idx int, path string) string {
	base := filepath.Base(path)
	safe := nonSafe.ReplaceAllString(base, "-")
	return fmt.Sprintf("SPDXRef-File-%d-%s", idx, safe)
}
