// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultIgnore is the set of paths excluded when no --ignore flag is provided.
var DefaultIgnore = []string{
	"vendor", "node_modules", ".git", "dist", "build", "*.pb.go", "*.gen.go",
}

// Options holds the resolved configuration for a single addlicense run.
type Options struct {
	License   string
	Author    string
	Template  string
	Ignore    []string
	Year      int
	Paths     []string
	CheckOnly bool
	Reuse     bool   // emit SPDX-FileCopyrightText: instead of Copyright
	Remove    bool   // strip existing license headers
	Update    bool   // remove existing header then inject new one
	DryRun    bool   // preview changes without writing
	Verbose   bool   // print skipped/ok files in addition to modified ones
	Quiet     bool   // suppress all stdout; errors still go to stderr
	Format    string // output format: "text" (default) or "json"
	Workers   int    // parallel workers; 0 = runtime.NumCPU()
	YearRange bool   // emit "YYYY-YYYY" copyright range when updating files with an existing year
	Dep5      bool   // generate .reuse/dep5 for files without inline header support
	Diff      bool   // emit JSON Lines with the rendered header for each file that would change (no writes)
	Sbom      string // output path for SPDX 2.3 document ("-" = stdout); empty = disabled
}

// fileConfig is the shape of .addlicenserc.yaml / .addlicenserc.json.
type fileConfig struct {
	License string   `yaml:"license"`
	Author  string   `yaml:"author"`
	Ignore  []string `yaml:"ignore"`
	Year    int      `yaml:"year"`
	Reuse   bool     `yaml:"reuse"`
}

// Load merges file config into opts; CLI flags already set on opts take precedence.
// Search order: CWD, then first element of opts.Paths (if it differs from CWD).
func Load(opts *Options) error {
	searchDirs := []string{"."}
	if len(opts.Paths) > 0 {
		abs, err := filepath.Abs(opts.Paths[0])
		if err == nil {
			cwd, _ := os.Getwd()
			if abs != cwd {
				searchDirs = append(searchDirs, abs)
			}
		}
	}

	for _, dir := range searchDirs {
		fc, err := findAndParse(dir)
		if err != nil {
			return err
		}
		if fc == nil {
			continue
		}
		merge(opts, fc)
		return nil
	}
	return nil
}

func merge(opts *Options, fc *fileConfig) {
	if opts.License == "" && fc.License != "" {
		opts.License = fc.License
	}
	if opts.Author == "" && fc.Author != "" {
		opts.Author = fc.Author
	}
	if len(opts.Ignore) == 0 && len(fc.Ignore) > 0 {
		opts.Ignore = fc.Ignore
	}
	if opts.Year == 0 && fc.Year != 0 {
		opts.Year = fc.Year
	}
	if !opts.Reuse && fc.Reuse {
		opts.Reuse = true
	}
}

// Normalize fills in zero-value fields with their defaults.
func (o *Options) Normalize() {
	if o.Year == 0 {
		o.Year = time.Now().Year()
	}
	if o.License == "" {
		o.License = "MIT"
	}
	if len(o.Ignore) == 0 {
		o.Ignore = DefaultIgnore
	}
}

var candidates = []string{
	".addlicenserc.yaml",
	".addlicenserc.yml",
	".addlicenserc.json",
	"addlicense.json",
}

func findAndParse(dir string) (*fileConfig, error) {
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		var fc fileConfig
		if err := yaml.Unmarshal(data, &fc); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		return &fc, nil
	}
	return nil, nil
}
