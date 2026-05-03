package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const historyDir = "config/regressions"

func TestSnapshotCommand_GeneratesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create minimal litmus.yaml
	err = os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()

	require.NoError(t, err)

	// Verify regression state file exists with ID and tests in the correct Package-First location
	require.FileExists(t, filepath.Join(historyDir, "regressions.litmus.yml"))

	// Verify regressions.litmus.yml contains ID
	data, err := os.ReadFile(filepath.Join(historyDir, "regressions.litmus.yml"))
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
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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
	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
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
	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
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
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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
	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
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
	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture", "--strict"})
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
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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
	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
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

	// Run snapshot with update (should succeed)
	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"update"})
	err = cmd.Execute()

	require.NoError(t, err)

	// Verify a version was created in history
	entries, err := os.ReadDir(historyDir)
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
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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
	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Count history entries (should be 1)
	entries, err := os.ReadDir(historyDir)
	require.NoError(t, err)
	mpkCount := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			mpkCount++
		}
	}
	require.Equal(t, 1, mpkCount, "should have exactly 1 baseline version after initial snapshot")

	// Run snapshot with update (no config changes, so no drift)
	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"update"})
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

	// Create minimal litmus.yaml so that the loader can find the workspace.
	// The command should then fail because no entrypoint can be discovered.
	err = os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
`), 0600)
	require.NoError(t, err)

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()

	// Should fail because no base or alertmanager config exists in the root.
	require.Error(t, err)
	require.Contains(t, err.Error(), "found 0 files matching")
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
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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
	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Check history is recorded
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
	cmd = withConfig(newHistoryCmd())
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
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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
	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
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
	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"update"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Read the second baseline
	entries, err = os.ReadDir(historyDir)
	require.NoError(t, err)
	require.Greater(t, len(entries), 0)

	// Rollback to first baseline using new history command
	cmd = withConfig(newHistoryCmd())
	cmd.SetArgs([]string{"rollback", firstID})
	err = cmd.Execute()
	require.NoError(t, err)
}

func TestSnapshotCommand_WithFragments(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tmpDir))

	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  fragments: "fragments/*"
  history: 3
global_labels:
  severity: "warning"
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

	// Fragment uses group to create a synthetic parent under scope=teams.
	// Namespace "db-team" → assembled receiver name is "db-team-db-critical".
	require.NoError(t, os.MkdirAll("config/fragments/db-team", 0755))
	require.NoError(t, os.WriteFile("config/fragments/db-team/fragment.yml", []byte(`
namespace: "db-team"
group:
  match:
    scope: "teams"
routes:
  - receiver: "db-critical"
    match:
      service: "mysql"
receivers:
  - name: "db-critical"
`), 0600))

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	require.NoError(t, cmd.Execute())

	data, err := os.ReadFile(filepath.Join("config", "regressions", "regressions.litmus.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "db-team-db-critical", "snapshot must capture routes from assembled fragments with namespace prefix")
}

func countHistoryEntries(t *testing.T) int {
	t.Helper()
	entries, _ := os.ReadDir(historyDir)
	count := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			count++
		}
	}
	return count
}

func getBaselineID(t *testing.T, ymlPath string) string {
	t.Helper()
	data, err := os.ReadFile(ymlPath)
	require.NoError(t, err)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "id:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		}
	}
	return ""
}

func TestSnapshotCapture_NoHistory_CreatesBaselineAndHistory(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	require.FileExists(t, filepath.Join(historyDir, "regressions.litmus.yml"))

	historyCount := countHistoryEntries(t)
	require.Equal(t, 1, historyCount, "should have exactly 1 history entry")

	firstID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))
	require.NotEmpty(t, firstID, "baseline should have an ID")
}

func TestSnapshotCapture_WithHistory_DriftDoesNotUpdateBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	initialID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))
	historyCountBefore := countHistoryEntries(t)

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

	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	currentID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))
	historyCountAfter := countHistoryEntries(t)

	require.Equal(t, initialID, currentID, "baseline ID should NOT change when drift detected during capture")
	require.Equal(t, historyCountBefore, historyCountAfter, "history count should NOT change when drift detected during capture")
}

func TestSnapshotCapture_WithHistory_NoDriftReportsBaselineCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	initialID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))

	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	currentID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))

	require.Equal(t, initialID, currentID, "baseline ID should not change when baseline is current")
}

func TestSnapshotUpdate_WithDrift_CreatesNewHistory(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	initialID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))
	historyCountBefore := countHistoryEntries(t)

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

	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"update"})
	err = cmd.Execute()
	require.NoError(t, err)

	currentID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))
	historyCountAfter := countHistoryEntries(t)

	require.NotEqual(t, initialID, currentID, "baseline ID should change after update with drift")
	require.Equal(t, historyCountBefore+1, historyCountAfter, "should create new history entry")
}

func TestSnapshotUpdate_WithNoDrift_ReportsNoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  history: 3
global_labels:
  severity: "warning"
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

	cmd := withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"capture"})
	err = cmd.Execute()
	require.NoError(t, err)

	initialID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))
	historyCountBefore := countHistoryEntries(t)

	cmd = withConfig(newSnapshotCmd())
	cmd.SetArgs([]string{"update"})
	err = cmd.Execute()
	require.NoError(t, err)

	currentID := getBaselineID(t, filepath.Join(historyDir, "regressions.litmus.yml"))
	historyCountAfter := countHistoryEntries(t)

	require.Equal(t, initialID, currentID, "baseline ID should not change")
	require.Equal(t, historyCountBefore, historyCountAfter, "history count should not change")
}
