package sanity

import (
	"fmt"

	"github.com/prometheus/alertmanager/config"
)

// InhibitionCycleDetector detects cycles in inhibition rules.
type InhibitionCycleDetector struct {
	rules []*config.InhibitRule
}

// NewInhibitionCycleDetector creates cycle detector for inhibition rules.
func NewInhibitionCycleDetector(rules []*config.InhibitRule) *InhibitionCycleDetector {
	return &InhibitionCycleDetector{rules: rules}
}

// DetectCycles returns list of detected inhibition cycles.
func (icd *InhibitionCycleDetector) DetectCycles() []string {
	var cycles []string

	// Build adjacency list: source matcher -> target matcher
	graph := make(map[string][]string)
	for _, rule := range icd.rules {
		if rule == nil {
			continue
		}
		source := icd.matcherKey(rule.SourceMatch)
		target := icd.matcherKey(rule.TargetMatch)
		if source != "" && target != "" {
			graph[source] = append(graph[source], target)
		}
	}

	// DFS-based cycle detection
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

// hasCycle performs DFS to detect cycles.
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

// matcherKey creates a unique key from matcher map.
func (icd *InhibitionCycleDetector) matcherKey(matcher map[string]string) string {
	if len(matcher) == 0 {
		return ""
	}
	key := ""
	for k, v := range matcher {
		key += k + "=" + v + "|"
	}
	return key
}

// OrphanReceiverDetector detects unused receivers.
type OrphanReceiverDetector struct {
	root      *config.Route
	receivers map[string]*config.Receiver
}

// NewOrphanReceiverDetector creates orphan detector.
func NewOrphanReceiverDetector(root *config.Route, receivers map[string]*config.Receiver) *OrphanReceiverDetector {
	return &OrphanReceiverDetector{
		root:      root,
		receivers: receivers,
	}
}

// DetectOrphans returns list of unused receivers.
func (ord *OrphanReceiverDetector) DetectOrphans() []string {
	used := make(map[string]bool)

	// Traverse routes and mark used receivers
	ord.markUsed(ord.root, used)

	var orphans []string
	for name := range ord.receivers {
		if !used[name] {
			orphans = append(orphans, fmt.Sprintf("Receiver %q is defined but never used", name))
		}
	}

	return orphans
}

// markUsed recursively marks receivers as used.
func (ord *OrphanReceiverDetector) markUsed(route *config.Route, used map[string]bool) {
	if route == nil {
		return
	}
	if route.Receiver != "" {
		used[route.Receiver] = true
	}

	for _, child := range route.Routes {
		ord.markUsed(child, used)
	}
}
