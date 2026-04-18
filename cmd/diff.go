package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newDiffCmd creates the diff command.
func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "diff",
		Short:        "Show behavioral changes compared to baseline",
		Long:         "Performs a structural comparison between the current configuration and the saved regression baseline.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunDiff()
		},
	}
}
