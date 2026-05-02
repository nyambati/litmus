package workspace

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

func New(dir string, logger logrus.FieldLogger) *Workspace {
	if logger == nil {
		l := logrus.New()
		l.Out = io.Discard
		logger = l
	}
	return &Workspace{
		dir:    dir,
		logger: logger,
	}
}

// Fragments returns the loaded child fragments.
func (w *Workspace) Fragments() []*fragment.Fragment { return w.fragments }

// Tests returns root-level test cases.
func (w *Workspace) Tests() []*types.TestCase { return w.tests }

// RootFragment returns the pre-assembly snapshot of the root fragment (for policy checks).
func (w *Workspace) RootFragment() *fragment.Fragment { return w.rootFragment }

func (w *Workspace) Config() (*amconfig.Config, error) {
	if w.root == nil {
		return nil, fmt.Errorf("workspace not assembled: call Assemble() first")
	}
	s := w.root.String()
	if s == "" {
		return nil, fmt.Errorf("failed to serialize alertmanager config (check stderr for encoding error)")
	}
	return amconfig.Load(s)
}

func (w *Workspace) ConfigString() string {
	if w.root == nil {
		return ""
	}
	return w.root.String()
}

func (w *Workspace) read() (*Metadata, error) {
	absPath, err := filepath.Abs(w.dir)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("stat workspace path %q: %w", absPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace path %q is not a directory", absPath)
	}

	basePath, err := resolveBaseFile(absPath)
	if err != nil {
		return nil, err
	}

	root, err := readBase(basePath)
	if err != nil {
		return nil, err
	}

	w.root = root

	tests, testFiles, err := readRootTests(filepath.Join(absPath, "tests"))
	if err != nil {
		return nil, err
	}

	w.tests = tests

	return &Metadata{
		Dir:       absPath,
		BaseFile:  basePath,
		TestFiles: testFiles,
	}, nil
}

func readRootTests(testsDir string) ([]*types.TestCase, []string, error) {
	info, err := os.Stat(testsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("stat tests dir %q: %w", testsDir, err)
	}
	if !info.IsDir() {
		return nil, nil, nil
	}

	var tests []*types.TestCase
	var files []string
	walkErr := filepath.WalkDir(testsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".yaml", ".yml":
		default:
			return nil
		}

		fi, infoErr := d.Info()
		if infoErr != nil {
			return fmt.Errorf("stat test file %q: %w", path, infoErr)
		}
		if fi.Size() > maxFileSize {
			return fmt.Errorf("test file %q exceeds size limit of %d bytes", path, maxFileSize)
		}

		data, err := os.ReadFile(path) //nolint:gosec // path comes from WalkDir, not user input
		if err != nil {
			return fmt.Errorf("read test file %q: %w", path, err)
		}
		var doc fragment.TestDoc
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("parse yaml in %q: %w", path, err)
		}
		for _, tc := range doc.Tests {
			tc.Type = "unit"
		}
		tests = append(tests, doc.Tests...)
		files = append(files, path)
		return nil
	})
	if walkErr != nil {
		return nil, nil, walkErr
	}
	return tests, files, nil
}

// resolveBaseFile finds the single unambiguous base config in dir.
// Accepted names (in priority order): base.yaml, base.yml, alertmanager.yaml, alertmanager.yml.
// Returns an error if zero or more than one matching file exists.
func resolveBaseFile(dir string) (string, error) {
	candidates := []string{
		filepath.Join(dir, "base.yaml"),
		filepath.Join(dir, "base.yml"),
		filepath.Join(dir, "alertmanager.yaml"),
		filepath.Join(dir, "alertmanager.yml"),
	}

	var found []string
	for _, c := range candidates {
		if fileExists(c) {
			found = append(found, c)
		}
	}

	switch len(found) {
	case 1:
		return found[0], nil
	case 0:
		return "", fmt.Errorf("workspace %q missing base config (base.yaml, base.yml, alertmanager.yaml, or alertmanager.yml)", dir)
	default:
		return "", fmt.Errorf("workspace %q has ambiguous base config: %v", dir, found)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func readBase(path string) (*types.AlertmanagerConfig, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat base file %q: %w", path, err)
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("base file %q exceeds size limit of %d bytes", path, maxFileSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read base file %q: %w", path, err)
	}

	var cfg types.AlertmanagerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse alertmanager config %q: %w", path, err)
	}
	return &cfg, nil
}
