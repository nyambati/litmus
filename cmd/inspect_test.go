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
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "Test 1",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
			Tags:   []string{"regression"},
		},
	}

	mpkFile, err := os.Create("test.mpk")
	require.NoError(t, err)
	err = codec.EncodeMsgPack(mpkFile, tests)
	require.NoError(t, err)
	err = mpkFile.Close()
	require.NoError(t, err)

	cmd := newInspectCmd()
	cmd.SetArgs([]string{"test.mpk"})
	err = cmd.Execute()

	require.NoError(t, err)
}

func TestInspectCommand_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "Test 1",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
			Tags:   []string{"regression"},
		},
	}

	mpkFile, err := os.Create("test.mpk")
	require.NoError(t, err)
	err = codec.EncodeMsgPack(mpkFile, tests)
	require.NoError(t, err)
	err = mpkFile.Close()
	require.NoError(t, err)

	cmd := newInspectCmd()
	cmd.SetArgs([]string{"test.mpk", "--format", "json"})
	err = cmd.Execute()

	require.NoError(t, err)
}

func TestInspectCommand_MissingFile(t *testing.T) {
	cmd := newInspectCmd()
	cmd.SetArgs([]string{"nonexistent.mpk"})
	err := cmd.Execute()

	require.Error(t, err)
}
