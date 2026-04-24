package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const historyDir = "regressions"

func TestSnapshotCommand_GeneratesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create minimal litmus.yaml
	err = os.WriteFile(".litmus.yaml", []byte(`
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
`), 0600)
	require.NoError(t, err)

	// Create minimal alertmanager.yml
	err = os.MkdirAll("config", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600)
	require.NoError(t, err)

	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)

	// Verify regression state file exists with ID and tests
	require.FileExists(t, filepath.Join("regressions", "regressions.litmus.yml"))

	// Verify regressions.litmus.yml contains ID
	data, err := os.ReadFile(filepath.Join("regressions", "regressions.litmus.yml"))
	require.NoError(t, err)
	require.Contains(t, string(data), "id:")
	require.Contains(t, string(data), "tests:")
}

func TestSnapshotCommand_DriftDetection_Graceful(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create config
	err = os.WriteFile(".litmus.yaml", []byte(`
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
`), 0600)
	require.NoError(t, err)

	err = os.MkdirAll("config", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600)
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
`), 0600)
	require.NoError(t, err)

	// Run snapshot again without --strict (should succeed with warning)
	cmd = newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)
}

func TestSnapshotCommand_Strict_DriftDetection(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create config
	err = os.WriteFile(".litmus.yaml", []byte(`
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
`), 0600)
	require.NoError(t, err)

	err = os.MkdirAll("config", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600)
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
`), 0600)
	require.NoError(t, err)

	// Run snapshot with --strict (should fail due to drift)
	cmd = newSnapshotCmd()
	cmd.SetArgs([]string{"--strict"})
	err = cmd.Execute()

	require.Error(t, err)
	require.Contains(t, err.Error(), "drift")
}

func TestSnapshotCommand_UpdateFlag(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create config
	err = os.WriteFile(".litmus.yaml", []byte(`
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
`), 0600)
	require.NoError(t, err)

	err = os.MkdirAll("config", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600)
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
`), 0600)
	require.NoError(t, err)

	// Run snapshot with --update (should succeed)
	cmd = newSnapshotCmd()
	cmd.SetArgs([]string{"--update"})
	err = cmd.Execute()

	require.NoError(t, err)

	// Verify a version was created in history
	entries, err := os.ReadDir("regressions")
	require.NoError(t, err)
	hasVersion := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			hasVersion = true
			break
		}
	}
	require.True(t, hasVersion, "should have created a versioned baseline")
}

func TestSnapshotCommand_UpdateWithNoDrift(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create config
	err = os.WriteFile(".litmus.yaml", []byte(`
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
`), 0600)
	require.NoError(t, err)

	err = os.MkdirAll("config", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600)
	require.NoError(t, err)

	// Generate initial baseline
	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	// Count history entries (should be 1)
	historyDir := "regressions"
	entries, err := os.ReadDir(historyDir)
	require.NoError(t, err)
	mpkCount := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			mpkCount++
		}
	}
	require.Equal(t, 1, mpkCount, "should have exactly 1 baseline version after initial snapshot")

	// Run snapshot with --update (no config changes, so no drift)
	cmd = newSnapshotCmd()
	cmd.SetArgs([]string{"--update"})
	err = cmd.Execute()

	require.NoError(t, err)

	// Count history entries again (should still be 1, no new backup created)
	entries, err = os.ReadDir(historyDir)
	require.NoError(t, err)
	mpkCount = 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			mpkCount++
		}
	}
	require.Equal(t, 1, mpkCount, "should still have exactly 1 baseline version (no new version created for no-drift case)")
}

func TestSnapshotCommand_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	// Should fail because config/alertmanager.yml is missing
	require.Error(t, err)
	require.Contains(t, err.Error(), "alertmanager.yml")
}

func TestSnapshotHistory_ListsBaselines(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create config
	err = os.WriteFile(".litmus.yaml", []byte(`
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
`), 0600)
	require.NoError(t, err)

	err = os.MkdirAll("config", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600)
	require.NoError(t, err)

	// Generate initial baseline
	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	// Check history is recorded
	historyDir := "regressions"
	entries, err := os.ReadDir(historyDir)
	require.NoError(t, err)
	versionCount := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			versionCount++
		}
	}
	require.Greater(t, versionCount, 0, "should have at least one version")

	// List history using new history command
	cmd = newHistoryCmd()
	cmd.SetArgs([]string{"list"})
	err = cmd.Execute()
	require.NoError(t, err)
}

func TestSnapshotRollback_RestoresPreviousBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create config
	err = os.WriteFile(".litmus.yaml", []byte(`
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
`), 0600)
	require.NoError(t, err)

	err = os.MkdirAll("config", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600)
	require.NoError(t, err)

	// Generate initial baseline
	cmd := newSnapshotCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	require.NoError(t, err)

	// Read the history to get the first baseline ID

	entries, err := os.ReadDir(historyDir)
	require.NoError(t, err)

	var firstID string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".mpk") && !strings.HasPrefix(name, "regressions.litmus") {
			firstID = strings.TrimSuffix(name, ".mpk")
			break
		}
	}
	require.NotEmpty(t, firstID, "should have created at least one version")

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
`), 0600)
	require.NoError(t, err)

	// Update baseline
	cmd = newSnapshotCmd()
	cmd.SetArgs([]string{"--update"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Read the second baseline
	entries, err = os.ReadDir(historyDir)
	require.NoError(t, err)
	require.Greater(t, len(entries), 0)

	// Rollback to first baseline using new history command
	cmd = newHistoryCmd()
	cmd.SetArgs([]string{"rollback", firstID})
	err = cmd.Execute()
	require.NoError(t, err)
}
