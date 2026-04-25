package cmd

import (
	"os"
	"testing"

	"github.com/nyambati/litmus/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Setup minimal config
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

	err = os.MkdirAll("config/tests", 0755)
	require.NoError(t, err)

	cmd := newCheckCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)
}

func TestCheckCommand_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Should fail because config/alertmanager.yml is missing (even with default litmus config)
	cmd := newCheckCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.Error(t, err)
}

func TestCheckCommand_TextOutput(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Setup minimal config
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

	err = os.MkdirAll("config/tests", 0755)
	require.NoError(t, err)

	cmd := newCheckCmd()
	cmd.SetArgs([]string{"--format", "text"})
	err = cmd.Execute()

	require.NoError(t, err)
}

func TestCheckCommand_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Setup minimal config
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

	err = os.MkdirAll("config/tests", 0755)
	require.NoError(t, err)
	err = os.WriteFile("config/tests/test1.yml", []byte(`
name: "test critical"
tags:
  - critical
state:
  active_alerts: []
  silences: []
alert:
  labels:
    alertname: "TestAlert"
expect:
  outcome: active
  receivers:
    - default
`), 0600)
	require.NoError(t, err)
	err = os.WriteFile("config/tests/test2.yml", []byte(`
name: "test smoke"
tags:
  - smoke
state:
  active_alerts: []
  silences: []
alert:
  labels:
    alertname: "TestAlert2"
expect:
  outcome: active
  receivers:
    - default
`), 0600)
	require.NoError(t, err)

	cmd := newCheckCmd()
	cmd.SetArgs([]string{"--tags", "critical"})
	err = cmd.Execute()

	require.NoError(t, err)
}

func TestCheckCommand_Policy_RequireTests_Root(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tmpDir))

	// Policy: require_tests=true, no skip_root
	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
policy:
  require_tests: true
`), 0600))
	require.NoError(t, os.MkdirAll("config/tests", 0755))
	require.NoError(t, os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600))

	// No tests in config/tests/ → root fails require_tests (exit code 3 = sanity failure)
	// Use cli.RunCheck directly — cmd.Execute() calls os.Exit for non-zero codes.
	code, err := cli.RunCheck("text", false, nil)
	require.NoError(t, err)
	assert.NotEqual(t, cli.CheckExitCode(0), code, "sanity must fail when root has no tests and require_tests=true")

	// Add a test file → root now satisfies require_tests
	require.NoError(t, os.WriteFile("config/tests/base.yml", []byte(`
name: "base test"
alert:
  labels:
    alertname: "Test"
expect:
  outcome: active
  receivers:
    - default
`), 0600))

	code, err = cli.RunCheck("text", false, nil)
	require.NoError(t, err)
	assert.Equal(t, cli.CheckExitCode(0), code)
}

func TestCheckCommand_Policy_SkipRoot(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	require.NoError(t, os.Chdir(tmpDir))

	// Policy: require_tests=true WITH skip_root — root violation must be suppressed
	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
policy:
  require_tests: true
  skip_root: true
`), 0600))
	require.NoError(t, os.MkdirAll("config/tests", 0755))
	require.NoError(t, os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600))

	// No tests in root, but skip_root=true → should pass
	code, err := cli.RunCheck("text", false, nil)
	require.NoError(t, err)
	assert.Equal(t, cli.CheckExitCode(0), code, "root violations must be suppressed with skip_root=true")
}
