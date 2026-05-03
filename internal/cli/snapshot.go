package cli

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	"github.com/nyambati/litmus/internal/workspace"
	"github.com/sirupsen/logrus"
)

// RunSnapshot captures current routing behavior as a regression baseline.
// If update is false and a baseline exists, drift is checked.
// In strict mode, drift causes an error and prints a diff.
// Otherwise, drift only prints a warning and does not block the snapshot creation.
func RunSnapshot(cfg *config.LitmusConfig, logger logrus.FieldLogger, update, strict bool) error {
	ws, err := workspace.Load(cfg.Workspace.Root, logger)
	if err != nil {
		return err
	}

	amCfg, err := ws.Config()
	if err != nil {
		return fmt.Errorf("failed to load alertmanager config: %w", err)
	}
	if amCfg.Route == nil {
		return fmt.Errorf("alertmanager config has no route defined")
	}

	ctx := context.Background()

	router := pipeline.NewRouter(amCfg.Route)
	runner := pipeline.NewRunner(stores.NewSilenceStore(nil), stores.NewAlertStore(), router, nil)

	walker := snapshot.NewRouteWalker(amCfg.Route)
	paths := walker.FindTerminalPaths()

	synthesizer := snapshot.NewSnapshotSynthesizer(runner)
	outcomes, err := synthesizer.DiscoverOutcomes(ctx, paths)
	if err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	if len(outcomes) == 0 {
		fmt.Fprintf(os.Stderr, "WARN: synthesis produced zero outcomes; baseline will be empty\n")
	}

	regTests := BuildRegressionTests(outcomes, cfg.GlobalLabels)

	var existing []*types.TestCase
	existingHistory, err := ListHistory(cfg.RegressionsDir())
	hasHistory := err == nil && len(existingHistory) > 0

	state, err := LoadRegressionState(cfg.RegressionsYamlFilePath())
	if err == nil && state != nil {
		existing = state.Tests
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
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
		}
	}

	if err := os.MkdirAll(cfg.RegressionsDir(), 0755); err != nil {
		return fmt.Errorf("creating regression directory: %w", err)
	}

	if !hasHistory {
		if _, err := ArchiveBaseline(cfg, regTests); err != nil {
			return fmt.Errorf("creating initial baseline: %w", err)
		}
		fmt.Println("✓ Baseline created") //nolint:forbidigo
		return nil
	}

	if hasDrift {
		if !update {
			fmt.Fprintf(os.Stderr, "WARN: drift detected in routing behavior; run 'litmus snapshot update' to accept changes, or 'litmus diff' to inspect\n")
			return nil
		}
		if _, err := ArchiveBaseline(cfg, regTests); err != nil {
			return fmt.Errorf("archiving baseline to history: %w", err)
		}
		fmt.Println("✓ Baseline updated") //nolint:forbidigo
		return nil
	}

	if update {
		fmt.Println("✓ No changes detected; baseline is up to date") //nolint:forbidigo
		return nil
	}

	fmt.Println("✓ Baseline is current") //nolint:forbidigo
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
