package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/types"
	"github.com/nyambati/litmus/internal/workspace"
)

const historyTimeFormat = "20060102-150405.000000"

var timeNow = time.Now

// ArchiveBaseline saves tests as a new history entry and writes regressions.litmus.yml.
func ArchiveBaseline(cfg *config.LitmusConfig, tests []*types.TestCase) (string, error) {
	id := timeNow().Format(historyTimeFormat)
	if err := os.MkdirAll(cfg.RegressionsDir(), 0o755); err != nil {
		return "", fmt.Errorf("creating history dir: %w", err)
	}

	// Archive the baseline with timestamp filename
	mpkPath := filepath.Join(cfg.RegressionsDir(), id+".mpk")
	f, err := os.Create(mpkPath)
	if err != nil {
		return "", fmt.Errorf("creating history entry: %w", err)
	}
	defer f.Close()

	if err := codec.EncodeMsgPack(f, tests); err != nil {
		return "", fmt.Errorf("encoding history entry: %w", err)
	}

	if err := snapshot.SaveRegressionState(cfg.RegressionsYamlFilePath(), &snapshot.RegressionState{ID: id, Tests: tests}); err != nil {
		return "", fmt.Errorf("writing regression state: %w", err)
	}

	// Clean up old entries based on keep policy
	if err := cleanupOldEntries(cfg); err != nil {
		return "", fmt.Errorf("cleaning up old entries: %w", err)
	}

	return id, nil
}

// cleanupOldEntries removes old entries keeping only the latest `keep` versions.
func cleanupOldEntries(cfg *config.LitmusConfig) error {
	ids, err := snapshot.ListHistory(cfg.RegressionsDir())
	if err != nil {
		return err
	}

	if len(ids) <= cfg.Workspace.History {
		return nil
	}

	// IDs are sorted newest-first; remove older ones
	toDelete := ids[cfg.Workspace.History:]

	for _, id := range toDelete {
		path := filepath.Join(cfg.RegressionsDir(), id+".mpk")
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("deleting old entry %q: %w", id, err)
		}
	}

	return nil
}

// RunHistoryList prints available baseline history entries.
func RunHistoryList(ctx context.Context) error {
	litmusConfig := config.ConfigFromContext(ctx)
	ids, err := snapshot.ListHistory(litmusConfig.RegressionsDir())
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		fmt.Fprintln(os.Stdout, "No baseline history found. Run 'litmus snapshot capture' to create one.")
		return nil
	}

	ws, err := workspace.Load(litmusConfig, config.LoggerFromContext(ctx))
	if err != nil {
		return err
	}

	var current string
	if ws.RegressionState != nil {
		current = ws.RegressionState.ID
	}

	fmt.Fprintln(os.Stdout, "Available baselines:")
	for _, id := range ids {
		var builder strings.Builder
		builder.WriteString("  ")
		builder.WriteString(id)
		if id == current {
			builder.WriteString(" (current)")
		}
		fmt.Fprintln(os.Stdout, builder.String())
	}
	fmt.Fprintf(os.Stdout, "\nUse 'litmus history rollback <id>' to restore a baseline.\n")
	return nil
}

// RunHistoryRollback restores the baseline identified by id.
func RunHistoryRollback(ctx context.Context, id string) error {
	litmusConfig := config.ConfigFromContext(ctx)
	ids, err := snapshot.ListHistory(litmusConfig.RegressionsDir())
	if err != nil {
		return err
	}

	if !slices.Contains(ids, id) {
		return fmt.Errorf("version %q not found; run 'litmus history list' to see available versions", id)
	}

	if err := snapshot.RollbackToEntry(litmusConfig, id); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ Rolled back baseline to %s\n", id)
	return nil
}
