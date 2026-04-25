package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/types"
	"github.com/spf13/cobra"
)

const historyTimeFormat = "20060102-150405"

// ArchiveBaseline saves tests as a new history entry and writes regressions.litmus.yml.
func ArchiveBaseline(cfg *config.LitmusConfig, tests []*types.TestCase) (string, error) {
	id := time.Now().Format(historyTimeFormat)
	if err := os.MkdirAll(cfg.Regression.Directory, 0755); err != nil {
		return "", fmt.Errorf("creating history dir: %w", err)
	}

	// Archive the baseline with timestamp filename
	mpkPath := filepath.Join(cfg.Regression.Directory, id+".mpk")
	f, err := os.Create(mpkPath)
	if err != nil {
		return "", fmt.Errorf("creating history entry: %w", err)
	}
	defer f.Close()

	if err := codec.EncodeMsgPack(f, tests); err != nil {
		return "", fmt.Errorf("encoding history entry: %w", err)
	}

	if err := SaveRegressionState(cfg.RegressionsYamlFilePath(), &RegressionState{ID: id, Tests: tests}); err != nil {
		return "", fmt.Errorf("writing regression state: %w", err)
	}

	// Clean up old entries based on keep policy
	if err := cleanupOldEntries(cfg); err != nil {
		return "", fmt.Errorf("cleaning up old entries: %w", err)
	}

	return id, nil
}

// ListHistory returns history entry IDs sorted newest-first.
func ListHistory(regressionDir string) ([]string, error) {
	entries, err := os.ReadDir(regressionDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading history dir: %w", err)
	}

	var ids []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".mpk") && !strings.HasPrefix(e.Name(), "regressions.litmus") {
			id := strings.TrimSuffix(e.Name(), ".mpk")
			ids = append(ids, id)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ids)))
	return ids, nil
}

// RollbackToEntry restores a history entry as the active baseline.
func RollbackToEntry(cfg *config.LitmusConfig, id string) error {
	srcMpk := filepath.Join(cfg.Regression.Directory, id+".mpk")

	// Load the tests from the historical baseline
	tests, err := LoadBaseline(srcMpk)
	if err != nil {
		return fmt.Errorf("loading history entry %q: %w", id, err)
	}

	if err := SaveRegressionState(cfg.RegressionsYamlFilePath(), &RegressionState{ID: id, Tests: tests}); err != nil {
		return fmt.Errorf("writing regression state: %w", err)
	}

	return nil
}

// cleanupOldEntries removes old entries keeping only the latest `keep` versions.
func cleanupOldEntries(cfg *config.LitmusConfig) error {
	ids, err := ListHistory(cfg.Regression.Directory)
	if err != nil {
		return err
	}

	if len(ids) <= cfg.Regression.Keep {
		return nil
	}

	// IDs are sorted newest-first; remove older ones
	toDelete := ids[cfg.Regression.Keep:]

	for _, id := range toDelete {
		path := filepath.Join(cfg.Regression.Directory, id+".mpk")
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("deleting old entry %q: %w", id, err)
		}
	}

	return nil
}

// RunHistoryList prints available baseline history entries.
func RunHistoryList(cmd *cobra.Command) error {
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	ids, err := ListHistory(litmusConfig.Regression.Directory)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		cmd.Println("No baseline history found. Run 'litmus snapshot --update' to create one.")
		return nil
	}

	// Load current ID from regressions.litmus.yml
	state, err := LoadRegressionState(litmusConfig.RegressionsYamlFilePath())
	var current string
	if err == nil {
		current = state.ID
	}

	cmd.Println("Available baselines:")
	for _, id := range ids {
		var builder strings.Builder
		builder.WriteString("  ")
		builder.WriteString(id)
		if id == current {
			builder.WriteString(" (current)")
		}
		cmd.Println(builder.String())
	}
	cmd.Printf("\nUse 'litmus history rollback <id>' to restore a baseline.\n")
	return nil
}

// RunHistoryRollback restores the baseline identified by id.
func RunHistoryRollback(cmd *cobra.Command, id string) error {
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	ids, err := ListHistory(litmusConfig.Regression.Directory)
	if err != nil {
		return err
	}

	if !slices.Contains(ids, id) {
		return fmt.Errorf("version %q not found; run 'litmus history list' to see available versions", id)
	}

	if err := RollbackToEntry(litmusConfig, id); err != nil {
		return err
	}

	cmd.Printf("✓ Rolled back baseline to %s\n", id)
	return nil
}
