package snapshot

import (
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
			wantVals: []string{"api-"},
		},
		{
			name:     "wildcard",
			pattern:  ".*",
			wantVals: []string{"litmus_match"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewLabelCombinationGenerator(tt.maxCombinations)
			combos := gen.GenerateCovering(tt.matchers)

			require.Len(t, combos, tt.wantCount)
			if tt.wantAllOptionsHit {
				// Verify all options appear in at least one combination
				coverage := verifyCoverage(combos, tt.matchers)
				if !coverage {
					// Debug: show which options weren't covered
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
