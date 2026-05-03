package sanity

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prometheus/alertmanager/config"
)

// InhibitionCycleDetector detects cycles in inhibition rules.
type InhibitionCycleDetector struct {
	rules []*config.InhibitRule
}

// NewInhibitionCycleDetector creates a cycle detector for the given inhibition rules.
func NewInhibitionCycleDetector(rules []*config.InhibitRule) *InhibitionCycleDetector {
	return &InhibitionCycleDetector{rules: rules}
}

// Name implements Check.
func (icd *InhibitionCycleDetector) Name() string { return "inhibition_cycles" }

// Run implements Check.
func (icd *InhibitionCycleDetector) Run(ctx CheckContext) []string {
	return NewInhibitionCycleDetector(ctx.Rules).DetectCycles()
}

// DetectCycles returns a list of detected inhibition cycles.
func (icd *InhibitionCycleDetector) DetectCycles() []string {
	var cycles []string

	graph := make(map[string][]string)
	for _, rule := range icd.rules {
		if rule == nil {
			continue
		}
		source := icd.ruleMatcherKey(rule.SourceMatch, rule.SourceMatchRE, rule.SourceMatchers)
		target := icd.ruleMatcherKey(rule.TargetMatch, rule.TargetMatchRE, rule.TargetMatchers)
		if source != "" && target != "" {
			graph[source] = append(graph[source], target)
		}
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for node := range graph {
		if !visited[node] {
			if icd.hasCycle(node, graph, visited, recStack) {
				cycles = append(cycles, fmt.Sprintf("Inhibition cycle detected involving %s", node))
			}
		}
	}

	return cycles
}

func (icd *InhibitionCycleDetector) hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if icd.hasCycle(neighbor, graph, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[node] = false
	return false
}

func (icd *InhibitionCycleDetector) ruleMatcherKey(
	exact map[string]string,
	regex config.MatchRegexps,
	matchers config.Matchers,
) string {
	parts := make([]string, 0, len(exact)+len(regex)+len(matchers))

	for k, v := range exact {
		parts = append(parts, k+"="+v)
	}
	for k, v := range regex {
		parts = append(parts, k+"=~"+v.String())
	}
	for _, matcher := range matchers {
		parts = append(parts, matcher.String())
	}

	if len(parts) == 0 {
		return ""
	}

	sort.Strings(parts)
	return strings.Join(parts, "|")
}
