package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nyambati/litmus/internal/engine/behavioral"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/spf13/viper"
)

const (
	// DefaultConfigName is the base name of the configuration file.
	defaultConfigName         = ".litmus"
	defaultRegressionYamlFile = "regressions.litmus.yml"
	defaultRegressionDir      = "regressions"
	defaultConfigDir          = "config"
	defaultConfigFile         = "alertmanager.yml"
	defaultConfigFileAlt      = "alertmanager.yaml"
	defaultConfigTemplatesDir = "templates"
	defaultTestsDir           = "tests"
	defaultFragmentsPattern   = "fragments/*"
	defaultHistoryKeep        = 5
)

// envPattern matches env(VAR_NAME) where VAR_NAME is an uppercase env var identifier.
var envPattern = regexp.MustCompile(`env\(([A-Za-z_][A-Za-z0-9_]*)\)`)

// LoadConfig initializes and returns the litmus configuration.
// It follows the precedence order: Flags > Env Vars > Config File > Defaults.
func LoadConfig() (*LitmusConfig, error) {
	//nolint:errcheck
	godotenv.Load()
	v := viper.New()

	// 1. Set Defaults
	v.SetDefault("workspace.root", defaultConfigDir)
	v.SetDefault("workspace.fragments", defaultFragmentsPattern)
	v.SetDefault("workspace.history", defaultHistoryKeep)
	v.SetDefault("mimir.address", "")
	v.SetDefault("mimir.tenant_id", "")
	v.SetDefault("mimir.api_key", "")
	v.SetDefault("policy.enforce.strict", true)

	// 2. Environment Variables
	v.SetEnvPrefix("LITMUS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 3. Config File
	v.SetConfigName(defaultConfigName)
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

// expandAlertmanagerConfig reads and env-expands alertmanager YAML without parsing.
// Returns the expanded raw YAML string.
func expandAlertmanagerConfig(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading alertmanager config: %w", err)
	}

	expanded, err := expandEnvVars(string(data))
	if err != nil {
		return "", fmt.Errorf("expanding env vars in alertmanager config: %w", err)
	}

	return expanded, nil
}

// loadAlertmanagerConfig reads, expands env(VAR) placeholders, and parses the
// Alertmanager YAML using alertmanager's own loader (applies validation and defaults).
// Returns the parsed config, raw expanded YAML, and any error.
func loadAlertmanagerConfig(path string) (*amconfig.Config, string, error) {
	expanded, err := expandAlertmanagerConfig(path)
	if err != nil {
		return nil, "", err
	}

	cfg, err := amconfig.Load(expanded)
	if err != nil {
		return nil, "", fmt.Errorf("parsing alertmanager config: %w", err)
	}

	return cfg, expanded, nil
}

// expandEnv applies env(VAR) substitution to all string fields in LitmusConfig.
func (c *LitmusConfig) expandEnv() error {
	fields := []*string{
		&c.Workspace.Root,
		&c.Workspace.Fragments,
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

// methods

// FilePath returns the path to the alertmanager config file, preferring .yml
// but falling back to .yaml if the .yml variant does not exist.
func (c *LitmusConfig) FilePath() string {
	primary := filepath.Join(c.Workspace.Root, defaultConfigFile)
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	alt := filepath.Join(c.Workspace.Root, defaultConfigFileAlt)
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return primary // return primary so callers get a meaningful "not found" path in errors
}

func (c *LitmusConfig) RegressionsDir() string {
	return filepath.Join(c.Workspace.Root, defaultRegressionDir)
}

func (c *LitmusConfig) RegressionsYamlFilePath() string {
	// Regressions always live in the root package's regressions/ directory
	return filepath.Join(c.Workspace.Root, defaultRegressionDir, defaultRegressionYamlFile)
}

func (c *LitmusConfig) TemplatesDir() string {
	return filepath.Join(c.Workspace.Root, defaultConfigTemplatesDir)
}

func (c *LitmusConfig) TestsDir() string {
	return filepath.Join(c.Workspace.Root, defaultTestsDir)
}

// LoadAssembledConfig loads the base Alertmanager config and all fragments,
// performing virtual assembly and namespacing.
func (c *LitmusConfig) LoadAssembledConfig() (*amconfig.Config, []*Fragment, string, error) {
	// 1. Load Base Config Package
	base, raw, err := loadAlertmanagerConfig(c.FilePath())
	if err != nil {
		return nil, nil, "", err
	}

	// 2. Discover Fragments
	fragments, err := LoadFragments(c.FragmentsPath())
	if err != nil {
		err := fmt.Errorf("WARN: loading fragments: %w", err)
		return nil, nil, "", err
	}

	// Capture base routes before assembly (shallow copy — assembly appends to the slice
	// in place and would otherwise pollute the root fragment with mounted fragment routes).
	baseRoutes := make([]*amconfig.Route, len(base.Route.Routes))
	copy(baseRoutes, base.Route.Routes)

	// Load tests from the root tests/ directory.
	rootTests, _ := behavioral.NewBehavioralTestLoader().LoadFromDirectory(c.TestsDir())

	rootFrag := &Fragment{
		Name:   "root",
		Tests:  rootTests,
		Routes: baseRoutes,
	}

	allFragments := append([]*Fragment{rootFrag}, fragments...)

	if len(fragments) == 0 {
		// No assembly needed, but always return allFragments so policy can run on root.
		return base, allFragments, raw, nil
	}

	// 3. Assemble
	assembler := NewAssembler(base)
	assembled, err := assembler.Assemble(fragments)
	if err != nil {
		return nil, nil, "", fmt.Errorf("assembling fragments: %w", err)
	}

	return assembled, allFragments, raw, nil
}

func (c *LitmusConfig) FragmentsPath() string {
	if filepath.IsAbs(c.Workspace.Fragments) {
		return c.Workspace.Fragments
	}
	return filepath.Join(c.Workspace.Root, c.Workspace.Fragments)
}

func (m *MimirConfig) Validate() error {
	// Validate required field
	if m.Address == "" {
		return fmt.Errorf("mimir address not configured: set LITMUS_MIMIR_ADDRESS env var, provide --address flag, or add mimir.address to .litmus.yaml")
	}
	return nil
}
