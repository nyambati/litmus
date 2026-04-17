package behavioral

import (
	"context"
	"fmt"
	"time"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
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
type BehavioralTestExecutor struct{}

// NewBehavioralTestExecutor creates test executor.
func NewBehavioralTestExecutor() *BehavioralTestExecutor {
	return &BehavioralTestExecutor{}
}

// Execute runs a behavioral test through the pipeline and verifies assertions.
func (bte *BehavioralTestExecutor) Execute(ctx context.Context, test *types.BehavioralTest, router *pipeline.Router) *TestResult {
	// Create stores from test state
	var silences []types.Silence
	var activeAlerts []types.AlertSample
	if test.State != nil {
		silences = test.State.Silences
		activeAlerts = test.State.ActiveAlerts
	}

	silenceStore := stores.NewSilenceStore(silences)
	alertStore := stores.NewAlertStore()

	// Populate alert store with active alerts
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
		alertStore.Put(alertmgrAlert)
	}

	// Create pipeline runner
	runner := pipeline.NewRunner(silenceStore, alertStore, router)

	// Convert test alert labels to model.LabelSet
	labelSet := make(model.LabelSet)
	for k, v := range test.Alert.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}

	// Execute through pipeline
	outcome, err := runner.Execute(ctx, labelSet)
	if err != nil {
		return &TestResult{
			Name:  test.Name,
			Pass:  false,
			Error: fmt.Sprintf("pipeline execution failed: %v", err),
		}
	}

	// Verify outcome
	if outcome.Status != test.Expect.Outcome {
		return &TestResult{
			Name:  test.Name,
			Pass:  false,
			Error: fmt.Sprintf("expected outcome %q, got %q", test.Expect.Outcome, outcome.Status),
		}
	}

	// Verify receivers if outcome is active
	if test.Expect.Outcome == "active" && len(test.Expect.Receivers) > 0 {
		if !receiversMatch(outcome.Receivers, test.Expect.Receivers) {
			return &TestResult{
				Name:  test.Name,
				Pass:  false,
				Error: fmt.Sprintf("expected receivers %v, got %v", test.Expect.Receivers, outcome.Receivers),
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
