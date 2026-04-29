package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func chdirTemp(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })
	require.NoError(t, os.Chdir(tmpDir))
}

func TestLoadConfig_Defaults(t *testing.T) {
	os.Clearenv()
	chdirTemp(t)

	err := os.WriteFile(".litmus.yaml", []byte{}, 0600)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll("config", 0755))
	require.NoError(t, os.WriteFile(filepath.Join("config", "alertmanager.yml"), []byte{}, 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, defaultConfigDir, cfg.Workspace.Root)
	assert.Equal(t, defaultFragmentsPattern, cfg.Workspace.Fragments)
	assert.Equal(t, defaultHistoryKeep, cfg.Workspace.History)
	assert.Equal(t, SanityModeFail, cfg.Sanity.NegativeOnlyRoutes)
}

func TestLoadConfig_NegativeOnlyRoutesSanityMode(t *testing.T) {
	os.Clearenv()
	chdirTemp(t)

	err := os.WriteFile(".litmus.yaml", []byte(`
sanity:
  negative_only_routes: warn
`), 0600)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll("config", 0755))
	require.NoError(t, os.WriteFile(filepath.Join("config", "alertmanager.yml"), []byte{}, 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, SanityModeWarn, cfg.Sanity.NegativeOnlyRoutes)
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	os.Clearenv()
	chdirTemp(t)

	os.Setenv("LITMUS_WORKSPACE_ROOT", "custom-root")
	os.Setenv("LITMUS_MIMIR_ADDRESS", "https://mimir.io")

	err := os.WriteFile(".litmus.yaml", []byte{}, 0600)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll("custom-root", 0755))
	require.NoError(t, os.WriteFile(filepath.Join("custom-root", "alertmanager.yml"), []byte{}, 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "custom-root", cfg.Workspace.Root)
	assert.Equal(t, "https://mimir.io", cfg.Mimir.Address)
}

func TestLoadConfig_EnvSubstitution(t *testing.T) {
	os.Clearenv()
	chdirTemp(t)

	os.Setenv("MY_MIMIR_TOKEN", "secret-token")

	content := `
mimir:
  address: "https://mimir.example.com"
  api_key: "env(MY_MIMIR_TOKEN)"
`
	err := os.WriteFile(".litmus.yaml", []byte(content), 0600)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll("config", 0755))
	require.NoError(t, os.WriteFile(filepath.Join("config", "alertmanager.yml"), []byte{}, 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://mimir.example.com", cfg.Mimir.Address)
	assert.Equal(t, "secret-token", cfg.Mimir.APIKey)
}

func TestLoadConfig_EnvSubstitution_Unset(t *testing.T) {
	os.Clearenv()
	chdirTemp(t)

	content := `
mimir:
  api_key: "env(MISSING_VAR)"
`
	err := os.WriteFile(".litmus.yaml", []byte(content), 0600)
	require.NoError(t, err)

	_, err = LoadConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MISSING_VAR")
}

func TestAlertmanagerConfig_EnvSubstitution(t *testing.T) {
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

	cfg, err := loadAlertmanagerConfigYAML(f.Name())
	require.NoError(t, err)
	assert.Equal(t, "test-key-123", cfg.Global.OpsGenieAPIKey)
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

func TestFilePath_EntrypointCandidates(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing files returns discovery error",
			wantErr:     true,
			errContains: "found 0 files matching",
		},
		{
			name:        "non yaml alertmanager file is ignored",
			files:       []string{"alertmanager.toml"},
			wantErr:     true,
			errContains: "found 0 files matching",
		},
		{
			name:        "database yaml does not match base substring",
			files:       []string{"database.yml"},
			wantErr:     true,
			errContains: "found 0 files matching",
		},
		{
			name:        "baseline yaml does not match base substring",
			files:       []string{"baseline.yml"},
			wantErr:     true,
			errContains: "found 0 files matching",
		},
		{
			name:        "alertmanager backup yaml does not match alertmanager substring",
			files:       []string{"alertmanager-backup.yml"},
			wantErr:     true,
			errContains: "found 0 files matching",
		},
		{
			name:        "unrelated yaml beside one entrypoint is ignored",
			files:       []string{"database.yml", "alertmanager.yml"},
			want:        filepath.Join("config", "alertmanager.yml"),
			errContains: "",
		},
		{
			name:  "alertmanager yml is found",
			files: []string{"alertmanager.yml"},
			want:  filepath.Join("config", "alertmanager.yml"),
		},
		{
			name:  "alertmanager yaml is found",
			files: []string{"alertmanager.yaml"},
			want:  filepath.Join("config", "alertmanager.yaml"),
		},
		{
			name:  "base yml is found",
			files: []string{"base.yml"},
			want:  filepath.Join("config", "base.yml"),
		},
		{
			name:  "base yaml is found",
			files: []string{"base.yaml"},
			want:  filepath.Join("config", "base.yaml"),
		},
		{
			name:        "multiple matching stems return discovery error",
			files:       []string{"base.yml", "alertmanager.yaml"},
			wantErr:     true,
			errContains: "found 2 files matching",
		},
		{
			name:        "multiple matching extensions return discovery error",
			files:       []string{"base.yaml", "base.yml"},
			wantErr:     true,
			errContains: "found 2 files matching",
		},
		{
			name:        "multiple matching candidates return discovery error",
			files:       []string{"base.yml", "base.yaml", "alertmanager.yaml", "alertmanager.yml"},
			wantErr:     true,
			errContains: "found 4 files matching",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldCwd, err := os.Getwd()
			require.NoError(t, err)
			t.Cleanup(func() { _ = os.Chdir(oldCwd) })
			require.NoError(t, os.Chdir(tmpDir))

			require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
`), 0600))
			require.NoError(t, os.MkdirAll("config", 0755))
			for _, file := range tt.files {
				require.NoError(t, os.WriteFile(filepath.Join("config", file), []byte{}, 0600))
			}

			cfg, err := LoadConfig()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.want, cfg.FilePath())
		})
	}
}

func TestLoadAssembledConfig_NoFragments(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tmpDir))

	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  fragments: "fragments/*"
`), 0600))
	require.NoError(t, os.MkdirAll("config", 0755))
	require.NoError(t, os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600))
	require.NoError(t, os.MkdirAll("config/fragments", 0755))

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assembled, fragments, _, err := cfg.LoadAssembledConfig()
	require.NoError(t, err)
	assert.NotNil(t, assembled)
	assert.Equal(t, "default", assembled.Route.Receiver)
	// no team fragments — root fragment always returned for policy checks
	require.Len(t, fragments, 1)
	assert.Equal(t, "root", fragments[0].Name)
}

