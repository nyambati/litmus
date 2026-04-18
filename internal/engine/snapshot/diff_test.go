package snapshot

import (
	"testing"

	"github.com/nyambati/litmus/internal/types"
)

func TestComputeDiff_Added(t *testing.T) {
	oldTests := []*types.RegressionTest{}
	newTests := []*types.RegressionTest{
		{
			Name:     "Route to team-a",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
	}

	diff := ComputeDiff(oldTests, newTests)
	if len(diff.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(diff.Deltas))
	}
	if diff.Deltas[0].Kind != types.DeltaAdded {
		t.Errorf("expected DeltaAdded, got %v", diff.Deltas[0].Kind)
	}
}

func TestComputeDiff_Removed(t *testing.T) {
	oldTests := []*types.RegressionTest{
		{
			Name:     "Route to team-a",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
	}
	newTests := []*types.RegressionTest{}

	diff := ComputeDiff(oldTests, newTests)
	if len(diff.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(diff.Deltas))
	}
	if diff.Deltas[0].Kind != types.DeltaRemoved {
		t.Errorf("expected DeltaRemoved, got %v", diff.Deltas[0].Kind)
	}
}

func TestComputeDiff_Modified(t *testing.T) {
	oldTests := []*types.RegressionTest{
		{
			Name:     "Route to team-a",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
	}
	newTests := []*types.RegressionTest{
		{
			Name:     "Route to team-a-new",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a-new"},
		},
	}

	diff := ComputeDiff(oldTests, newTests)
	if len(diff.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(diff.Deltas))
	}
	if diff.Deltas[0].Kind != types.DeltaModified {
		t.Errorf("expected DeltaModified, got %v", diff.Deltas[0].Kind)
	}
	if diff.Deltas[0].Expected[0] != "receiver-a" {
		t.Errorf("expected Expected='receiver-a', got %v", diff.Deltas[0].Expected)
	}
	if diff.Deltas[0].Actual[0] != "receiver-a-new" {
		t.Errorf("expected Actual='receiver-a-new', got %v", diff.Deltas[0].Actual)
	}
}

func TestComputeDiff_Mixed(t *testing.T) {
	oldTests := []*types.RegressionTest{
		{
			Name:     "Route 1",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
		{
			Name:     "Route 2",
			Labels:   []map[string]string{{"team": "b"}},
			Expected: []string{"receiver-b"},
		},
		{
			Name:     "Route 3",
			Labels:   []map[string]string{{"team": "c"}},
			Expected: []string{"receiver-c"},
		},
	}
	newTests := []*types.RegressionTest{
		// Route 1: removed
		// Route 2: modified
		{
			Name:     "Route 2 Modified",
			Labels:   []map[string]string{{"team": "b"}},
			Expected: []string{"receiver-b-new"},
		},
		// Route 3: unchanged
		{
			Name:     "Route 3",
			Labels:   []map[string]string{{"team": "c"}},
			Expected: []string{"receiver-c"},
		},
		// Route 4: added
		{
			Name:     "Route 4",
			Labels:   []map[string]string{{"team": "d"}},
			Expected: []string{"receiver-d"},
		},
	}

	diff := ComputeDiff(oldTests, newTests)
	if len(diff.Deltas) != 3 {
		t.Fatalf("expected 3 deltas (1 removed, 1 modified, 1 added), got %d", len(diff.Deltas))
	}

	// Count each kind
	kinds := make(map[types.DeltaKind]int)
	for _, delta := range diff.Deltas {
		kinds[delta.Kind]++
	}

	if kinds[types.DeltaRemoved] != 1 {
		t.Errorf("expected 1 removed, got %d", kinds[types.DeltaRemoved])
	}
	if kinds[types.DeltaModified] != 1 {
		t.Errorf("expected 1 modified, got %d", kinds[types.DeltaModified])
	}
	if kinds[types.DeltaAdded] != 1 {
		t.Errorf("expected 1 added, got %d", kinds[types.DeltaAdded])
	}
}
