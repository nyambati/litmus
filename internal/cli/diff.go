package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/utils"
	"github.com/nyambati/litmus/internal/workspace"
	"github.com/sirupsen/logrus"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// RunDiff compares current config against the baseline and prints a structural diff.
func RunDiff(cfg *config.LitmusConfig, logger logrus.FieldLogger) error {
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

	currentTests := snapshot.BuildRegressionTests(outcomes, cfg.GlobalLabels)

	if ws.RegressionState == nil {
		return fmt.Errorf("no baseline found — run 'litmus snapshot capture' to create one")
	}

	state := ws.RegressionState

	existingTests := state.Tests
	if len(existingTests) == 0 {
		return fmt.Errorf("baseline is empty")
	}

	diff := snapshot.ComputeDiff(existingTests, currentTests)
	PrintDiffReport(diff)

	return nil
}

// PrintDiffReport outputs a color-coded structural delta.
//
//nolint:forbidigo
func PrintDiffReport(diff *snapshot.RegressionDiff) {
	if len(diff.Deltas) == 0 {
		fmt.Println("No behavioral changes detected.")
		return
	}

	// Sort deltas for stable output
	sort.Slice(diff.Deltas, func(i, j int) bool {
		return labelKeyForSort(diff.Deltas[i].Labels) < labelKeyForSort(diff.Deltas[j].Labels)
	})

	for _, delta := range diff.Deltas {
		switch delta.Kind {
		case snapshot.DeltaAdded:
			fmt.Printf("%s[+] ADDED:   Route to %s%s\n", colorGreen, formatReceivers(delta.Actual), colorReset)
			fmt.Printf("    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Printf("    Outcome: %s\n", formatReceivers(delta.Actual))

		case snapshot.DeltaRemoved:
			fmt.Printf("%s[-] REMOVED: Route to %s%s\n", colorRed, formatReceivers(delta.Expected), colorReset)
			fmt.Printf("    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Printf("    Old:     %s\n", formatReceivers(delta.Expected))

		case snapshot.DeltaModified:
			fmt.Printf("%s[!] MODIFIED: Behavior for Labels%s\n", colorYellow, colorReset)
			fmt.Printf("    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Printf("    %s- Expected: %s%s\n", colorRed, formatReceivers(delta.Expected), colorReset)
			fmt.Printf("    %s+ Actual:   %s%s\n", colorGreen, formatReceivers(delta.Actual), colorReset)
		}
		fmt.Println()
	}
}

func labelKeyForSort(m map[string]string) string {
	return utils.LabelFormat(m)
}
