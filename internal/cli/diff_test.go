package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nyambati/litmus/internal/types"
)

func TestPrintDiffReport_AllCases(t *testing.T) {
	diff := &types.RegressionDiff{
		Deltas: []types.RegressionDelta{
			{
				Kind:   types.DeltaAdded,
				Labels: map[string]string{"team": "new-team", "severity": "critical"},
				Actual: []string{"new-receiver"},
			},
			{
				Kind:     types.DeltaRemoved,
				Labels:   map[string]string{"team": "old-team", "severity": "warning"},
				Expected: []string{"old-receiver"},
			},
			{
				Kind:     types.DeltaModified,
				Labels:   map[string]string{"team": "existing-team", "severity": "info"},
				Expected: []string{"receiver-old"},
				Actual:   []string{"receiver-new"},
			},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintDiffReport(diff)

	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	result := string(output)

	// Verify all three cases are present
	if !strings.Contains(result, "[+] ADDED") {
		t.Error("ADDED case not found in output")
	}
	if !strings.Contains(result, "[-] REMOVED") {
		t.Error("REMOVED case not found in output")
	}
	if !strings.Contains(result, "[!] MODIFIED") {
		t.Error("MODIFIED case not found in output")
	}

	// Verify specific content
	if !strings.Contains(result, "new-receiver") {
		t.Error("new-receiver not in ADDED output")
	}
	if !strings.Contains(result, "old-receiver") {
		t.Error("old-receiver not in REMOVED output")
	}
	if !strings.Contains(result, "receiver-old") && !strings.Contains(result, "receiver-new") {
		t.Error("Modified receivers not in output")
	}
}

func TestPrintDiffReport_NoChanges(t *testing.T) {
	diff := &types.RegressionDiff{Deltas: []types.RegressionDelta{}}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintDiffReport(diff)

	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	result := string(output)

	if !strings.Contains(result, "No behavioral changes detected") {
		t.Error("Expected 'No behavioral changes detected' message")
	}
}
