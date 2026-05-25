// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// File represents a source file found by the scanner.
type File struct {
	Path string
	Ext  string
}

// Walk recursively collects files under roots, skipping ignored patterns.
func Walk(roots []string, ignore []string) ([]File, error) {
	var files []File

	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			name := d.Name()
			relPath, _ := filepath.Rel(root, path)

			if d.IsDir() {
				if shouldIgnore(name, relPath, ignore) {
					return filepath.SkipDir
				}
				return nil
			}

			if isSymlink(path) {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(name))
			if ext == "" {
				return nil
			}

			if shouldIgnore(name, relPath, ignore) {
				return nil
			}

			files = append(files, File{Path: path, Ext: ext})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

func shouldIgnore(name, relPath string, patterns []string) bool {
	rel := filepath.ToSlash(relPath)
	for _, p := range patterns {
		pat := filepath.ToSlash(p)
		// Match relative path — handles ** doublestar patterns across directories.
		if matched, _ := doublestar.Match(pat, rel); matched {
			return true
		}
		// Match basename — *.pb.go applies at any depth without a path prefix.
		if matched, _ := doublestar.Match(pat, name); matched {
			return true
		}
		// Substring match for plain non-glob patterns (e.g. "generated" → "auto_generated.go").
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}
