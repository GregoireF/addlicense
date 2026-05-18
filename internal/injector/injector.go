// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package injector

import (
	"os"
	"strings"
)

const scanLines = 20

// HasHeader reports whether the file already contains a license header.
// It checks the first scanLines lines for SPDX identifiers or copyright notices.
func HasHeader(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	lines := strings.SplitN(string(data), "\n", scanLines+1)
	if len(lines) > scanLines {
		lines = lines[:scanLines]
	}
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "spdx-license-identifier") ||
			strings.Contains(lower, "copyright") {
			return true, nil
		}
	}
	return false, nil
}

// Inject prepends header to the file, respecting shebangs and file encoding.
// If the file already has a header (idempotent check), it returns without modifying.
func Inject(path, header string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(existing)

	// Preserve shebang on the first line.
	var shebang, rest string
	if strings.HasPrefix(content, "#!") {
		idx := strings.Index(content, "\n")
		if idx >= 0 {
			shebang = content[:idx+1]
			rest = content[idx+1:]
		} else {
			shebang = content
		}
	} else {
		rest = content
	}

	var out strings.Builder
	if shebang != "" {
		out.WriteString(shebang)
		out.WriteString("\n")
	}
	out.WriteString(header)
	out.WriteString("\n")
	out.WriteString(rest)

	return os.WriteFile(path, []byte(out.String()), 0o644)
}
