package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnapshotCommand_GeneratesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Create minimal litmus.yaml
	err := os.WriteFile(".litmus.yaml", []byte(`
config:
  directory: "config"
  file: "alertmanager.yml"
global_labels:
  severity: "warning"
regression:
  max_samples: 5
  directory: "regressions"
tests:
  directory: "tests/"
`), 0644)
	require.NoError(t, err)

	// Create minimal alertmanager.yml
	os.MkdirAll("config", 0755)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0644)
	require.NoError(t, err)

	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)
	require.FileExists(t, filepath.Join("regressions", "regressions.litmus.mpk"))
	require.FileExists(t, filepath.Join("regressions", "regressions.litmus.yml"))
}

func TestSnapshotCommand_DriftDetection(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Create config
	err := os.WriteFile(".litmus.yaml", []byte(`
config:
  directory: "config"
  file: "alertmanager.yml"
global_labels:
  severity: "warning"
regression:
  max_samples: 5
  directory: "regressions"
tests:
  directory: "tests/"
`), 0644)
	require.NoError(t, err)

	os.MkdirAll("config", 0755)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0644)
	require.NoError(t, err)

	// Generate initial baseline
	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	// Modify alertmanager config
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
  routes:
    - receiver: 'api-team'
      match:
        service: 'api'
receivers:
  - name: 'default'
  - name: 'api-team'
`), 0644)
	require.NoError(t, err)

	// Run snapshot again without --update (should fail due to drift)
	cmd = newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.Error(t, err)
	require.Contains(t, err.Error(), "drift")
}

func TestSnapshotCommand_UpdateFlag(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Create config
	err := os.WriteFile(".litmus.yaml", []byte(`
config:
  directory: "config"
  file: "alertmanager.yml"
global_labels:
  severity: "warning"
regression:
  max_samples: 5
  directory: "regressions"
tests:
  directory: "tests/"
`), 0644)
	require.NoError(t, err)

	os.MkdirAll("config", 0755)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0644)
	require.NoError(t, err)

	// Generate initial baseline
	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	// Modify alertmanager config
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
  routes:
    - receiver: 'api-team'
      match:
        service: 'api'
receivers:
  - name: 'default'
  - name: 'api-team'
`), 0644)
	require.NoError(t, err)

	// Run snapshot with --update (should succeed)
	cmd = newSnapshotCmd()
	cmd.SetArgs([]string{"--update"})
	err = cmd.Execute()

	require.NoError(t, err)
	require.FileExists(t, filepath.Join("regressions", "regressions.litmus.mpk"))
}

func TestSnapshotCommand_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	// Should fail because config/alertmanager.yml is missing
	require.Error(t, err)
	require.Contains(t, err.Error(), "alertmanager.yml")
}
