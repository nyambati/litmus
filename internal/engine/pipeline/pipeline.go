package pipeline

import (
	"context"
	"fmt"

	"github.com/nyambati/litmus/internal/labelmatcher"
	"github.com/nyambati/litmus/internal/stores"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

// Outcome represents result of executing alert through pipeline.
type Outcome struct {
	Status    string   // one of StatusSilenced, StatusInhibited, StatusActive
	Receivers []string // List of receivers if active
}

// Pipeline status constants
const (
	StatusSilenced  = "silenced"
	StatusInhibited = "inhibited"
	StatusActive    = "active"
)

// Runner is unified execution pipeline for routing and suppression.
type Runner struct {
	silenceStore *stores.SilenceStore
	alertStore   *stores.AlertStore
	router       *Router
	inhibitRules []amconfig.InhibitRule
}

// NewRunner creates pipeline runner with stores and config.
func NewRunner(silenceStore *stores.SilenceStore, alertStore *stores.AlertStore, router *Router, inhibitRules []amconfig.InhibitRule) *Runner {
	return &Runner{
		silenceStore: silenceStore,
		alertStore:   alertStore,
		router:       router,
		inhibitRules: inhibitRules,
	}
}

// Execute processes alert through pipeline: silence -> inhibit -> route.
func (r *Runner) Execute(ctx context.Context, labels model.LabelSet) (*Outcome, error) {
	if r == nil {
		return nil, fmt.Errorf("runner is nil")
	}

	if r.silenceStore != nil && r.silenceStore.Mutes(ctx, labels) {
		return &Outcome{Status: StatusSilenced}, nil
	}

	if r.alertStore != nil {
		iter := r.alertStore.GetPending()
		alertChan := iter.Next()
		for activeAlert := range alertChan {
			activeLabels := model.LabelSet(activeAlert.Labels)
			if r.isInhibited(activeLabels, labels) {
				iter.Close()
				return &Outcome{Status: StatusInhibited}, nil
			}
		}
		iter.Close()
		if err := iter.Err(); err != nil {
			return nil, fmt.Errorf("checking inhibition: %w", err)
		}
	}

	var receivers []string
	if r.router != nil {
		receivers = r.router.Match(labels)
	}

	return &Outcome{
		Status:    StatusActive,
		Receivers: receivers,
	}, nil
}

// isInhibited checks if target is inhibited by source using configured rules.
func (r *Runner) isInhibited(source, target model.LabelSet) bool {
	for _, rule := range r.inhibitRules {
		if !labelmatcher.SourceMatches(rule, source) {
			continue
		}
		if !labelmatcher.TargetMatches(rule, target) {
			continue
		}
		if labelmatcher.EqualLabelsMatch(rule.Equal, source, target) {
			return true
		}
	}
	return false
}
