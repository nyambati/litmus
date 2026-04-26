package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitCommand_CreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)
	require.FileExists(t, ".litmus.yaml")
	require.DirExists(t, "config")
	require.DirExists(t, "config/templates")
	require.DirExists(t, "config/regressions")
	require.DirExists(t, "config/tests")
}

func TestInitCommand_LitmusYAML(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)

	content, err := os.ReadFile(".litmus.yaml")
	require.NoError(t, err)
	require.Contains(t, string(content), "workspace:")
	require.Contains(t, string(content), "root: \"config\"")
	require.Contains(t, string(content), "history: 5")
	require.Contains(t, string(content), "global_labels:")
}

func TestInitCommand_TestsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.NoError(t, err)
	require.FileExists(t, filepath.Join("config", "tests", "README.md"))
}

func TestInitCommand_DoesNotOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a pre-existing litmus.yaml
	existingContent := "existing: true\n"
	err = os.WriteFile(".litmus.yaml", []byte(existingContent), 0600)
	require.NoError(t, err)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")

	// Verify original content is preserved
	content, _ := os.ReadFile(".litmus.yaml")
	require.Equal(t, existingContent, string(content))
}
