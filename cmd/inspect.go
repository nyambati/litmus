package cmd

import (
	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newInspectCmd creates the inspect command.
func newInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "inspect <file.mpk>",
		Short:        "Inspect binary regression baseline",
		Long:         "Loads a MessagePack baseline and displays it as YAML or JSON",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			return cli.RunInspect(args[0], format)
		},
	}

	cmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json")
	return cmd
}
