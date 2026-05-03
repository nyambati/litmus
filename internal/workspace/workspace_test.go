package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/fixtures"
	"github.com/nyambati/litmus/internal/types"
)

func writeWSFixture(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
	return path
}

// --- resolveBaseFile ---

func TestResolveBaseFile_AcceptsBaseYaml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))

	got, err := resolveBaseFile(dir)
	if err != nil {
		t.Fatalf("resolveBaseFile = %v, want nil", err)
	}
	if !strings.HasSuffix(got, "base.yaml") {
		t.Errorf("resolveBaseFile = %q, want suffix base.yaml", got)
	}
}

func TestResolveBaseFile_AcceptsBaseYml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yml", fixtures.MustRead("workspace/base-simple.yaml"))

	got, err := resolveBaseFile(dir)
	if err != nil {
		t.Fatalf("resolveBaseFile = %v, want nil", err)
	}
	if !strings.HasSuffix(got, "base.yml") {
		t.Errorf("resolveBaseFile = %q, want suffix base.yml", got)
	}
}

// Regression: workspace used to reject alertmanager.yaml — must now accept it.
func TestResolveBaseFile_AcceptsAlertmanagerYaml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "alertmanager.yaml", fixtures.MustRead("workspace/base-simple.yaml"))

	got, err := resolveBaseFile(dir)
	if err != nil {
		t.Fatalf("resolveBaseFile = %v, want nil (alertmanager.yaml must be accepted)", err)
	}
	if !strings.HasSuffix(got, "alertmanager.yaml") {
		t.Errorf("resolveBaseFile = %q, want suffix alertmanager.yaml", got)
	}
}

func TestResolveBaseFile_AcceptsAlertmanagerYml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "alertmanager.yml", fixtures.MustRead("workspace/base-simple.yaml"))

	got, err := resolveBaseFile(dir)
	if err != nil {
		t.Fatalf("resolveBaseFile = %v, want nil (alertmanager.yml must be accepted)", err)
	}
	if !strings.HasSuffix(got, "alertmanager.yml") {
		t.Errorf("resolveBaseFile = %q, want suffix alertmanager.yml", got)
	}
}

func TestResolveBaseFile_MissingErrors(t *testing.T) {
	dir := t.TempDir()
	_, err := resolveBaseFile(dir)
	if err == nil {
		t.Fatal("resolveBaseFile = nil, want missing-base error")
	}
	if !strings.Contains(err.Error(), "missing base config") {
		t.Errorf("err = %q, want 'missing base config' substring", err)
	}
}

func TestResolveBaseFile_AmbiguousErrors(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))
	writeWSFixture(t, dir, "alertmanager.yml", fixtures.MustRead("workspace/base-simple.yaml"))

	_, err := resolveBaseFile(dir)
	if err == nil {
		t.Fatal("resolveBaseFile = nil, want ambiguous error")
	}
	if !strings.Contains(err.Error(), "ambiguous base config") {
		t.Errorf("err = %q, want 'ambiguous base config' substring", err)
	}
}

func TestResolveBaseFile_DirectoryIsNotAFile(t *testing.T) {
	dir := t.TempDir()
	// Create a directory named base.yaml — must not be treated as the base file.
	if err := os.Mkdir(filepath.Join(dir, "base.yaml"), 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	_, err := resolveBaseFile(dir)
	if err == nil {
		t.Fatal("resolveBaseFile = nil, want missing-base error when base.yaml is a directory")
	}
	if !strings.Contains(err.Error(), "missing base config") {
		t.Errorf("err = %q, want 'missing base config' substring", err)
	}
}

// --- read ---

func TestWorkspaceRead_HappyPathBaseYaml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if ws.Config == nil {
		t.Fatal("root nil, want populated AlertmanagerConfig")
	}
	if ws.Config.Route == nil {
		t.Fatal("root.Route nil")
	}
	if ws.Config.Route.Receiver != "default" {
		t.Errorf("root.Route.Receiver = %q, want %q", ws.Config.Route.Receiver, "default")
	}
	if !strings.HasSuffix(meta.BaseFile, "base.yaml") {
		t.Errorf("meta.BaseFile = %q, want suffix base.yaml", meta.BaseFile)
	}
}

func TestWorkspaceRead_HappyPathAlertmanagerYml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "alertmanager.yml", fixtures.MustRead("workspace/base-simple.yaml"))

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if ws.Config == nil {
		t.Fatal("root nil, want populated AlertmanagerConfig")
	}
	if !strings.HasSuffix(meta.BaseFile, "alertmanager.yml") {
		t.Errorf("meta.BaseFile = %q, want suffix alertmanager.yml", meta.BaseFile)
	}
}

func TestWorkspaceRead_NoTestsDir(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	if _, err := ws.read(); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests()) != 0 {
		t.Errorf("Tests length = %d, want 0 (no tests dir)", len(ws.Tests()))
	}
}

