package cmd

import (
	"context"

	"github.com/nyambati/litmus/internal/config"
	"github.com/spf13/cobra"
)

// withConfig injects a PersistentPreRunE onto cmd that loads litmus config
// into context, mirroring what rootCmd does in production. All child
// subcommands inherit the hook, so tests that call newSnapshotCmd() or
// newHistoryCmd() directly still receive a non-nil config.
func withConfig(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentPreRunE = func(c *cobra.Command, _ []string) error {
		cfg, err := config.New()
		if err != nil {
			return err
		}
		c.SetContext(context.WithValue(c.Context(), config.ConfigKey{}, cfg))
		return nil
	}
	return cmd
}
