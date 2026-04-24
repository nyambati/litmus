package cli

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
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

	// Load existing baseline from regressions.litmus.yml
	var existing []*types.TestCase
	ymlPath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.yml")
	state, err := LoadRegressionState(ymlPath)
	if err == nil && state != nil {
		existing = state.Tests
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading existing baseline: %w", err)
	}

	hasDrift := false
	if existing != nil {
		d := snapshot.ComputeDiff(existing, regTests)
		if len(d.Deltas) > 0 {
			hasDrift = true
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

	// Decide whether to update baseline
	shouldUpdate := existing == nil || (update && hasDrift)
	if update && existing != nil && !hasDrift {
		fmt.Println("✓ No changes detected; baseline is up to date") //nolint:forbidigo
		return nil
	}

	// Archive baseline to history only on actual updates (new baseline or drift+update)
	// This creates a new timestamped MPK and updates the ID
	if shouldUpdate {
		if _, err := ArchiveBaseline(litmusConfig, regTests); err != nil {
			return fmt.Errorf("archiving baseline to history: %w", err)
		}
	} else {
		// Not creating new MPK, but update tests in state file with existing ID
		ymlPath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.yml")
		state, err := LoadRegressionState(ymlPath)
		if err == nil {
			// Keep existing ID, update tests only
			state.Tests = regTests
			if err := SaveRegressionState(ymlPath, state); err != nil {
				return fmt.Errorf("updating tests in state: %w", err)
			}
		}
	}

	fmt.Println("✓ Snapshot processed successfully") //nolint:forbidigo

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
			Type:   "regression",
			Name:   fmt.Sprintf("Route to %s", strings.Join(outcome.Receivers, ", ")),
			Labels: []map[string]string{labels},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: outcome.Receivers},
			Tags:   []string{"regression"},
		})
	}
	return tests
}
