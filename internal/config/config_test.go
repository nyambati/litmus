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

	cfg, err := New()
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

	cfg, err := New()
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

	cfg, err := New()
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

	cfg, err := New()
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

	_, err = New()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MISSING_VAR")
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

			cfg, err := New()
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
