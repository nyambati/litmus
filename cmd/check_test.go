package cmd

import (
	"os"
	"testing"

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
config:
  directory: "config"
  file: "alertmanager.yml"
global_labels:
  severity: "warning"
regression:
  keep: 3
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

	err = os.MkdirAll("tests", 0755)
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
config:
  directory: "config"
  file: "alertmanager.yml"
global_labels:
  severity: "warning"
regression:
  keep: 3
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

	err = os.MkdirAll("tests", 0755)
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
config:
  directory: "config"
  file: "alertmanager.yml"
global_labels:
  severity: "warning"
regression:
  keep: 3
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

	err = os.MkdirAll("tests", 0755)
	require.NoError(t, err)
	err = os.WriteFile("tests/test1.yml", []byte(`
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
	err = os.WriteFile("tests/test2.yml", []byte(`
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
