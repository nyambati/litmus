package sanity

import (
	"fmt"

	litconfig "github.com/nyambati/litmus/internal/config"
	amconfig "github.com/prometheus/alertmanager/config"
)

// RequireTestsChecker enforces that each fragment has at least one test per route.
type RequireTestsChecker struct{}

// NewRequireTestsChecker creates a new RequireTestsChecker.
func NewRequireTestsChecker() *RequireTestsChecker {
	return &RequireTestsChecker{}
}

// Name implements Check.
func (rtc *RequireTestsChecker) Name() litconfig.SanityCheck {
	return litconfig.CheckRequireTests
}

// Run implements Check.
func (rtc *RequireTestsChecker) Run(ctx CheckContext) []string {
	if !ctx.Policy.Require.Tests {
		return nil
	}

	skipTests := containsPolicyType(ctx.Policy.SkipRoot, litconfig.PolicyTypeTests)

	var issues []string
	for _, frag := range ctx.Fragments {
		if frag.Namespace == "root" && skipTests {
			continue
		}

		routeCount := countRoutes(frag.Routes)
		if routeCount == 0 {
			continue
		}

		testCount := len(frag.Tests)
		if testCount < routeCount {
			issues = append(issues, fmt.Sprintf(
				"fragment %q has %d test(s) but %d route(s); need at least 1 test per route (policy: require.tests=true)",
				frag.Namespace, testCount, routeCount,
			))
		}
	}

	return issues
}

// countRoutes recursively counts all routes in the tree.
func countRoutes(routes []*amconfig.Route) int {
	total := 0
	for _, r := range routes {
		total++
		total += countRoutes(r.Routes)
	}
	return total
}
