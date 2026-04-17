package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup minimal config
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

	os.MkdirAll("tests", 0755)

	cmd := newCheckCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)
}

func TestCheckCommand_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Should fail because config/alertmanager.yml is missing (even with default litmus config)
	cmd := newCheckCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	require.Error(t, err)
}

func TestCheckCommand_TextOutput(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Setup minimal config
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

	os.MkdirAll("tests", 0755)

	cmd := newCheckCmd()
	cmd.SetArgs([]string{"--format", "text"})
	err = cmd.Execute()

	require.NoError(t, err)
}
