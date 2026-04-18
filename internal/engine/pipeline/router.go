package pipeline

import (
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

// Router walks an Alertmanager route tree, returning all receivers that
// would handle a given label set under the standard routing algorithm.
type Router struct {
	root *config.Route
}

// NewRouter creates a Router for the given route tree root.
func NewRouter(root *config.Route) *Router {
	return &Router{root: root}
}

// Match returns the ordered list of receiver names that would handle labels.
func (r *Router) Match(labels model.LabelSet) []string {
	routes := r.walk(r.root, labels)
	receivers := make([]string, 0, len(routes))
	for _, route := range routes {
		if route.Receiver != "" {
			receivers = append(receivers, route.Receiver)
		}
	}
	return receivers
}

// walk implements the Alertmanager routing algorithm recursively.
// Returns the matched leaf routes (or this route as fallback).
func (r *Router) walk(route *config.Route, labels model.LabelSet) []*config.Route {
	if !routeMatches(route, labels) {
		return nil
	}

	var matched []*config.Route
	for _, child := range route.Routes {
		childResults := r.walk(child, labels)
		if len(childResults) > 0 {
			matched = append(matched, childResults...)
			if !child.Continue {
				return matched
			}
		}
	}

	if len(matched) == 0 {
		return []*config.Route{route}
	}
	return matched
}

// routeMatches checks all three matcher formats against the label set.
// An empty matcher set matches everything (e.g. root route).
func routeMatches(route *config.Route, labels model.LabelSet) bool {
	for k, v := range route.Match {
		if string(labels[model.LabelName(k)]) != v {
			return false
		}
	}
	for k, re := range route.MatchRE {
		if !re.MatchString(string(labels[model.LabelName(k)])) {
			return false
		}
	}
	for _, m := range route.Matchers {
		if !m.Matches(string(labels[model.LabelName(m.Name)])) {
			return false
		}
	}
	return true
}
