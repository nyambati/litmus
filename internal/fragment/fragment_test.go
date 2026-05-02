package fragment

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nyambati/litmus/internal/fixtures"
	"github.com/nyambati/litmus/internal/types"
	"github.com/prometheus/alertmanager/config"
	"go.yaml.in/yaml/v3"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		wantFragments int
		wantNames     []string
		wantErr       string
	}{
		{
			name: "loads fragments from subdirectories only",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				writeFragmentFixture(t, filepath.Join(root, "fragment1"), "fragment.yaml", `namespace: "fragment1"`)
				writeFragmentFixture(t, root, "config.yaml", `namespace: "config"`)
				return root
			},
			wantFragments: 1,
			wantNames:     []string{"fragment1"},
		},
		{
			name: "loads multiple fragments from separate directories",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				writeFragmentFixture(t, filepath.Join(root, "database"), "fragment.yaml", `namespace: "database"`)
				writeFragmentFixture(t, filepath.Join(root, "payments"), "fragment.yaml", `namespace: "payments"`)
				writeFragmentFixture(t, filepath.Join(root, "monitoring"), "fragment.yaml", `namespace: "monitoring"`)
				return root
			},
			wantFragments: 3,
			wantNames:     []string{"database", "payments", "monitoring"},
		},
		{
			name: "ignores files in root directory",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				writeFragmentFixture(t, root, "root-file.yaml", `namespace: "root"`)
				writeFragmentFixture(t, filepath.Join(root, "fragment"), "fragment.yaml", `namespace: "fragment"`)
				return root
			},
			wantFragments: 1,
			wantNames:     []string{"fragment"},
		},
		{
			name: "returns error for non-existent directory",
			setup: func(t *testing.T) string {
				t.Helper()
				return "/non/existent/path"
			},
			wantErr: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.setup(t)

			result, err := Load(root)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Load() error = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error = %q, want error containing %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() returned error: %v", err)
			}

			if len(result.Fragments) != tt.wantFragments {
				t.Fatalf("Load() returned %d fragments, want %d", len(result.Fragments), tt.wantFragments)
			}

			gotNames := make([]string, 0, len(result.Fragments))
			for _, f := range result.Fragments {
				gotNames = append(gotNames, f.Fragment.Namespace)
			}

			for _, want := range tt.wantNames {
				if !slices.Contains(gotNames, want) {
					t.Fatalf("Load() did not return fragment with name %q", want)
				}
			}
		})
	}
}

func TestFragmentRead(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		wantNS        string
		wantRoutes    int
		wantReceivers int
		wantInhibit   int
		wantFiles     int
		wantErr       string
	}{
		{
			name: "reads single yaml file",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "fragment.yaml", fixtures.MustRead("fragment/database-full.yaml"))
				return dir
			},
			wantNS:        "db",
			wantRoutes:    1,
			wantReceivers: 1,
			wantFiles:     1,
		},
		{
			name: "reads multiple yaml files and merges",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "01-base.yaml", fixtures.MustRead("fragment/merge-base.yaml"))
				writeFragmentFixture(t, dir, "02-receivers.yaml", fixtures.MustRead("fragment/merge-receivers.yaml"))
				return dir
			},
			wantNS:        "db",
			wantRoutes:    1,
			wantReceivers: 1,
			wantInhibit:   1,
			wantFiles:     2,
		},
		{
			name: "defaults namespace to directory name when not declared",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				subdir := filepath.Join(dir, "myfragment")
				if err := os.MkdirAll(subdir, 0o755); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				writeFragmentFixture(t, subdir, "fragment.yaml", fixtures.MustRead("fragment/routes-receivers-bare.yaml"))
				return subdir
			},
			wantNS:        "myfragment",
			wantRoutes:    1,
			wantReceivers: 1,
			wantFiles:     1,
		},
		{
			name: "returns error for invalid YAML",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "bad.yaml", fixtures.MustRead("fragment/invalid.yaml"))
				return dir
			},
			wantErr: "yaml: line",
		},
		{
			name: "returns error on namespace conflict across files",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "01.yaml", `namespace: "db"`)
				writeFragmentFixture(t, dir, "02.yaml", `namespace: "payments"`)
				return dir
			},
			wantErr: "conflicting fragment namespace",
		},
		{
			name: "returns error on group conflict across files",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "01.yaml", fixtures.MustRead("fragment/group-team-a.yaml"))
				writeFragmentFixture(t, dir, "02.yaml", fixtures.MustRead("fragment/group-team-b.yaml"))
				return dir
			},
			wantErr: "conflicting fragment group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)

			frag := New(dir)
			meta, err := frag.Read()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Read() error = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Read() error = %q, want error containing %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Read() returned error: %v", err)
			}

			if frag.Namespace != tt.wantNS {
				t.Errorf("Namespace = %q, want %q", frag.Namespace, tt.wantNS)
			}
			if len(frag.Routes) != tt.wantRoutes {
				t.Errorf("Routes length = %d, want %d", len(frag.Routes), tt.wantRoutes)
			}
			if len(frag.Receivers) != tt.wantReceivers {
				t.Errorf("Receivers length = %d, want %d", len(frag.Receivers), tt.wantReceivers)
			}
			if len(frag.InhibitRules) != tt.wantInhibit {
				t.Errorf("InhibitRules length = %d, want %d", len(frag.InhibitRules), tt.wantInhibit)
			}
			if len(meta.Files) != tt.wantFiles {
				t.Errorf("Files length = %d, want %d", len(meta.Files), tt.wantFiles)
			}
		})
	}
}

