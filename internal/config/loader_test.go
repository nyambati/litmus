package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
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

	content := `
mimir:
  address: "https://mimir.example.com"
  api_key: "env(MY_MIMIR_TOKEN)"
`
	err := os.WriteFile(".litmus.yaml", []byte(content), 0644)
	require.NoError(t, err)
	defer os.Remove(".litmus.yaml")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://mimir.example.com", cfg.Mimir.Address)
	assert.Equal(t, "secret-token", cfg.Mimir.APIKey)
}

func TestLoadConfig_EnvSubstitution_Unset(t *testing.T) {
	os.Clearenv()

	content := `
mimir:
  api_key: "env(MISSING_VAR)"
`
	err := os.WriteFile(".litmus.yaml", []byte(content), 0644)
	require.NoError(t, err)
	defer os.Remove(".litmus.yaml")

	_, err = LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MISSING_VAR")
}

func TestLoadAlertmanagerConfig_EnvSubstitution(t *testing.T) {
	os.Setenv("AM_OPSGENIE_KEY", "test-key-123")
	defer os.Unsetenv("AM_OPSGENIE_KEY")

	content := `
global:
  resolve_timeout: 5m
  opsgenie_api_key: "env(AM_OPSGENIE_KEY)"
route:
  receiver: 'default'
receivers:
  - name: 'default'
`
	f, err := os.CreateTemp("", "alertmanager-*.yml")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString(content)
	require.NoError(t, err)
	f.Close()

	cfg, err := LoadAlertmanagerConfig(f.Name())
	require.NoError(t, err)
	assert.Equal(t, "test-key-123", string(cfg.Global.OpsGenieAPIKey))
}

func TestExpandEnvVars(t *testing.T) {
	os.Setenv("MY_VAR", "hello")
	os.Setenv("OTHER_VAR", "world")
	defer os.Unsetenv("MY_VAR")
	defer os.Unsetenv("OTHER_VAR")

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "no placeholder",
			input: "plain string",
			want:  "plain string",
		},
		{
			name:  "single placeholder",
			input: "env(MY_VAR)",
			want:  "hello",
		},
		{
			name:  "placeholder in url",
			input: "https://example.com/env(MY_VAR)/path",
			want:  "https://example.com/hello/path",
		},
		{
			name:  "multiple placeholders",
			input: "env(MY_VAR) env(OTHER_VAR)",
			want:  "hello world",
		},
		{
			name:  "lowercase var name uppercased before lookup",
			input: "env(my_var)",
			want:  "hello",
		},
		{
			name:  "mixed case var name uppercased before lookup",
			input: "env(My_Var)",
			want:  "hello",
		},
		{
			name:    "unset var returns error",
			input:   "env(UNDEFINED_VAR)",
			wantErr: true,
		},
		{
			name:  "alertmanager template blocks untouched",
			input: `{{ template "slack.title" . }}`,
			want:  `{{ template "slack.title" . }}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandEnvVars(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
