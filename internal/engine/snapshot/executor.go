package snapshot

import (
	"context"
	"fmt"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	"github.com/prometheus/common/model"
)

// RegressionResult holds execution result for a regression test.
type RegressionResult struct {
	Name  string
	Pass  bool
	Error string
}

// RegressionTestExecutor executes regression tests through the pipeline.
type RegressionTestExecutor struct{}

// NewRegressionTestExecutor creates regression test executor.
func NewRegressionTestExecutor() *RegressionTestExecutor {
	return &RegressionTestExecutor{}
}

// Execute runs regression tests against the router.
// Regression tests are expected to match exactly the receivers.
func (rte *RegressionTestExecutor) Execute(ctx context.Context, tests []*types.RegressionTest, router *pipeline.Router) []*RegressionResult {
	results := make([]*RegressionResult, 0, len(tests))

	// Regression tests run with empty silence and alert stores (pure routing)
	silenceStore := stores.NewSilenceStore(nil)
	alertStore := stores.NewAlertStore()
	runner := pipeline.NewRunner(silenceStore, alertStore, router)

	for _, test := range tests {
		pass := true
		var errMsg string

		// A regression test might have multiple label sets to check
		for _, labels := range test.Labels {
			labelSet := make(model.LabelSet)
			for k, v := range labels {
				labelSet[model.LabelName(k)] = model.LabelValue(v)
			}

			outcome, err := runner.Execute(ctx, labelSet)
			if err != nil {
				pass = false
				errMsg = fmt.Sprintf("pipeline execution failed: %v", err)
				break
			}

			if !receiversMatch(outcome.Receivers, test.Expected) {
				pass = false
				errMsg = fmt.Sprintf("expected receivers %v, got %v", test.Expected, outcome.Receivers)
				break
			}
		}

		results = append(results, &RegressionResult{
			Name:  test.Name,
			Pass:  pass,
			Error: errMsg,
		})
	}

	return results
}

// receiversMatch checks if actual receivers exactly match expected receivers.
func receiversMatch(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
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