func TestFragmentMerge(t *testing.T) {
	tests := []struct {
		name       string
		dst        Fragment
		src        Fragment
		wantErr    string
		wantNS     string
		wantRoutes int
	}{
		{
			name: "merges empty destination with source values",
			dst:  Fragment{},
			src: Fragment{
				Namespace: "db",
				Routes:    []*config.Route{{Receiver: "test"}},
			},
			wantNS:     "db",
			wantRoutes: 1,
		},
		{
			name: "allows empty identity fields on source",
			dst: Fragment{
				Namespace: "db",
			},
			src: Fragment{
				Routes: []*config.Route{{Receiver: "test"}},
			},
			wantNS:     "db",
			wantRoutes: 1,
		},
		{
			name:    "rejects namespace conflict",
			dst:     Fragment{Namespace: "db"},
			src:     Fragment{Namespace: "payments"},
			wantErr: "conflicting fragment namespace",
		},
		{
			name: "rejects group conflict",
			dst: Fragment{
				Group: &FragmentGroup{
					Receiver: "team-a",
					Match:    map[string]string{"team": "a"},
				},
			},
			src: Fragment{
				Group: &FragmentGroup{
					Receiver: "team-b",
					Match:    map[string]string{"team": "b"},
				},
			},
			wantErr: "conflicting fragment group",
		},
		{
			name: "allows empty groups",
			dst:  Fragment{},
			src:  Fragment{},
		},
		{
			name: "allows empty group on destination",
			dst:  Fragment{},
			src: Fragment{
				Group: &FragmentGroup{
					Receiver: "team-db",
					Match:    map[string]string{"team": "db"},
				},
			},
		},
		{
			name: "allows empty group on source",
			dst: Fragment{
				Group: &FragmentGroup{
					Receiver: "team-db",
					Match:    map[string]string{"team": "db"},
				},
			},
			src: Fragment{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dst.merge(&tt.src)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("merge() error = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("merge() error = %q, want error containing %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("merge() returned error: %v", err)
			}

			if tt.wantNS != "" && tt.dst.Namespace != tt.wantNS {
				t.Errorf("Namespace = %q, want %q", tt.dst.Namespace, tt.wantNS)
			}
			if tt.wantRoutes > 0 && len(tt.dst.Routes) != tt.wantRoutes {
				t.Errorf("Routes length = %d, want %d", len(tt.dst.Routes), tt.wantRoutes)
			}
		})
	}
}

