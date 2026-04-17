package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
// If update is false and a baseline exists, drift causes an error.
func RunSnapshot(update bool) error {
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
	outcomes := synthesizer.DiscoverOutcomes(ctx, paths)

	regTests := buildRegressionTests(outcomes, litmusConfig.GlobalLabels)

	baselinePath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.mpk")

	if !update {
		existing, err := LoadBaseline(baselinePath)
		if err == nil && existing != nil {
			if hasChanges(existing, regTests) {
				return fmt.Errorf("drift detected in routing behavior: use --update to accept changes")
			}
		}
	}

	if err := os.MkdirAll(litmusConfig.Regression.Directory, 0755); err != nil {
		return fmt.Errorf("creating regression directory: %w", err)
	}

	mpkFile, err := os.Create(baselinePath)
	if err != nil {
		return fmt.Errorf("creating baseline file: %w", err)
	}
	defer mpkFile.Close()

	if err := codec.EncodeMsgPack(mpkFile, regTests); err != nil {
		return fmt.Errorf("writing baseline: %w", err)
	}

	ymlPath := strings.Replace(baselinePath, "mpk", "yml", 1)
	ymlData, err := yaml.Marshal(regTests)
	if err != nil {
		return fmt.Errorf("marshaling YAML mirror: %w", err)
	}
	if err := os.WriteFile(ymlPath, ymlData, 0644); err != nil {
		return fmt.Errorf("writing YAML mirror: %w", err)
	}

	fmt.Printf("✓ Generated baseline: %s\n", baselinePath)
	fmt.Printf("✓ YAML mirror: %s\n", ymlPath)
	return nil
}

func buildRegressionTests(outcomes []*snapshot.SynthesisResult, globalLabels map[string]string) []*types.RegressionTest {
	tests := make([]*types.RegressionTest, 0, len(outcomes))
	for _, outcome := range outcomes {
		labels := make(map[string]string)
		for k, v := range globalLabels {
			labels[k] = v
		}
		for k, v := range outcome.Labels {
			labels[k] = v
		}
		tests = append(tests, &types.RegressionTest{
			Name:     fmt.Sprintf("Route to %s", strings.Join(outcome.Receivers, ", ")),
			Labels:   []map[string]string{labels},
			Expected: outcome.Receivers,
			Tags:     []string{"regression"},
		})
	}
	return tests
}

// hasChanges detects drift by comparing names, expected receivers, and label sets.
func hasChanges(existing, current []*types.RegressionTest) bool {
	if len(existing) != len(current) {
		return true
	}
	existingIdx := make(map[string]*types.RegressionTest, len(existing))
	for _, e := range existing {
		existingIdx[e.Name] = e
	}
	for _, t := range current {
		e, ok := existingIdx[t.Name]
		if !ok {
			return true
		}
		if !receiversEqual(e.Expected, t.Expected) {
			return true
		}
		if !labelSetsEqual(e.Labels, t.Labels) {
			return true
		}
	}
	return false
}

func receiversEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ac := append([]string{}, a...)
	bc := append([]string{}, b...)
	sort.Strings(ac)
	sort.Strings(bc)
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

func labelSetsEqual(a, b []map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for k, v := range a[i] {
			if b[i][k] != v {
				return false
			}
		}
	}
	return true
}
