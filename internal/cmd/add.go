package cmd

import (
	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var opts config.Options

	cmd := &cobra.Command{
		Use:   "add [flags] [path...]",
		Short: "Add license headers to files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Paths = args
			return runner.Run(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.License, "license", "l", "MIT", "SPDX license identifier (e.g. MIT, Apache-2.0)")
	cmd.Flags().StringVarP(&opts.Author, "author", "a", "", "Copyright holder name")
	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Path to a custom header template file")
	cmd.Flags().StringSliceVarP(&opts.Ignore, "ignore", "i", config.DefaultIgnore, "Comma-separated list of patterns to ignore")
	cmd.Flags().IntVarP(&opts.Year, "year", "y", 0, "Copyright year (defaults to current year)")

	return cmd
}
