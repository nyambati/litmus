/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"

	"github.com/nyambati/litmus/internal/config"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "litmus",
	Short:   "Litmus - Alertmanager Validator",
	Long:    "Litmus validates Alertmanager configurations through regression and behavioral testing",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.New()
		if err != nil {
			return err
		}
		ctx := context.WithValue(cmd.Context(), config.ConfigKey{}, cfg)
		cmd.SetContext(ctx)
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("litmus version {{.Version}}\n")
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newSnapshotCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newDiffCmd())
	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newInspectCmd())
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newServeCmd())
}
