package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// RunDiff compares current config against the baseline and prints a structural diff.
func RunDiff() error {
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

	currentTests := buildRegressionTests(outcomes, litmusConfig.GlobalLabels)

	baselinePath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.mpk")
	existingTests, err := LoadBaseline(baselinePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no baseline found — run 'litmus snapshot' to create one")
		}
		return fmt.Errorf("loading baseline: %w", err)
	}

	diff := snapshot.ComputeDiff(existingTests, currentTests)
	PrintDiffReport(diff)

	return nil
}

// PrintDiffReport outputs a color-coded structural delta.
//
//nolint:forbidigo
func PrintDiffReport(diff *types.RegressionDiff) {
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
		case types.DeltaAdded:
			fmt.Printf("%s[+] ADDED:   Route to %s%s\n", colorGreen, formatReceivers(delta.Actual), colorReset)
			fmt.Printf("    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Printf("    Outcome: %s\n", formatReceivers(delta.Actual))

		case types.DeltaRemoved:
			fmt.Printf("%s[-] REMOVED: Route to %s%s\n", colorRed, formatReceivers(delta.Expected), colorReset)
			fmt.Printf("    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Printf("    Old:     %s\n", formatReceivers(delta.Expected))

		case types.DeltaModified:
			fmt.Printf("%s[!] MODIFIED: Behavior for Labels%s\n", colorYellow, colorReset)
			fmt.Printf("    Labels:  %s\n", formatLabels(delta.Labels))
			fmt.Printf("    %s- Expected: %s%s\n", colorRed, formatReceivers(delta.Expected), colorReset)
			fmt.Printf("    %s+ Actual:   %s%s\n", colorGreen, formatReceivers(delta.Actual), colorReset)
		}
		fmt.Println()
	}
}

func labelKeyForSort(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(m[k])
	}
	return b.String()
}
