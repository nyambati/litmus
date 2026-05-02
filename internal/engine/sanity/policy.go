package sanity

import (
	"fmt"

	litconfig "github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/fragment"
	labelmatcher "github.com/nyambati/litmus/internal/labelmatcher"
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

// Name implements Check.
func (pc *PolicyChecker) Name() string { return "policy_violations" }

// Run implements Check.
func (pc *PolicyChecker) Run(ctx CheckContext) []string {
	return NewPolicyChecker(ctx.Policy).Check(ctx.Fragments)
}

// Check returns a list of policy violations across the given fragments.
func (pc *PolicyChecker) Check(fragments []*fragment.Fragment) []string {
	if !pc.policy.RequireTests && len(pc.policy.Enforce.Matchers) == 0 {
		return nil
	}

	var issues []string
	for _, frag := range fragments {
		if frag.Namespace == "root" && pc.policy.SkipRoot {
			continue
		}

		if pc.policy.RequireTests && len(frag.Tests) == 0 {
			issues = append(issues, fmt.Sprintf(
				"fragment %q has no tests (policy: require_tests=true)", frag.Namespace,
			))
		}

		if len(pc.policy.Enforce.Matchers) > 0 {
			var groupInherited map[string]struct{}
			if frag.Group != nil {
				groupInherited = labelmatcher.LabelNamesFromStringMap(frag.Group.Match)
			}
			issues = append(issues, pc.checkRoutes(frag.Namespace, frag.Routes, groupInherited)...)
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
	issues := make([]string, 0, len(routes))
	for _, route := range routes {
		issues = append(issues, pc.checkRoute(fragName, route, inherited)...)
	}
	return issues
}

// checkRoute checks a single route and its children for policy violations.
func (pc *PolicyChecker) checkRoute(fragName string, route *amconfig.Route, inherited map[string]struct{}) []string {
	union := labelmatcher.UnionLabelNames(inherited, labelmatcher.LabelNamesFromRoute(route))

	if pc.isCoveredByPolicy(union) {
		return nil
	}

	if pc.isLeafRoute(route) {
		return pc.reportLeafViolation(fragName, route, union)
	}

	return pc.checkChildRoutes(fragName, route, union)
}

// isCoveredByPolicy checks if the accumulated label names satisfy the enforce policy.
func (pc *PolicyChecker) isCoveredByPolicy(labelNames map[string]struct{}) bool {
	return pc.isCovered(labelNames)
}

// isLeafRoute checks if the route has no children.
func (pc *PolicyChecker) isLeafRoute(route *amconfig.Route) bool {
	return len(route.Routes) == 0
}

// reportLeafViolation creates a violation for a leaf route that doesn't satisfy the policy.
func (pc *PolicyChecker) reportLeafViolation(fragName string, route *amconfig.Route, union map[string]struct{}) []string {
	return []string{pc.formatViolation(fragName, route.Receiver, pc.missingMatchers(union))}
}

// checkChildRoutes processes child routes with the accumulated union of label names.
func (pc *PolicyChecker) checkChildRoutes(fragName string, route *amconfig.Route, union map[string]struct{}) []string {
	childIssues := pc.checkRoutes(fragName, route.Routes, union)
	if len(childIssues) == 0 {
		return nil
	}
	return childIssues
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
