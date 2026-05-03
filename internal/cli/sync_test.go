package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout redirects os.Stdout for the duration of fn and returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	old := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	fn()

	require.NoError(t, w.Close())
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data)
}

func TestPrintYAML_NoOutput_PrintsToStdout(t *testing.T) {
	yaml := "route:\n  receiver: default\n"

	out := captureStdout(t, func() {
		err := printYAML(yaml, "")
		require.NoError(t, err)
	})

	assert.Contains(t, out, yaml, "YAML must appear on stdout when no --output is given")
}

func TestPrintYAML_WithOutput_WritesFileAndDoesNotPrintYAML(t *testing.T) {
	yaml := "route:\n  receiver: default\n"
	dest := filepath.Join(t.TempDir(), "out.yaml")

	out := captureStdout(t, func() {
		err := printYAML(yaml, dest)
		require.NoError(t, err)
	})

	// File must contain the YAML.
	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, yaml, string(data))

	// stdout must show only the confirmation message, not the full YAML.
	assert.Contains(t, out, dest, "stdout must mention the output file path")
	assert.False(t, strings.Contains(out, "receiver: default"),
		"full YAML must NOT appear on stdout when --output is set")
}
