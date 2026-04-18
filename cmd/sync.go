package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newSyncCmd creates the sync command.
func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "sync",
		Short:        "Sync validated config to Grafana Mimir",
		Long:         "Validates the alertmanager configuration and pushes it to Grafana Mimir's /api/v1/alerts endpoint.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			address, _ := cmd.Flags().GetString("address")
			tenantID, _ := cmd.Flags().GetString("tenant-id")
			apiKey, _ := cmd.Flags().GetString("api-key")
			skipValidate, _ := cmd.Flags().GetBool("skip-validate")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			return cli.RunSync(address, tenantID, apiKey, skipValidate, dryRun)
		},
	}

	cmd.Flags().String("address", "", "Mimir API address (overrides LITMUS_MIMIR_ADDRESS)")
	cmd.Flags().String("tenant-id", "", "Mimir tenant ID (overrides LITMUS_MIMIR_TENANT_ID)")
	cmd.Flags().String("api-key", "", "Mimir API key (overrides LITMUS_MIMIR_API_KEY)")
	cmd.Flags().Bool("skip-validate", false, "Skip sanity checks before push")
	cmd.Flags().Bool("dry-run", false, "Validate only, do not push")

	return cmd
}
