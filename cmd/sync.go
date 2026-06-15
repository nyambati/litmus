package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newSyncCmd creates the sync command.
func newSyncCmd() *cobra.Command {
	opts := cli.SyncOptions{}

	cmd := &cobra.Command{
		Use:          "sync",
		Short:        "Sync validated config to Grafana Mimir",
		Long:         "Validates the alertmanager configuration and pushes it to Grafana Mimir's /api/v1/alerts endpoint.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunSync(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Address, "address", "", "Mimir API address (overrides LITMUS_MIMIR_ADDRESS)")
	cmd.Flags().StringVar(&opts.TenantID, "tenant-id", "", "Mimir tenant ID (overrides LITMUS_MIMIR_TENANT_ID)")
	cmd.Flags().StringVar(&opts.APIKey, "api-key", "", "Mimir API key (overrides LITMUS_MIMIR_API_KEY)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Render config to stdout or file without syncing to Mimir")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (use with --dry-run)")

	return cmd
}
