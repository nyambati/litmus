package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nyambati/litmus/internal/utils"
	"github.com/spf13/viper"
)

const (
	// DefaultConfigName is the base name of the configuration file.
	defaultConfigName         = ".litmus"
	defaultRegressionYamlFile = "regressions.litmus.yml"
	defaultRegressionDir      = "regressions"
	defaultConfigDir          = "config"
	defaultConfigTemplatesDir = "templates"
	defaultTestsDir           = "tests"
	defaultFragmentsPattern   = "fragments/*"
	defaultHistoryKeep        = 5
)

// ConfigKey is the context key for LitmusConfig.
type ConfigKey struct{}

// FromContext retrieves LitmusConfig from context.
func FromContext(ctx context.Context) *LitmusConfig {
	if cfg, ok := ctx.Value(ConfigKey{}).(*LitmusConfig); ok {
		return cfg
	}
	return nil
}

// LoadConfig is an alias for New to maintain compatibility with existing callers.
func LoadConfig() (*LitmusConfig, error) {
	return New()
}

func New() (*LitmusConfig, error) {
	//nolint:errcheck
	godotenv.Load()
	v := viper.New()

	v.SetDefault("workspace.root", defaultConfigDir)
	v.SetDefault("workspace.fragments", defaultFragmentsPattern)
	v.SetDefault("workspace.history", defaultHistoryKeep)
	v.SetDefault("mimir.address", "")
	v.SetDefault("mimir.tenant_id", "")
	v.SetDefault("mimir.api_key", "")
	v.SetDefault("policy.enforce.strict", true)
	v.SetDefault("sanity.orphan_receivers", SanityModeFail)
	v.SetDefault("sanity.dead_receivers", SanityModeFail)
	v.SetDefault("sanity.shadowed_routes", SanityModeFail)
	v.SetDefault("sanity.inhibition_cycles", SanityModeFail)
	v.SetDefault("sanity.policy_violations", SanityModeFail)
	v.SetDefault("sanity.negative_only_routes", SanityModeFail)

	v.SetEnvPrefix("LITMUS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigName(defaultConfigName)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		_, ok := errors.AsType[viper.ConfigFileNotFoundError](err)
		if !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		fmt.Fprintf(os.Stderr, "WARN: .litmus.yaml not found, using defaults\n")
	}

	var cfg LitmusConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.expandEnv(); err != nil {
		return nil, fmt.Errorf("failed to expand environment variables: %w", err)
	}

	entrypoint, err := findEntrypoint(cfg.Workspace.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to find base config: %w", err)
	}

	cfg.Workspace.entrypoint = entrypoint

	return &cfg, nil
}

func (c *LitmusConfig) expandEnv() error {
	fields := []*string{
		&c.Workspace.Root,
		&c.Workspace.Fragments,
		&c.Mimir.Address,
		&c.Mimir.TenantID,
		&c.Mimir.APIKey,
	}
	for _, f := range fields {
		expanded, err := utils.ExpandEnvVars(*f)
		if err != nil {
			return err
		}
		*f = expanded
	}
	for k, v := range c.GlobalLabels {
		expanded, err := utils.ExpandEnvVars(v)
		if err != nil {
			return err
		}
		c.GlobalLabels[k] = expanded
	}
	return nil
}

func (c *LitmusConfig) FilePath() string {
	return c.Workspace.entrypoint
}

func (c *LitmusConfig) RegressionsDir() string {
	return filepath.Join(c.Workspace.Root, defaultRegressionDir)
}

func (c *LitmusConfig) RegressionsYamlFilePath() string {
	return filepath.Join(c.Workspace.Root, defaultRegressionDir, defaultRegressionYamlFile)
}

func (c *LitmusConfig) TemplatesDir() string {
	return filepath.Join(c.Workspace.Root, defaultConfigTemplatesDir)
}

func (c *LitmusConfig) TestsDir() string {
	return filepath.Join(c.Workspace.Root, defaultTestsDir)
}

func (c *LitmusConfig) FragmentsPath() string {
	if filepath.IsAbs(c.Workspace.Fragments) {
		return c.Workspace.Fragments
	}
	return filepath.Join(c.Workspace.Root, c.Workspace.Fragments)
}

func (m *MimirConfig) Validate() error {
	if m.Address == "" {
		return fmt.Errorf("mimir address not configured: set LITMUS_MIMIR_ADDRESS env var, provide --address flag, or add mimir.address to .litmus.yaml")
	}
	return nil
}

func findEntrypoint(root string) (string, error) {
	// Use two explicit globs instead of "*ml" which matches .toml, .html, .xml, etc.
	yamlFiles, err := filepath.Glob(filepath.Join(root, "*.yaml"))
	if err != nil {
		return "", err
	}
	ymlFiles, err := filepath.Glob(filepath.Join(root, "*.yml"))
	if err != nil {
		return "", err
	}
	allFiles := make([]string, 0, len(yamlFiles)+len(ymlFiles))
	allFiles = append(allFiles, yamlFiles...)
	allFiles = append(allFiles, ymlFiles...)

	baseFileRegex := regexp.MustCompile(`^(base|alertmanager)\.ya?ml$`)
	var matches []string
	for _, file := range allFiles {
		if baseFileRegex.MatchString(filepath.Base(file)) {
			matches = append(matches, file)
		}
	}

	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("found 0 files matching base or alertmanager in root directory: %s", root)
	}
	return "", fmt.Errorf("found %d files matching base or alertmanager (%s) in root directory: %s",
		len(matches), strings.Join(matches, ","), root)
}
