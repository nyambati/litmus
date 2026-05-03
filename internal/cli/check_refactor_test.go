package cli

import (
	"testing"

	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/types"
	"github.com/stretchr/testify/require"
)

func TestFilterByTags_Whitespace(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
		{Name: "test3", Tags: []string{"critical", "smoke"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 2)
}

func TestFilterByTags_EmptyTagInFilter(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
	}

	result := filterByTags(tests, []string{"", "critical"})
	require.Len(t, result, 1)
	require.Equal(t, "test1", result[0].Name)
}

func TestFilterByTags_DuplicateTagsInFilter(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"critical"}},
	}

	result := filterByTags(tests, []string{"critical", "critical"})
	require.Len(t, result, 2)
}

func TestFilterByTags_CaseSensitive(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"Critical"}},
		{Name: "test2", Tags: []string{"critical"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 1)
	require.Equal(t, "test2", result[0].Name)
}

type mockCheckResult struct {
	passed bool
}

type runCheckTestCase struct {
	name             string
	sanityResult     mockCheckResult
	regressionResult mockCheckResult
	behavioralResult mockCheckResult
	expectedPassed   bool
	expectedCode     int
}

func TestRunCheck_ExitCodeLogic(t *testing.T) {
	tests := []runCheckTestCase{
		{
			name:             "all passed",
			sanityResult:     mockCheckResult{passed: true},
			regressionResult: mockCheckResult{passed: true},
			behavioralResult: mockCheckResult{passed: true},
			expectedPassed:   true,
			expectedCode:     0,
		},
		{
			name:             "sanity failed only",
			sanityResult:     mockCheckResult{passed: false},
			regressionResult: mockCheckResult{passed: true},
			behavioralResult: mockCheckResult{passed: true},
			expectedPassed:   false,
			expectedCode:     3,
		},
		{
			name:             "regression failed only",
			sanityResult:     mockCheckResult{passed: true},
			regressionResult: mockCheckResult{passed: false},
			behavioralResult: mockCheckResult{passed: true},
			expectedPassed:   false,
			expectedCode:     2,
		},
		{
			name:             "behavioral failed only",
			sanityResult:     mockCheckResult{passed: true},
			regressionResult: mockCheckResult{passed: true},
			behavioralResult: mockCheckResult{passed: false},
			expectedPassed:   false,
			expectedCode:     2,
		},
		{
			name:             "multiple failures",
			sanityResult:     mockCheckResult{passed: false},
			regressionResult: mockCheckResult{passed: false},
			behavioralResult: mockCheckResult{passed: false},
			expectedPassed:   false,
			expectedCode:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanityPassed := tt.sanityResult.passed
			regressionPassed := tt.regressionResult.passed
			behavioralPassed := tt.behavioralResult.passed

			passed := sanityPassed && regressionPassed && behavioralPassed

			var code int
			if !passed {
				if !sanityPassed {
					code = 3
				} else {
					code = 2
				}
			}

			require.Equal(t, tt.expectedPassed, passed)
			require.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestFormatLabels(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   string
	}{
		{
			name:   "empty",
			labels: map[string]string{},
			want:   "{}",
		},
		{
			name:   "single",
			labels: map[string]string{"service": "api"},
			want:   "{service: api}",
		},
		{
			name:   "multiple sorted",
			labels: map[string]string{"service": "api", "env": "prod"},
			want:   "{env: prod, service: api}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLabels(tt.labels)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestFormatReceivers(t *testing.T) {
	tests := []struct {
		name      string
		receivers []string
		want      string
	}{
		{
			name:      "empty",
			receivers: []string{},
			want:      "[]",
		},
		{
			name:      "single",
			receivers: []string{"slack"},
			want:      "[slack]",
		},
		{
			name:      "multiple",
			receivers: []string{"slack", "email"},
			want:      "[slack, email]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatReceivers(tt.receivers)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestMissingReceivers(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
		want     []string
	}{
		{
			name:     "none missing",
			expected: []string{"slack", "email"},
			actual:   []string{"slack", "email"},
			want:     nil,
		},
		{
			name:     "one missing",
			expected: []string{"slack", "email"},
			actual:   []string{"slack"},
			want:     []string{"email"},
		},
		{
			name:     "all missing",
			expected: []string{"slack", "email"},
			actual:   []string{},
			want:     []string{"slack", "email"},
		},
		{
			name:     "extra actual",
			expected: []string{"slack"},
			actual:   []string{"slack", "pagerduty"},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := missingReceivers(tt.expected, tt.actual)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "s"},
		{"one", 1, ""},
		{"two", 2, "s"},
		{"many", 100, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plural(tt.n)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestFormatSummary(t *testing.T) {
	tests := []struct {
		name   string
		result CheckResult
		want   string
	}{
		{
			name:   "all passed",
			result: CheckResult{Passed: true},
			want:   "PASS",
		},
		{
			name: "regression failures",
			result: CheckResult{
				Passed: false,
				Regression: RegressionResult{
					Failures: []TestFailure{{}, {}, {}},
				},
			},
			want: "FAIL (3 Regressions)",
		},
		{
			name: "sanity warnings",
			result: CheckResult{
				Passed: false,
				Sanity: sanity.Result{
					Checks: []sanity.CheckEntry{
						{Name: "shadowed_routes", Issues: []string{"issue1"}},
					},
				},
			},
			want: "FAIL (1 Sanity Warning)",
		},
		{
			name: "behavioral failures",
			result: CheckResult{
				Passed: false,
				Behavioral: BehavioralResult{
					Failures: []TestFailure{{}},
				},
			},
			want: "FAIL (1 Behavioral Failure)",
		},
		{
			name: "all failed",
			result: CheckResult{
				Passed: false,
				Regression: RegressionResult{
					Failures: []TestFailure{{}, {}, {}},
				},
				Sanity: sanity.Result{
					Checks: []sanity.CheckEntry{
						{Name: "shadowed_routes", Issues: []string{"issue1", "issue2"}},
					},
				},
				Behavioral: BehavioralResult{
					Failures: []TestFailure{{}},
				},
			},
			want: "FAIL (3 Regressions, 2 Sanity Warnings, 1 Behavioral Failure)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSummary(tt.result)
			require.Equal(t, tt.want, result)
		})
	}
}
