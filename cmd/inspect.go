package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
			return runInspect(args[0], format)
		},
	}

	cmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json")
	return cmd
}

func runInspect(filePath string, format string) error {
	// Load baseline file
	tests, err := loadBaseline(filePath)
	if err != nil {
		return fmt.Errorf("loading baseline: %w", err)
	}

	// Output in requested format
	if format == "json" {
		data, _ := json.MarshalIndent(tests, "", "  ")
		fmt.Println(string(data))
	} else {
		data, _ := yaml.Marshal(tests)
		fmt.Println(string(data))
	}

	return nil
}