func TestWorkspaceRead_LoadsTestsFromDir(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))
	writeWSFixture(t, filepath.Join(dir, "tests"), "case.yaml", fixtures.MustRead("workspace/tests/root-case.yaml"))

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests()) != 1 {
		t.Fatalf("Tests length = %d, want 1", len(ws.Tests()))
	}
	if ws.Tests()[0].Name != "root1" {
		t.Errorf("Tests[0].Name = %q, want %q", ws.Tests()[0].Name, "root1")
	}
	if ws.Tests()[0].Type != "unit" {
		t.Errorf("Tests[0].Type = %q, want %q", ws.Tests()[0].Type, "unit")
	}
	if len(meta.TestFiles) != 1 {
		t.Errorf("meta.TestFiles length = %d, want 1", len(meta.TestFiles))
	}
}

func TestWorkspaceRead_LoadsNestedTests(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))
	writeWSFixture(t, filepath.Join(dir, "tests"), "top.yaml", fixtures.MustRead("workspace/tests/root-case.yaml"))
	writeWSFixture(t, filepath.Join(dir, "tests", "sub"), "deep.yaml", fixtures.MustRead("workspace/tests/sub/nested.yaml"))

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	if _, err := ws.read(); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests()) != 2 {
		t.Fatalf("Tests length = %d, want 2 (top + nested)", len(ws.Tests()))
	}
}

func TestWorkspaceRead_MissingBaseErrors(t *testing.T) {
	dir := t.TempDir()
	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	_, err := ws.read()
	if err == nil {
		t.Fatal("read = nil, want missing-base error")
	}
	if !strings.Contains(err.Error(), "missing base config") {
		t.Errorf("err = %q, want 'missing base config' substring", err)
	}
}

func TestWorkspaceRead_MissingDirErrors(t *testing.T) {
	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      "/non/existent/workspace/__test__",
			Fragments: "fragments",
		},
	}, nil)
	_, err := ws.read()
	if err == nil {
		t.Fatal("read = nil, want stat error")
	}
	if !strings.Contains(err.Error(), "stat workspace") {
		t.Errorf("err = %q, want 'stat workspace' substring", err)
	}
}

func TestWorkspaceRead_NotADirectoryErrors(t *testing.T) {
	dir := t.TempDir()
	filePath := writeWSFixture(t, dir, "not-a-dir.yaml", "x: 1")

	_, err := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      filePath,
			Fragments: "fragments",
		},
	}, nil).read()
	if err == nil {
		t.Fatal("read = nil, want not-a-directory error")
	}
	if !strings.Contains(err.Error(), "is not a directory") {
		t.Errorf("err = %q, want 'is not a directory'", err)
	}
}

func TestWorkspaceRead_InvalidBaseErrors(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-invalid.yaml"))

	_, err := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil).read()
	if err == nil {
		t.Fatal("read = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "parse alertmanager config") {
		t.Errorf("err = %q, want 'parse alertmanager config' substring", err)
	}
}

func TestWorkspaceRead_IgnoresNonYAMLInTestsDir(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))
	writeWSFixture(t, filepath.Join(dir, "tests"), "case.yaml", fixtures.MustRead("workspace/tests/root-case.yaml"))
	writeWSFixture(t, filepath.Join(dir, "tests"), "README.md", "# notes")
	writeWSFixture(t, filepath.Join(dir, "tests"), "scratch.txt", "ignore me")

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests()) != 1 {
		t.Errorf("Tests length = %d, want 1 (non-yaml ignored)", len(ws.Tests()))
	}
	if len(meta.TestFiles) != 1 {
		t.Errorf("meta.TestFiles length = %d, want 1", len(meta.TestFiles))
	}
}

// --- Assemble: RootFragment ---

const rootNamespace = "root"

func TestAssemble_RootFragmentCapturedBeforeChildMerge(t *testing.T) {
	dir := t.TempDir()

	// Root config has one direct child route.
	writeWSFixture(t, dir, "base.yaml", `
route:
  receiver: default
  routes:
    - receiver: root-critical
      match:
        severity: critical
receivers:
  - name: default
  - name: root-critical
`)
	// A child fragment adds its own route.
	writeWSFixture(t, filepath.Join(dir, "fragments", "db"), "fragment.yaml", `
namespace: db
routes:
  - receiver: db-alert
receivers:
  - name: db-alert
`)

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)

	if err := ws.Assemble(); err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	if len(ws.Fragments) == 0 {
		t.Fatal("should have one fragment, found 0")
	}
	if ws.Fragments[0].Namespace != rootNamespace {
		t.Errorf("RootFragment.Namespace = %q, want %q", rootNamespace, ws.Fragments[0].Namespace)
	}
	// Snapshot must contain only the root's own route (root-critical), not db-alert.
	if len(ws.Fragments[0].Routes) != 1 {
		t.Fatalf("RootFragment.Routes length = %d, want 1", len(ws.Fragments[0].Routes))
	}
	if ws.Fragments[0].Routes[0].Receiver != "root-critical" {
		t.Errorf("RootFragment.Routes[0].Receiver = %q, want \"root-critical\"", ws.Fragments[0].Routes[0].Receiver)
	}
	// The child fragment route must NOT appear in the snapshot.
	for _, r := range ws.Fragments[0].Routes {
		if r.Receiver == "db-db-alert" || r.Receiver == "db-alert" {
			t.Errorf("RootFragment must not contain child fragment route %q", r.Receiver)
		}
	}
}

