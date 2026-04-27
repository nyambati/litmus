package labelmatcher

import (
	"testing"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func mustMatcher(t *testing.T, typ labels.MatchType, name, value string) *labels.Matcher {
	t.Helper()
	m, err := labels.NewMatcher(typ, name, value)
	require.NoError(t, err)
	return m
}

func TestSourceMatches_ExactMatch(t *testing.T) {
	rule := config.InhibitRule{
		SourceMatch: map[string]string{"service": "api", "env": "prod"},
	}
	labels := model.LabelSet{
		"service": "api",
		"env":     "prod",
		"region":  "us-east",
	}

	result := SourceMatches(rule, labels)
	require.True(t, result)
}

func TestSourceMatches_ExactMatch_Fails(t *testing.T) {
	rule := config.InhibitRule{
		SourceMatch: map[string]string{"service": "api"},
	}
	labels := model.LabelSet{
		"service": "db",
	}

	result := SourceMatches(rule, labels)
	require.False(t, result)
}

func TestSourceMatchers_RegexMatch(t *testing.T) {
	m := mustMatcher(t, labels.MatchRegexp, "service", "api.*")
	rule := config.InhibitRule{
		SourceMatchers: []*labels.Matcher{m},
	}
	labels := model.LabelSet{
		"service": "api-gateway",
	}

	result := SourceMatches(rule, labels)
	require.True(t, result)
}

func TestSourceMatchers_RegexMatch_Fails(t *testing.T) {
	m := mustMatcher(t, labels.MatchRegexp, "service", "api.*")
	rule := config.InhibitRule{
		SourceMatchers: []*labels.Matcher{m},
	}
	labels := model.LabelSet{
		"service": "db-primary",
	}

	result := SourceMatches(rule, labels)
	require.False(t, result)
}

func TestSourceMatches_MixedExactAndRegex(t *testing.T) {
	m := mustMatcher(t, labels.MatchRegexp, "service", "api.*")
	rule := config.InhibitRule{
		SourceMatch:    map[string]string{"env": "prod"},
		SourceMatchers: []*labels.Matcher{m},
	}

	tests := []struct {
		name     string
		labels   model.LabelSet
		expected bool
	}{
		{"both match", model.LabelSet{"env": "prod", "service": "api-gateway"}, true},
		{"exact match but regex fails", model.LabelSet{"env": "prod", "service": "db-primary"}, false},
		{"regex match but exact fails", model.LabelSet{"env": "dev", "service": "api-gateway"}, false},
		{"both fail", model.LabelSet{"env": "dev", "service": "db-primary"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SourceMatches(rule, tt.labels)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSourceMatches_EmptyMatchers(t *testing.T) {
	rule := config.InhibitRule{}
	labels := model.LabelSet{"service": "api"}

	result := SourceMatches(rule, labels)
	require.True(t, result)
}

func TestTargetMatches_ExactMatch(t *testing.T) {
	rule := config.InhibitRule{
		TargetMatch: map[string]string{"severity": "warning"},
	}
	labels := model.LabelSet{
		"severity": "warning",
		"service":  "api",
	}

	result := TargetMatches(rule, labels)
	require.True(t, result)
}

func TestTargetMatches_ExactMatch_Fails(t *testing.T) {
	rule := config.InhibitRule{
		TargetMatch: map[string]string{"severity": "warning"},
	}
	labels := model.LabelSet{
		"severity": "critical",
	}

	result := TargetMatches(rule, labels)
	require.False(t, result)
}

func TestTargetMatchers_RegexMatch(t *testing.T) {
	m := mustMatcher(t, labels.MatchRegexp, "severity", "warn.*")
	rule := config.InhibitRule{
		TargetMatchers: []*labels.Matcher{m},
	}
	labels := model.LabelSet{
		"severity": "warning",
	}

	result := TargetMatches(rule, labels)
	require.True(t, result)
}

func TestTargetMatchers_RegexMatch_Fails(t *testing.T) {
	m := mustMatcher(t, labels.MatchRegexp, "severity", "warn.*")
	rule := config.InhibitRule{
		TargetMatchers: []*labels.Matcher{m},
	}
	labels := model.LabelSet{
		"severity": "critical",
	}

	result := TargetMatches(rule, labels)
	require.False(t, result)
}

func TestTargetMatchers_MixedEmptyMatchers(t *testing.T) {
	m := mustMatcher(t, labels.MatchRegexp, "severity", "warn.*")
	rule := config.InhibitRule{
		TargetMatchers: []*labels.Matcher{m},
	}
	labels := model.LabelSet{
		"severity": "warning",
	}

	result := TargetMatches(rule, labels)
	require.True(t, result)
}

func TestEqualLabelsMatch_AllEqual(t *testing.T) {
	equal := model.LabelNames{"service", "env"}
	source := model.LabelSet{"service": "api", "env": "prod", "region": "us-east"}
	target := model.LabelSet{"service": "api", "env": "prod", "region": "eu-west"}

	result := EqualLabelsMatch(equal, source, target)
	require.True(t, result)
}

func TestEqualLabelsMatch_OneNotEqual(t *testing.T) {
	equal := model.LabelNames{"service", "env"}
	source := model.LabelSet{"service": "api", "env": "prod"}
	target := model.LabelSet{"service": "api", "env": "dev"}

	result := EqualLabelsMatch(equal, source, target)
	require.False(t, result)
}

func TestEqualLabelsMatch_Empty(t *testing.T) {
	equal := model.LabelNames{}
	source := model.LabelSet{"service": "api"}
	target := model.LabelSet{"service": "db"}

	result := EqualLabelsMatch(equal, source, target)
	require.True(t, result)
}

func TestEqualLabelsMatch_PartialOverlap(t *testing.T) {
	equal := model.LabelNames{"service"}
	source := model.LabelSet{"service": "api", "env": "prod"}
	target := model.LabelSet{"service": "api", "env": "dev"}

	result := EqualLabelsMatch(equal, source, target)
	require.True(t, result)
}

func TestEqualLabelsMatch_LabelInSourceNotInTarget(t *testing.T) {
	equal := model.LabelNames{"service", "env"}
	source := model.LabelSet{"service": "api", "env": "prod"}
	target := model.LabelSet{"service": "api"} // missing "env"

	result := EqualLabelsMatch(equal, source, target)
	require.False(t, result)
}

func TestEqualLabelsMatch_LabelInTargetNotInSource(t *testing.T) {
	equal := model.LabelNames{"service", "env"}
	source := model.LabelSet{"service": "api"} // missing "env"
	target := model.LabelSet{"service": "api", "env": "prod"}

	result := EqualLabelsMatch(equal, source, target)
	require.False(t, result)
}

func TestEqualLabelsMatch_SpecialCharactersInValue(t *testing.T) {
	equal := model.LabelNames{"service", "env"}
	source := model.LabelSet{"service": "my.api-service", "env": "prod.v1.0"}
	target := model.LabelSet{"service": "my.api-service", "env": "prod.v1.0"}

	result := EqualLabelsMatch(equal, source, target)
	require.True(t, result)
}

func TestLabelNamesFromRoute_AllMatcherTypes(t *testing.T) {
	m := mustMatcher(t, labels.MatchEqual, "team", "platform")
	route := &config.Route{
		Match:    map[string]string{"env": "prod"},
		Matchers: []*labels.Matcher{m},
	}

	result := LabelNamesFromRoute(route)

	require.Contains(t, result, "env")
	require.Contains(t, result, "team")
	require.Len(t, result, 2)
}

func TestLabelNamesFromRoute_Empty(t *testing.T) {
	route := &config.Route{}

	result := LabelNamesFromRoute(route)

	require.Empty(t, result)
}

func TestLabelNamesFromRoute_SingleMatch(t *testing.T) {
	route := &config.Route{
		Match: map[string]string{"env": "prod"},
	}

	result := LabelNamesFromRoute(route)

	require.Contains(t, result, "env")
	require.Len(t, result, 1)
}

func TestLabelNamesFromRoute_Dedup(t *testing.T) {
	route := &config.Route{
		Match:    map[string]string{"env": "prod"},
		MatchRE:  map[string]config.Regexp{},
		Matchers: []*labels.Matcher{mustMatcher(t, labels.MatchEqual, "team", "platform")},
	}

	result := LabelNamesFromRoute(route)

	require.Contains(t, result, "env")
	require.Contains(t, result, "team")
	require.Len(t, result, 2)
}

func TestLabelNamesFromRoute_EmptyMatch(t *testing.T) {
	route := &config.Route{
		Match: map[string]string{"key": ""},
	}

	result := LabelNamesFromRoute(route)

	require.Contains(t, result, "key")
	require.Len(t, result, 1)
}

func TestUnionLabelNames_BothEmpty(t *testing.T) {
	a := map[string]struct{}{}
	b := map[string]struct{}{}

	result := UnionLabelNames(a, b)

	require.Empty(t, result)
}

func TestUnionLabelNames_Disjoint(t *testing.T) {
	a := map[string]struct{}{"env": {}}
	b := map[string]struct{}{"team": {}}

	result := UnionLabelNames(a, b)

	require.Contains(t, result, "env")
	require.Contains(t, result, "team")
	require.Len(t, result, 2)
}

func TestUnionLabelNames_NoOverlap(t *testing.T) {
	a := map[string]struct{}{"env": {}, "service": {}}
	b := map[string]struct{}{"team": {}, "region": {}}

	result := UnionLabelNames(a, b)

	require.Contains(t, result, "env")
	require.Contains(t, result, "service")
	require.Contains(t, result, "team")
	require.Contains(t, result, "region")
	require.Len(t, result, 4)
}

func TestUnionLabelNames_Overlapping(t *testing.T) {
	a := map[string]struct{}{"env": {}, "service": {}}
	b := map[string]struct{}{"service": {}, "team": {}}

	result := UnionLabelNames(a, b)

	require.Contains(t, result, "env")
	require.Contains(t, result, "service")
	require.Contains(t, result, "team")
	require.Len(t, result, 3)
}

func TestUnionLabelNames_FirstEmpty(t *testing.T) {
	a := map[string]struct{}{}
	b := map[string]struct{}{"env": {}}

	result := UnionLabelNames(a, b)

	require.Contains(t, result, "env")
	require.Len(t, result, 1)
}

func TestLabelNamesFromStringMap_Multiple(t *testing.T) {
	m := map[string]string{
		"env":     "prod",
		"service": "api",
		"region":  "us-east",
	}

	result := LabelNamesFromStringMap(m)

	require.Contains(t, result, "env")
	require.Contains(t, result, "service")
	require.Contains(t, result, "region")
	require.Len(t, result, 3)
}

func TestLabelNamesFromStringMap_Single(t *testing.T) {
	m := map[string]string{
		"env": "prod",
	}

	result := LabelNamesFromStringMap(m)

	require.Contains(t, result, "env")
	require.Len(t, result, 1)
}

func TestLabelNamesFromStringMap_Empty(t *testing.T) {
	m := map[string]string{}

	result := LabelNamesFromStringMap(m)

	require.Empty(t, result)
}

func TestLabelNamesFromStringMap_SpecialCharacters(t *testing.T) {
	m := map[string]string{
		"my.service": "value",
		"env-1":      "prod",
		"region_v2":  "us",
	}

	result := LabelNamesFromStringMap(m)

	require.Contains(t, result, "my.service")
	require.Contains(t, result, "env-1")
	require.Contains(t, result, "region_v2")
	require.Len(t, result, 3)
}
