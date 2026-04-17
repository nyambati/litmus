package main

import (
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0-alpha"

func main() {
	rootCmd := &cobra.Command{
		Use:     "litmus",
		Short:   "Litmus - Alertmanager Validator",
		Long:    "Litmus validates Alertmanager configurations through regression and behavioral testing",
		Version: version,
	}

	rootCmd.SetVersionTemplate("litmus version {{.Version}}\n")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newSnapshotCmd())
	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newInspectCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
