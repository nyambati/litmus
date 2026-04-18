package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/spf13/viper"
)

// DefaultConfigName is the base name of the configuration file.
const DefaultConfigName = ".litmus"

// envPattern matches env(VAR_NAME) where VAR_NAME is an uppercase env var identifier.
var envPattern = regexp.MustCompile(`env\(([A-Za-z_][A-Za-z0-9_]*)\)`)

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
		var notFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundErr) {
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

	// 5. Post-process: env(VAR) substitution across all string fields
	if err := cfg.expandEnv(); err != nil {
		return nil, fmt.Errorf("failed to expand environment variables: %w", err)
	}

	return &cfg, nil
}

// LoadAlertmanagerConfig reads, expands env(VAR) placeholders, and parses the
// Alertmanager YAML using alertmanager's own loader (applies validation and defaults).
func LoadAlertmanagerConfig(path string) (*amconfig.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading alertmanager config: %w", err)
	}

	expanded, err := expandEnvVars(string(data))
	if err != nil {
		return nil, fmt.Errorf("expanding env vars in alertmanager config: %w", err)
	}

	cfg, err := amconfig.Load(expanded)
	if err != nil {
		return nil, fmt.Errorf("parsing alertmanager config: %w", err)
	}

	return cfg, nil
}

// expandEnv applies env(VAR) substitution to all string fields in LitmusConfig.
func (c *LitmusConfig) expandEnv() error {
	fields := []*string{
		&c.Config.Directory,
		&c.Config.File,
		&c.Config.Templates,
		&c.Regression.Directory,
		&c.Tests.Directory,
		&c.Mimir.Address,
		&c.Mimir.TenantID,
		&c.Mimir.APIKey,
	}
	for _, f := range fields {
		expanded, err := expandEnvVars(*f)
		if err != nil {
			return err
		}
		*f = expanded
	}
	for k, v := range c.GlobalLabels {
		expanded, err := expandEnvVars(v)
		if err != nil {
			return err
		}
		c.GlobalLabels[k] = expanded
	}
	return nil
}

// expandEnvVars replaces env(VAR_NAME) expressions with the corresponding
// environment variable values. Returns an error if a referenced variable is unset.
// Fails fast on first error (does not partially substitute).
func expandEnvVars(s string) (string, error) {
	if !strings.Contains(s, "env(") {
		return s, nil
	}
	var expandErr error
	result := envPattern.ReplaceAllStringFunc(s, func(match string) string {
		// If error already occurred, skip further processing
		if expandErr != nil {
			return match
		}
		sub := envPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		val, ok := os.LookupEnv(strings.ToUpper(sub[1]))
		if !ok {
			expandErr = fmt.Errorf("env var %q referenced in config but not set", strings.ToUpper(sub[1]))
			return match
		}
		return val
	})
	if expandErr != nil {
		return "", expandErr
	}
	return result, nil
}
