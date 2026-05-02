package sanity

import (
	"fmt"
	"strings"

	"github.com/prometheus/alertmanager/config"
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
	reqVal      string // the positive exact value that was required
	reqSrc      string // receiver of the route that required it
	isNeg       bool   // true = conflict is a negation of reqVal (exact or regex)
	negSrc      string // receiver that set the neg-exact matcher (exact neg only)
	conflictVal string // conflicting positive value (pos-pos exact only)
	// regex fields — populated when one side of the contradiction is a regex matcher
	isRegex  bool   // true = one side is a regex matcher
	regexPat string // the regex pattern (human-readable, without anchors)
	regexSrc string // receiver that set the regex matcher
}

// Name implements Check.
func (d *DeadReceiverDetector) Name() string { return "dead_receivers" }

// Run implements Check.
func (d *DeadReceiverDetector) Run(ctx CheckContext) []string {
	return NewDeadReceiverDetector(ctx.Route).Detect()
}

// Detect returns one issue string per dead route. When a contradiction is found,
// recursion into that subtree stops — descendants are implied dead and not re-reported.
func (d *DeadReceiverDetector) Detect() []string {
	if d.root == nil {
		return nil
	}
	inherited := collectMatchers(d.root)
	crumb := []string{d.root.Receiver}
	var issues []string
	for _, child := range d.root.Routes {
		issues = append(issues, d.walk(child, inherited, crumb)...)
	}
	return issues
}

func (d *DeadReceiverDetector) walk(route *config.Route, inherited []sanityMatcher, crumb []string) []string {
	if route == nil {
		return nil
	}
	own := collectMatchers(route)
	combined := append(append([]sanityMatcher{}, inherited...), own...)
	currentCrumb := append(append([]string{}, crumb...), route.Receiver)

	if c, ok := findContradiction(combined); ok {
		return []string{formatDeadIssue(route.Receiver, c, currentCrumb)}
	}

	var issues []string
	for _, child := range route.Routes {
		issues = append(issues, d.walk(child, combined, currentCrumb)...)
	}
	return issues
}

// formatDeadIssue renders the human-readable issue string for a dead route.
func formatDeadIssue(receiver string, c contradiction, crumb []string) string {
	var base string
	if c.isRegex {
		base = formatRegexDeadIssue(receiver, c)
	} else {
		base = formatExactDeadIssue(receiver, c)
	}
	return base + " " + formatCrumb(crumb)
}

func formatExactDeadIssue(receiver string, c contradiction) string {
	if c.isNeg {
		switch {
		case c.reqSrc == receiver && c.negSrc == receiver:
			// both pos and neg matchers are on this route
			return fmt.Sprintf(
				"Route to %q can never be reached: label %q is both required and excluded by this route",
				receiver, c.label,
			)
		case c.reqSrc == receiver:
			// pos on this route, neg from ancestor
			return fmt.Sprintf(
				"Route to %q can never be reached: label %q requires %q but %q excludes it",
				receiver, c.label, c.reqVal, c.negSrc,
			)
		default:
			// pos from ancestor, neg on this route
			return fmt.Sprintf(
				"Route to %q can never be reached: label %q requires %q (set by %q) but this route excludes it",
				receiver, c.label, c.reqVal, c.reqSrc,
			)
		}
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

func formatRegexDeadIssue(receiver string, c contradiction) string {
	if c.isNeg {
		if c.regexSrc == receiver {
			return fmt.Sprintf(
				"Route to %q can never be reached: label %q requires %q (set by %q) but this route excludes it with regex %q",
				receiver, c.label, c.reqVal, c.reqSrc, c.regexPat,
			)
		}
		return fmt.Sprintf(
			"Route to %q can never be reached: label %q is excluded by regex %q (set by %q) but this route requires %q",
			receiver, c.label, c.regexPat, c.regexSrc, c.reqVal,
		)
	}
	if c.regexSrc == receiver {
		return fmt.Sprintf(
			"Route to %q can never be reached: label %q was set to %q by %q which does not match regex %q required by this route",
			receiver, c.label, c.reqVal, c.reqSrc, c.regexPat,
		)
	}
	return fmt.Sprintf(
		"Route to %q can never be reached: label %q requires regex %q (set by %q) but this route requires %q which does not match",
		receiver, c.label, c.regexPat, c.regexSrc, c.reqVal,
	)
}

func formatCrumb(crumb []string) string {
	return "[path: " + strings.Join(crumb, " → ") + "]"
}

// findContradiction scans matchers for irreconcilable conflicts:
//  1. Two positive exact matchers on the same label with different values.
//  2. A positive exact matcher cancelled by a negative exact matcher on the same label+value.
//  3. A positive exact value that does not satisfy a positive regex on the same label.
//  4. A positive exact value that is excluded by a negative regex on the same label.
func findContradiction(matchers []sanityMatcher) (contradiction, bool) {
	posExact := make(map[string]sanityMatcher) // label → first positive exact matcher seen

	// Pass 1: collect positive exact matchers; detect pos-pos exact conflicts.
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

	// Pass 2: detect negative-exact cancelling a required positive-exact value.
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
				negSrc: m.source,
			}, true
		}
	}

	// Pass 3: detect regex contradicting a positive exact value.
	for _, m := range matchers {
		if !m.isRegex || m.compiled == nil {
			continue
		}
		pos, ok := posExact[m.name]
		if !ok {
			continue
		}
		matched := m.compiled.MatchString(pos.value)
		if !m.isNeg && !matched {
			return contradiction{
				label:    m.name,
				reqVal:   pos.value,
				reqSrc:   pos.source,
				isRegex:  true,
				regexPat: m.value,
				regexSrc: m.source,
			}, true
		}
		if m.isNeg && matched {
			return contradiction{
				label:    m.name,
				reqVal:   pos.value,
				reqSrc:   pos.source,
				isNeg:    true,
				isRegex:  true,
				regexPat: m.value,
				regexSrc: m.source,
			}, true
		}
	}

	return contradiction{}, false
}
