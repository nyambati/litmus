package sanity

import (
	"fmt"

	"github.com/prometheus/alertmanager/config"
)

// ShadowedRouteDetector detects unreachable routes.
type ShadowedRouteDetector struct {
	root *config.Route
}

// NewShadowedRouteDetector creates detector for given route tree.
func NewShadowedRouteDetector(root *config.Route) *ShadowedRouteDetector {
	return &ShadowedRouteDetector{root: root}
}

// Detect returns list of shadowed route issues.
func (srd *ShadowedRouteDetector) Detect() []string {
	paths := newRouteInspector(srd.root).findPaths()

	var issues []string
	for i, path := range paths {
		for j := 0; j < i; j++ {
			if srd.isShadowed(path, paths[j]) {
				issues = append(issues, fmt.Sprintf(
					"Route to %q is shadowed by earlier route to %q",
					path.receiver, paths[j].receiver,
				))
				break
			}
		}
	}
	return issues
}

// isShadowed returns true when parent completely shadows child — i.e., every alert
// matching child would also match parent, making child unreachable. Requires:
//   - parent has continue:false (otherwise Alertmanager evaluates child regardless)
//   - parent's positive matchers ⊆ child's positive matchers (parent broader)
//   - parent has no negative matcher that conflicts with a child positive matcher
//     on the same label+value (mutual exclusion → routes are disjoint, not shadowed)
func (srd *ShadowedRouteDetector) isShadowed(child, parent *sanityPath) bool {
	if parent.continueOnMatch {
		return false
	}
	childPos := make(map[string]string)
	for _, m := range child.matchers {
		if !m.isNeg {
			childPos[m.name] = m.value
		}
	}

	// Parent's positive matchers must all appear in child's positive matchers
	// (parent broader: fewer/equal constraints → larger match set).
	for _, m := range parent.matchers {
		if m.isNeg {
			continue
		}
		if cv, ok := childPos[m.name]; !ok || cv != m.value {
			return false
		}
	}

	// Parent's negative matcher on a label where child has a matching positive
	// matcher means the two routes are mutually exclusive on that label —
	// parent excludes the very alerts child requires, so parent cannot shadow child.
	for _, m := range parent.matchers {
		if !m.isNeg {
			continue
		}
		if cv, ok := childPos[m.name]; ok && cv == m.value {
			return false
		}
	}

	return true
}
