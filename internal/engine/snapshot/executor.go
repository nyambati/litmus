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

// RegressionResult holds execution result for a single regression test.
type RegressionResult struct {
	Name     string
	Pass     bool
	Error    string
	Labels   map[string]string // failing label set; nil on pass
	Expected []string
	Actual   []string
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

	silenceStore := stores.NewSilenceStore(nil)
	alertStore := stores.NewAlertStore()
	runner := pipeline.NewRunner(silenceStore, alertStore, router, nil)

	for _, test := range tests {
		result := &RegressionResult{Name: test.Name, Pass: true}

		if len(test.Labels) == 0 {
			result.Pass = false
			result.Error = "test has no label sets to validate"
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

			if !matching.ExactMatch(outcome.Receivers, test.Expected) {
				result.Pass = false
				result.Labels = labels
				result.Expected = test.Expected
				result.Actual = outcome.Receivers
				break
			}
		}

		results = append(results, result)
	}

	return results
}