func TestFragmentValidation(t *testing.T) {
	tests := []struct {
		name     string
		fragment *Fragment
		wantErr  string
	}{
		{
			name: "passes when all routes use defined receivers",
			fragment: &Fragment{
				Namespace: "test",
				Routes:    []*config.Route{{Receiver: "critical"}, {Receiver: "warning"}},
				Receivers: []*types.Receiver{
					{Name: "critical"},
					{Name: "warning"},
				},
			},
		},
		{
			name: "passes with no routes",
			fragment: &Fragment{
				Namespace: "test",
				Receivers: []*types.Receiver{{Name: "test"}},
			},
		},
		{
			name: "passes with no receivers",
			fragment: &Fragment{
				Namespace: "test",
				Routes:    []*config.Route{},
			},
		},
		{
			name: "fails when route uses undefined receiver",
			fragment: &Fragment{
				Namespace: "test",
				Routes:    []*config.Route{{Receiver: "undefined"}},
				Receivers: []*types.Receiver{
					{Name: "defined"},
				},
			},
			wantErr: "undefined receiver",
		},
		{
			name: "fails when multiple routes use undefined receivers",
			fragment: &Fragment{
				Namespace: "test",
				Routes: []*config.Route{
					{Receiver: "undefined1"},
					{Receiver: "undefined2"},
				},
				Receivers: []*types.Receiver{
					{Name: "defined"},
				},
			},
			wantErr: "undefined receiver",
		},
		{
			name: "fails when route receiver is empty",
			fragment: &Fragment{
				Namespace: "test",
				Routes:    []*config.Route{{Receiver: ""}},
				Receivers: []*types.Receiver{
					{Name: "test"},
				},
			},
			wantErr: "undefined receiver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fragment.validate()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("validate() error = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("validate() error = %q, want error containing %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("validate() returned error: %v", err)
			}
		})
	}
}

func TestHasConflictString(t *testing.T) {
	tests := []struct {
		name string
		src  string
		dst  string
		want bool
	}{
		{"both empty", "", "", false},
		{"src empty", "", "database", false},
		{"dst empty", "database", "", false},
		{"same values", "database", "database", false},
		{"different values", "payments", "database", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasConflictString(tt.src, tt.dst); got != tt.want {
				t.Fatalf("hasConflictString(%q, %q) = %v, want %v", tt.src, tt.dst, got, tt.want)
			}
		})
	}
}

func TestIsEmptyGroup(t *testing.T) {
	tests := []struct {
		name  string
		group *FragmentGroup
		want  bool
	}{
		{"nil group", nil, true},
		{"empty group", &FragmentGroup{}, true},
		{"only receiver set", &FragmentGroup{Receiver: "team-db"}, false},
		{"only match set", &FragmentGroup{Match: map[string]string{"team": "db"}}, false},
		{"both set", &FragmentGroup{Receiver: "team-db", Match: map[string]string{"team": "db"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmptyGroup(tt.group); got != tt.want {
				t.Fatalf("isEmptyGroup(%#v) = %v, want %v", tt.group, tt.want, got)
			}
		})
	}
}

func TestHasConflictGroup(t *testing.T) {
	tests := []struct {
		name string
		src  *FragmentGroup
		dst  *FragmentGroup
		want bool
	}{
		{
			name: "both nil",
			src:  nil,
			dst:  nil,
			want: false,
		},
		{
			name: "dst nil",
			src:  &FragmentGroup{Receiver: "team-db"},
			dst:  nil,
			want: false,
		},
		{
			name: "src nil",
			src:  nil,
			dst:  &FragmentGroup{Receiver: "team-db"},
			want: false,
		},
		{
			name: "both empty",
			src:  &FragmentGroup{},
			dst:  &FragmentGroup{},
			want: false,
		},
		{
			name: "same values",
			src:  &FragmentGroup{Receiver: "team-db", Match: map[string]string{"team": "db"}},
			dst:  &FragmentGroup{Receiver: "team-db", Match: map[string]string{"team": "db"}},
			want: false,
		},
		{
			name: "different receiver",
			src:  &FragmentGroup{Receiver: "team-a"},
			dst:  &FragmentGroup{Receiver: "team-b"},
			want: true,
		},
		{
			name: "different match",
			src:  &FragmentGroup{Match: map[string]string{"team": "a"}},
			dst:  &FragmentGroup{Match: map[string]string{"team": "b"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasConflictGroup(tt.src, tt.dst); got != tt.want {
				t.Fatalf("hasConflictGroup(%#v, %#v) = %v, want %v", tt.src, tt.dst, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantDir string
	}{
		{"absolute path", "/some/path", "/some/path"},
		{"relative path", "relative/path", "relative/path"},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frag := New(tt.dir)
			if frag.dir != tt.wantDir {
				t.Errorf("New(%q).dir = %q, want %q", tt.dir, frag.dir, tt.wantDir)
			}
		})
	}
}

func TestFragmentGroupDefault(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantGroup bool
	}{
		{
			name: "group remains nil when not set in any file",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "fragment.yaml", fixtures.MustRead("fragment/routes-critical.yaml"))
				return dir
			},
			wantGroup: false,
		},
		{
			name: "group is set when present in file",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "fragment.yaml", fixtures.MustRead("fragment/group-default.yaml"))
				return dir
			},
			wantGroup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)

			frag := New(dir)
			_, err := frag.Read()

			if err != nil {
				t.Fatalf("Read() returned error: %v", err)
			}

			hasGroup := frag.Group != nil
			if hasGroup != tt.wantGroup {
				t.Errorf("Group present = %v, want %v", hasGroup, tt.wantGroup)
			}

			if tt.wantGroup && frag.Group.Receiver != "default-receiver" {
				t.Errorf("Group.Receiver = %q, want %q", frag.Group.Receiver, "default-receiver")
			}
		})
	}
}

