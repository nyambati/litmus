package sanity

import (
	"fmt"

	litconfig "github.com/nyambati/litmus/internal/config"
)

// RequireTestsChecker enforces that each fragment has at least one test.
type RequireTestsChecker struct{}

// NewRequireTestsChecker creates a new RequireTestsChecker.
func NewRequireTestsChecker() *RequireTestsChecker {
	return &RequireTestsChecker{}
}

// Name implements Check.
func (rtc *RequireTestsChecker) Name() litconfig.SanityCheck {
	return litconfig.CheckPolicyViolations
}

// Run implements Check.
func (rtc *RequireTestsChecker) Run(ctx CheckContext) []string {
	if !ctx.Policy.Require.Tests {
		return nil
	}

	var issues []string
	skipTests := containsPolicyType(ctx.Policy.SkipRoot, litconfig.PolicyTypeTests)

	for _, frag := range ctx.Fragments {
		if frag.Namespace == "root" && skipTests {
			continue
		}

		if len(frag.Tests) == 0 {
			issues = append(issues, fmt.Sprintf(
				"fragment %q has no tests (policy: require.tests=true)", frag.Namespace,
			))
		}
	}

	return issues
}
