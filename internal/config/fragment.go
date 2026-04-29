package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nyambati/litmus/internal/types"
	"gopkg.in/yaml.v3"
)

const (
	testSuffixYML  = "-tests.yml"
	testSuffixYAML = "-tests.yaml"
)

// LoadFragments discovers and loads all fragments matching the configured pattern.
func LoadFragments(pattern string) ([]*Fragment, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("resolving fragments pattern: %w", err)
	}
	sort.Strings(matches)

	var fragments []*Fragment
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			// Folder package mode
			frag, err := loadFragmentFromFolder(path)
			if err != nil {
				return nil, fmt.Errorf("loading fragment folder %s: %w", path, err)
			}
			if frag != nil {
				fragments = append(fragments, frag)
			}
		}

		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			// Single file mode
			if strings.HasSuffix(path, testSuffixYML) || strings.HasSuffix(path, testSuffixYAML) {
				continue // Skip sibling test files, they are loaded by the definition file logic
			}
			frag, err := loadFragmentFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("loading fragment file %s: %w", path, err)
			}
			fragments = append(fragments, frag)
		}
	}

	return fragments, nil
}

func loadFragmentFromFile(path string) (*Fragment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var frag Fragment
	if err := yaml.Unmarshal(data, &frag); err != nil {
		return nil, err
	}

	if frag.Name == "" {
		frag.Name = filepath.Base(path)
	}

	// Sibling test file discovery
	base := strings.TrimSuffix(path, filepath.Ext(path))
	for _, suffix := range []string{testSuffixYML, testSuffixYAML} {
		siblingTest := base + suffix
		if _, err := os.Stat(siblingTest); err == nil {
			tests, err := loadTestsFromFile(siblingTest)
			if err == nil {
				frag.Tests = append(frag.Tests, tests...)
			}
			break
		}
	}

	return &frag, nil
}

func loadTestsFromFile(path string) ([]*types.TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// The test file can be a single test or a list of tests
	// We'll try both
	var single types.TestCase
	if err := yaml.Unmarshal(data, &single); err == nil && single.Name != "" {
		if single.Type == "" {
			single.Type = "unit"
		}
		return []*types.TestCase{&single}, nil
	}

	var list []*types.TestCase
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	for _, t := range list {
		if t.Type == "" {
			t.Type = "unit"
		}
	}

	return list, nil
}

func loadFragmentFromFolder(path string) (*Fragment, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	composite := &Fragment{
		Name: filepath.Base(path),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		// Skip sibling test files
		if strings.HasSuffix(entry.Name(), "-tests.yml") || strings.HasSuffix(entry.Name(), "-tests.yaml") {
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		part, err := loadFragmentFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("loading part %s: %w", entry.Name(), err)
		}

		// Merge components
		composite.Routes = append(composite.Routes, part.Routes...)
		composite.Receivers = append(composite.Receivers, part.Receivers...)
		composite.InhibitRules = append(composite.InhibitRules, part.InhibitRules...)
		composite.Tests = append(composite.Tests, part.Tests...)

		// Merge metadata (first non-empty value wins)
		if part.Namespace != "" && composite.Namespace == "" {
			composite.Namespace = part.Namespace
		}
		if part.Group != nil {
			if composite.Group == nil {
				composite.Group = part.Group
			} else if !maps.Equal(composite.Group.Match, part.Group.Match) {
				return nil, fmt.Errorf("conflicting group definitions in folder package %s", path)
			}
		}
	}

	// Look for tests/ directory
	testsDir := filepath.Join(path, "tests")
	if info, err := os.Stat(testsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(testsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				ext := filepath.Ext(entry.Name())
				if ext == ".yml" || ext == ".yaml" {
					tests, err := loadTestsFromFile(filepath.Join(testsDir, entry.Name()))
					if err == nil {
						composite.Tests = append(composite.Tests, tests...)
					}
				}
			}
		}
	}

	return composite, nil
}