func TestWorkspaceRead_TestsDirAsFileSilentlyIgnored(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))
	writeWSFixture(t, dir, "tests", "this is a file not a dir")

	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      dir,
			Fragments: "fragments",
		},
	}, nil)
	if _, err := ws.read(); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests()) != 0 {
		t.Errorf("Tests length = %d, want 0 (tests-as-file silently ignored)", len(ws.Tests()))
	}
}

// func TestAssemble_Idempotent(t *testing.T) {
// 	dir := t.TempDir()
// 	writeWSFixture(t, dir, "base.yaml", `
// route:
//   receiver: default
// receivers:
//   - name: default
// `)
// 	writeWSFixture(t, filepath.Join(dir, "fragments", "db"), "fragment.yaml", `
// namespace: db
// routes:
//   - receiver: db-alert
// receivers:
//   - name: db-alert
// `)

// 	ws := New(&config.LitmusConfig{
// 		Workspace: config.WorkspaceConfig{
// 			Root:      dir,
// 			Fragments: "fragments",
// 		},
// 	}, nil)

// 	first := *ws

// 	if err := ws.Assemble(); err != nil {
// 		t.Fatalf("first Assemble: %v", err)
// 	}

// 	firstFragments := len(first.Fragments)
// 	firstReceivers := len(first.Config.Receivers)

// 	if err := ws.Assemble(); err != nil {
// 		t.Fatalf("second Assemble: %v", err)
// 	}

// 	if len(ws.Fragments) != firstFragments {
// 		t.Errorf("Fragments after second Assemble = %d, want %d (idempotency broken)", len(ws.Fragments), firstFragments)
// 	}
// 	if len(ws.Config.Receivers) != firstReceivers {
// 		t.Errorf("Receivers after second Assemble = %d, want %d (idempotency broken)", len(ws.Config.Receivers), firstReceivers)
// 	}
// }

func TestAMConfig_ReturnsErrorWhenNotAssembled(t *testing.T) {
	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      t.TempDir(),
			Fragments: "fragments",
		},
	}, nil)
	_, err := ws.AMConfig()
	if err == nil {
		t.Fatal("AMConfig() = nil error, want error")
	}
	if !strings.Contains(err.Error(), "workspace not assembled") {
		t.Errorf("AMConfig() error = %q, want 'workspace not assembled'", err)
	}
}

func TestAMConfig_PropagatesSerializationError(t *testing.T) {
	// Inject a config that references an unset env var so Marshal() fails.
	ws := New(&config.LitmusConfig{
		Workspace: config.WorkspaceConfig{
			Root:      t.TempDir(),
			Fragments: "fragments",
		},
	}, nil)
	ws.Config = &types.AlertmanagerConfig{
		Receivers: []*types.Receiver{
			{
				Name:           "r",
				WebhookConfigs: []map[string]any{{"url": "env(litmus_test_unset_amconfig_var)"}},
			},
		},
	}
	_, err := ws.AMConfig()
	if err == nil {
		t.Fatal("AMConfig() = nil error, want serialization error")
	}
	if !strings.Contains(err.Error(), "serializing alertmanager config") {
		t.Errorf("AMConfig() error = %q, want wrapped 'serializing alertmanager config'", err)
	}
}

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

func TestLoadRegressionState_MissingFile_WrapsError(t *testing.T) {
	_, err := readRegressionState(missingPath)
	if err == nil {
		t.Fatal("GetRegressionState() = nil error, want error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("GetRegressionState() error chain must include os.ErrNotExist, got: %v", err)
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Errorf("LoadRegressionState() error must contain path, got: %v", err)
	}
}

func TestLoadRegressionState_BadYAML_WrapsError(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.yml")
	if err := os.WriteFile(f, []byte("[\nbad yaml"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := readRegressionState(f)
	if err == nil {
		t.Fatal("LoadRegressionState() = nil error, want parse error")
	}
	if !strings.Contains(err.Error(), f) {
		t.Errorf("LoadRegressionState() error must contain path, got: %v", err)
	}
}

func TestSaveRegressionState_BadPath_WrapsError(t *testing.T) {
	err := SaveRegressionState("/nonexistent/dir/state.yml", &types.RegressionState{})
	if err == nil {
		t.Fatal("SaveRegressionState() = nil error, want error")
	}
	if !strings.Contains(err.Error(), "/nonexistent/dir/state.yml") {
		t.Errorf("SaveRegressionState() error must contain path, got: %v", err)
	}
}
