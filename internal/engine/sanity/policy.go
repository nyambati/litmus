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
func (pc *PolicyChecker) Check(fragments []*litconfig.Fragment) []string {
	if !pc.policy.RequireTests && len(pc.policy.Enforce.Matchers) == 0 {
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

		if len(pc.policy.Enforce.Matchers) > 0 {
			var groupInherited map[string]struct{}
			if frag.Group != nil {
				groupInherited = labelNamesFromStringMap(frag.Group.Match)
			}
			issues = append(issues, pc.checkRoutes(frag.Name, frag.Routes, groupInherited)...)
		}
	}
	return issues
}

// checkRoutes walks the route tree with accumulated inherited label names.
// A route is covered when its union (inherited + own) satisfies the enforce policy.
// Covered routes skip their descendants entirely.
// Uncovered routes with children are cleared if all children resolve the coverage gap.
// Uncovered leaf routes are always violations.
func (pc *PolicyChecker) checkRoutes(fragName string, routes []*amconfig.Route, inherited map[string]struct{}) []string {
	var issues []string
	for _, route := range routes {
		union := unionLabelNames(inherited, labelNamesFromRoute(route))

		if pc.isCovered(union) {
			// This route satisfies the policy — skip descendants (they inherit coverage).
			continue
		}

		if len(route.Routes) == 0 {
			// Leaf route with incomplete coverage — definite violation.
			issues = append(issues, pc.formatViolation(fragName, route.Receiver, pc.missingMatchers(union)))
			continue
		}

		// Has children — check them with the accumulated union.
		childIssues := pc.checkRoutes(fragName, route.Routes, union)
		if len(childIssues) == 0 {
			// All children resolved the coverage gap — parent violation cleared.
			continue
		}
		// At least one child branch remains uncovered — propagate child issues.
		issues = append(issues, childIssues...)
	}
	return issues
}

// missingMatchers returns which required matchers are absent from the accumulated label set.
// In strict (AND) mode: returns each required label not present in the union.
// In non-strict (OR) mode: violation means none matched, so all required labels are returned.
func (pc *PolicyChecker) missingMatchers(labelNames map[string]struct{}) []string {
	if pc.policy.Enforce.Strict {
		var missing []string
		for _, required := range pc.policy.Enforce.Matchers {
			if _, ok := labelNames[required]; !ok {
				missing = append(missing, required)
			}
		}
		return missing
	}
	// OR mode: violated only when none present — return all as missing.
	return pc.policy.Enforce.Matchers
}

// isCovered reports whether the accumulated label names satisfy the enforce policy.
// strict=true (AND): every required label must be present.
// strict=false (OR): at least one required label must be present.
func (pc *PolicyChecker) isCovered(labelNames map[string]struct{}) bool {
	if pc.policy.Enforce.Strict {
		for _, required := range pc.policy.Enforce.Matchers {
			if _, ok := labelNames[required]; !ok {
				return false
			}
		}
		return true
	}
	for _, required := range pc.policy.Enforce.Matchers {
		if _, ok := labelNames[required]; ok {
			return true
		}
	}
	return false
}

// labelNamesFromRoute returns the set of label names present on a single route's own matchers.
func labelNamesFromRoute(route *amconfig.Route) map[string]struct{} {
	names := make(map[string]struct{}, len(route.Match)+len(route.MatchRE)+len(route.Matchers))
	for k := range route.Match {
		names[k] = struct{}{}
	}
	for k := range route.MatchRE {
		names[k] = struct{}{}
	}
	for _, m := range route.Matchers {
		names[m.Name] = struct{}{}
	}
	return names
}

// unionLabelNames merges two label name sets into a new set.
func unionLabelNames(a, b map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		result[k] = struct{}{}
	}
	for k := range b {
		result[k] = struct{}{}
	}
	return result
}

func (pc *PolicyChecker) formatViolation(fragName, receiver string, missing []string) string {
	mode := "strict"
	if !pc.policy.Enforce.Strict {
		mode = "non-strict"
	}
	return fmt.Sprintf(
		"fragment %q: route to receiver %q is missing required matchers %v (policy: enforce_matchers, mode: %s)",
		fragName, receiver, missing, mode,
	)
}

// labelNamesFromStringMap returns the set of label names from an exact-match map.
func labelNamesFromStringMap(m map[string]string) map[string]struct{} {
	names := make(map[string]struct{}, len(m))
	for k := range m {
		names[k] = struct{}{}
	}
	return names
}
