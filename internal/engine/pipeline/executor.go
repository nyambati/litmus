package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/nyambati/litmus/internal/engine/matching"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	alertmgr "github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

// TestExecutor runs both unit (behavioral) and regression test cases through
// the pipeline. Branching is driven by TestCase.Type ("unit" or "regression").
type TestExecutor struct {
	inhibitRules []amconfig.InhibitRule
}

// NewTestExecutor creates a TestExecutor. Pass inhibit rules for unit tests;
// regression tests ignore them.
func NewTestExecutor(inhibitRules []amconfig.InhibitRule) *TestExecutor {
	return &TestExecutor{inhibitRules: inhibitRules}
}

// Execute runs a single test case through the pipeline and returns a result.
// It dispatches on test.Type: "unit" uses state + SubsetMatch + outcome check;
// "regression" uses empty state + ExactMatch across all label sets.
func (e *TestExecutor) Execute(ctx context.Context, test *types.TestCase, router *Router) *types.TestResult {
	if test == nil {
		return &types.TestResult{Pass: false, Error: "test case is nil"}
	}
	if router == nil {
		return &types.TestResult{Name: test.Name, Type: test.Type, Pass: false, Error: "router is nil"}
	}
	switch test.Type {
	case "regression":
		return e.executeRegression(ctx, test, router)
	default:
		return e.executeUnit(ctx, test, router)
	}
}

// ExecuteAll runs a batch of test cases and returns one result per case.
func (e *TestExecutor) ExecuteAll(ctx context.Context, tests []*types.TestCase, router *Router) []*types.TestResult {
	results := make([]*types.TestResult, 0, len(tests))
	for _, test := range tests {
		if test == nil {
			continue
		}
		results = append(results, e.Execute(ctx, test, router))
	}
	return results
}

func (e *TestExecutor) executeUnit(ctx context.Context, test *types.TestCase, router *Router) *types.TestResult {
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
		if err := alertStore.Put(&alertmgr.Alert{
			Alert: model.Alert{Labels: labelSet, StartsAt: time.Now()},
		}); err != nil {
			return &types.TestResult{
				Name:  test.Name,
				Type:  test.Type,
				Pass:  false,
				Error: fmt.Sprintf("seeding active alert: %v", err),
			}
		}
	}

	runner := NewRunner(silenceStore, alertStore, router, e.inhibitRules)

	if test.Alert == nil {
		return &types.TestResult{Name: test.Name, Type: test.Type, Pass: false, Error: "test has no alert defined"}
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
		return &types.TestResult{Name: test.Name, Type: test.Type, Pass: false, Error: "test has no expect defined"}
	}

	if outcome.Status != test.Expect.Outcome {
		return &types.TestResult{
			Name:  test.Name,
			Type:  test.Type,
			Pass:  false,
			Error: fmt.Sprintf("Expected outcome: %q\n \t   - Actual outcome: %q", test.Expect.Outcome, outcome.Status),
		}
	}

	if test.Expect.Outcome == "active" && len(test.Expect.Receivers) > 0 {
		if !matching.SubsetMatch(outcome.Receivers, test.Expect.Receivers) {
			return &types.TestResult{
				Name:  test.Name,
				Type:  test.Type,
				Pass:  false,
				Error: fmt.Sprintf("Expected receivers: %v\n \t   - Actual receivers: %v", test.Expect.Receivers, outcome.Receivers),
			}
		}
	}

	return &types.TestResult{Name: test.Name, Type: test.Type, Pass: true}
}

func (e *TestExecutor) executeRegression(ctx context.Context, test *types.TestCase, router *Router) *types.TestResult {
	result := &types.TestResult{Name: test.Name, Type: test.Type, Pass: true}

	if len(test.Labels) == 0 {
		result.Pass = false
		result.Error = "test has no label sets to validate"
		return result
	}
	if test.Expect == nil {
		result.Pass = false
		result.Error = "test has no expect defined"
		return result
	}

	silenceStore := stores.NewSilenceStore(nil)
	alertStore := stores.NewAlertStore()
	runner := NewRunner(silenceStore, alertStore, router, nil)

	for _, labels := range test.Labels {
		labelSet := make(model.LabelSet)
		for k, v := range labels {
			labelSet[model.LabelName(k)] = model.LabelValue(v)
		}
		outcome, err := runner.Execute(ctx, labelSet)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Sprintf("pipeline execution failed: %v", err)
			result.Labels = labels
			break
		}
		if !matching.ExactMatch(outcome.Receivers, test.Expect.Receivers) {
			result.Pass = false
			result.Labels = labels
			result.Expected = test.Expect.Receivers
			result.Actual = outcome.Receivers
			break
		}
	}

	return result
}
