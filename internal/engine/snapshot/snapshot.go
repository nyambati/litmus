package snapshot

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/prometheus/common/model"
)

// SynthesisResult holds outcome for synthesized alert.
type SynthesisResult struct {
	Labels    map[string]string
	Receivers []string
}

// SnapshotSynthesizer generates regression test baselines from route paths.
type SnapshotSynthesizer struct {
	runner       *pipeline.Runner
	expander     *RegexExpander
	combGen      *LabelCombinationGenerator
	failureCount int
	failureLimit int
}

// NewSnapshotSynthesizer creates synthesizer for snapshot generation.
func NewSnapshotSynthesizer(runner *pipeline.Runner) *SnapshotSynthesizer {
	return &SnapshotSynthesizer{
		runner:       runner,
		expander:     NewRegexExpander(),
		combGen:      NewLabelCombinationGenerator(5),
		failureLimit: 100, // Allow up to 100 failures before returning error
	}
}

// DiscoverOutcomes executes synthesized alerts through pipeline to discover outcomes.
func (ss *SnapshotSynthesizer) DiscoverOutcomes(ctx context.Context, paths []*RoutePath) ([]*SynthesisResult, error) {
	var results []*SynthesisResult
	seen := make(map[string]bool) // Dedup by outcome
	ss.failureCount = 0

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
				log.Printf("synthesis: pipeline execution failed for labels %v: %v", labels, err)
				ss.failureCount++
				if ss.failureCount > ss.failureLimit {
					return nil, fmt.Errorf("synthesis failed: exceeded maximum failures (%d)", ss.failureLimit)
				}
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

	if ss.failureCount > 0 {
		log.Printf("synthesis completed with %d pipeline execution failures", ss.failureCount)
	}

	return results, nil
}

// expandMatchers converts LabelSets to label options for combination generation.
func (ss *SnapshotSynthesizer) expandMatchers(matchers []model.LabelSet) map[string][]string {
	opts := make(map[string][]string)

	for _, labelSet := range matchers {
		for k, v := range labelSet {
			vals := ss.expander.ExpandAlternations(string(v))
			if len(vals) > 0 {
				opts[string(k)] = vals
			}
		}
	}

	return opts
}

// outcomeKey creates a stable unique key for a receiver list.
func (ss *SnapshotSynthesizer) outcomeKey(receivers []string) string {
	sorted := append([]string{}, receivers...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}
