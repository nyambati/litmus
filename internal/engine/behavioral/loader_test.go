package behavioral

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBehavioralTestLoader_LoadFromFile(t *testing.T) {
	// Create temp test file
	testYAML := `
name: "API service routing"
tags:
  - routing
state:
  active_alerts:
    - labels:
        service: "api"
        severity: "critical"
  silences:
    - labels:
        service: "maintenance"
      comment: "scheduled maintenance"
alert:
  labels:
    service: "api"
    severity: "critical"
expect:
  outcome: "active"
  receivers:
    - "api-team"
    - "on-call"
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")
	err := os.WriteFile(testFile, []byte(testYAML), 0644)
	require.NoError(t, err)

	loader := NewBehavioralTestLoader()
	tests, err := loader.LoadFromFile(testFile)

	require.NoError(t, err)
	require.Len(t, tests, 1)

	test := tests[0]
	require.Equal(t, "API service routing", test.Name)
	require.Len(t, test.Tags, 1)
	require.Equal(t, "routing", test.Tags[0])
	require.NotNil(t, test.State)
	require.Len(t, test.State.ActiveAlerts, 1)
	require.Len(t, test.State.Silences, 1)
	require.Equal(t, "active", test.Expect.Outcome)
	require.Contains(t, test.Expect.Receivers, "api-team")
}

func TestBehavioralTestLoader_LoadFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "test1.yml",
			yaml: `
name: "Test 1"
tags: ["test"]
alert:
  labels:
    service: "api"
expect:
  outcome: "active"
  receivers: ["api-team"]
`,
		},
		{
			name: "test2.yml",
			yaml: `
name: "Test 2"
tags: ["test"]
alert:
  labels:
    service: "db"
expect:
  outcome: "active"
  receivers: ["db-team"]
`,
		},
	}

	for _, tc := range tests {
		testFile := filepath.Join(tmpDir, tc.name)
		err := os.WriteFile(testFile, []byte(tc.yaml), 0644)
		require.NoError(t, err)
	}

	loader := NewBehavioralTestLoader()
	allTests, err := loader.LoadFromDirectory(tmpDir)

	require.NoError(t, err)
	require.Len(t, allTests, 2)

	names := make(map[string]bool)
	for _, test := range allTests {
		names[test.Name] = true
	}
	require.True(t, names["Test 1"])
	require.True(t, names["Test 2"])
}
