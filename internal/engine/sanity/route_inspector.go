package sanity

import (
	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
)

// sanityMatcher captures a single label constraint with its match polarity.
// Negative is true for != and !~ matchers.
type sanityMatcher struct {
	name  string
	value string
	isNeg bool
}

// sanityPath holds a root-to-leaf path through the route tree with typed matchers.
type sanityPath struct {
	receiver        string
	matchers        []sanityMatcher
	continueOnMatch bool // mirrors the route's continue field
}

// routeInspector walks the route tree collecting typed matchers for sanity analysis.
type routeInspector struct {
	root *config.Route
}

func newRouteInspector(root *config.Route) *routeInspector {
	return &routeInspector{root: root}
}

// findPaths returns all leaf paths with their cumulative typed matchers.
func (ri *routeInspector) findPaths() []*sanityPath {
	var paths []*sanityPath
	ri.walk(ri.root, nil, &paths)
	return paths
}

func (ri *routeInspector) walk(route *config.Route, inherited []sanityMatcher, paths *[]*sanityPath) {
	current := append([]sanityMatcher{}, inherited...)

	for k, v := range route.Match {
		current = append(current, sanityMatcher{name: k, value: v, isNeg: false})
	}

	for k, v := range route.MatchRE {
		current = append(current, sanityMatcher{name: k, value: v.String(), isNeg: false})
	}

	for _, m := range route.Matchers {
		isNeg := m.Type == labels.MatchNotEqual || m.Type == labels.MatchNotRegexp
		current = append(current, sanityMatcher{name: m.Name, value: m.Value, isNeg: isNeg})
	}

	if len(route.Routes) == 0 {
		*paths = append(*paths, &sanityPath{
			receiver:        route.Receiver,
			matchers:        current,
			continueOnMatch: route.Continue,
		})
		return
	}

	for _, child := range route.Routes {
		ri.walk(child, current, paths)
	}
}
