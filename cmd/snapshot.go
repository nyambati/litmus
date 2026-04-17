package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

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
	// Load configs
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	alertConfigPath := filepath.Join(litmusConfig.Config.Directory, litmusConfig.Config.File)
	alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
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

	baselinePath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.mpk")

	// Check for drift if baseline exists and not updating
	if !update {
		existing, err := loadBaseline(baselinePath)
		if err == nil && existing != nil {
			if hasChanges(existing, regTests) {
				return fmt.Errorf("drift detected in routing behavior: use --update to accept changes")
			}
		}
	}

	// Ensure regression directory exists
	if err := os.MkdirAll(litmusConfig.Regression.Directory, 0755); err != nil {
		return fmt.Errorf("creating regression directory: %w", err)
	}

	// Write baseline files
	mpkFile, err := os.Create(baselinePath)
	if err != nil {
		return fmt.Errorf("creating baseline file: %w", err)
	}
	defer mpkFile.Close()

	if err := codec.EncodeMsgPack(mpkFile, regTests); err != nil {
		return fmt.Errorf("writing baseline: %w", err)
	}

	// Write YAML mirror
	ymlPath := strings.Replace(baselinePath, "mpk", "yml", 1)
	ymlData, _ := yaml.Marshal(regTests)
	if err := os.WriteFile(ymlPath, ymlData, 0644); err != nil {
		return fmt.Errorf("writing YAML mirror: %w", err)
	}

	fmt.Printf("✓ Generated baseline: %s\n", baselinePath)
	fmt.Printf("✓ YAML mirror: %s\n", ymlPath)
	return nil
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
