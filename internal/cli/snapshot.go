package cli

import (
	"context"
	"fmt"
	"os"

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
	ws, err := workspace.Load(cfg, logger)
	if err != nil {
		return err
	}

	amCfg, err := ws.Config()
	if err != nil {
		return fmt.Errorf("failed to load alertmanager config: %w", err)
	}

	ctx := context.Background()

	router := pipeline.NewRouter(amCfg.Route)
	runner := pipeline.NewRunner(stores.NewSilenceStore(nil), stores.NewAlertStore(), router, nil)

	walker := snapshot.NewRouteWalker(amCfg.Route)
	paths := walker.FindTerminalPaths()

	synthesizer := snapshot.NewSnapshotSynthesizer(runner, logger)
	outcomes, err := synthesizer.DiscoverOutcomes(ctx, paths)
	if err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	if len(outcomes) == 0 {
		fmt.Fprintf(os.Stderr, "WARN: synthesis produced zero outcomes; baseline will be empty\n")
	}

	regTests := snapshot.BuildRegressionTests(outcomes, cfg.GlobalLabels)

	var existing []*types.TestCase
	existingHistory, histErr := snapshot.ListHistory(cfg.RegressionsDir())
	if histErr != nil {
		return fmt.Errorf("listing regression history: %w", histErr)
	}
	hasHistory := len(existingHistory) > 0

	if ws.RegressionState != nil {
		existing = ws.RegressionState.Tests
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

	if err := os.MkdirAll(cfg.RegressionsDir(), 0o755); err != nil {
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
