package config

import (
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
)

type (
	// MimirConfig defines the connection parameters for Grafana Mimir.
	MimirConfig struct {
		Address  string `yaml:"address" mapstructure:"address"`
		TenantID string `yaml:"tenant_id" mapstructure:"tenant_id"`
		APIKey   string `yaml:"api_key" mapstructure:"api_key"`
	}

	// WorkspaceConfig defines the package-based layout and history settings.
	WorkspaceConfig struct {
		Root      string `yaml:"root" mapstructure:"root"`
		Fragments string `yaml:"fragments" mapstructure:"fragments"`
		History   int    `yaml:"history" mapstructure:"history"`
	}

	// EnforceConfig defines matcher enforcement rules for fragment routes.
	EnforceConfig struct {
		Strict   bool     `yaml:"strict"   mapstructure:"strict"`
		Matchers []string `yaml:"matchers" mapstructure:"matchers"`
	}

	// PolicyConfig defines global rules for fragments.
	PolicyConfig struct {
		RequireTests bool          `yaml:"require_tests" mapstructure:"require_tests"`
		SkipRoot     bool          `yaml:"skip_root"     mapstructure:"skip_root"`
		Enforce      EnforceConfig `yaml:"enforce"      mapstructure:"enforce"`
	}

	// LitmusConfig is the root configuration object.
	LitmusConfig struct {
		Workspace    WorkspaceConfig   `yaml:"workspace" mapstructure:"workspace"`
		Policy       PolicyConfig      `yaml:"policy" mapstructure:"policy"`
		GlobalLabels map[string]string `yaml:"global_labels" mapstructure:"global_labels"`
		Mimir        MimirConfig       `yaml:"mimir" mapstructure:"mimir"`
	}

	// FragmentGroup defines a synthetic parent route created during assembly.
	// Match uses exact label matching only — no regex support on the synthetic parent.
	FragmentGroup struct {
		Match    map[string]string `yaml:"match"`
		Receiver string            `yaml:"receiver"`
	}

	// Fragment represents a team-level configuration fragment.
	Fragment struct {
		Name         string                 `yaml:"name"`
		Namespace    string                 `yaml:"namespace"`
		Group        *FragmentGroup         `yaml:"group"`
		Routes       []*amconfig.Route      `yaml:"routes"`
		Receivers    []amconfig.Receiver    `yaml:"receivers"`
		InhibitRules []amconfig.InhibitRule `yaml:"inhibit_rules"`
		// Tests are discovered from sibling *-tests.yml files and tests/ subdirectory.
		// Never parsed from the fragment definition file itself.
		Tests []*types.TestCase `yaml:"-"`
	}
)
