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
	"gopkg.in/yaml.v3"
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

var envPattern = regexp.MustCompile(`env\(([A-Za-z_][A-Za-z0-9_]*)\)`)

func LoadConfig() (*LitmusConfig, error) {
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

	v.SetEnvPrefix("LITMUS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigName(defaultConfigName)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if errors.Is(err, &notFoundErr) {
			fmt.Fprintf(os.Stderr, "WARN: .litmus.yaml not found, using defaults\n")
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg LitmusConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.expandEnv(); err != nil {
		return nil, fmt.Errorf("failed to expand environment variables: %w", err)
	}

	return &cfg, nil
}

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

func loadAlertmanagerConfigYAML(path string) (*AlertmanagerConfig, error) {
	expanded, err := expandAlertmanagerConfig(path)
	if err != nil {
		return nil, err
	}

	var cfg AlertmanagerConfig
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing alertmanager config: %w", err)
	}

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

func expandEnvVars(s string) (string, error) {
	if !strings.Contains(s, "env(") {
		return s, nil
	}
	var expandErr error
	result := envPattern.ReplaceAllStringFunc(s, func(match string) string {
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

func (c *LitmusConfig) FilePath() string {
	primary := filepath.Join(c.Workspace.Root, defaultConfigFile)
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	alt := filepath.Join(c.Workspace.Root, defaultConfigFileAlt)
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return primary
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

func (c *LitmusConfig) LoadAssembledConfig() (*AlertmanagerConfig, []*Fragment, *amconfig.Config, error) {
	base, err := loadAlertmanagerConfigYAML(c.FilePath())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading alertmanager config: %w", err)
	}

	fragments, err := LoadFragments(c.FragmentsPath())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading fragments: %w", err)
	}

	var baseRoutes []*amconfig.Route
	if base.Route != nil {
		baseRoutes = make([]*amconfig.Route, len(base.Route.Routes))
		copy(baseRoutes, base.Route.Routes)
	}

	rootTests, _ := behavioral.NewBehavioralTestLoader().LoadFromDirectory(c.TestsDir())

	rootFrag := &Fragment{
		Name:   "root",
		Tests:  rootTests,
		Routes: baseRoutes,
	}

	allFragments := append([]*Fragment{rootFrag}, fragments...)

	if len(fragments) == 0 {
		amCfg, _ := ToAMConfig(base)
		return base, allFragments, amCfg, nil
	}

	assembler := NewAssembler(base)
	assembled, err := assembler.Assemble(fragments)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("assembling fragments: %w", err)
	}

	amCfg, _ := ToAMConfig(assembled)
	return assembled, allFragments, amCfg, nil
}
