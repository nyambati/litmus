package cli

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/utils"
	"github.com/nyambati/litmus/internal/workspace"
)

// RunDiff compares current config against the baseline and prints a structural diff.
func RunDiff(ctx context.Context) error {
	cfg := config.ConfigFromContext(ctx)
	logger := config.LoggerFromContext(ctx)
	ws, err := workspace.Load(cfg, logger)
	if err != nil {
		return err
	}

	amCfg, err := ws.Config()
	if err != nil {
		return fmt.Errorf("failed to load alertmanager config: %w", err)
	}

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

// PrintDiffReport outputs a structural delta. Color is applied automatically
// when stdout is a terminal (see styler).
func PrintDiffReport(diff *snapshot.RegressionDiff) {
	st := newStyler(os.Stdout)

	if len(diff.Deltas) == 0 {
		fmt.Fprintln(os.Stdout, "No behavioral changes detected.")
		return
	}

	// Sort deltas for stable output
	sort.Slice(diff.Deltas, func(i, j int) bool {
		return labelKeyForSort(diff.Deltas[i].Labels) < labelKeyForSort(diff.Deltas[j].Labels)
	})

	for _, delta := range diff.Deltas {
		switch delta.Kind {
		case snapshot.DeltaAdded:
			fmt.Fprintln(os.Stdout, st.green("[+] ADDED:   Route to "+formatReceivers(delta.Actual)))
			fmt.Fprintf(os.Stdout, "    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Fprintf(os.Stdout, "    Outcome: %s\n", formatReceivers(delta.Actual))

		case snapshot.DeltaRemoved:
			fmt.Fprintln(os.Stdout, st.red("[-] REMOVED: Route to "+formatReceivers(delta.Expected)))
			fmt.Fprintf(os.Stdout, "    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Fprintf(os.Stdout, "    Old:     %s\n", formatReceivers(delta.Expected))

		case snapshot.DeltaModified:
			fmt.Fprintln(os.Stdout, st.yellow("[!] MODIFIED: Behavior for Labels"))
			fmt.Fprintf(os.Stdout, "    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Fprintf(os.Stdout, "    %s\n", st.red("- Expected: "+formatReceivers(delta.Expected)))
			fmt.Fprintf(os.Stdout, "    %s\n", st.green("+ Actual:   "+formatReceivers(delta.Actual)))
		}
		fmt.Fprintln(os.Stdout)
	}
}

func labelKeyForSort(m map[string]string) string {
	return utils.LabelFormat(m)
}
