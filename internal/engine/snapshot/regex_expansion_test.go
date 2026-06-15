package snapshot

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegexExpansion_ExpandAlternations(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		wantVals []string
	}{
		{
			name:     "simple alternation",
			pattern:  "(api|db)",
			wantVals: []string{"api", "db"},
		},
		{
			name:     "three options",
			pattern:  "(api|db|cache)",
			wantVals: []string{"api", "db", "cache"},
		},
		{
			name:     "no alternation",
			pattern:  "api",
			wantVals: []string{"api"},
		},
		{
			name:     "anchored prefix",
			pattern:  "^api-.*",
			wantVals: []string{"api-" + wildcardPlaceholder},
		},
		{
			name:     "pure wildcard dropped",
			pattern:  ".*",
			wantVals: nil,
		},
		{
			name:     "pure plus wildcard dropped",
			pattern:  ".+",
			wantVals: nil,
		},
		{
			name:     "non-capturing group single value",
			pattern:  "(?:latam-compliance-team)$",
			wantVals: []string{"latam-compliance-team"},
		},
		{
			name:     "non-capturing group alternation",
			pattern:  "(?:routing|assignment)$",
			wantVals: []string{"routing", "assignment"},
		},
		{
			name:     "bare alternation no parens",
			pattern:  "production|production.*",
			wantVals: []string{"production", "production" + wildcardPlaceholder},
		},
		{
			name:     "bare alternation with wildcards",
			pattern:  "prd.*|infra01|infra03|production|production.*",
			wantVals: []string{"prd" + wildcardPlaceholder, "infra01", "infra03", "production", "production" + wildcardPlaceholder},
		},
		{
			name:     "suffix wildcard",
			pattern:  "prd.*",
			wantVals: []string{"prd" + wildcardPlaceholder},
		},
		{
			name:     "suffix plus wildcard",
			pattern:  "prd.+",
			wantVals: []string{"prd" + wildcardPlaceholder},
		},
		{
			name:     "prefix wildcard",
			pattern:  ".*-eu-sim",
			wantVals: []string{wildcardPlaceholder + "-eu-sim"},
		},
		{
			name:     "infix wildcard",
			pattern:  "pre-.*-suf",
			wantVals: []string{"pre-" + wildcardPlaceholder + "-suf"},
		},
		{
			name:     "prefix plus wildcard",
			pattern:  ".+-eu-sim",
			wantVals: []string{wildcardPlaceholder + "-eu-sim"},
		},
		// Alertmanager wraps MatchRE patterns as ^(?:value)$ — these must clean to bare values.
		{
			name:     "alertmanager compiled single value",
			pattern:  "^(?:apac-compliance-team)$",
			wantVals: []string{"apac-compliance-team"},
		},
		{
			name:     "alertmanager compiled alternation",
			pattern:  "^(?:critical|warning)$",
			wantVals: []string{"critical", "warning"},
		},
		{
			name:     "alertmanager compiled alternation with wildcards",
			pattern:  "^(?:prd.*|infra01|production)$",
			wantVals: []string{"prd" + wildcardPlaceholder, "infra01", "production"},
		},
		{
			name:     "alertmanager compiled prefix wildcard",
			pattern:  "^(?:.*-eu-sim)$",
			wantVals: []string{wildcardPlaceholder + "-eu-sim"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := NewRegexExpander()
			vals := exp.ExpandAlternations(tt.pattern)
			require.Equal(t, tt.wantVals, vals)
		})
	}
}

