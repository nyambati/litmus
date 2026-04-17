package behavioral

import (
	"context"
	"fmt"
	"time"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	alertmgr "github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

// TestResult holds execution result for a behavioral test.
type TestResult struct {
	Name  string
	Pass  bool
	Error string
}

// BehavioralTestExecutor executes behavioral tests through the pipeline.
type BehavioralTestExecutor struct {
	inhibitRules []amconfig.InhibitRule
}

// NewBehavioralTestExecutor creates test executor.
func NewBehavioralTestExecutor(inhibitRules []amconfig.InhibitRule) *BehavioralTestExecutor {
	return &BehavioralTestExecutor{inhibitRules: inhibitRules}
}

// Execute runs a behavioral test through the pipeline and verifies assertions.
func (bte *BehavioralTestExecutor) Execute(ctx context.Context, test *types.BehavioralTest, router *pipeline.Router) *TestResult {
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
			return &TestResult{
				Name:  test.Name,
				Pass:  false,
				Error: fmt.Sprintf("seeding active alert: %v", err),
			}
		}
	}

	runner := pipeline.NewRunner(silenceStore, alertStore, router, bte.inhibitRules)

	labelSet := make(model.LabelSet)
	for k, v := range test.Alert.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}

	outcome, err := runner.Execute(ctx, labelSet)
	if err != nil {
		return &TestResult{
			Name:  test.Name,
			Pass:  false,
			Error: fmt.Sprintf("pipeline execution failed: %v", err),
		}
	}

	if outcome.Status != test.Expect.Outcome {
		return &TestResult{
			Name: test.Name,
			Pass: false,
			Error: fmt.Sprintf(
				"Expected outcome: %q\n \t   - Actual outcome: %q",
				test.Expect.Outcome,
				outcome.Status,
			),
		}
	}

	if test.Expect.Outcome == "active" && len(test.Expect.Receivers) > 0 {
		if !receiversMatch(outcome.Receivers, test.Expect.Receivers) {
			return &TestResult{
				Name: test.Name,
				Pass: false,
				Error: fmt.Sprintf(
					"Expected receivers: %v\n \t   - Actual receivers: %v",
					test.Expect.Receivers,
					outcome.Receivers,
				),
			}
		}
	}

	return &TestResult{
		Name: test.Name,
		Pass: true,
	}
}

// receiversMatch checks if actual receivers contain all expected receivers.
func receiversMatch(actual, expected []string) bool {
	actualMap := make(map[string]bool)
	for _, r := range actual {
		actualMap[r] = true
	}
	for _, r := range expected {
		if !actualMap[r] {
			return false
		}
	}
	return true
}
