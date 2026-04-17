package pipeline

import (
	"context"

	"github.com/prometheus/common/model"
	"litmus/internal/stores"
)

// Outcome represents result of executing alert through pipeline.
type Outcome struct {
	Status    string   // "silenced", "inhibited", or "active"
	Receivers []string // List of receivers if active
}

// Runner is unified execution pipeline for routing and suppression.
type Runner struct {
	silenceStore *stores.SilenceStore
	alertStore   *stores.AlertStore
	receivers    []string // Default receivers for routing
}

// NewRunner creates pipeline runner with stores and config.
func NewRunner(silenceStore *stores.SilenceStore, alertStore *stores.AlertStore, receivers []string) *Runner {
	return &Runner{
		silenceStore: silenceStore,
		alertStore:   alertStore,
		receivers:    receivers,
	}
}

// Execute processes alert through pipeline: silence -> inhibit -> route.
func (r *Runner) Execute(ctx context.Context, labels model.LabelSet) (*Outcome, error) {
	// Check if silenced
	if r.silenceStore.Mutes(ctx, labels) {
		return &Outcome{Status: "silenced"}, nil
	}

	// Check if inhibited by any active alert
	iter := r.alertStore.GetPending()
	alertChan := iter.Next()
	for next := range alertChan {
		// Check if this alert inhibits the incoming one (simple label match)
		inhibitLabels := model.LabelSet(next.Labels)
		if r.labelsMatch(inhibitLabels, labels) {
			iter.Close()
			return &Outcome{Status: "inhibited"}, nil
		}
	}
	iter.Close()

	// Active: route to receivers
	return &Outcome{
		Status:    "active",
		Receivers: r.receivers,
	}, nil
}

// labelsMatch checks if inhibitor labels match all target labels.
func (r *Runner) labelsMatch(inhibitor, target model.LabelSet) bool {
	for k, v := range inhibitor {
		if target[k] != v {
			return false
		}
	}
	return true
}
