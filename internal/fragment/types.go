package fragment

import (
	"github.com/nyambati/litmus/internal/types"
	"github.com/prometheus/alertmanager/config"
)

type (
	FragmentGroup struct {
		Match    map[string]string `yaml:"match"`
		Receiver string            `yaml:"receiver"`
	}

	InhibitRule struct {
		SourceMatch map[string]string `yaml:"source_match"`
		TargetMatch map[string]string `yaml:"target_match"`
		Equal       []string          `yaml:"equal"`
	}

	Fragment struct {
		// Namespace is the fragment's identity and the prefix applied to all receiver
		// and route receiver names during assembly. Auto-set to the directory basename
		// if not declared in YAML.
		Namespace    string               `yaml:"namespace,omitempty"`
		Group        *FragmentGroup       `yaml:"group,omitempty"`
		Routes       []*config.Route      `yaml:"routes,omitempty"`
		Receivers    []*types.Receiver    `yaml:"receivers,omitempty"`
		InhibitRules []config.InhibitRule `yaml:"inhibit_rules,omitempty"`
		// Deprecated. Remove before v1.0 release.
		MuteTimeIntervals []config.MuteTimeInterval `yaml:"mute_time_intervals,omitempty" json:"mute_time_intervals,omitempty"`
		TimeIntervals     []config.TimeInterval     `yaml:"time_intervals,omitempty" json:"time_intervals,omitempty"`
		// Tests are discovered from sibling *-tests.yml files and tests/ subdirectory
		// via dedicated parsing (see testDoc). Never serialized as part of the
		// Fragment YAML — yaml:"-" suppresses both encode and decode.
		Tests []*types.TestCase `yaml:"-"`
		dir   string            `yaml:"-"`
	}

	Metadata struct {
		Dir   string   `yaml:"dir,omitempty"`
		Files []string `yaml:"files,omitempty"`
	}

	AugmentedFragment struct {
		Fragment *Fragment         `yaml:"fragment,omitempty"`
		Metadata *Metadata         `yaml:"metadata,omitempty"`
		Tests    []*types.TestCase `yaml:"tests,omitempty"`
	}

	LoadResult struct {
		Fragments []AugmentedFragment `yaml:"fragments,omitempty"`
	}

	TestDoc struct {
		Tests []*types.TestCase `yaml:"tests"`
	}
)
