package sanity

import (
	"fmt"
	"strings"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
)

// NegativeOnlyRouteDetector flags routes whose own matchers are exclusively negative.
type NegativeOnlyRouteDetector struct {
	root *config.Route
}

// NewNegativeOnlyRouteDetector creates a detector for negative-only route matchers.
func NewNegativeOnlyRouteDetector(root *config.Route) *NegativeOnlyRouteDetector {
	return &NegativeOnlyRouteDetector{root: root}
}

// Name implements Check.
func (d *NegativeOnlyRouteDetector) Name() string { return "negative_only_routes" }

// Run implements Check.
func (d *NegativeOnlyRouteDetector) Run(ctx CheckContext) []string {
	return NewNegativeOnlyRouteDetector(ctx.Route).Detect()
}

// Detect returns one issue per route that has only negative matchers.
func (d *NegativeOnlyRouteDetector) Detect() []string {
	if d.root == nil {
		return nil
	}
	return d.walk(d.root, nil)
}

func (d *NegativeOnlyRouteDetector) walk(route *config.Route, crumb []string) []string {
	if route == nil {
		return nil
	}
	currentCrumb := append(append([]string{}, crumb...), route.Receiver)

	var issues []string
	if isNegativeOnlyRoute(route) {
		issues = append(issues, formatNegativeOnlyIssue(route, currentCrumb))
	}
	for _, child := range route.Routes {
		issues = append(issues, d.walk(child, currentCrumb)...)
	}
	return issues
}

func isNegativeOnlyRoute(route *config.Route) bool {
	if route == nil || len(route.Match) > 0 || len(route.MatchRE) > 0 {
		return false
	}
	seenMatcher := false
	for _, m := range route.Matchers {
		if m == nil {
			continue
		}
		seenMatcher = true
		if !isNegativeMatcher(m) {
			return false
		}
	}
	return seenMatcher
}

func isNegativeMatcher(m *labels.Matcher) bool {
	return m.Type == labels.MatchNotEqual || m.Type == labels.MatchNotRegexp
}

func formatNegativeOnlyIssue(route *config.Route, crumb []string) string {
	return fmt.Sprintf(
		"Route to %q has only negative matchers (%s); add at least one positive matcher to make the route intent explicit %s",
		route.Receiver,
		strings.Join(negativeMatcherStrings(route), ", "),
		formatCrumb(crumb),
	)
}

func negativeMatcherStrings(route *config.Route) []string {
	matchers := make([]string, 0, len(route.Matchers))
	for _, m := range route.Matchers {
		if m != nil && isNegativeMatcher(m) {
			matchers = append(matchers, m.String())
		}
	}
	return matchers
}
