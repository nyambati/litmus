package snapshot

import (
	"context"
	"fmt"

	"github.com/nyambati/litmus/internal/engine/matching"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	"github.com/prometheus/common/model"
)

// RegressionTestExecutor executes regression tests through the pipeline.
type RegressionTestExecutor struct{}

// NewRegressionTestExecutor creates regression test executor.
func NewRegressionTestExecutor() *RegressionTestExecutor {
	return &RegressionTestExecutor{}
}

// Execute runs regression TestCases against the router.
func (rte *RegressionTestExecutor) Execute(ctx context.Context, tests []*types.TestCase, router *pipeline.Router) []*types.TestResult {
	if router == nil {
		return []*types.TestResult{{Pass: false, Error: "router is nil"}}
	}
	results := make([]*types.TestResult, 0, len(tests))

	silenceStore := stores.NewSilenceStore(nil)
	alertStore := stores.NewAlertStore()
	runner := pipeline.NewRunner(silenceStore, alertStore, router, nil)

	for _, test := range tests {
		if test == nil {
			continue
		}
		result := &types.TestResult{Name: test.Name, Type: test.Type, Pass: true}

		if len(test.Labels) == 0 {
			result.Pass = false
			result.Error = "test has no label sets to validate"
			results = append(results, result)
			continue
		}

		if test.Expect == nil {
			result.Pass = false
			result.Error = "test has no expect defined"
			results = append(results, result)
			continue
		}

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

		results = append(results, result)
	}

	return results
}
