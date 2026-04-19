package cmd

import (
	"github.com/nyambati/litmus/internal/server"
	"github.com/spf13/cobra"
)

// newServeCmd creates the serve command.
func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Litmus web interface",
		Long:  "Starts a web server providing a visual interface for route exploration and test management",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			dev, _ := cmd.Flags().GetBool("dev")
			return server.RunUIServer(port, dev)
		},
	}

	cmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	cmd.Flags().Bool("dev", false, "Enable development mode")
	return cmd
}
