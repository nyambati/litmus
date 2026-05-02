package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nyambati/litmus/internal/fixtures"
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

// --- workspace.read ---

func TestWorkspaceRead_HappyPathBaseYaml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))

	ws := New(dir)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if ws.root == nil {
		t.Fatal("root nil, want populated AlertmanagerConfig")
	}
	if ws.root.Route == nil {
		t.Fatal("root.Route nil")
	}
	if ws.root.Route.Receiver != "default" {
		t.Errorf("root.Route.Receiver = %q, want %q", ws.root.Route.Receiver, "default")
	}
	if !strings.HasSuffix(meta.BaseFile, "base.yaml") {
		t.Errorf("meta.BaseFile = %q, want suffix base.yaml", meta.BaseFile)
	}
}

func TestWorkspaceRead_HappyPathAlertmanagerYml(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "alertmanager.yml", fixtures.MustRead("workspace/base-simple.yaml"))

	ws := New(dir)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if ws.root == nil {
		t.Fatal("root nil, want populated AlertmanagerConfig")
	}
	if !strings.HasSuffix(meta.BaseFile, "alertmanager.yml") {
		t.Errorf("meta.BaseFile = %q, want suffix alertmanager.yml", meta.BaseFile)
	}
}

func TestWorkspaceRead_NoTestsDir(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))

	ws := New(dir)
	if _, err := ws.read(); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests) != 0 {
		t.Errorf("Tests length = %d, want 0 (no tests dir)", len(ws.Tests))
	}
}

func TestWorkspaceRead_LoadsTestsFromDir(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))
	writeWSFixture(t, filepath.Join(dir, "tests"), "case.yaml", fixtures.MustRead("workspace/tests/root-case.yaml"))

	ws := New(dir)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests) != 1 {
		t.Fatalf("Tests length = %d, want 1", len(ws.Tests))
	}
	if ws.Tests[0].Name != "root1" {
		t.Errorf("Tests[0].Name = %q, want %q", ws.Tests[0].Name, "root1")
	}
	if ws.Tests[0].Type != "unit" {
		t.Errorf("Tests[0].Type = %q, want %q", ws.Tests[0].Type, "unit")
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

	ws := New(dir)
	if _, err := ws.read(); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests) != 2 {
		t.Fatalf("Tests length = %d, want 2 (top + nested)", len(ws.Tests))
	}
}

func TestWorkspaceRead_MissingBaseErrors(t *testing.T) {
	dir := t.TempDir()
	ws := New(dir)
	_, err := ws.read()
	if err == nil {
		t.Fatal("read = nil, want missing-base error")
	}
	if !strings.Contains(err.Error(), "missing base config") {
		t.Errorf("err = %q, want 'missing base config' substring", err)
	}
}

func TestWorkspaceRead_MissingDirErrors(t *testing.T) {
	ws := New("/non/existent/workspace/__test__")
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

	_, err := New(filePath).read()
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

	_, err := New(dir).read()
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

	ws := New(dir)
	meta, err := ws.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests) != 1 {
		t.Errorf("Tests length = %d, want 1 (non-yaml ignored)", len(ws.Tests))
	}
	if len(meta.TestFiles) != 1 {
		t.Errorf("meta.TestFiles length = %d, want 1", len(meta.TestFiles))
	}
}

func TestWorkspaceRead_TestsDirAsFileSilentlyIgnored(t *testing.T) {
	dir := t.TempDir()
	writeWSFixture(t, dir, "base.yaml", fixtures.MustRead("workspace/base-simple.yaml"))
	writeWSFixture(t, dir, "tests", "this is a file not a dir")

	ws := New(dir)
	if _, err := ws.read(); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ws.Tests) != 0 {
		t.Errorf("Tests length = %d, want 0 (tests-as-file silently ignored)", len(ws.Tests))
	}
}
