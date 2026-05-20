// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package injector

import (
	"os"
	"strings"
)

const (
	scanLines    = 20
	spdxKey      = "spdx-license-identifier"
	copyrightKey = "copyright"
)

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
		if strings.Contains(lower, spdxKey) || strings.Contains(lower, copyrightKey) {
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

// Remove strips the license header from path using the provided comment style.
// lineComment is the line-comment prefix (e.g. "//"), or empty for block-comment languages.
// blockOpen/blockClose delimit block comments (e.g. "<!--", "-->").
// The header is only removed when at least one of its comment lines contains
// "spdx-license-identifier" or "copyright" (case-insensitive).
// Returns true if the file was modified.
func Remove(path, lineComment, blockOpen, blockClose string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	result, changed := stripHeader(string(data), lineComment, blockOpen, blockClose)
	if !changed {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(result), 0o644)
}

// stripHeader returns content with the leading license header removed, and
// whether a change was made. It handles both line-comment and block-comment styles.
func stripHeader(content, lineComment, blockOpen, blockClose string) (string, bool) {
	lines := strings.Split(content, "\n")

	// Skip shebang line if present.
	start := 0
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		start = 1
	}

	var end int
	var found bool

	if lineComment != "" {
		end, found = findLineCommentHeader(lines, start, lineComment)
	} else if blockOpen != "" {
		end, found = findBlockCommentHeader(lines, start, blockOpen, blockClose)
	}

	if !found {
		return content, false
	}

	// Strip any blank lines immediately following the removed header block
	// (Inject always appends a blank line after the header).
	for end < len(lines) && strings.TrimSpace(lines[end]) == "" {
		end++
	}

	result := lines[:start:start]
	result = append(result, lines[end:]...)
	return strings.Join(result, "\n"), true
}

// findLineCommentHeader collects consecutive comment lines from lines[start:],
// and returns (end index, true) when at least one line contains SPDX/copyright.
func findLineCommentHeader(lines []string, start int, prefix string) (int, bool) {
	trimmedPrefix := strings.TrimSpace(prefix)
	end := start
	for end < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[end]), trimmedPrefix) {
		end++
	}
	if end == start || !isLicenseBlock(lines[start:end]) {
		return 0, false
	}
	return end, true
}

// findBlockCommentHeader looks for a block comment starting at lines[start]
// and returns (end index, true) when the block contains SPDX/copyright.
func findBlockCommentHeader(lines []string, start int, blockOpen, blockClose string) (int, bool) {
	if start >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[start]), blockOpen) {
		return 0, false
	}
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], blockClose) {
			if isLicenseBlock(lines[start : i+1]) {
				return i + 1, true
			}
			return 0, false
		}
	}
	return 0, false
}

// isLicenseBlock reports whether any line in the slice contains a license marker.
func isLicenseBlock(lines []string) bool {
	for _, l := range lines {
		lower := strings.ToLower(l)
		if strings.Contains(lower, spdxKey) || strings.Contains(lower, copyrightKey) {
			return true
		}
	}
	return false
}