func TestMetadataCollection(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (string, string)
		wantDir   string
		wantFiles int
	}{
		{
			name: "metadata contains directory path and file list",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				dir := t.TempDir()
				writeFragmentFixture(t, dir, "fragment.yaml", `namespace: "test"`)
				writeFragmentFixture(t, dir, "routes.yaml", fixtures.MustRead("fragment/routes-receivers-bare.yaml"))
				return dir, dir
			},
			wantFiles: 2,
		},
		{
			name: "metadata directory is parent when given file path",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				dir := t.TempDir()
				file := writeFragmentFixture(t, dir, "fragment.yaml", `namespace: "test"`)
				return file, dir
			},
			wantFiles: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, wantDir := tt.setup(t)

			frag := New(path)
			meta, err := frag.Read()

			if err != nil {
				t.Fatalf("Read() returned error: %v", err)
			}

			if meta.Dir != wantDir {
				t.Errorf("Meta.Dir = %q, want %q", meta.Dir, wantDir)
			}

			if len(meta.Files) != tt.wantFiles {
				t.Errorf("Meta.Files length = %d, want %d", len(meta.Files), tt.wantFiles)
			}
		})
	}
}

// --- Test-file tagging hardening ---
//
// Old behavior: strings.Contains(file, "tests") tagged any file whose path
// contained the substring "tests" — including spurious matches like
// "contests/", "tests-other/", and any tempdir whose name happened to
// contain those letters. New behavior: tag only files whose base name
// matches "*-tests.{yaml,yml}" OR that live under a directory component
// named exactly "tests".

