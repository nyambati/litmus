package snapshot

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const missingPath = "/nonexistent/litmus/test/path/file.yml"

func TestLoadBaseline_MissingFile_WrapsError(t *testing.T) {
	_, err := LoadBaseline(missingPath)
	if err == nil {
		t.Fatal("LoadBaseline() = nil error, want error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("LoadBaseline() error chain must include os.ErrNotExist, got: %v", err)
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Errorf("LoadBaseline() error must contain path, got: %v", err)
	}
}

func TestLoadBaselineYAML_MissingFile_WrapsError(t *testing.T) {
	_, err := LoadBaselineYAML(missingPath)
	if err == nil {
		t.Fatal("LoadBaselineYAML() = nil error, want error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("LoadBaselineYAML() error chain must include os.ErrNotExist, got: %v", err)
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Errorf("LoadBaselineYAML() error must contain path, got: %v", err)
	}
}

func TestLoadBaselineYAML_BadYAML_WrapsError(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.yml")
	if err := os.WriteFile(f, []byte("[\nbad yaml"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadBaselineYAML(f)
	if err == nil {
		t.Fatal("LoadBaselineYAML() = nil error, want parse error")
	}
	if !strings.Contains(err.Error(), f) {
		t.Errorf("LoadBaselineYAML() error must contain path, got: %v", err)
	}
}

func TestSaveRegressionState_BadPath_WrapsError(t *testing.T) {
	err := SaveRegressionState("/nonexistent/dir/state.yml", &RegressionState{})
	if err == nil {
		t.Fatal("SaveRegressionState() = nil error, want error")
	}
	if !strings.Contains(err.Error(), "/nonexistent/dir/state.yml") {
		t.Errorf("SaveRegressionState() error must contain path, got: %v", err)
	}
}
