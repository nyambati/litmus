package config

type (
	// MimirConfig defines the connection parameters for Grafana Mimir.
	MimirConfig struct {
		Address  string `yaml:"address" mapstructure:"address"`
		TenantID string `yaml:"tenant_id" mapstructure:"tenant_id"`
		APIKey   string `yaml:"api_key" mapstructure:"api_key"`
	}

	// WorkspaceConfig defines the package-based layout and history settings.
	WorkspaceConfig struct {
		Root       string `yaml:"root" mapstructure:"root"`
		entrypoint string `yaml:"entrypoint" mapstructure:"entrypoint"`
		Fragments  string `yaml:"fragments" mapstructure:"fragments"`
		History    int    `yaml:"history" mapstructure:"history"`
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

	// SanityMode defines whether a sanity check should warn or fail.
	SanityMode string

	// SanityConfig defines sanity check behavior modes.
	SanityConfig struct {
		OrphanReceivers    SanityMode `yaml:"orphan_receivers"     mapstructure:"orphan_receivers"`
		DeadReceivers      SanityMode `yaml:"dead_receivers"       mapstructure:"dead_receivers"`
		ShadowedRoutes     SanityMode `yaml:"shadowed_routes"      mapstructure:"shadowed_routes"`
		InhibitionCycles   SanityMode `yaml:"inhibition_cycles"    mapstructure:"inhibition_cycles"`
		PolicyViolations   SanityMode `yaml:"policy_violations"    mapstructure:"policy_violations"`
		NegativeOnlyRoutes SanityMode `yaml:"negative_only_routes" mapstructure:"negative_only_routes"`
	}

	// LitmusConfig is the root configuration object.
	LitmusConfig struct {
		Workspace    WorkspaceConfig   `yaml:"workspace" mapstructure:"workspace"`
		Policy       PolicyConfig      `yaml:"policy" mapstructure:"policy"`
		Sanity       SanityConfig      `yaml:"sanity" mapstructure:"sanity"`
		GlobalLabels map[string]string `yaml:"global_labels" mapstructure:"global_labels"`
		Mimir        MimirConfig       `yaml:"mimir" mapstructure:"mimir"`
	}
)

const (
	SanityModeFail SanityMode = "fail"
	SanityModeWarn SanityMode = "warn"
)

func (m SanityMode) IsFail() bool {
	return m == SanityModeFail
}
