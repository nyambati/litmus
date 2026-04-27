package matching

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExactMatch(t *testing.T) {
	tests := []struct {
		name     string
		actual   []string
		expected []string
		want     bool
	}{
		{
			name:     "both empty",
			actual:   []string{},
			expected: []string{},
			want:     true,
		},
		{
			name:     "single identical",
			actual:   []string{"slack"},
			expected: []string{"slack"},
			want:     true,
		},
		{
			name:     "single different",
			actual:   []string{"slack"},
			expected: []string{"email"},
			want:     false,
		},
		{
			name:     "same elements different order",
			actual:   []string{"slack", "email"},
			expected: []string{"email", "slack"},
			want:     true,
		},
		{
			name:     "different lengths - actual longer",
			actual:   []string{"slack", "email", "pagerduty"},
			expected: []string{"slack", "email"},
			want:     false,
		},
		{
			name:     "different lengths - expected longer",
			actual:   []string{"slack"},
			expected: []string{"slack", "email"},
			want:     false,
		},
		{
			name:     "multiple same elements",
			actual:   []string{"slack", "email", "pagerduty"},
			expected: []string{"email", "pagerduty", "slack"},
			want:     true,
		},
		{
			name:     "one duplicate in actual",
			actual:   []string{"slack", "slack"},
			expected: []string{"slack"},
			want:     false,
		},
		{
			name:     "one duplicate in expected",
			actual:   []string{"slack"},
			expected: []string{"slack", "slack"},
			want:     false,
		},
		{
			name:     "case sensitive - different cases",
			actual:   []string{"Slack"},
			expected: []string{"slack"},
			want:     false,
		},
		{
			name:     "with special characters",
			actual:   []string{"slack-channel", "email+prod"},
			expected: []string{"email+prod", "slack-channel"},
			want:     true,
		},
		{
			name:     "completely different",
			actual:   []string{"slack"},
			expected: []string{"email", "pagerduty"},
			want:     false,
		},
		{
			name:     "empty actual non-empty expected",
			actual:   []string{},
			expected: []string{"slack"},
			want:     false,
		},
		{
			name:     "non-empty actual empty expected",
			actual:   []string{"slack"},
			expected: []string{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExactMatch(tt.actual, tt.expected)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestSubsetMatch(t *testing.T) {
	tests := []struct {
		name     string
		actual   []string
		expected []string
		want     bool
	}{
		{
			name:     "both empty",
			actual:   []string{},
			expected: []string{},
			want:     true,
		},
		{
			name:     "empty expected - always true",
			actual:   []string{"slack", "email"},
			expected: []string{},
			want:     true,
		},
		{
			name:     "exact match",
			actual:   []string{"slack", "email"},
			expected: []string{"slack", "email"},
			want:     true,
		},
		{
			name:     "actual has extra - subset true",
			actual:   []string{"slack", "email", "pagerduty"},
			expected: []string{"slack", "email"},
			want:     true,
		},
		{
			name:     "expected has extra - subset false",
			actual:   []string{"slack"},
			expected: []string{"slack", "email"},
			want:     false,
		},
		{
			name:     "completely different",
			actual:   []string{"slack"},
			expected: []string{"email", "pagerduty"},
			want:     false,
		},
		{
			name:     "order independent",
			actual:   []string{"email", "slack"},
			expected: []string{"slack"},
			want:     true,
		},
		{
			name:     "case sensitive",
			actual:   []string{"Slack"},
			expected: []string{"slack"},
			want:     false,
		},
		{
			name:     "with special characters",
			actual:   []string{"slack-channel", "email+prod"},
			expected: []string{"slack-channel"},
			want:     true,
		},
		{
			name:     "with special characters not matching",
			actual:   []string{"slack-channel"},
			expected: []string{"email+prod"},
			want:     false,
		},
		{
			name:     "single actual missing expected",
			actual:   []string{"slack"},
			expected: []string{"email"},
			want:     false,
		},
		{
			name:     "empty actual with non-empty expected",
			actual:   []string{},
			expected: []string{"slack"},
			want:     false,
		},
		{
			name:     "duplicates in actual still matches",
			actual:   []string{"slack", "slack", "email"},
			expected: []string{"slack"},
			want:     true,
		},
		{
			name:     "duplicates in expected still matches",
			actual:   []string{"slack", "email"},
			expected: []string{"slack", "slack"},
			want:     true,
		},
		{
			name:     "subset with all special characters",
			actual:   []string{"a.b", "c-d", "e_f"},
			expected: []string{"a.b", "c-d"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubsetMatch(tt.actual, tt.expected)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestExactMatch_PropertyBased(t *testing.T) {
	t.Run("symmetry - ExactMatch(a,b) == ExactMatch(b,a)", func(t *testing.T) {
		pairs := []struct {
			a []string
			b []string
		}{
			{[]string{"slack", "email"}, []string{"email", "slack"}},
			{[]string{"a", "b", "c"}, []string{"c", "b", "a"}},
			{[]string{}, []string{}},
		}
		for _, p := range pairs {
			resultA := ExactMatch(p.a, p.b)
			resultB := ExactMatch(p.b, p.a)
			require.Equal(t, resultA, resultB, "Expected symmetry for a=%v, b=%v", p.a, p.b)
		}
	})

	t.Run("idempotence - ExactMatch(a,a) == true for non-empty", func(t *testing.T) {
		cases := [][]string{
			{"slack"},
			{"slack", "email"},
			{"a", "b", "c"},
		}
		for _, c := range cases {
			result := ExactMatch(c, c)
			require.True(t, result, "Expected ExactMatch(a,a) to be true for %v", c)
		}
	})
}

func TestSubsetMatch_PropertyBased(t *testing.T) {
	t.Run("symmetry - SubsetMatch(a,b) != SubsetMatch(b,a) in general", func(t *testing.T) {
		actual := []string{"slack", "email", "pagerduty"}
		expected := []string{"slack", "email"}

		result := SubsetMatch(actual, expected)
		require.True(t, result, "SubsetMatch(actual, subset) should be true")

		resultReverse := SubsetMatch(expected, actual)
		require.False(t, resultReverse, "SubsetMatch(subset, actual) should be false")
	})

	t.Run("empty expected always true", func(t *testing.T) {
		emptyExpected := []string{}
		nonEmptyActual := []string{"slack", "email"}

		result := SubsetMatch(nonEmptyActual, emptyExpected)
		require.True(t, result, "SubsetMatch with empty expected should always be true")
	})

	t.Run("empty actual with non-empty expected always false", func(t *testing.T) {
		emptyActual := []string{}
		nonEmptyExpected := []string{"slack"}

		result := SubsetMatch(emptyActual, nonEmptyExpected)
		require.False(t, result, "SubsetMatch with empty actual and non-empty expected should be false")
	})
}
