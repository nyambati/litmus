package behavioral

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"litmus/internal/types"
)

// BehavioralTestLoader loads human-authored test scenarios from YAML.
type BehavioralTestLoader struct{}

// NewBehavioralTestLoader creates test loader.
func NewBehavioralTestLoader() *BehavioralTestLoader {
	return &BehavioralTestLoader{}
}

// LoadFromFile loads behavioral tests from single YAML file.
func (btl *BehavioralTestLoader) LoadFromFile(path string) ([]*types.BehavioralTest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var test types.BehavioralTest
	if err := yaml.Unmarshal(data, &test); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	return []*types.BehavioralTest{&test}, nil
}

// LoadFromDirectory loads all YAML files from directory.
func (btl *BehavioralTestLoader) LoadFromDirectory(dir string) ([]*types.BehavioralTest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var allTests []*types.BehavioralTest

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process YAML files
		if filepath.Ext(entry.Name()) != ".yml" && filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		tests, err := btl.LoadFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", entry.Name(), err)
		}

		allTests = append(allTests, tests...)
	}

	return allTests, nil
}
