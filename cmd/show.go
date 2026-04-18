package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newShowCmd creates the show command (stub for future implementation).
func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Visualize routing path for labels (not yet implemented)",
		Long:  "Visualize the routing path that a given set of labels would take through the Alertmanager configuration tree.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("show command not yet implemented - see docs/backlog.md")
		},
	}
}
