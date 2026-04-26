package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFragments(t *testing.T) {
	tmpDir := t.TempDir()

	type fileContent struct {
		path    string
		content string
	}

	tests := []struct {
		name          string
		setupFiles    []fileContent
		pattern       string
		wantName      string
		wantNamespace string
		wantReceivers int
		wantRoutes    int
		wantTests     int
	}{
		{
			name: "Single File Fragment",
			setupFiles: []fileContent{
				{
					path: "db.yml",
					content: `
name: "db-fragment"
namespace: "db"
receivers: [{name: "critical"}]`,
				},
			},
			pattern:       "*.yml",
			wantName:      "db-fragment",
			wantNamespace: "db",
			wantReceivers: 1,
		},
		{
			name: "Folder Package with Merged Files",
			setupFiles: []fileContent{
				{
					path: "database/receivers.yml",
					content: `
namespace: "db"
receivers: [{name: "critical"}]`,
				},
				{
					path: "database/routes.yml",
					content: `
routes:
  - receiver: "critical"
    matchers: [ service=mysql ]`,
				},
				{
					path: "database/tests/unit.yml",
					content: `
- name: "test mysql"
  expect: { outcome: active }`,
				},
			},
			pattern:       "*",
			wantName:      "database",
			wantNamespace: "db",
			wantReceivers: 1,
			wantRoutes:    1,
			wantTests:     1,
		},
		{
			name: "Test Discovery - Sibling and Folder",
			setupFiles: []fileContent{
				{path: "net/fragment.yml", content: `name: "net"`},
				{
					path: "net/fragment-tests.yml",
					content: `- name: "sibling test"
  expect: { outcome: active }`,
				},
				{
					path: "net/tests/sub-test.yml",
					content: `- name: "folder test"
  expect: { outcome: active }`,
				},
			},
			pattern:   "*",
			wantName:  "net",
			wantTests: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and setup files for each sub-test
			caseDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(caseDir, 0755)
			require.NoError(t, err)

			for _, f := range tt.setupFiles {
				fullPath := filepath.Join(caseDir, f.path)
				err := os.MkdirAll(filepath.Dir(fullPath), 0755)
				require.NoError(t, err)
				err = os.WriteFile(fullPath, []byte(f.content), 0600)
				require.NoError(t, err)
			}

			frags, err := LoadFragments(filepath.Join(caseDir, tt.pattern))
			require.NoError(t, err)
			require.NotEmpty(t, frags)

			f := frags[0]
			assert.Equal(t, tt.wantName, f.Name)
			assert.Equal(t, tt.wantNamespace, f.Namespace)
			assert.Len(t, f.Receivers, tt.wantReceivers)
			assert.Len(t, f.Routes, tt.wantRoutes)
			assert.Len(t, f.Tests, tt.wantTests)
		})
	}
}

func TestLoadTestsFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantCount int
	}{
		{
			name: "Single Object Test",
			content: `
name: "single"
expect: { outcome: active }`,
			wantCount: 1,
		},
		{
			name: "List of Tests",
			content: `
- name: "t1"
- name: "t2"`,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".yml")
			err := os.WriteFile(path, []byte(tt.content), 0600)
			require.NoError(t, err)

			got, err := loadTestsFromFile(path)
			require.NoError(t, err)
			assert.Len(t, got, tt.wantCount)
		})
	}
}

func TestLoadFragments_SiblingTestYAMLExtension(t *testing.T) {
	// Sibling file with .yaml extension (not .yml) must also be discovered.
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "frag.yml"), []byte(`name: "net"`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "frag-tests.yaml"), []byte(`
- name: "yaml ext test"
  expect: {outcome: active}
`), 0600))

	frags, err := LoadFragments(filepath.Join(tmpDir, "*.yml"))
	require.NoError(t, err)
	require.Len(t, frags, 1)
	assert.Len(t, frags[0].Tests, 1, "sibling with .yaml extension must be discovered")
	assert.Equal(t, "yaml ext test", frags[0].Tests[0].Name)
}

func TestLoadFragments_GroupConflict_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	fragDir := filepath.Join(tmpDir, "db-team")
	require.NoError(t, os.MkdirAll(fragDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(fragDir, "receivers.yml"), []byte(`
group:
  match: {scope: teams}
receivers: [{name: critical}]
`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(fragDir, "routes.yml"), []byte(`
group:
  match: {scope: platform}
routes:
  - receiver: critical
`), 0600))

	_, err := LoadFragments(filepath.Join(tmpDir, "*"))
	assert.Error(t, err, "conflicting group definitions in folder package must return an error")
}

func TestLoadFragments_MalformedYAML_ReturnsError(t *testing.T) {
	// A folder package containing a file with invalid YAML must propagate an error.
	tmpDir := t.TempDir()
	fragDir := filepath.Join(tmpDir, "bad-frag")
	require.NoError(t, os.MkdirAll(fragDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(fragDir, "broken.yml"), []byte(`: invalid: yaml:`), 0600))

	_, err := LoadFragments(filepath.Join(tmpDir, "*"))
	assert.Error(t, err, "malformed YAML in fragment folder must return an error")
}

func TestLoadFragments_AbsolutePattern(t *testing.T) {
	// Absolute glob pattern must work regardless of working directory.
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "sre.yml"), []byte(`
name: "sre"
namespace: "sre"
receivers: [{name: "pagerduty"}]
`), 0600))

	// Pattern is absolute — points directly to tmpDir/*.yml
	frags, err := LoadFragments(filepath.Join(tmpDir, "*.yml"))
	require.NoError(t, err)
	require.Len(t, frags, 1)
	assert.Equal(t, "sre", frags[0].Namespace)
}