func TestReadTestFileTagging(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantTests int
	}{
		{
			name: "file under tests/ subdirectory is tagged",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				writeFragmentFixture(t, root, "fragment.yaml", fixtures.MustRead("fragment/name-x.yaml"))
				writeFragmentFixture(t, filepath.Join(root, "tests"), "case.yaml", fixtures.MustRead("tests/case-t1.yaml"))
				return root
			},
			wantTests: 1,
		},
		{
			name: "sibling -tests.yaml is tagged",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				writeFragmentFixture(t, root, "fragment.yaml", fixtures.MustRead("fragment/name-x.yaml"))
				writeFragmentFixture(t, root, "fragment-tests.yaml", fixtures.MustRead("tests/case-t1.yaml"))
				return root
			},
			wantTests: 1,
		},
		{
			name: "sibling -tests.yml is tagged",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				writeFragmentFixture(t, root, "fragment.yaml", fixtures.MustRead("fragment/name-x.yaml"))
				writeFragmentFixture(t, root, "fragment-tests.yml", fixtures.MustRead("tests/case-t1.yaml"))
				return root
			},
			wantTests: 1,
		},
		{
			name: "directory whose name contains 'tests' as substring does not tag",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				dir := filepath.Join(root, "contests")
				writeFragmentFixture(t, dir, "fragment.yaml", fixtures.MustRead("fragment/name-x-with-tests.yaml"))
				return dir
			},
			wantTests: 0,
		},
		{
			name: "directory 'tests-other' does not tag",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				dir := filepath.Join(root, "tests-other")
				writeFragmentFixture(t, dir, "fragment.yaml", fixtures.MustRead("fragment/name-x-with-tests.yaml"))
				return dir
			},
			wantTests: 0,
		},
		{
			name: "regular fragment.yaml drops Tests field",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				writeFragmentFixture(t, root, "fragment.yaml", fixtures.MustRead("fragment/name-x-with-tests.yaml"))
				return root
			},
			wantTests: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			frag := New(dir)
			if _, err := frag.Read(); err != nil {
				t.Fatalf("Read() returned error: %v", err)
			}
			if len(frag.Tests) != tt.wantTests {
				t.Fatalf("Tests length = %d, want %d", len(frag.Tests), tt.wantTests)
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/x/tests/foo.yaml", true},
		{"/x/tests/sub/foo.yaml", true},
		{"/x/foo-tests.yaml", true},
		{"/x/foo-tests.yml", true},
		{"/x/foo-tests-extra.yaml", true},
		{"/x/-tests-prefix.yml", true},
		{"/x/contests/foo.yaml", false},
		{"/x/tests-other/foo.yaml", false},
		{"/x/foo.yaml", false},
		{"/x/mytests.yaml", false},
		{"/x/contests.yaml", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isTestFile(tt.path); got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// --- Defensive validate (no panic on malformed receiver) ---

func TestValidateRejectsMalformedReceiver(t *testing.T) {
	tests := []struct {
		name     string
		fragment *Fragment
		wantErr  string
	}{
		{
			name: "receiver missing name field returns error",
			fragment: &Fragment{
				Namespace: "test",
				Receivers: []*types.Receiver{{WebhookConfigs: []map[string]any{{"url": "http://example.com"}}}},
			},
			wantErr: "name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fragment.validate()
			if err == nil {
				t.Fatalf("validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("validate() error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

// --- Error wrapping with context ---

func TestReadWrapsStatError(t *testing.T) {
	frag := New("/non/existent/path/__hardening__")
	_, err := frag.Read()
	if err == nil {
		t.Fatal("Read() = nil, want error")
	}
	if !strings.Contains(err.Error(), "stat fragment") {
		t.Errorf("Read() error = %q, want wrapped 'stat fragment' prefix", err)
	}
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Read() error = %q, must still contain 'no such file or directory'", err)
	}
}

func TestLoadWrapsReadDirError(t *testing.T) {
	_, err := Load("/non/existent/path/__hardening__")
	if err == nil {
		t.Fatal("Load() = nil, want error")
	}
	if !strings.Contains(err.Error(), "read fragments dir") {
		t.Errorf("Load() error = %q, want wrapped 'read fragments dir' prefix", err)
	}
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Load() error = %q, must still contain 'no such file or directory'", err)
	}
}

func TestReadWrapsYAMLError(t *testing.T) {
	dir := t.TempDir()
	writeFragmentFixture(t, dir, "bad.yaml", fixtures.MustRead("fragment/invalid.yaml"))

	frag := New(dir)
	_, err := frag.Read()
	if err == nil {
		t.Fatal("Read() = nil, want error")
	}
	if !strings.Contains(err.Error(), "bad.yaml") {
		t.Errorf("Read() error = %q, want file path %q in error", err, "bad.yaml")
	}
	if !strings.Contains(err.Error(), "yaml: line") {
		t.Errorf("Read() error = %q, must still contain 'yaml: line' substring", err)
	}
}

// --- Nil return on error ---

func TestLoadReturnsNilOnError(t *testing.T) {
	result, err := Load("/non/existent/path/__hardening__")
	if err == nil {
		t.Fatal("Load() = nil err, want error")
	}
	if result != nil {
		t.Errorf("Load() result = %#v, want nil on error", result)
	}
}

func TestReadReturnsNilOnStatError(t *testing.T) {
	frag := New("/non/existent/path/__hardening__")
	meta, err := frag.Read()
	if err == nil {
		t.Fatal("Read() = nil err, want error")
	}
	if meta != nil {
		t.Errorf("Read() meta = %#v, want nil on error", meta)
	}
}

func TestReadReturnsNilOnYAMLError(t *testing.T) {
	dir := t.TempDir()
	writeFragmentFixture(t, dir, "bad.yaml", fixtures.MustRead("fragment/invalid.yaml"))

	frag := New(dir)
	meta, err := frag.Read()
	if err == nil {
		t.Fatal("Read() = nil err, want error")
	}
	if meta != nil {
		t.Errorf("Read() meta = %#v, want nil on YAML error", meta)
	}
}

func TestLoadReturnsNilOnFragmentReadError(t *testing.T) {
	root := t.TempDir()
	subdir := filepath.Join(root, "broken")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writeFragmentFixture(t, subdir, "bad.yaml", fixtures.MustRead("fragment/invalid.yaml"))

	result, err := Load(root)
	if err == nil {
		t.Fatal("Load() = nil err, want error")
	}
	if result != nil {
		t.Errorf("Load() result = %#v, want nil on per-fragment error", result)
	}
}

// --- Tests must never serialize on Fragment ---

func TestFragmentYAMLOmitsTests(t *testing.T) {
	frag := &Fragment{
		Namespace: "x",
		Tests: []*types.TestCase{
			{Name: "t1", Type: "unit"},
		},
	}
	out, err := yaml.Marshal(frag)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(out), "tests") {
		t.Errorf("YAML output must omit tests key, got:\n%s", out)
	}
	if strings.Contains(string(out), "t1") {
		t.Errorf("YAML output must omit test data, got:\n%s", out)
	}
}

func TestNonTestFileDropsTestsKey(t *testing.T) {
	dir := t.TempDir()
	writeFragmentFixture(t, dir, "fragment.yaml", fixtures.MustRead("fragment/name-x-with-leaked.yaml"))

	frag := New(dir)
	if _, err := frag.Read(); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(frag.Tests) != 0 {
		t.Errorf("Tests must be empty on non-test file, got %d entries", len(frag.Tests))
	}
}

func TestTestFileParsedViaTestDoc(t *testing.T) {
	root := t.TempDir()
	writeFragmentFixture(t, root, "fragment.yaml", fixtures.MustRead("fragment/name-real.yaml"))
	writeFragmentFixture(t, filepath.Join(root, "tests"), "case.yaml", fixtures.MustRead("tests/polluting-name.yaml"))

	frag := New(root)
	if _, err := frag.Read(); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if frag.Namespace != "real" {
		t.Errorf("Fragment.Namespace = %q, want %q (test file must not pollute Fragment fields)", frag.Namespace, "real")
	}
	if len(frag.Tests) != 1 {
		t.Fatalf("Tests length = %d, want 1", len(frag.Tests))
	}
	if frag.Tests[0].Name != "t1" {
		t.Errorf("Tests[0].Name = %q, want %q", frag.Tests[0].Name, "t1")
	}
	if frag.Tests[0].Type != "unit" {
		t.Errorf("Tests[0].Type = %q, want %q", frag.Tests[0].Type, "unit")
	}
}

func TestAugmentedFragmentSerializesTests(t *testing.T) {
	root := t.TempDir()
	writeFragmentFixture(t, filepath.Join(root, "frag"), "fragment.yaml", fixtures.MustRead("fragment/name-frag.yaml"))
	writeFragmentFixture(t, filepath.Join(root, "frag", "tests"), "case.yaml", fixtures.MustRead("tests/case-case1.yaml"))

	result, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	out, err := yaml.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), "tests:") {
		t.Errorf("YAML output must contain tests key on AugmentedFragment, got:\n%s", out)
	}
	if !strings.Contains(string(out), "case1") {
		t.Errorf("YAML output must contain test data, got:\n%s", out)
	}
}

func TestReadSkipsNonYAMLFiles(t *testing.T) {
	dir := t.TempDir()
	writeFragmentFixture(t, dir, "fragment.yaml", fixtures.MustRead("fragment/database-full.yaml"))
	writeFragmentFixture(t, dir, "README.md", "# notes")
	writeFragmentFixture(t, dir, "notes.txt", "scratch")
	writeFragmentFixture(t, dir, "config.json", `{"x":1}`)

	frag := New(dir)
	meta, err := frag.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(meta.Files) != 1 {
		t.Errorf("Files length = %d, want 1 (only fragment.yaml should be tracked)", len(meta.Files))
	}
	if frag.Namespace != "db" {
		t.Errorf("Namespace = %q, want %q", frag.Namespace, "db")
	}
}

func TestReadWrapsYAMLErrorInTestFile(t *testing.T) {
	root := t.TempDir()
	writeFragmentFixture(t, root, "fragment.yaml", fixtures.MustRead("fragment/name-x.yaml"))
	writeFragmentFixture(t, filepath.Join(root, "tests"), "bad.yaml", fixtures.MustRead("fragment/invalid.yaml"))

	frag := New(root)
	_, err := frag.Read()
	if err == nil {
		t.Fatal("Read = nil err, want yaml error from test-file branch")
	}
	if !strings.Contains(err.Error(), "parse yaml") {
		t.Errorf("err = %q, want wrapped 'parse yaml' prefix", err)
	}
	if !strings.Contains(err.Error(), "bad.yaml") {
		t.Errorf("err = %q, want file path 'bad.yaml' in error", err)
	}
}

// keep config import live for any future expansion
var _ = config.Route{}

func writeFragmentFixture(t *testing.T, dir, name, contents string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) returned error: %v", path, err)
	}
	return path
}
