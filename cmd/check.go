package cmd

import (
	"os"

	"github.com/nyambati/litmus/internal/cli"
	"github.com/spf13/cobra"
)

// newCheckCmd creates the check command.
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "check",
		Short:        "Validate alertmanager configuration",
		Long:         "Runs sanity linter, regression tests, and behavioral unit tests",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			diff, _ := cmd.Flags().GetBool("diff")
			tags, _ := cmd.Flags().GetStringSlice("tags")
			code, err := cli.RunCheck(format, diff, tags)
			if err != nil {
				return err
			}
			if code != 0 {
				os.Exit(int(code))
			}
			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format: text or json")
	cmd.Flags().BoolP("diff", "d", false, "Show detailed behavioral delta for regression failures")
	cmd.Flags().StringSliceP("tags", "t", nil, "run only tests matching these tags (comma-separated)")
	return cmd
}
