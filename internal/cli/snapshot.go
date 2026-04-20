package cli

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	"gopkg.in/yaml.v3"
)

// RunSnapshot captures current routing behavior as a regression baseline.
// If update is false and a baseline exists, drift is checked.
// In strict mode, drift causes an error and prints a diff.
// Otherwise, drift only prints a warning and does not block the snapshot creation.
func RunSnapshot(update, strict bool) error {
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	alertConfigPath := filepath.Join(litmusConfig.Config.Directory, litmusConfig.Config.File)
	alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
	if err != nil {
		return fmt.Errorf("loading alertmanager config: %w", err)
	}

	ctx := context.Background()

	router := pipeline.NewRouter(alertConfig.Route)
	runner := pipeline.NewRunner(stores.NewSilenceStore(nil), stores.NewAlertStore(), router, nil)

	walker := snapshot.NewRouteWalker(alertConfig.Route)
	paths := walker.FindTerminalPaths()

	synthesizer := snapshot.NewSnapshotSynthesizer(runner)
	outcomes, err := synthesizer.DiscoverOutcomes(ctx, paths)
	if err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	if len(outcomes) == 0 {
		fmt.Fprintf(os.Stderr, "WARN: synthesis produced zero outcomes; baseline will be empty\n")
	}

	regTests := BuildRegressionTests(outcomes, litmusConfig.GlobalLabels)

	baselinePath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.mpk")

	existing, err := LoadBaseline(baselinePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading existing baseline: %w", err)
	}

	if existing != nil {
		d := snapshot.ComputeDiff(existing, regTests)
		if len(d.Deltas) > 0 {
			if strict {
				PrintDiffReport(d)
				return fmt.Errorf("drift detected in routing behavior")
			}
			if !update {
				fmt.Fprintf(os.Stderr, "WARN: drift detected in routing behavior; run with --update to accept changes, or 'litmus diff' to inspect\n")
			}
		}
	}

	if err := os.MkdirAll(litmusConfig.Regression.Directory, 0755); err != nil {
		return fmt.Errorf("creating regression directory: %w", err)
	}

	// Protect mpk baseline: only write if it doesn't exist OR update is explicit
	if existing == nil || update {
		mpkFile, err := os.Create(baselinePath)
		if err != nil {
			return fmt.Errorf("creating baseline file: %w", err)
		}
		defer mpkFile.Close()

		if err := codec.EncodeMsgPack(mpkFile, regTests); err != nil {
			return fmt.Errorf("writing baseline: %w", err)
		}
	}

	// Mirror is not protected; always reflects current state
	ymlPath := strings.Replace(baselinePath, "mpk", "yml", 1)
	ymlData, err := yaml.Marshal(regTests)
	if err != nil {
		return fmt.Errorf("marshaling YAML mirror: %w", err)
	}

	if err := os.WriteFile(ymlPath, ymlData, 0600); err != nil {
		return fmt.Errorf("writing YAML mirror: %w", err)
	}

	fmt.Println("✓ Snapshot processed successffuly") //nolint:forbidigo

	return nil
}

// BuildRegressionTests converts synthesis outcomes into executable regression test cases.
func BuildRegressionTests(outcomes []*snapshot.SynthesisResult, globalLabels map[string]string) []*types.TestCase {
	tests := make([]*types.TestCase, 0, len(outcomes))
	for _, outcome := range outcomes {
		labels := make(map[string]string)
		maps.Copy(labels, globalLabels)
		maps.Copy(labels, outcome.Labels)
		tests = append(tests, &types.TestCase{
			Type:     "regression",
			Name:     fmt.Sprintf("Route to %s", strings.Join(outcome.Receivers, ", ")),
			Labels:   []map[string]string{labels},
			Expected: outcome.Receivers,
			Tags:     []string{"regression"},
		})
	}
	return tests
}
