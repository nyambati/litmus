package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newSnapshotCmd creates the snapshot command.
func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Capture regression baseline",
		Long:  "Captures current alertmanager routing behavior as regression baseline. Use --update to accept changes.",
		// Default to capture when no subcommand is specified (backward compatible)
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if a subcommand was invoked
			if len(args) > 0 {
				return cmd.Usage()
			}
			update, _ := cmd.Flags().GetBool("update")
			strict, _ := cmd.Flags().GetBool("strict")
			return cli.RunSnapshot(update, strict)
		},
	}

	// Add flags for backward compatibility
	cmd.Flags().BoolP("update", "u", false, "Update baseline with current behavior")
	cmd.Flags().BoolP("strict", "s", false, "Fail and show diff if drift is detected (useful for CI)")
	return cmd
}
