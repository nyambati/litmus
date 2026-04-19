package behavioral

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/types"
	"gopkg.in/yaml.v3"
)

// BehavioralTestLoader loads human-authored test scenarios from YAML.
type BehavioralTestLoader struct{}

// NewBehavioralTestLoader creates test loader.
func NewBehavioralTestLoader() *BehavioralTestLoader {
	return &BehavioralTestLoader{}
}

// LoadFromFile loads a single unit TestCase from a YAML file.
func (btl *BehavioralTestLoader) LoadFromFile(path string) (*types.TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var test types.TestCase
	if err := yaml.Unmarshal(data, &test); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if test.Type == "" {
		test.Type = "unit"
	}

	return &test, nil
}

// LoadFromDirectory loads all unit TestCases from YAML files in a directory.
func (btl *BehavioralTestLoader) LoadFromDirectory(dir string) ([]*types.TestCase, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var allTests []*types.TestCase

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		test, err := btl.LoadFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", entry.Name(), err)
		}

		allTests = append(allTests, test)
	}

	return allTests, nil
}
