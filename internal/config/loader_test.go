package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Ensure no env vars interfere
	os.Clearenv()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "config", cfg.Config.Directory)
	assert.Equal(t, "alertmanager.yml", cfg.Config.File)
	assert.Equal(t, "templates/", cfg.Config.Templates)
	assert.Equal(t, "regressions", cfg.Regression.Directory)
	assert.Equal(t, 5, cfg.Regression.MaxSamples)
	assert.Equal(t, "tests", cfg.Tests.Directory)
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	os.Clearenv()
	os.Setenv("LITMUS_CONFIG_DIRECTORY", "custom-config")
	os.Setenv("LITMUS_MIMIR_ADDRESS", "https://mimir.io")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "custom-config", cfg.Config.Directory)
	assert.Equal(t, "https://mimir.io", cfg.Mimir.Address)
}

func TestLoadConfig_EnvSubstitution(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_MIMIR_TOKEN", "secret-token")

	// Create a temporary .litmus.yaml
	content := `
mimir:
  address: "https://mimir.example.com"
  api_key: "{{ env \"MY_MIMIR_TOKEN\" }}"
`
	err := os.WriteFile(".litmus.yaml", []byte(content), 0644)
	require.NoError(t, err)
	defer os.Remove(".litmus.yaml")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://mimir.example.com", cfg.Mimir.Address)
	assert.Equal(t, "secret-token", cfg.Mimir.APIKey)
}
