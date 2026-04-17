package cmd

import (
	"os"
	"testing"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/types"
	"github.com/stretchr/testify/require"
)

func TestInspectCommand_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Create test baseline
	tests := []*types.RegressionTest{
		{
			Name:     "Test 1",
			Labels:   []map[string]string{{"service": "api"}},
			Expected: []string{"api-team"},
			Tags:     []string{"regression"},
		},
	}

	mpkFile, _ := os.Create("test.mpk")
	codec.EncodeMsgPack(mpkFile, tests)
	mpkFile.Close()

	cmd := newInspectCmd()
	cmd.SetArgs([]string{"test.mpk"})
	err := cmd.Execute()

	require.NoError(t, err)
}

func TestInspectCommand_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Create test baseline
	tests := []*types.RegressionTest{
		{
			Name:     "Test 1",
			Labels:   []map[string]string{{"service": "api"}},
			Expected: []string{"api-team"},
			Tags:     []string{"regression"},
		},
	}

	mpkFile, _ := os.Create("test.mpk")
	codec.EncodeMsgPack(mpkFile, tests)
	mpkFile.Close()

	cmd := newInspectCmd()
	cmd.SetArgs([]string{"test.mpk", "--format", "json"})
	err := cmd.Execute()

	require.NoError(t, err)
}

func TestInspectCommand_MissingFile(t *testing.T) {
	cmd := newInspectCmd()
	cmd.SetArgs([]string{"nonexistent.mpk"})
	err := cmd.Execute()

	require.Error(t, err)
}
