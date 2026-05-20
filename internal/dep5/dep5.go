// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

// Package dep5 generates REUSE-compliant .reuse/dep5 bulk-licence declarations
// for files that cannot carry inline SPDX headers (images, fonts, binaries).
package dep5

import (
	"fmt"
	"strings"
)

// Build returns the content of a .reuse/dep5 document.
// paths must be Unix-style relative paths from the repository root.
// If paths is empty, only the mandatory Format header paragraph is emitted —
// the document is still valid but declares no additional files.
func Build(paths []string, year int, author, license string) string {
	var b strings.Builder
	b.WriteString("Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/\n")
	if author != "" {
		b.WriteString("Upstream-Contact: " + author + "\n")
	}

	if len(paths) == 0 {
		return b.String()
	}

	b.WriteString("\n")
	for i, p := range paths {
		if i == 0 {
			fmt.Fprintf(&b, "Files: %s\n", p)
		} else {
			fmt.Fprintf(&b, "       %s\n", p)
		}
	}
	if author != "" {
		fmt.Fprintf(&b, "Copyright: %d %s\n", year, author)
	} else {
		fmt.Fprintf(&b, "Copyright: %d\n", year)
	}
	fmt.Fprintf(&b, "License: %s\n", license)

	return b.String()
}
