package snapshot

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/types"
	"gopkg.in/yaml.v3"
)

// RegressionState holds the current active baseline ID and its tests.
type RegressionState struct {
	ID    string            `yaml:"id"`
	Tests []*types.TestCase `yaml:"tests"`
}

// RegressionDiff holds the comparison result between two regression baselines.
type RegressionDiff struct {
	Deltas []RegressionDelta `json:"deltas"`
}

// RegressionDelta represents a single change between two regression baselines.
type RegressionDelta struct {
	Kind     DeltaKind         `json:"kind"`
	Labels   map[string]string `json:"labels"`
	Expected []string          `json:"expected,omitempty"`
	Actual   []string          `json:"actual,omitempty"`
}

// DeltaKind represents the kind of change in a regression delta.
type DeltaKind string

const (
	DeltaAdded    DeltaKind = "added"
	DeltaModified DeltaKind = "modified"
	DeltaRemoved  DeltaKind = "removed"
)

// BuildRegressionTests converts synthesis outcomes into executable regression test cases.
func BuildRegressionTests(outcomes []*SynthesisResult, globalLabels map[string]string) []*types.TestCase {
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

// ListHistory returns history entry IDs sorted newest-first.
func ListHistory(regressionDir string) ([]string, error) {
	entries, err := os.ReadDir(regressionDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading history dir: %w", err)
	}

	var ids []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			id := strings.TrimSuffix(e.Name(), ".mpk")
			ids = append(ids, id)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	return ids, nil
}

// RollbackToEntry restores a history entry as the active baseline.
func RollbackToEntry(cfg *config.LitmusConfig, id string) error {
	srcMpk := filepath.Join(cfg.RegressionsDir(), id+".mpk")

	// Load the tests from the historical baseline
	tests, err := LoadBaseline(srcMpk)
	if err != nil {
		return fmt.Errorf("loading history entry %q: %w", id, err)
	}

	if err := SaveRegressionState(cfg.RegressionsYamlFilePath(), &RegressionState{ID: id, Tests: tests}); err != nil {
		return fmt.Errorf("writing regression state: %w", err)
	}

	return nil
}

// LoadBaseline reads a msgpack regression baseline from disk.
func LoadBaseline(path string) ([]*types.TestCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var tests []*types.TestCase
	if err := codec.DecodeMsgPack(file, &tests); err != nil {
		return nil, fmt.Errorf("decoding baseline %q: %w", path, err)
	}
	return tests, nil
}

// LoadBaselineYAML reads a YAML regression baseline from disk.
func LoadBaselineYAML(path string) ([]*types.TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tests []*types.TestCase
	if err := yaml.Unmarshal(data, &tests); err != nil {
		return nil, fmt.Errorf("parsing baseline YAML %q: %w", path, err)
	}
	return tests, nil
}

// SaveRegressionState writes the regression state (ID + tests) to regressions.litmus.yml.
func SaveRegressionState(path string, state *RegressionState) error {
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("serializing regression state: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}

	return nil
}
