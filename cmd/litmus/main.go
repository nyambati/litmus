package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0-alpha"

func main() {
	rootCmd := &cobra.Command{
		Use:   "litmus",
		Short: "Litmus - Alertmanager Validator",
		Long:  "Litmus validates Alertmanager configurations through regression and behavioral testing",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Litmus Alertmanager Validator v%s\n", version)
		},
	}

	rootCmd.AddCommand(newInitCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
