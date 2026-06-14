package cmd

import (
	"os"

	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newCheckCmd creates the check command.
func newCheckCmd() *cobra.Command {
	opts := &cli.CheckOptions{}

	cmd := &cobra.Command{
		Use:          "check",
		Short:        "Validate alertmanager configuration",
		Long:         "Runs sanity linter, regression tests, and behavioral unit tests",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := cli.RunCheck(cmd.Context(), opts)
			if err != nil {
				return err
			}
			if code != 0 {
				os.Exit(int(code))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", "text", "Output format: text or json")
	cmd.Flags().BoolVarP(&opts.Diff, "diff", "d", false, "Show detailed behavioral delta for regression failures")
	cmd.Flags().StringSliceVarP(&opts.Tags, "tags", "t", nil, "run only tests matching these tags (comma-separated)")
	return cmd
}
