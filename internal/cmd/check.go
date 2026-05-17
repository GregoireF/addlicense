package cmd

import (
	"github.com/GregoireF/addlicense/internal/config"
	"github.com/GregoireF/addlicense/internal/runner"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	var opts config.Options

	cmd := &cobra.Command{
		Use:   "check [flags] [path...]",
		Short: "Check that all files have license headers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Paths = args
			opts.CheckOnly = true
			return runner.Run(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.License, "license", "l", "MIT", "SPDX license identifier")
	cmd.Flags().StringSliceVarP(&opts.Ignore, "ignore", "i", config.DefaultIgnore, "Comma-separated list of patterns to ignore")

	return cmd
}
