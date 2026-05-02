package sanity

import (
	"fmt"

	"github.com/prometheus/alertmanager/config"
)

// OrphanReceiverDetector detects receivers defined but never referenced by any route.
type OrphanReceiverDetector struct {
	root      *config.Route
	receivers map[string]*config.Receiver
}

// NewOrphanReceiverDetector creates a detector for the given route tree and receiver map.
func NewOrphanReceiverDetector(root *config.Route, receivers map[string]*config.Receiver) *OrphanReceiverDetector {
	return &OrphanReceiverDetector{root: root, receivers: receivers}
}

// Name implements Check.
func (ord *OrphanReceiverDetector) Name() string { return "orphan_receivers" }

// Run implements Check.
func (ord *OrphanReceiverDetector) Run(ctx CheckContext) []string {
	return NewOrphanReceiverDetector(ctx.Route, ctx.Receivers).DetectOrphans()
}

// DetectOrphans returns a list of receiver names that are never referenced.
func (ord *OrphanReceiverDetector) DetectOrphans() []string {
	used := make(map[string]bool)
	ord.markUsed(ord.root, used)

	var orphans []string
	for name := range ord.receivers {
		if !used[name] {
			orphans = append(orphans, fmt.Sprintf("Receiver %q is defined but never used", name))
		}
	}

	return orphans
}

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
