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
	Warnings  []string
}

// SnapshotSynthesizer generates regression test baselines from route paths.
type SnapshotSynthesizer struct {
	runner       *pipeline.Runner
	expander     *RegexExpander
	combGen      *LabelCombinationGenerator
	diagnostics  []string
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
	if ss == nil || ss.runner == nil {
		return nil, fmt.Errorf("synthesizer or runner is nil")
	}
	var results []*SynthesisResult
	seen := make(map[string]bool) // Dedup by outcome
	ss.failureCount = 0
	ss.diagnostics = nil

	for _, path := range paths {
		if path == nil {
			continue
		}
		// Convert matchers to label options for expansion
		labelOpts := ss.expandMatchers(path.Matchers)
		warnings := ss.pathWarnings(path)

		if len(labelOpts) == 0 && len(path.IgnoredMatchers) > 0 {
			ss.diagnostics = append(ss.diagnostics, fmt.Sprintf(
				"skipped route %q: negative matchers are not synthesizable from the route tree (%s)",
				path.Receiver,
				strings.Join(path.IgnoredMatchers, ", "),
			))
			continue
		}

		// Generate covering set
		combos := ss.combGen.GenerateCovering(labelOpts)
		matchedRoute := false

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

			if path.Receiver != "" && !containsReceiver(outcome.Receivers, path.Receiver) {
				continue
			}
			matchedRoute = true

			// Dedup by outcome (receiver list)
			outcomeKey := ss.outcomeKey(outcome.Receivers) + "|" + strings.Join(warnings, "|")
			if !seen[outcomeKey] {
				seen[outcomeKey] = true
				results = append(results, &SynthesisResult{
					Labels:    labels,
					Receivers: outcome.Receivers,
					Warnings:  append([]string(nil), warnings...),
				})
			}
		}

		if !matchedRoute && len(warnings) > 0 {
			ss.diagnostics = append(ss.diagnostics, fmt.Sprintf(
				"skipped route %q: generated labels from positive matchers did not reliably exercise the route because negative matchers were ignored",
				path.Receiver,
			))
		}
	}

	if ss.failureCount > 0 {
		log.Printf("synthesis completed with %d pipeline execution failures", ss.failureCount)
	}

	return results, nil
}

// Diagnostics returns synthesis warnings and skip reasons accumulated during the last run.
func (ss *SnapshotSynthesizer) Diagnostics() []string {
	if ss == nil {
		return nil
	}
	return append([]string(nil), ss.diagnostics...)
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

func (ss *SnapshotSynthesizer) pathWarnings(path *RoutePath) []string {
	if path == nil || len(path.IgnoredMatchers) == 0 {
		return nil
	}
	return []string{fmt.Sprintf(
		"incomplete synthesis coverage: ignored negative matchers %s",
		strings.Join(path.IgnoredMatchers, ", "),
	)}
}

// outcomeKey creates a stable unique key for a receiver list.
func (ss *SnapshotSynthesizer) outcomeKey(receivers []string) string {
	sorted := append([]string{}, receivers...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}

func containsReceiver(receivers []string, receiver string) bool {
	for _, r := range receivers {
		if r == receiver {
			return true
		}
	}
	return false
}
