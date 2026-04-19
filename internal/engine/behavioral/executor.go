package behavioral

import (
	"context"
	"fmt"
	"time"

	"github.com/nyambati/litmus/internal/engine/matching"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	alertmgr "github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

// BehavioralTestExecutor executes behavioral tests through the pipeline.
type BehavioralTestExecutor struct {
	inhibitRules []amconfig.InhibitRule
}

// NewBehavioralTestExecutor creates test executor.
func NewBehavioralTestExecutor(inhibitRules []amconfig.InhibitRule) *BehavioralTestExecutor {
	return &BehavioralTestExecutor{inhibitRules: inhibitRules}
}

// Execute runs a unit TestCase through the pipeline and verifies assertions.
func (bte *BehavioralTestExecutor) Execute(ctx context.Context, test *types.TestCase, router *pipeline.Router) *types.TestResult {
	var silences []types.Silence
	var activeAlerts []types.AlertSample
	if test.State != nil {
		silences = test.State.Silences
		activeAlerts = test.State.ActiveAlerts
	}

	silenceStore := stores.NewSilenceStore(silences)
	alertStore := stores.NewAlertStore()

	for _, alert := range activeAlerts {
		labelSet := make(model.LabelSet)
		for k, v := range alert.Labels {
			labelSet[model.LabelName(k)] = model.LabelValue(v)
		}
		alertmgrAlert := &alertmgr.Alert{
			Alert: model.Alert{
				Labels:   labelSet,
				StartsAt: time.Now(),
			},
		}
		if err := alertStore.Put(alertmgrAlert); err != nil {
			return &types.TestResult{
				Name:  test.Name,
				Type:  test.Type,
				Pass:  false,
				Error: fmt.Sprintf("seeding active alert: %v", err),
			}
		}
	}

	runner := pipeline.NewRunner(silenceStore, alertStore, router, bte.inhibitRules)

	if test.Alert == nil {
		return &types.TestResult{
			Name:  test.Name,
			Type:  test.Type,
			Pass:  false,
			Error: "test has no alert defined",
		}
	}

	labelSet := make(model.LabelSet)
	for k, v := range test.Alert.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}

	outcome, err := runner.Execute(ctx, labelSet)
	if err != nil {
		return &types.TestResult{
			Name:  test.Name,
			Type:  test.Type,
			Pass:  false,
			Error: fmt.Sprintf("pipeline execution failed: %v", err),
		}
	}

	if test.Expect == nil {
		return &types.TestResult{
			Name:  test.Name,
			Type:  test.Type,
			Pass:  false,
			Error: "test has no expect defined",
		}
	}

	if outcome.Status != test.Expect.Outcome {
		return &types.TestResult{
			Name: test.Name,
			Type: test.Type,
			Pass: false,
			Error: fmt.Sprintf(
				"Expected outcome: %q\n \t   - Actual outcome: %q",
				test.Expect.Outcome,
				outcome.Status,
			),
		}
	}

	if test.Expect.Outcome == "active" && len(test.Expect.Receivers) > 0 {
		if !matching.SubsetMatch(outcome.Receivers, test.Expect.Receivers) {
			return &types.TestResult{
				Name: test.Name,
				Type: test.Type,
				Pass: false,
				Error: fmt.Sprintf(
					"Expected receivers: %v\n \t   - Actual receivers: %v",
					test.Expect.Receivers,
					outcome.Receivers,
				),
			}
		}
	}

	return &types.TestResult{
		Name: test.Name,
		Type: test.Type,
		Pass: true,
	}
}
