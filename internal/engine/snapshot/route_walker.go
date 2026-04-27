package snapshot

import (
	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/prometheus/common/model"
)

// RoutePath represents a path from root to terminal route with cumulative matchers.
type RoutePath struct {
	Receiver        string           // Terminal receiver name
	Matchers        []model.LabelSet // Cumulative positive matchers from root to leaf
	IgnoredMatchers []string         // Negative matchers ignored during synthesis
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
	if rw.root == nil {
		return nil
	}
	var paths []*RoutePath
	rw.walk(rw.root, []model.LabelSet{}, nil, &paths)
	return paths
}

// walk recursively traverses route tree, accumulating matchers.
func (rw *RouteWalker) walk(route *config.Route, matchers []model.LabelSet, ignored []string, paths *[]*RoutePath) {
	if route == nil {
		return
	}
	currentMatchers := append([]model.LabelSet{}, matchers...)
	currentIgnored := append([]string{}, ignored...)

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
			if m == nil {
				continue
			}
			if m.Type == labels.MatchNotEqual || m.Type == labels.MatchNotRegexp {
				currentIgnored = append(currentIgnored, m.String())
				continue
			}
			labelSet[model.LabelName(m.Name)] = model.LabelValue(m.Value)
		}
		if len(labelSet) > 0 {
			currentMatchers = append(currentMatchers, labelSet)
		}
	}

	// A matched parent route remains a valid routing outcome when none of its
	// children match, so snapshots must include that fallback path as well.
	if route.Receiver != "" && len(route.Routes) > 0 {
		*paths = append(*paths, &RoutePath{
			Receiver:        route.Receiver,
			Matchers:        currentMatchers,
			IgnoredMatchers: currentIgnored,
		})
	}

	// Terminal route: has no children
	if len(route.Routes) == 0 {
		*paths = append(*paths, &RoutePath{
			Receiver:        route.Receiver,
			Matchers:        currentMatchers,
			IgnoredMatchers: currentIgnored,
		})
		return
	}

	// Recursive: walk children
	for _, child := range route.Routes {
		rw.walk(child, currentMatchers, currentIgnored, paths)
	}
}
