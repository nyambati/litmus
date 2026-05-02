package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/nyambati/litmus/internal/config"
	"github.com/spf13/cobra"
)

// newSnapshotCmd creates the snapshot command group.
func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "snapshot",
		Short:        "Manage regression baseline",
		Long:         "Manage alertmanager routing regression baseline.",
		SilenceUsage: true,
	}

	cmd.AddCommand(newSnapshotCaptureCmd())
	cmd.AddCommand(newSnapshotUpdateCmd())
	return cmd
}

// newSnapshotCaptureCmd creates the snapshot capture subcommand.
func newSnapshotCaptureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "capture",
		Short:        "Capture current routing behavior as baseline",
		Long:         "Captures current alertmanager routing behavior as regression baseline.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.FromContext(cmd.Context())
			strict, _ := cmd.Flags().GetBool("strict")
			return cli.RunSnapshot(cfg, false, strict)
		},
	}

	cmd.Flags().BoolP("strict", "s", false, "Fail and show diff if drift is detected (useful for CI)")
	return cmd
}

// newSnapshotUpdateCmd creates the snapshot update subcommand.
func newSnapshotUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update",
		Short:        "Update baseline with current routing behavior",
		Long:         "Updates baseline with current alertmanager routing behavior.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.FromContext(cmd.Context())
			strict, _ := cmd.Flags().GetBool("strict")
			return cli.RunSnapshot(cfg, true, strict)
		},
	}

	cmd.Flags().BoolP("strict", "s", false, "Fail and show diff if drift is detected (useful for CI)")
	return cmd
}
