package sanity

import (
	"fmt"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"litmus/internal/engine/snapshot"
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
	walker := snapshot.NewRouteWalker(srd.root)
	paths := walker.FindTerminalPaths()

	var issues []string

	for i, path := range paths {
		// Check if this path is shadowed by any earlier path
		for j := 0; j < i; j++ {
			if srd.isShadowed(path, paths[j]) {
				issues = append(issues, fmt.Sprintf(
					"Route to %q is shadowed by earlier route to %q: matchers %v are subset of %v",
					path.Receiver, paths[j].Receiver, path.Matchers, paths[j].Matchers,
				))
				break
			}
		}
	}

	return issues
}

// isShadowed checks if child path is shadowed by parent path.
// Child is shadowed if parent's matchers are a superset of child's matchers.
func (srd *ShadowedRouteDetector) isShadowed(child, parent *snapshot.RoutePath) bool {
	return srd.isSubset(parent.Matchers, child.Matchers)
}

// isSubset checks if parentMatchers is a superset of childMatchers.
// Returns true if every key-value pair in childMatchers appears in parentMatchers.
func (srd *ShadowedRouteDetector) isSubset(parentMatchers, childMatchers []model.LabelSet) bool {
	// Merge all matchers for each
	parentLabels := srd.mergeMatchers(parentMatchers)
	childLabels := srd.mergeMatchers(childMatchers)

	// Check if all child labels are in parent labels with same values
	for k, v := range childLabels {
		if pv, ok := parentLabels[k]; !ok || pv != v {
			return false
		}
	}
	return true
}

// mergeMatchers combines multiple LabelSets into one.
func (srd *ShadowedRouteDetector) mergeMatchers(matchers []model.LabelSet) model.LabelSet {
	merged := make(model.LabelSet)
	for _, matcher := range matchers {
		for k, v := range matcher {
			merged[k] = v
		}
	}
	return merged
}
