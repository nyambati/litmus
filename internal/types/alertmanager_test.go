package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func minimalConfig() *AlertmanagerConfig {
	return &AlertmanagerConfig{
		Receivers: []*Receiver{
			{Name: "default"},
		},
	}
}

func TestAlertmanagerConfig_Marshal_ValidConfig(t *testing.T) {
	cfg := minimalConfig()

	data, err := cfg.Marshal()

	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "receivers:")
	assert.Contains(t, string(data), "default")
}

func TestAlertmanagerConfig_Marshal_EnvVarExpansion(t *testing.T) {
	t.Setenv("LITMUS_TEST_WEBHOOK", "http://example.com/hook")

	cfg := &AlertmanagerConfig{
		Receivers: []*Receiver{
			{
				Name: "webhook-receiver",
				WebhookConfigs: []map[string]any{
					{"url": "env(litmus_test_webhook)"},
				},
			},
		},
	}

	data, err := cfg.Marshal()

	require.NoError(t, err)
	assert.Contains(t, string(data), "http://example.com/hook",
		"Marshal must expand env() references in the output")
	assert.NotContains(t, string(data), "env(litmus_test_webhook)")
}

func TestAlertmanagerConfig_Marshal_MissingEnvVar(t *testing.T) {
	cfg := &AlertmanagerConfig{
		Receivers: []*Receiver{
			{
				Name: "webhook-receiver",
				WebhookConfigs: []map[string]any{
					{"url": "env(litmus_test_unset_var_xyz)"},
				},
			},
		},
	}

	_, err := cfg.Marshal()

	require.Error(t, err, "Marshal must return an error when a referenced env var is not set")
	assert.Contains(t, strings.ToLower(err.Error()), "env var")
}

func TestAlertmanagerConfig_String_BackwardCompatible(t *testing.T) {
	cfg := minimalConfig()

	s := cfg.String()

	assert.NotEmpty(t, s, "String() must return non-empty YAML for a valid config")
	assert.Contains(t, s, "receivers:")
}

func TestAlertmanagerConfig_String_ReturnsEmptyOnMissingEnvVar(t *testing.T) {
	cfg := &AlertmanagerConfig{
		Receivers: []*Receiver{
			{
				Name: "r",
				WebhookConfigs: []map[string]any{
					{"url": "env(litmus_test_unset_string_var_xyz)"},
				},
			},
		},
	}

	// String() swallows errors (backward compat) — returns "" on failure.
	s := cfg.String()
	assert.Empty(t, s, "String() must return empty string when env expansion fails")
}
