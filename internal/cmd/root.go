// SPDX-License-Identifier: MIT
// Copyright 2026 Grégoire Favreau

package cmd

import (
	"fmt"

	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the root Cobra command wired up with all flags and version info.
func NewRootCmd(version, commit, date string) *cobra.Command {
	var opts config.Options

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
  addlicense --ignore dist,vendor .`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Paths = args
			return runner.Run(opts)
		},
	}

	// Core
	root.Flags().StringVarP(&opts.License, "license", "l", "", "SPDX license identifier (e.g. MIT, Apache-2.0)")
	root.Flags().StringVarP(&opts.Author, "author", "a", "", "Copyright holder name (comma-separated for multiple: \"Alice, Bob\")")
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

	root.SetVersionTemplate(fmt.Sprintf("addlicense %s\nbuild: %s @ %s\n", version, commit, date))

	return root
}
