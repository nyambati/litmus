package config

type (
	Config struct {
		Directory string `yaml:"directory"`
		File      string `yaml:"file"`
		Templates string `yaml:"templates"`
	}

	RegressionConfig struct {
		MaxSamples int    `yaml:"max_samples"`
		Directory  string `yaml:"directory"`
	}

	TestsConfig struct {
		Directory string `yaml:"directory"`
	}

	LitmusConfig struct {
		Config       Config            `yaml:"config"`
		GlobalLabels map[string]string `yaml:"global_labels"`
		Regression   RegressionConfig  `yaml:"regression"`
		Tests        TestsConfig       `yaml:"tests"`
	}
)
