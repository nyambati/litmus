package labelmatcher

import (
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

// SourceMatches checks if labels match the source matchers of an inhibition rule.
func SourceMatches(rule config.InhibitRule, labels model.LabelSet) bool {
	for k, v := range rule.SourceMatch {
		if string(labels[model.LabelName(k)]) != v {
			return false
		}
	}
	for _, m := range rule.SourceMatchers {
		if !m.Matches(string(labels[model.LabelName(m.Name)])) {
			return false
		}
	}
	return true
}

// TargetMatches checks if labels match the target matchers of an inhibition rule.
func TargetMatches(rule config.InhibitRule, labels model.LabelSet) bool {
	for k, v := range rule.TargetMatch {
		if string(labels[model.LabelName(k)]) != v {
			return false
		}
	}
	for _, m := range rule.TargetMatchers {
		if !m.Matches(string(labels[model.LabelName(m.Name)])) {
			return false
		}
	}
	return true
}

// EqualLabelsMatch checks if labels have equal values for the specified label names.
func EqualLabelsMatch(equal model.LabelNames, source, target model.LabelSet) bool {
	for _, name := range equal {
		if source[name] != target[name] {
			return false
		}
	}
	return true
}

// LabelNamesFromRoute returns the set of label names present on a single route's own matchers.
func LabelNamesFromRoute(route *config.Route) map[string]struct{} {
	names := make(map[string]struct{}, len(route.Match)+len(route.MatchRE)+len(route.Matchers))
	for k := range route.Match {
		names[k] = struct{}{}
	}
	for k := range route.MatchRE {
		names[k] = struct{}{}
	}
	for _, m := range route.Matchers {
		names[m.Name] = struct{}{}
	}
	return names
}

// UnionLabelNames merges two label name sets into a new set.
func UnionLabelNames(a, b map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		result[k] = struct{}{}
	}
	for k := range b {
		result[k] = struct{}{}
	}
	return result
}

// LabelNamesFromStringMap returns the set of label names from an exact-match map.
func LabelNamesFromStringMap(m map[string]string) map[string]struct{} {
	names := make(map[string]struct{}, len(m))
	for k := range m {
		names[k] = struct{}{}
	}
	return names
}
