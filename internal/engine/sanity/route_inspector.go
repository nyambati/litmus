package sanity

import (
	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
)

// sanityMatcher captures a single label constraint with its match polarity.
// isNeg is true for != and !~ matchers. isRegex is true for =~ and !~ matchers.
// source is the receiver name of the route that introduced this matcher.
type sanityMatcher struct {
	name    string
	value   string
	isNeg   bool
	isRegex bool
	source  string
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
	if ri.root == nil {
		return nil
	}
	var paths []*sanityPath
	ri.walk(ri.root, nil, &paths)
	return paths
}

func (ri *routeInspector) walk(route *config.Route, inherited []sanityMatcher, paths *[]*sanityPath) {
	if route == nil {
		return
	}
	current := append([]sanityMatcher{}, inherited...)

	src := route.Receiver
	for k, v := range route.Match {
		current = append(current, sanityMatcher{name: k, value: v, source: src})
	}

	for k, v := range route.MatchRE {
		current = append(current, sanityMatcher{name: k, value: v.String(), isRegex: true, source: src})
	}

	for _, m := range route.Matchers {
		isNeg := m.Type == labels.MatchNotEqual || m.Type == labels.MatchNotRegexp
		isRegex := m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp
		current = append(current, sanityMatcher{name: m.Name, value: m.Value, isNeg: isNeg, isRegex: isRegex, source: src})
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
