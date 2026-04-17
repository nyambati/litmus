package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitCommand_CreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	require.NoError(t, err)
	require.FileExists(t, ".litmus.yaml")
	require.DirExists(t, "tests")
	require.FileExists(t, ".gitattributes")
}

func TestInitCommand_LitmusYAML(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	require.NoError(t, err)

	content, err := os.ReadFile(".litmus.yaml")
	require.NoError(t, err)
	require.Contains(t, string(content), "config_file:")
	require.Contains(t, string(content), "global_labels:")
	require.Contains(t, string(content), "regression:")
	require.Contains(t, string(content), "tests:")
}

func TestInitCommand_TestsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	require.NoError(t, err)
	require.FileExists(t, filepath.Join("tests", "README.md"))
}

func TestInitCommand_GitAttributes(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	require.NoError(t, err)

	content, err := os.ReadFile(".gitattributes")
	require.NoError(t, err)
	require.Contains(t, string(content), "*.mpk")
}

func TestInitCommand_DoesNotOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Create a pre-existing litmus.yaml
	existingContent := "existing: true\n"
	err := os.WriteFile(".litmus.yaml", []byte(existingContent), 0644)
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
