package snapshot

import (
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

// RoutePath represents a path from root to terminal route with cumulative matchers.
type RoutePath struct {
	Receiver string              // Terminal receiver name
	Matchers []model.LabelSet    // Cumulative matchers from root to leaf
}

// RouteWalker traverses alertmanager route tree to find terminal paths.
type RouteWalker struct {
	root *config.Route
}

// NewRouteWalker creates walker for given route tree.
func NewRouteWalker(root *config.Route) *RouteWalker {
	return &RouteWalker{root: root}
}

// FindTerminalPaths returns all leaf routes and their matcher paths.
func (rw *RouteWalker) FindTerminalPaths() []*RoutePath {
	var paths []*RoutePath
	rw.walk(rw.root, []model.LabelSet{}, &paths)
	return paths
}

// walk recursively traverses route tree, accumulating matchers.
func (rw *RouteWalker) walk(route *config.Route, matchers []model.LabelSet, paths *[]*RoutePath) {
	currentMatchers := matchers

	if len(route.Match) > 0 {
		labelSet := make(model.LabelSet)
		for k, v := range route.Match {
			labelSet[model.LabelName(k)] = model.LabelValue(v)
		}
		currentMatchers = append(currentMatchers, labelSet)
	}

	// MatchRE: deprecated regex format — store pattern string as value.
	if len(route.MatchRE) > 0 {
		labelSet := make(model.LabelSet)
		for k, v := range route.MatchRE {
			labelSet[model.LabelName(k)] = model.LabelValue(v.String())
		}
		currentMatchers = append(currentMatchers, labelSet)
	}

	// Matchers: modern format — each matcher carries Name and Value.
	if len(route.Matchers) > 0 {
		labelSet := make(model.LabelSet)
		for _, m := range route.Matchers {
			labelSet[model.LabelName(m.Name)] = model.LabelValue(m.Value)
		}
		currentMatchers = append(currentMatchers, labelSet)
	}

	// Terminal route: has no children
	if len(route.Routes) == 0 {
		*paths = append(*paths, &RoutePath{
			Receiver: route.Receiver,
			Matchers: currentMatchers,
		})
		return
	}

	// Recursive: walk children
	for _, child := range route.Routes {
		rw.walk(child, currentMatchers, paths)
	}
}
