package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nyambati/litmus/internal/config"
	"github.com/stretchr/testify/require"
)

// minimalWorkspaceConfig builds a LitmusConfig backed by a temp directory
// containing a bare alertmanager.yml with a single route + receiver.
// The workspace is clean — no regressions directory.
func minimalWorkspaceConfig(t *testing.T) (*config.LitmusConfig, string) {
	t.Helper()
	root := t.TempDir()

	err := os.WriteFile(filepath.Join(root, "alertmanager.yml"), []byte(`
route:
  receiver: default
receivers:
  - name: default
`), 0o600)
	require.NoError(t, err)

	cfg := &config.LitmusConfig{}
	cfg.Workspace.Root = root
	return cfg, root
}

func TestRunSnapshot_ListHistoryErrorPropagated(t *testing.T) {
	cfg, root := minimalWorkspaceConfig(t)

	// Create regressions as a file (not a dir) so os.ReadDir fails with ENOTDIR,
	// which is a real error — not ErrNotExist. Before the fix, this error was
	// silently swallowed because err was overwritten by the next statement.
	regressionsPath := filepath.Join(root, "regressions")
	require.NoError(t, os.WriteFile(regressionsPath, []byte("not a dir"), 0o600))

	ctx := context.WithValue(context.Background(), config.ConfigKey{}, cfg)
	err := RunSnapshot(ctx, &SnapshotOptions{})

	require.Error(t, err, "RunSnapshot must propagate ListHistory errors")
	require.True(t,
		strings.Contains(err.Error(), "listing regression history"),
		"error must mention 'listing regression history', got: %s", err.Error(),
	)
}
