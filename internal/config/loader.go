package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DefaultConfigName is the base name of the configuration file.
const DefaultConfigName = ".litmus"

// LoadConfig initializes and returns the litmus configuration.
// It follows the precedence order: Flags > Env Vars > Config File > Defaults.
func LoadConfig() (*LitmusConfig, error) {
	v := viper.New()

	// 1. Set Defaults
	v.SetDefault("config.directory", "config")
	v.SetDefault("config.file", "alertmanager.yml")
	v.SetDefault("config.templates", "templates/")
	v.SetDefault("regression.directory", "regressions")
	v.SetDefault("regression.max_samples", 5)
	v.SetDefault("tests.directory", "tests")
	v.SetDefault("mimir.address", "")
	v.SetDefault("mimir.tenant_id", "")
	v.SetDefault("mimir.api_key", "")

	// 2. Environment Variables
	v.SetEnvPrefix("LITMUS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 3. Config File
	v.SetConfigName(DefaultConfigName)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(os.Stderr, "WARN: .litmus.yaml not found, using defaults\n")
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// 4. Unmarshal
	var cfg LitmusConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 5. Post-process: Environment variable substitution
	if err := cfg.expandEnv(); err != nil {
		return nil, fmt.Errorf("failed to expand environment variables: %w", err)
	}

	return &cfg, nil
}

// LoadAlertmanagerConfig reads and parses the Alertmanager YAML configuration.
func LoadAlertmanagerConfig(path string) (*amconfig.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading alertmanager config: %w", err)
	}

	var cfg amconfig.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing alertmanager config: %w", err)
	}

	return &cfg, nil
}

// expandEnv processes fields that might contain {{ env "VAR" }} templates.
func (c *LitmusConfig) expandEnv() error {
	funcMap := template.FuncMap{
		"env": func(key string) string {
			return os.Getenv(key)
		},
	}

	expand := func(s string) (string, error) {
		if !strings.Contains(s, "{{") {
			return s, nil
		}
		tmpl, err := template.New("config").Funcs(funcMap).Parse(s)
		if err != nil {
			return s, err
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			return s, err
		}
		return buf.String(), nil
	}

	var err error
	if c.Mimir.Address, err = expand(c.Mimir.Address); err != nil {
		return err
	}
	if c.Mimir.TenantID, err = expand(c.Mimir.TenantID); err != nil {
		return err
	}
	if c.Mimir.APIKey, err = expand(c.Mimir.APIKey); err != nil {
		return err
	}

	return nil
}
