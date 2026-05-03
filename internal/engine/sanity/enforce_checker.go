package sanity

import (
	"fmt"

	litconfig "github.com/nyambati/litmus/internal/config"
	labelmatcher "github.com/nyambati/litmus/internal/labelmatcher"
	amconfig "github.com/prometheus/alertmanager/config"
)

// EnforceChecker enforces that routes have required label matchers.
type EnforceChecker struct{}

// NewEnforceChecker creates a new EnforceChecker.
func NewEnforceChecker() *EnforceChecker {
	return &EnforceChecker{}
}

// Name implements Check.
func (ec *EnforceChecker) Name() litconfig.SanityCheck {
	return litconfig.CheckPolicyViolations
}

// Run implements Check.
func (ec *EnforceChecker) Run(ctx CheckContext) []string {
	if len(ctx.Policy.Enforce.Matchers) == 0 {
		return nil
	}

	var issues []string
	skipEnforce := containsPolicyType(ctx.Policy.SkipRoot, "enforce")

	for _, frag := range ctx.Fragments {
		if frag.Namespace == "root" && skipEnforce {
			continue
		}

		var groupInherited map[string]struct{}
		if frag.Group != nil {
			groupInherited = labelmatcher.LabelNamesFromStringMap(frag.Group.Match)
		}
		issues = append(issues, ec.checkRoutes(ctx.Policy, frag.Namespace, frag.Routes, groupInherited)...)
	}

	return issues
}

// containsPolicyType checks if a slice of PolicyType contains a specific item.
func containsPolicyType(slice []litconfig.PolicyType, item string) bool {
	for _, v := range slice {
		if string(v) == item {
			return true
		}
	}
	return false
}

// checkRoutes walks the route tree with accumulated inherited label names.
func (ec *EnforceChecker) checkRoutes(policy litconfig.PolicyConfig, fragName string, routes []*amconfig.Route, inherited map[string]struct{}) []string {
	issues := make([]string, 0, len(routes))
	for _, route := range routes {
		issues = append(issues, ec.checkRoute(policy, fragName, route, inherited)...)
	}
	return issues
}

// checkRoute checks a single route and its children for policy violations.
func (ec *EnforceChecker) checkRoute(policy litconfig.PolicyConfig, fragName string, route *amconfig.Route, inherited map[string]struct{}) []string {
	union := labelmatcher.UnionLabelNames(inherited, labelmatcher.LabelNamesFromRoute(route))

	if ec.isCovered(policy, union) {
		return nil
	}

	if ec.isLeafRoute(route) {
		return ec.reportLeafViolation(policy, fragName, route, union)
	}

	return ec.checkChildRoutes(policy, fragName, route, union)
}

// isLeafRoute checks if the route has no children.
func (ec *EnforceChecker) isLeafRoute(route *amconfig.Route) bool {
	return len(route.Routes) == 0
}

// reportLeafViolation creates a violation for a leaf route that doesn't satisfy the policy.
func (ec *EnforceChecker) reportLeafViolation(policy litconfig.PolicyConfig, fragName string, route *amconfig.Route, union map[string]struct{}) []string {
	return []string{ec.formatViolation(policy, fragName, route.Receiver, ec.missingMatchers(policy, union))}
}

// checkChildRoutes processes child routes with the accumulated union of label names.
func (ec *EnforceChecker) checkChildRoutes(policy litconfig.PolicyConfig, fragName string, route *amconfig.Route, union map[string]struct{}) []string {
	childIssues := ec.checkRoutes(policy, fragName, route.Routes, union)
	if len(childIssues) == 0 {
		return nil
	}
	return childIssues
}

// missingMatchers returns which required matchers are absent from the accumulated label set.
func (ec *EnforceChecker) missingMatchers(policy litconfig.PolicyConfig, labelNames map[string]struct{}) []string {
	if policy.Enforce.Strict {
		var missing []string
		for _, required := range policy.Enforce.Matchers {
			if _, ok := labelNames[required]; !ok {
				missing = append(missing, required)
			}
		}
		return missing
	}
	return policy.Enforce.Matchers
}

// isCovered reports whether the accumulated label names satisfy the enforce policy.
func (ec *EnforceChecker) isCovered(policy litconfig.PolicyConfig, labelNames map[string]struct{}) bool {
	if policy.Enforce.Strict {
		for _, required := range policy.Enforce.Matchers {
			if _, ok := labelNames[required]; !ok {
				return false
			}
		}
		return true
	}
	for _, required := range policy.Enforce.Matchers {
		if _, ok := labelNames[required]; ok {
			return true
		}
	}
	return false
}

func (ec *EnforceChecker) formatViolation(policy litconfig.PolicyConfig, fragName, receiver string, missing []string) string {
	mode := "strict"
	if !policy.Enforce.Strict {
		mode = "non-strict"
	}
	return fmt.Sprintf(
		"fragment %q: route to receiver %q is missing required matchers %v (policy: enforce_matchers, mode: %s)",
		fragName, receiver, missing, mode,
	)
}
