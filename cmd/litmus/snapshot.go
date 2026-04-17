package main

import (
	"fmt"
	"os"

	"github.com/prometheus/alertmanager/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"litmus/internal/codec"
	"litmus/internal/engine/snapshot"
	"litmus/internal/types"
)

// LitmusConfig represents the litmus.yaml configuration.
type LitmusConfig struct {
	ConfigFile   string            `yaml:"config_file"`
	GlobalLabels map[string]string `yaml:"global_labels"`
	Regression   RegressionConfig  `yaml:"regression"`
	Tests        TestsConfig       `yaml:"tests"`
}

// RegressionConfig represents regression settings.
type RegressionConfig struct {
	MaxSamples   int    `yaml:"max_samples"`
	BaselinePath string `yaml:"baseline_path"`
}

// TestsConfig represents test settings.
type TestsConfig struct {
	Directory string `yaml:"directory"`
}

// newSnapshotCmd creates the snapshot command.
func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Generate regression baseline from route tree",
		Long:  "Captures current alertmanager routing behavior as regression baseline. Use --update to accept changes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			update, _ := cmd.Flags().GetBool("update")
			return runSnapshot(update)
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolP("update", "u", false, "Update baseline with current behavior")
	return cmd
}

func runSnapshot(update bool) error {
	// Load litmus.yaml
	litmusConfig, err := loadLitmusConfig("litmus.yaml")
	if err != nil {
		return fmt.Errorf("loading litmus.yaml: %w", err)
	}

	// Load alertmanager config
	alertConfig, err := loadAlertmanagerConfig(litmusConfig.ConfigFile)
	if err != nil {
		return fmt.Errorf("loading alertmanager config: %w", err)
	}

	// Discover routes
	walker := snapshot.NewRouteWalker(alertConfig.Route)
	paths := walker.FindTerminalPaths()

	// Generate placeholder regression tests from routes
	regTests := make([]*types.RegressionTest, 0)
	for _, path := range paths {
		labels := litmusConfig.GlobalLabels
		if labels == nil {
			labels = make(map[string]string)
		}

		regTests = append(regTests, &types.RegressionTest{
			Name:     fmt.Sprintf("Route to %s", path.Receiver),
			Labels:   []map[string]string{labels},
			Expected: []string{path.Receiver},
			Tags:     []string{"regression"},
		})
	}

	// Check for drift if baseline exists and not updating
	if !update {
		existing, err := loadBaseline(litmusConfig.Regression.BaselinePath)
		if err == nil && existing != nil {
			if hasChanges(existing, regTests) {
				return fmt.Errorf("drift detected in routing behavior: use --update to accept changes")
			}
		}
	}

	// Write baseline files
	mpkFile, err := os.Create(litmusConfig.Regression.BaselinePath)
	if err != nil {
		return fmt.Errorf("creating baseline file: %w", err)
	}
	defer mpkFile.Close()

	if err := codec.EncodeMsgPack(mpkFile, regTests); err != nil {
		return fmt.Errorf("writing baseline: %w", err)
	}

	// Write YAML mirror
	ymlPath := litmusConfig.Regression.BaselinePath + ".yml"
	ymlData, _ := yaml.Marshal(regTests)
	if err := os.WriteFile(ymlPath, ymlData, 0644); err != nil {
		return fmt.Errorf("writing YAML mirror: %w", err)
	}

	fmt.Printf("✓ Generated baseline: %s\n", litmusConfig.Regression.BaselinePath)
	fmt.Printf("✓ YAML mirror: %s\n", ymlPath)
	return nil
}

func loadLitmusConfig(path string) (*LitmusConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg LitmusConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadAlertmanagerConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadBaseline(path string) ([]*types.RegressionTest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tests []*types.RegressionTest
	if err := codec.DecodeMsgPack(file, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}

func hasChanges(existing, current []*types.RegressionTest) bool {
	if len(existing) != len(current) {
		return true
	}

	// Simple comparison: just check if any test names changed
	for _, t := range current {
		found := false
		for _, e := range existing {
			if t.Name == e.Name {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}
