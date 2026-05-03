package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func TestAlertmanagerConfig_EnvSubstitution(t *testing.T) {
// 	os.Setenv("AM_OPSGENIE_KEY", "test-key-123")
// 	defer os.Unsetenv("AM_OPSGENIE_KEY")

// 	content := `
// global:
//   resolve_timeout: 5m
//   opsgenie_api_key: "env(AM_OPSGENIE_KEY)"
// route:
//   receiver: 'default'
// receivers:
//   - name: 'default'
//  `
// 	f, err := os.CreateTemp("", "alertmanager-*.yml")
// 	require.NoError(t, err)
// 	defer os.Remove(f.Name())
// 	_, err = f.WriteString(content)
// 	require.NoError(t, err)
// 	f.Close()

// 	cfg, err := loadAlertmanagerConfigYAML(f.Name())
// 	require.NoError(t, err)
// 	assert.Equal(t, "test-key-123", cfg.Global.OpsGenieAPIKey)
// }

func TestLabelFormat(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  string
	}{
		{"nil map", nil, ""},
		{"empty map", map[string]string{}, ""},
		{"single entry", map[string]string{"env": "prod"}, "env=prod"},
		{
			"multiple entries sorted alphabetically",
			map[string]string{"service": "api", "env": "prod", "team": "platform"},
			"env=prod,service=api,team=platform",
		},
		{
			"stable output",
			map[string]string{"z": "last", "a": "first", "m": "middle"},
			"a=first,m=middle,z=last",
		},
		{
			"matches legacy groupKey output",
			map[string]string{"scope": "teams", "region": "us-east"},
			"region=us-east,scope=teams",
		},
		{
			"matches legacy LabelKey output",
			map[string]string{"alertname": "HighCPU", "severity": "critical"},
			"alertname=HighCPU,severity=critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LabelFormat(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabelFormat_Idempotent(t *testing.T) {
	m := map[string]string{"b": "2", "a": "1", "c": "3"}
	first := LabelFormat(m)
	for range 10 {
		require.Equal(t, first, LabelFormat(m), "LabelFormat must be idempotent")
	}
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
			got, err := ExpandEnvVars(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
