// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package injector

import (
	"bufio"
	"os"
	"strings"
)

const scanLines = 20

// HasHeader reports whether the file already contains a license header.
// It checks the first scanLines lines for SPDX identifiers or copyright notices.
func HasHeader(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	n := 0
	for scanner.Scan() && n < scanLines {
		line := strings.ToLower(scanner.Text())
		if strings.Contains(line, "spdx-license-identifier") ||
			strings.Contains(line, "copyright") {
			return true, nil
		}
		n++
	}
	return false, scanner.Err()
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
