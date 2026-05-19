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
		Long: `addlicense adds, updates and checks license headers across your project.

Examples:
  addlicense --license MIT .
  addlicense --license MIT --author "Grégoire" .
  addlicense --check .
  addlicense --template ./header.txt .
  addlicense --ignore dist,vendor .`,
		Version:      version,
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Paths = args
			return runner.Run(opts)
		},
	}

	root.Flags().StringVarP(&opts.License, "license", "l", "", "SPDX license identifier (e.g. MIT, Apache-2.0)")
	root.Flags().StringVarP(&opts.Author, "author", "a", "", "Copyright holder name")
	root.Flags().StringVarP(&opts.Template, "template", "t", "", "Path to a custom header template file")
	root.Flags().StringSliceVarP(&opts.Ignore, "ignore", "i", nil, "Comma-separated list of patterns to ignore")
	root.Flags().IntVarP(&opts.Year, "year", "y", 0, "Copyright year (defaults to current year)")
	root.Flags().BoolVarP(&opts.CheckOnly, "check", "c", false, "Check headers without modifying files (exit 1 if missing)")

	root.SetVersionTemplate(fmt.Sprintf("addlicense %s\nbuild: %s @ %s\n", version, commit, date))

	return root
}
