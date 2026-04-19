package snapshot

import (
	"testing"

	"github.com/nyambati/litmus/internal/types"
	"github.com/stretchr/testify/require"
)

func TestComputeDiff_Added(t *testing.T) {
	oldTests := []*types.TestCase{}
	newTests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route to team-a",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
	}

	diff := ComputeDiff(oldTests, newTests)
	require.Len(t, diff.Deltas, 1)
	require.Equal(t, types.DeltaAdded, diff.Deltas[0].Kind)
}

func TestComputeDiff_Removed(t *testing.T) {
	oldTests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route to team-a",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
	}
	newTests := []*types.TestCase{}

	diff := ComputeDiff(oldTests, newTests)
	require.Len(t, diff.Deltas, 1)
	require.Equal(t, types.DeltaRemoved, diff.Deltas[0].Kind)
}

func TestComputeDiff_Modified(t *testing.T) {
	oldTests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route to team-a",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
	}
	newTests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route to team-a-new",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a-new"},
		},
	}

	diff := ComputeDiff(oldTests, newTests)
	require.Len(t, diff.Deltas, 1)
	require.Equal(t, types.DeltaModified, diff.Deltas[0].Kind)
	require.Equal(t, "receiver-a", diff.Deltas[0].Expected[0])
	require.Equal(t, "receiver-a-new", diff.Deltas[0].Actual[0])
}

func TestComputeDiff_Unchanged(t *testing.T) {
	tests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route to team-a",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
	}

	diff := ComputeDiff(tests, tests)
	require.Empty(t, diff.Deltas)
}

func TestComputeDiff_Mixed(t *testing.T) {
	oldTests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route 1",
			Labels:   []map[string]string{{"team": "a"}},
			Expected: []string{"receiver-a"},
		},
		{
			Type:     "regression",
			Name:     "Route 2",
			Labels:   []map[string]string{{"team": "b"}},
			Expected: []string{"receiver-b"},
		},
		{
			Type:     "regression",
			Name:     "Route 3",
			Labels:   []map[string]string{{"team": "c"}},
			Expected: []string{"receiver-c"},
		},
	}
	newTests := []*types.TestCase{
		// Route 1: removed
		// Route 2: modified
		{
			Type:     "regression",
			Name:     "Route 2 Modified",
			Labels:   []map[string]string{{"team": "b"}},
			Expected: []string{"receiver-b-new"},
		},
		// Route 3: unchanged
		{
			Type:     "regression",
			Name:     "Route 3",
			Labels:   []map[string]string{{"team": "c"}},
			Expected: []string{"receiver-c"},
		},
		// Route 4: added
		{
			Type:     "regression",
			Name:     "Route 4",
			Labels:   []map[string]string{{"team": "d"}},
			Expected: []string{"receiver-d"},
		},
	}

	diff := ComputeDiff(oldTests, newTests)
	require.Len(t, diff.Deltas, 3)

	kinds := make(map[types.DeltaKind]int)
	for _, delta := range diff.Deltas {
		kinds[delta.Kind]++
	}

	require.Equal(t, 1, kinds[types.DeltaRemoved])
	require.Equal(t, 1, kinds[types.DeltaModified])
	require.Equal(t, 1, kinds[types.DeltaAdded])
}