func TestLabelCombinations_BalancedCovering(t *testing.T) {
	tests := []struct {
		name              string
		matchers          map[string][]string
		maxCombinations   int
		wantCount         int
		wantAllOptionsHit bool
	}{
		{
			name: "single option per matcher",
			matchers: map[string][]string{
				"service": {"api"},
				"env":     {"prod"},
			},
			maxCombinations:   5,
			wantCount:         1,
			wantAllOptionsHit: true,
		},
		{
			name: "small cartesian product",
			matchers: map[string][]string{
				"service": {"api", "db"},
				"env":     {"prod", "staging"},
			},
			maxCombinations:   5,
			wantCount:         4, // 2 * 2
			wantAllOptionsHit: true,
		},
		{
			name: "large cartesian needs covering set",
			matchers: map[string][]string{
				"service":  {"api", "db"},
				"env":      {"prod", "staging"},
				"severity": {"critical", "warning"},
			},
			maxCombinations:   4,
			wantCount:         4,
			wantAllOptionsHit: true,
		},
		{
			name: "very large matcher set does not OOM",
			matchers: map[string][]string{
				"l1": {"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10"},
				"l2": {"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10"},
				"l3": {"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10"},
				"l4": {"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10"},
				"l5": {"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10"},
				"l6": {"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "v10"},
			},
			maxCombinations:   20,
			wantCount:         20,
			wantAllOptionsHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewLabelCombinationGenerator(tt.maxCombinations)
			combos := gen.GenerateCovering(tt.matchers)

			require.LessOrEqual(t, len(combos), tt.maxCombinations)
			if tt.wantCount > 0 {
				require.GreaterOrEqual(t, len(combos), 1)
			}
			if tt.wantAllOptionsHit {
				coverage := verifyCoverage(combos, tt.matchers)
				if !coverage {
					for k, vals := range tt.matchers {
						for _, v := range vals {
							found := false
							for _, combo := range combos {
								if combo[k] == v {
									found = true
									break
								}
							}
							if !found {
								t.Logf("Not covered: %s=%s", k, v)
							}
						}
					}
				}
				require.True(t, coverage)
			}
		})
	}
}

// TestLabelCombinations_Deterministic ensures GenerateCovering produces an
// identical ordered result across runs. Map iteration order is randomized by
// the Go runtime, so without sorted keys the synthesized baseline would drift
// between runs and trigger spurious snapshot regressions in CI.
func TestLabelCombinations_Deterministic(t *testing.T) {
	cases := []struct {
		name            string
		matchers        map[string][]string
		maxCombinations int
	}{
		{
			name: "full cartesian product",
			matchers: map[string][]string{
				"service":  {"api", "db"},
				"env":      {"prod", "staging"},
				"severity": {"critical", "warning"},
			},
			maxCombinations: 100,
		},
		{
			name: "minimal covering set",
			matchers: map[string][]string{
				"l1": {"v1", "v2", "v3", "v4", "v5"},
				"l2": {"v1", "v2", "v3", "v4", "v5"},
				"l3": {"v1", "v2", "v3", "v4", "v5"},
			},
			maxCombinations: 8,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewLabelCombinationGenerator(tc.maxCombinations)
			want := gen.GenerateCovering(tc.matchers)

			// Re-run several times; map iteration order varies per run, so a
			// non-deterministic implementation would eventually diverge.
			for i := range 50 {
				got := NewLabelCombinationGenerator(tc.maxCombinations).GenerateCovering(tc.matchers)
				require.Equal(t, want, got, "run %d diverged from first result", i)
			}

			// No duplicate combinations within a single result.
			seen := make(map[string]bool)
			for _, combo := range want {
				key := comboKey(sortedKeys(combo), combo)
				require.False(t, seen[key], "duplicate combination: %v", combo)
				seen[key] = true
			}
		})
	}
}

// TestLabelCombinations_EmptyValueKeysDoNotPanic ensures keys with no concrete
// values are dropped rather than triggering an index/division-by-zero panic in
// either the cartesian or minimal-covering path.
func TestLabelCombinations_EmptyValueKeysDoNotPanic(t *testing.T) {
	matchers := map[string][]string{
		"service": {"api", "db"},
		"empty":   {},
	}

	// Small product -> cartesian path.
	require.NotPanics(t, func() {
		combos := NewLabelCombinationGenerator(100).GenerateCovering(matchers)
		for _, c := range combos {
			_, ok := c["empty"]
			require.False(t, ok, "empty-value key must not appear in combinations")
		}
	})

	// Force the minimal-covering path with a tight limit.
	require.NotPanics(t, func() {
		NewLabelCombinationGenerator(1).GenerateCovering(map[string][]string{
			"a":     {"1", "2", "3"},
			"b":     {"1", "2", "3"},
			"empty": {},
		})
	})
}

// sortedKeys returns the map keys in sorted order for stable comboKey input.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// verifyCoverage checks if all options appear in generated combinations.
func verifyCoverage(combos []map[string]string, matchers map[string][]string) bool {
	for k, vals := range matchers {
		for _, v := range vals {
			found := false
			for _, combo := range combos {
				if combo[k] == v {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}
