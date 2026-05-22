// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the root Cobra command wired up with all flags and version info.
func NewRootCmd(version, commit, date string) *cobra.Command {
	var opts config.Options
	var authorFile string

	root := &cobra.Command{
		Use:   "addlicense [flags] [path...]",
		Short: "Fast, minimal license header manager",
		Long: `addlicense adds, updates, removes and checks license headers across your project.

Examples:
  addlicense --license MIT .
  addlicense --license MIT --author "Grégoire" .
  addlicense --license MIT --author "Alice, Bob" .
  addlicense --check .
  addlicense --remove .
  addlicense --update --license EUPL-1.2 .
  addlicense --update --year-range .
  addlicense --diff .
  addlicense --dry-run .
  addlicense --dep5 --license MIT --author "Acme" .
  addlicense --check --format json .
  addlicense --template ./header.txt .
  addlicense --ignore dist,vendor .
  addlicense --sbom sbom.spdx .
  addlicense --sbom - . | head`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if authorFile != "" {
				if opts.Author != "" {
					return fmt.Errorf("--author and --author-file are mutually exclusive")
				}
				data, err := os.ReadFile(authorFile)
				if err != nil {
					return fmt.Errorf("--author-file: %w", err)
				}
				authors := parseAuthorFile(string(data))
				if len(authors) == 0 {
					return fmt.Errorf("--author-file: no authors found in %s", authorFile)
				}
				opts.Author = strings.Join(authors, ", ")
			}
			opts.Paths = args
			return runner.Run(opts)
		},
	}

	// Core
	root.Flags().StringVarP(&opts.License, "license", "l", "", "SPDX license identifier (e.g. MIT, Apache-2.0)")
	root.Flags().StringVarP(&opts.Author, "author", "a", "", "Copyright holder name (comma-separated for multiple: \"Alice, Bob\")")
	root.Flags().StringVar(&authorFile, "author-file", "", "Path to a file listing copyright holders, one per line (mutually exclusive with --author)")
	root.Flags().StringVarP(&opts.Template, "template", "t", "", "Path to a custom header template file")
	root.Flags().StringSliceVarP(&opts.Ignore, "ignore", "i", nil, "Comma-separated list of patterns to ignore")
	root.Flags().IntVarP(&opts.Year, "year", "y", 0, "Copyright year (defaults to current year)")

	// Modes
	root.Flags().BoolVarP(&opts.CheckOnly, "check", "c", false, "Check headers without modifying files (exit 1 if any are missing)")
	root.Flags().BoolVarP(&opts.Remove, "remove", "R", false, "Strip existing license headers")
	root.Flags().BoolVarP(&opts.Update, "update", "u", false, "Replace existing headers with the new one (remove + inject)")
	root.Flags().BoolVarP(&opts.DryRun, "dry-run", "n", false, "Preview changes without writing to disk")
	root.Flags().BoolVar(&opts.Diff, "diff", false, "Emit JSON Lines with the rendered header for each file that would change (no writes; exit 1 if any file would be modified)")

	// Compliance
	root.Flags().BoolVarP(&opts.Reuse, "reuse", "r", false, "Emit SPDX-FileCopyrightText: instead of Copyright (REUSE/FSFE compliance)")

	// Output
	root.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Print every file, including already-licensed ones")
	root.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress all stdout; errors still go to stderr")
	root.Flags().StringVarP(&opts.Format, "format", "f", "text", "Output format: text or json (JSON Lines)")

	// Performance
	root.Flags().IntVar(&opts.Workers, "workers", 0, "Parallel workers (default: number of CPUs)")

	// Header management
	root.Flags().BoolVar(&opts.YearRange, "year-range", false, "Preserve original copyright year when updating (emits \"YYYY-YYYY\" range)")
	root.Flags().BoolVar(&opts.Dep5, "dep5", false, "Generate .reuse/dep5 for files that cannot carry inline headers (REUSE compliance)")

	// SBOM / compliance
	root.Flags().StringVar(&opts.Sbom, "sbom", "", "Generate a SPDX 2.3 tag-value document at this path (\"-\" for stdout). Mutually exclusive with --check/--remove/--update/--diff/--dep5/--dry-run")

	root.SetVersionTemplate(fmt.Sprintf("addlicense %s\nbuild: %s @ %s\n", version, commit, date))

	return root
}

// parseAuthorFile reads copyright holders from file content, one per line.
// Lines starting with '#' and blank lines are ignored.
func parseAuthorFile(content string) []string {
	lines := strings.Split(content, "\n")
	authors := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		authors = append(authors, line)
	}
	return authors
}