func TestLoadAssembledConfig_WithFragments(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tmpDir))

	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  fragments: "fragments/*"
`), 0600))
	require.NoError(t, os.MkdirAll("config/fragments", 0755))
	require.NoError(t, os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join("config", "fragments", "db.yml"), []byte(`
name: "db-team"
namespace: "db"
group:
  match:
    scope: "teams"
routes:
  - receiver: "critical"
    match:
      service: "mysql"
receivers:
  - name: "critical"
`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join("config", "fragments", "db-tests.yml"), []byte(`
- name: "mysql test"
  expect: {outcome: active}
`), 0600))

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assembled, fragments, _, err := cfg.LoadAssembledConfig()
	require.NoError(t, err)

	// Namespace applied: receiver renamed db-critical
	receiverNames := make([]string, 0, len(assembled.Receivers))
	for _, r := range assembled.Receivers {
		receiverNames = append(receiverNames, r.Name)
	}
	assert.Contains(t, receiverNames, "db-critical")

	// Group creates synthetic parent with scope=teams; db-critical route is its child
	require.Len(t, assembled.Route.Routes, 1)
	teamsRoute := assembled.Route.Routes[0]
	assert.Equal(t, map[string]string{"scope": "teams"}, teamsRoute.Match)
	require.Len(t, teamsRoute.Routes, 1)
	assert.Equal(t, "db-critical", teamsRoute.Routes[0].Receiver)

	// root frag + db-team frag returned; root has no pre-assembly sub-routes
	require.Len(t, fragments, 2)
	rootFrag := fragments[0]
	assert.Equal(t, "root", rootFrag.Name)
	assert.Empty(t, rootFrag.Routes, "base config has no sub-routes")

	// Fragment test included
	var testNames []string
	for _, frag := range fragments {
		for _, tc := range frag.Tests {
			testNames = append(testNames, tc.Name)
		}
	}
	assert.Contains(t, testNames, "mysql test")
}
