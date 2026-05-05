package workspace

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	amtemplate "github.com/prometheus/alertmanager/template"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

// New creates a new workspace from a LitmusConfig.
func New(cfg *config.LitmusConfig, logger logrus.FieldLogger) *Workspace {
	if logger == nil {
		l := logrus.New()
		l.Out = io.Discard
		logger = l
	}
	return &Workspace{
		cfg:    cfg,
		dir:    cfg.Workspace.Root,
		logger: logger,
	}
}

func (w *Workspace) Config() (*amconfig.Config, error) {
	if w.config == nil {
		return nil, fmt.Errorf("workspace not assembled: call Assemble() first")
	}
	data, err := w.config.Marshal()
	if err != nil {
		return nil, fmt.Errorf("serializing alertmanager config: %w", err)
	}
	cfg, err := amconfig.Load(string(data))
	if err != nil {
		return nil, fmt.Errorf("loading alertmanager config: %w", err)
	}

	if cfg.Route == nil {
		return nil, fmt.Errorf("alertmanager config has no route defined")
	}

	return cfg, nil
}

// Templates validates and loads template files referenced in the config.
// It checks for file existence and validates Go template syntax.
func (w *Workspace) Templates() (map[string]string, error) {
	if w.config == nil {
		return nil, fmt.Errorf("workspace not assembled: call Assemble() first")
	}
	if w.cfg == nil {
		return nil, fmt.Errorf("missing litmus configuration")
	}

	templates := make(map[string]string)

	for _, filename := range w.config.Templates {
		filePath := filepath.Join(w.cfg.TemplatesDir(), filename)
		if filepath.IsAbs(filename) {
			filePath = filename
		}

		// Check file existence
		if _, err := os.Stat(filePath); err != nil {
			return nil, fmt.Errorf("template file %q: %w", filename, err)
		}

		// Read file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading template %q: %w", filename, err)
		}

		// Validate Go template syntax
		t, err := amtemplate.New()
		if err != nil {
			return nil, fmt.Errorf("creating template parser for %q: %w", filename, err)
		}
		if err := t.Parse(strings.NewReader(string(data))); err != nil {
			return nil, fmt.Errorf("template %q has invalid syntax: %w", filename, err)
		}

		key := filepath.Base(filename)
		templates[key] = string(data)
	}

	return templates, nil
}

func (w *Workspace) ConfigString() string {
	if w.config == nil {
		return ""
	}
	return w.config.String()
}

// Tests returns all behavioral test cases from all loaded fragments.
func (w *Workspace) Tests() []*types.TestCase {
	var tests []*types.TestCase
	for _, f := range w.Fragments {
		tests = append(tests, f.Tests...)
	}
	return tests
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

	cfg, err := readBase(basePath)
	if err != nil {
		return nil, err
	}

	w.config = cfg

	tests, testFiles, err := readRootTests(filepath.Join(absPath, "tests"))
	if err != nil {
		return nil, err
	}

	w.Fragments = append(w.Fragments, getRootFragment(w.config, tests))

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
		parsed, err := fragment.ParseTestDoc(data)
		if err != nil {
			return fmt.Errorf("parse yaml in %q: %w", path, err)
		}
		tests = append(tests, parsed...)
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

// rootSnapshot captures the root's own routes and receivers before assembly
// merges child fragment data into root. The snapshot is used so PolicyChecker
// can evaluate the root independently without seeing fragment contributions.
func getRootFragment(root *types.AlertmanagerConfig, tests []*types.TestCase) *fragment.Fragment {
	frag := &fragment.Fragment{
		Namespace: "root",
		Tests:     tests,
	}
	if root.Route != nil && len(root.Route.Routes) > 0 {
		frag.Routes = append([]*amconfig.Route{}, root.Route.Routes...)
	}
	if len(root.Receivers) > 0 {
		frag.Receivers = append([]*types.Receiver{}, root.Receivers...)
	}
	return frag
}

// GetRegressionState reads the regression state (ID + tests) from regressions.litmus.yml.
func readRegressionState(path string) (*snapshot.RegressionState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state snapshot.RegressionState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing regression state %q: %w", path, err)
	}
	return &state, nil
}

// loadRegressionState loads the regression state into the workspace.
func (w *Workspace) loadRegressionState() error {
	if w.cfg == nil {
		return fmt.Errorf("missing litmus configuration")
	}
	regPath := w.cfg.RegressionsYamlFilePath()
	state, err := readRegressionState(regPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	w.RegressionState = state
	return nil
}

// EnsureRegressionState ensures the regression state is loaded into the workspace.
func (w *Workspace) EnsureRegressionState() error {
	return w.loadRegressionState()
}

// SaveRegressionState writes the regression state (ID + tests) to regressions.litmus.yml.
func (w *Workspace) SaveRegressionState(state *snapshot.RegressionState) error {
	if w.cfg == nil {
		return fmt.Errorf("no config set on workspace")
	}
	return snapshot.SaveRegressionState(w.cfg.RegressionsYamlFilePath(), state)
}
