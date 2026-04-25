package sanity

import (
	"fmt"

	litconfig "github.com/nyambati/litmus/internal/config"
	amconfig "github.com/prometheus/alertmanager/config"
)

// PolicyChecker enforces workspace-level policy rules on fragment packages.
type PolicyChecker struct {
	policy litconfig.PolicyConfig
}

// NewPolicyChecker creates a checker for the given policy.
func NewPolicyChecker(policy litconfig.PolicyConfig) *PolicyChecker {
	return &PolicyChecker{policy: policy}
}

// Check returns a list of policy violations across the given fragments.
//
// For root, Routes contains only the children of the catch-all root route
// (base.Route.Routes), so the catch-all itself — which has no matchers by
// design — is never evaluated. Use skip_root: true to exempt root entirely.
func (pc *PolicyChecker) Check(fragments []*litconfig.Fragment) []string {
	if !pc.policy.RequireTests && len(pc.policy.EnforceMatchers) == 0 {
		return nil
	}

	var issues []string
	for _, frag := range fragments {
		if frag.Name == "root" && pc.policy.SkipRoot {
			continue
		}

		if pc.policy.RequireTests && len(frag.Tests) == 0 {
			issues = append(issues, fmt.Sprintf(
				"fragment %q has no tests (policy: require_tests=true)", frag.Name,
			))
		}

		if len(pc.policy.EnforceMatchers) > 0 {
			for _, route := range frag.Routes {
				if !pc.routeHasRequiredMatcher(route) {
					issues = append(issues, fmt.Sprintf(
						"fragment %q: route to receiver %q is missing a required matcher %v (policy: enforce_matchers)",
						frag.Name, route.Receiver, pc.policy.EnforceMatchers,
					))
				}
			}
		}
	}
	return issues
}

// routeHasRequiredMatcher returns true if the route contains at least one matcher
// whose label name appears in the policy's enforce_matchers list.
func (pc *PolicyChecker) routeHasRequiredMatcher(route *amconfig.Route) bool {
	for _, required := range pc.policy.EnforceMatchers {
		if _, ok := route.Match[required]; ok {
			return true
		}
		if _, ok := route.MatchRE[required]; ok {
			return true
		}
		for _, m := range route.Matchers {
			if m.Name == required {
				return true
			}
		}
	}
	return false
}
