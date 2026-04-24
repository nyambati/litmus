package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newHistoryCmd creates the history command.
func newHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Manage regression baseline versions",
		Long:  "List, view, and rollback to previous regression baseline versions.",
	}

	cmd.AddCommand(newHistoryListCmd())
	cmd.AddCommand(newHistoryRollbackCmd())

	return cmd
}

// newHistoryListCmd creates the history list subcommand.
func newHistoryListCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list",
		Short:        "List available baseline versions",
		Long:         "Display all baseline versions with the currently active version marked.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunHistoryList(cmd)
		},
	}
}

// newHistoryRollbackCmd creates the history rollback subcommand.
func newHistoryRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "rollback <version>",
		Short:        "Restore a previous baseline version",
		Long:         "Restore the active baseline to a previous numbered version from history.",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunHistoryRollback(cmd, args[0])
		},
	}
}
