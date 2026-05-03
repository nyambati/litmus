package config

type (
	// MimirConfig defines the connection parameters for Grafana Mimir.
	MimirConfig struct {
		Address  string `yaml:"address" mapstructure:"address" validate:"required,url"`
		TenantID string `yaml:"tenant_id" mapstructure:"tenant_id"`
		APIKey   string `yaml:"api_key" mapstructure:"api_key"`
	}

	// WorkspaceConfig defines the package-based layout and history settings.
	WorkspaceConfig struct {
		Root       string `yaml:"root" mapstructure:"root" validate:"required"`
		entrypoint string
		Fragments  string `yaml:"fragments" mapstructure:"fragments"`
		History    int    `yaml:"history" mapstructure:"history"`
	}

	// RequireConfig defines test requirement rules.
	RequireConfig struct {
		Tests      bool `yaml:"tests"      mapstructure:"tests"`
		Regression bool `yaml:"regression" mapstructure:"regression"`
	}

	// EnforceConfig defines matcher enforcement rules for fragment routes.
	EnforceConfig struct {
		Strict   bool     `yaml:"strict"   mapstructure:"strict"`
		Matchers []string `yaml:"matchers" mapstructure:"matchers"`
	}

	PolicyType string

	// PolicyConfig defines global rules for fragments.
	PolicyConfig struct {
		Require  RequireConfig `yaml:"require"  mapstructure:"require"`
		SkipRoot []PolicyType  `yaml:"skip_root" mapstructure:"skip_root"`
		Enforce  EnforceConfig `yaml:"enforce"  mapstructure:"enforce"`
	}

	// SanityMode defines whether a sanity check should warn or fail.
	SanityMode string

	SanityCheck string

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
	SanityModeFail          SanityMode  = "fail"
	SanityModeWarn          SanityMode  = "warn"
	CheckPolicyViolations   SanityCheck = "policy_violations"
	CheckOrphanReceivers    SanityCheck = "orphan_receivers"
	CheckDeadReceivers      SanityCheck = "dead_receivers"
	CheckShadowedRoutes     SanityCheck = "shadowed_routes"
	CheckInhibitionCycles   SanityCheck = "inhibition_cycles"
	CheckNegativeOnlyRoutes SanityCheck = "negative_only_routes"
)

func (m SanityMode) IsFail() bool {
	return m == SanityModeFail
}

// ModeFor returns the configured SanityMode for the named check.
// Unknown check names default to SanityModeFail.
func (c SanityConfig) ModeFor(name SanityCheck) SanityMode {
	modes := map[SanityCheck]SanityMode{
		CheckOrphanReceivers:    c.OrphanReceivers,
		CheckDeadReceivers:      c.DeadReceivers,
		CheckShadowedRoutes:     c.ShadowedRoutes,
		CheckInhibitionCycles:   c.InhibitionCycles,
		CheckPolicyViolations:   c.PolicyViolations,
		CheckNegativeOnlyRoutes: c.NegativeOnlyRoutes,
	}
	if m, ok := modes[name]; ok && m != "" {
		return m
	}
	return SanityModeFail
}
