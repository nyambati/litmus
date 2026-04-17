/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0-alpha"

var rootCmd = &cobra.Command{
	Use:     "litmus",
	Short:   "Litmus - Alertmanager Validator",
	Long:    "Litmus validates Alertmanager configurations through regression and behavioral testing",
	Version: version,
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
	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newInspectCmd())
}
