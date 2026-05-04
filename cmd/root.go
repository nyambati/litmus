/*
Copyright © 2026 Thomas Nyambati <thomasnyambati@gmail.com>
*/
package cmd

import (
	"context"
	"io"
	"os"

	"github.com/nyambati/litmus/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "litmus",
	Short: "Litmus - Alertmanager Validator",
	Long:  "Litmus validates Alertmanager configurations through regression and behavioral testing",
	// Version:      version,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.New()
		if err != nil {
			return err
		}

		log := logrus.New()
		log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			log.SetLevel(logrus.DebugLevel)
		} else {
			log.SetOutput(io.Discard)
		}

		ctx := context.WithValue(cmd.Context(), config.ConfigKey{}, cfg)
		ctx = context.WithValue(ctx, config.LoggerKey{}, logrus.FieldLogger(log))
		cmd.SetContext(ctx)
		return nil
	},
}

// SetVersion injects the build-time version into the root command.
func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("litmus version {{.Version}}\n")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose (debug) logging")
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newSnapshotCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newDiffCmd())
	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newInspectCmd())
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newServeCmd())
}
