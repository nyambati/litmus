package sanity

import (
	"regexp"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
)

// sanityMatcher captures a single label constraint with its match polarity.
// isNeg is true for != and !~ matchers. isRegex is true for =~ and !~ matchers.
// compiled holds the anchored regexp for regex matchers (nil for exact matchers).
// source is the receiver name of the route that introduced this matcher.
type sanityMatcher struct {
	name     string
	value    string
	isNeg    bool
	isRegex  bool
	compiled *regexp.Regexp
	source   string
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
	current = append(current, collectMatchers(route)...)

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

// collectMatchers extracts a route's own matchers as sanityMatchers,
// tagging each with the route's receiver as the source.
func collectMatchers(route *config.Route) []sanityMatcher {
	src := route.Receiver
	ms := make([]sanityMatcher, 0, len(route.Match)+len(route.MatchRE)+len(route.Matchers))
	for k, v := range route.Match {
		ms = append(ms, sanityMatcher{name: k, value: v, source: src})
	}
	for k, v := range route.MatchRE {
		ms = append(ms, sanityMatcher{name: k, value: v.String(), isRegex: true, compiled: v.Regexp, source: src})
	}
	for _, m := range route.Matchers {
		isNeg := m.Type == labels.MatchNotEqual || m.Type == labels.MatchNotRegexp
		isRegex := m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp
		var compiled *regexp.Regexp
		if isRegex {
			compiled, _ = regexp.Compile("^(?:" + m.Value + ")$")
		}
		ms = append(ms, sanityMatcher{name: m.Name, value: m.Value, isNeg: isNeg, isRegex: isRegex, compiled: compiled, source: src})
	}
	return ms
}
