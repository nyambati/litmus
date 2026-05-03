package sanity

import (
	"fmt"

	litconfig "github.com/nyambati/litmus/internal/config"
	amconfig "github.com/prometheus/alertmanager/config"
)

// RegressionChecker enforces that regression baseline exists and has sufficient coverage.
type RegressionChecker struct{}

// NewRegressionChecker creates a new RegressionChecker.
func NewRegressionChecker() *RegressionChecker {
	return &RegressionChecker{}
}

// Name implements Check.
func (rc *RegressionChecker) Name() litconfig.SanityCheck {
	return litconfig.CheckPolicyViolations
}

// Run implements Check.
func (rc *RegressionChecker) Run(ctx CheckContext) []string {
	if !ctx.Policy.Require.Regression {
		return nil
	}

	if ctx.RegressionState == nil {
		return []string{"no regression baseline found, run 'litmus snapshot'"}
	}

	if len(ctx.RegressionState.Tests) == 0 {
		return []string{"no regression tests found (policy: require.regression=true)"}
	}

	routeCount := countTerminalRoutes(ctx.Route)
	testCount := len(ctx.RegressionState.Tests)

	if testCount < routeCount {
		return []string{fmt.Sprintf("regression tests (%d) fewer than routes (%d)", testCount, routeCount)}
	}

	return nil
}

// countTerminalRoutes counts all terminal routes (routes with no children).
func countTerminalRoutes(route *amconfig.Route) int {
	if route == nil {
		return 0
	}

	if len(route.Routes) == 0 {
		return 1
	}

	count := 0
	for _, child := range route.Routes {
		count += countTerminalRoutes(child)
	}
	return count
}
