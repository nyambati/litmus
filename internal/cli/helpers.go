package cli

import (
	"os"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/types"
	"gopkg.in/yaml.v3"
)

// LoadBaseline reads a msgpack regression baseline from disk.
func LoadBaseline(path string) ([]*types.TestCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var tests []*types.TestCase
	if err := codec.DecodeMsgPack(file, &tests); err != nil {
		return nil, err
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
		return nil, err
	}
	return tests, nil
}

// RegressionState holds the current active baseline ID and its tests.
type RegressionState struct {
	ID    string            `yaml:"id"`
	Tests []*types.TestCase `yaml:"tests"`
}

// LoadRegressionState reads the regression state (ID + tests) from regressions.litmus.yml.
func LoadRegressionState(path string) (*RegressionState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state RegressionState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveRegressionState writes the regression state (ID + tests) to regressions.litmus.yml.
func SaveRegressionState(path string, state *RegressionState) error {
	data, err := yaml.Marshal(state)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}

	return nil
}
