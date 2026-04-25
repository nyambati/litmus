package sanity

import (
	"fmt"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
)

// DeadReceiverDetector flags routes whose cumulative ancestor matchers make them
// permanently unreachable — a vertical complement to the sibling-focused ShadowedRouteDetector.
type DeadReceiverDetector struct {
	root *config.Route
}

// NewDeadReceiverDetector creates a detector for the given route tree.
func NewDeadReceiverDetector(root *config.Route) *DeadReceiverDetector {
	return &DeadReceiverDetector{root: root}
}

// contradiction describes a pair of irreconcilable matchers on the same label.
type contradiction struct {
	label       string
	reqVal      string // the positive value that was required
	reqSrc      string // receiver of the route that required it
	isNeg       bool   // true = conflict is a negation of reqVal; false = a different positive value
	conflictVal string // the conflicting positive value (only meaningful when !isNeg)
}

// Detect returns one issue string per dead route. When a contradiction is found,
// recursion into that subtree stops — descendants are implied dead and not re-reported.
func (d *DeadReceiverDetector) Detect() []string {
	if d.root == nil {
		return nil
	}
	inherited := collectMatchers(d.root)
	var issues []string
	for _, child := range d.root.Routes {
		issues = append(issues, d.walk(child, inherited)...)
	}
	return issues
}

func (d *DeadReceiverDetector) walk(route *config.Route, inherited []sanityMatcher) []string {
	if route == nil {
		return nil
	}
	own := collectMatchers(route)
	combined := append(append([]sanityMatcher{}, inherited...), own...)

	if c, ok := findContradiction(combined); ok {
		return []string{formatDeadIssue(route.Receiver, c)}
	}

	var issues []string
	for _, child := range route.Routes {
		issues = append(issues, d.walk(child, combined)...)
	}
	return issues
}

// formatDeadIssue renders the human-readable issue string, distinguishing three cases:
//   - pos vs pos, same route:     both conflicting matchers on this route itself
//   - pos vs pos, cross-route:    ancestor required one value, this route requires another
//   - pos vs neg (± same route):  a required value is excluded
func formatDeadIssue(receiver string, c contradiction) string {
	if c.isNeg {
		if c.reqSrc == receiver {
			return fmt.Sprintf(
				"Route to %q can never be reached: label %q is both required and excluded by this route",
				receiver, c.label,
			)
		}
		return fmt.Sprintf(
			"Route to %q can never be reached: label %q requires %q (set by %q) but this route excludes it",
			receiver, c.label, c.reqVal, c.reqSrc,
		)
	}
	if c.reqSrc == receiver {
		return fmt.Sprintf(
			"Route to %q can never be reached: label %q has conflicting matchers %q and %q on this route",
			receiver, c.label, c.reqVal, c.conflictVal,
		)
	}
	return fmt.Sprintf(
		"Route to %q can never be reached: label %q set to %q by %q, conflicts with %q (this route)",
		receiver, c.label, c.reqVal, c.reqSrc, c.conflictVal,
	)
}

// collectMatchers extracts this route's own matchers (not inherited) as sanityMatchers,
// tagging each with the route's receiver as the source.
func collectMatchers(route *config.Route) []sanityMatcher {
	src := route.Receiver
	capacity := len(route.Match) + len(route.MatchRE) + len(route.Matchers)
	var ms = make([]sanityMatcher, 0, capacity)
	for k, v := range route.Match {
		ms = append(ms, sanityMatcher{name: k, value: v, source: src})
	}
	for k, v := range route.MatchRE {
		ms = append(ms, sanityMatcher{name: k, value: v.String(), isRegex: true, source: src})
	}
	for _, m := range route.Matchers {
		isNeg := m.Type == labels.MatchNotEqual || m.Type == labels.MatchNotRegexp
		isRegex := m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp
		ms = append(ms, sanityMatcher{name: m.Name, value: m.Value, isNeg: isNeg, isRegex: isRegex, source: src})
	}
	return ms
}

// findContradiction scans matchers (exact only — regex skipped) for two kinds of conflicts:
//  1. Two positive exact matchers on the same label with different values.
//  2. A positive exact matcher cancelled by a negative exact matcher on the same label+value.
func findContradiction(matchers []sanityMatcher) (contradiction, bool) {
	posExact := make(map[string]sanityMatcher) // label → first positive exact matcher seen

	for _, m := range matchers {
		if m.isRegex || m.isNeg {
			continue
		}
		if existing, ok := posExact[m.name]; ok {
			if existing.value != m.value {
				return contradiction{
					label:       m.name,
					reqVal:      existing.value,
					reqSrc:      existing.source,
					conflictVal: m.value,
				}, true
			}
		} else {
			posExact[m.name] = m
		}
	}

	for _, m := range matchers {
		if m.isRegex || !m.isNeg {
			continue
		}
		if pos, ok := posExact[m.name]; ok && pos.value == m.value {
			return contradiction{
				label:  m.name,
				reqVal: pos.value,
				reqSrc: pos.source,
				isNeg:  true,
			}, true
		}
	}

	return contradiction{}, false
}
