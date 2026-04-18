package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newSnapshotCmd creates the snapshot command.
func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "snapshot",
		Short:        "Generate regression baseline from route tree",
		Long:         "Captures current alertmanager routing behavior as regression baseline. Use --update to accept changes.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			update, _ := cmd.Flags().GetBool("update")
			diff, _ := cmd.Flags().GetBool("diff")
			return cli.RunSnapshot(update, diff)
		},
	}

	cmd.Flags().BoolP("update", "u", false, "Update baseline with current behavior")
	cmd.Flags().BoolP("diff", "d", false, "Show diff between existing baseline and current routing")
	return cmd
}
