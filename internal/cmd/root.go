package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "addlicense",
		Short: "Fast, minimal license header manager",
		Long: `addlicense adds, updates and checks license headers across your project.

Examples:
  addlicense --license MIT .
  addlicense --license MIT --author "Grégoire" .
  addlicense --check .
  addlicense --template ./header.txt .`,
		SilenceUsage: true,
	}

	root.AddCommand(newAddCmd())
	root.AddCommand(newCheckCmd())

	return root
}
