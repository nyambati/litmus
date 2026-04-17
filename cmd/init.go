package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newInitCmd creates the init command.
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "init",
		Short:        "Initialize a new litmus workspace",
		Long:         "Creates .litmus.yaml, tests/ directory, and .gitattributes for a new workspace",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunInit()
		},
	}
}
