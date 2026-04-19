package cli

import (
	"os"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/types"
	"gopkg.in/yaml.v3"
)

// LoadBaseline reads a msgpack regression baseline from disk.
func LoadBaseline(path string) ([]*types.RegressionTest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var tests []*types.RegressionTest
	if err := codec.DecodeMsgPack(file, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}

// LoadBaselineYAML reads a YAML regression baseline from disk.
func LoadBaselineYAML(path string) ([]*types.RegressionTest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tests []*types.RegressionTest
	if err := yaml.Unmarshal(data, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}
