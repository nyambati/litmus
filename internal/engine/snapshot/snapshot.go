package snapshot

import (
	"context"

	"github.com/prometheus/common/model"
	"litmus/internal/engine/pipeline"
)

// SynthesisResult holds outcome for synthesized alert.
type SynthesisResult struct {
	Labels    map[string]string
	Receivers []string
}

// SnapshotSynthesizer generates regression test baselines from route paths.
type SnapshotSynthesizer struct {
	runner   *pipeline.Runner
	expander *RegexExpander
	combGen  *LabelCombinationGenerator
}

// NewSnapshotSynthesizer creates synthesizer for snapshot generation.
func NewSnapshotSynthesizer(runner *pipeline.Runner) *SnapshotSynthesizer {
	return &SnapshotSynthesizer{
		runner:   runner,
		expander: NewRegexExpander(),
		combGen:  NewLabelCombinationGenerator(5),
	}
}

// DiscoverOutcomes executes synthesized alerts through pipeline to discover outcomes.
func (ss *SnapshotSynthesizer) DiscoverOutcomes(ctx context.Context, paths []*RoutePath) []*SynthesisResult {
	var results []*SynthesisResult
	seen := make(map[string]bool) // Dedup by outcome

	for _, path := range paths {
		// Convert matchers to label options for expansion
		labelOpts := ss.expandMatchers(path.Matchers)

		// Generate covering set
		combos := ss.combGen.GenerateCovering(labelOpts)

		// Execute each through pipeline
		for _, labels := range combos {
			labelSet := make(model.LabelSet)
			for k, v := range labels {
				labelSet[model.LabelName(k)] = model.LabelValue(v)
			}

			outcome, err := ss.runner.Execute(ctx, labelSet)
			if err != nil {
				continue
			}

			// Dedup by outcome (receiver list)
			outcomeKey := ss.outcomeKey(outcome.Receivers)
			if !seen[outcomeKey] {
				seen[outcomeKey] = true
				results = append(results, &SynthesisResult{
					Labels:    labels,
					Receivers: outcome.Receivers,
				})
			}
		}
	}

	return results
}

// expandMatchers converts LabelSets to label options for combination generation.
func (ss *SnapshotSynthesizer) expandMatchers(matchers []model.LabelSet) map[string][]string {
	opts := make(map[string][]string)

	for _, labelSet := range matchers {
		for k, v := range labelSet {
			vals := ss.expander.ExpandAlternations(string(v))
			opts[string(k)] = vals
		}
	}

	return opts
}

// outcomeKey creates unique key for receiver list.
func (ss *SnapshotSynthesizer) outcomeKey(receivers []string) string {
	// Simple join - in production would sort for consistency
	var key string
	for i, r := range receivers {
		if i > 0 {
			key += ","
		}
		key += r
	}
	return key
}
