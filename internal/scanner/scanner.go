// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

			if d.IsDir() {
				if shouldIgnore(name, ignore) {
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

			if shouldIgnore(name, ignore) || shouldIgnore(path, ignore) {
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

func shouldIgnore(name string, patterns []string) bool {
	for _, p := range patterns {
		matched, err := filepath.Match(p, name)
		if err == nil && matched {
			return true
		}
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
