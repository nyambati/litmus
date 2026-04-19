package snapshot

import (
	"sort"
	"strings"

	"github.com/nyambati/litmus/internal/engine/matching"
	"github.com/nyambati/litmus/internal/types"
)

// ComputeDiff compares two sets of regression TestCases and identifies deltas.
// Tests are matched by their first label set (assuming canonical synthesis).
func ComputeDiff(oldTests, newTests []*types.TestCase) *types.RegressionDiff {
	diff := &types.RegressionDiff{Deltas: []types.RegressionDelta{}}

	oldIdx := indexByLabels(oldTests)
	newIdx := indexByLabels(newTests)

	// Identify Added and Modified
	for labelKey, newTest := range newIdx {
		if len(newTest.Labels) == 0 {
			continue
		}
		oldTest, exists := oldIdx[labelKey]
		if !exists {
			diff.Deltas = append(diff.Deltas, types.RegressionDelta{
				Kind:   types.DeltaAdded,
				Labels: newTest.Labels[0],
				Actual: newTest.Expected,
			})
			continue
		}

		if !matching.ExactMatch(newTest.Expected, oldTest.Expected) {
			diff.Deltas = append(diff.Deltas, types.RegressionDelta{
				Kind:     types.DeltaModified,
				Labels:   newTest.Labels[0],
				Expected: oldTest.Expected,
				Actual:   newTest.Expected,
			})
		}
	}

	// Identify Removed
	for labelKey, oldTest := range oldIdx {
		if len(oldTest.Labels) == 0 {
			continue
		}
		if _, exists := newIdx[labelKey]; !exists {
			diff.Deltas = append(diff.Deltas, types.RegressionDelta{
				Kind:     types.DeltaRemoved,
				Labels:   oldTest.Labels[0],
				Expected: oldTest.Expected,
			})
		}
	}

	return diff
}

// indexByLabels creates a map keyed by a canonical string representation of labels.
func indexByLabels(tests []*types.TestCase) map[string]*types.TestCase {
	idx := make(map[string]*types.TestCase)
	for _, t := range tests {
		if len(t.Labels) == 0 {
			continue
		}
		key := labelKey(t.Labels[0])
		idx[key] = t
	}
	return idx
}

// labelKey produces a stable string key for a label map.
func labelKey(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(m[k])
	}
	return b.String()
}
