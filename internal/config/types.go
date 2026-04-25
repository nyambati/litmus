package config

type (
	// Config defines the Alertmanager configuration locations.
	Config struct {
		Directory string `yaml:"directory" mapstructure:"directory"`
		File      string `yaml:"file" mapstructure:"file"`
		Templates string `yaml:"templates" mapstructure:"templates"`
	}

	// MimirConfig defines the connection parameters for Grafana Mimir.
	MimirConfig struct {
		Address  string `yaml:"address" mapstructure:"address"`
		TenantID string `yaml:"tenant_id" mapstructure:"tenant_id"`
		APIKey   string `yaml:"api_key" mapstructure:"api_key"`
	}

	// RegressionConfig defines the regression test parameters.
	RegressionConfig struct {
		Directory string `yaml:"directory" mapstructure:"directory"`
		Keep      int    `yaml:"keep" mapstructure:"keep"`
	}

	// TestsConfig defines the behavioral unit test parameters.
	TestsConfig struct {
		Directory string `yaml:"directory" mapstructure:"directory"`
	}

	// LitmusConfig is the root configuration object.
	LitmusConfig struct {
		Config       Config            `yaml:"config" mapstructure:"config"`
		GlobalLabels map[string]string `yaml:"global_labels" mapstructure:"global_labels"`
		Mimir        MimirConfig       `yaml:"mimir" mapstructure:"mimir"`
		Regression   RegressionConfig  `yaml:"regression" mapstructure:"regression"`
		Tests        TestsConfig       `yaml:"tests" mapstructure:"tests"`
	}
)
