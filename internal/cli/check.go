package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	"github.com/nyambati/litmus/internal/workspace"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
)

const divider = "--------------------------------------------------"

// CheckExitCode signals which exit code the caller should use.
type CheckExitCode int

// CheckResult holds results from a full validation run.
type CheckResult struct {
	Passed     bool             `json:"passed"`
	ConfigPath string           `json:"config_path"`
	Sanity     sanity.Result    `json:"sanity"`
	Regression RegressionResult `json:"regression"`
	Behavioral BehavioralResult `json:"behavioral"`
	Duration   time.Duration    `json:"duration_ns"`
	ExitCode   CheckExitCode    `json:"exit_code"`
}

// TestFailure holds structured detail for a single test failure.
type TestFailure struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Error    string            `json:"error,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	Expected []string          `json:"expected,omitempty"`
	Actual   []string          `json:"actual,omitempty"`
}

// RegressionResult holds regression test results.
type RegressionResult struct {
	Passed     bool          `json:"passed"`
	Tests      int           `json:"tests"`
	TotalTests int           `json:"total_tests"`
	PassCount  int           `json:"pass_count"`
	Failures   []TestFailure `json:"failures,omitempty"`
}

// BehavioralResult holds behavioral test results.
type BehavioralResult struct {
	Passed     bool          `json:"passed"`
	Tests      int           `json:"tests"`
	TotalTests int           `json:"total_tests"`
	PassCount  int           `json:"pass_count"`
	Failures   []TestFailure `json:"failures,omitempty"`
}

// RunCheck loads config, runs all validation stages, prints results, and returns
// the exit code the CLI layer should pass to os.Exit (0 = all passed).
func RunCheck(cfg *config.LitmusConfig, logger logrus.FieldLogger, format string, showDiff bool, tags []string) (CheckExitCode, error) {
	start := time.Now()
	ws, err := workspace.Load(cfg.Workspace.Root, logger)
	if err != nil {
		return 1, err
	}

	amCfg, err := ws.Config()
	if err != nil {
		return 1, fmt.Errorf("failed to load alertmanager config: %w", err)
	}

	if amCfg.Route == nil {
		return 1, fmt.Errorf("alertmanager config has no route defined")
	}

	router := pipeline.NewRouter(amCfg.Route)

	fragments := ws.Fragments()
	if ws.RootFragment() != nil {
		fragments = append([]*fragment.Fragment{ws.RootFragment()}, fragments...)
	}
	ctx := context.Background()
	sanityResult := runSanityChecks(cfg, fragments, amCfg)
	regressionResult := RunRegressionTests(ctx, cfg, router, tags)
	behavioralResult := RunBehavioralTests(ctx, cfg, ws.Fragments(), ws.Tests(), router, amCfg.InhibitRules, tags)

	result := buildCheckResult(cfg.FilePath(), sanityResult, regressionResult, behavioralResult, time.Since(start))

	code := calculateExitCode(result)
	result.ExitCode = code

	if err := outputResults(result, format, showDiff); err != nil {
		return 1, err
	}

	return code, nil
}

// runSanityChecks executes all sanity checks including policy enforcement.
func runSanityChecks(litmusConfig *config.LitmusConfig, fragments []*fragment.Fragment, amCfg *amconfig.Config) sanity.Result {
	receiversMap := make(map[string]*amconfig.Receiver)
	for i := range amCfg.Receivers {
		receiversMap[amCfg.Receivers[i].Name] = &amCfg.Receivers[i]
	}
	rules := make([]*amconfig.InhibitRule, 0, len(amCfg.InhibitRules))
	for i := range amCfg.InhibitRules {
		rules = append(rules, &amCfg.InhibitRules[i])
	}
	ctx := sanity.CheckContext{
		Route:     amCfg.Route,
		Receivers: receiversMap,
		Rules:     rules,
		Policy:    litmusConfig.Policy,
		Fragments: fragments,
	}
	return sanity.Run(ctx, litmusConfig.Sanity)
}

// buildCheckResult assembles the final check result from all test results.
func buildCheckResult(configPath string, sanityResult sanity.Result, regressionResult RegressionResult, behavioralResult BehavioralResult, duration time.Duration) CheckResult {
	passed := sanityResult.Passed && regressionResult.Passed && behavioralResult.Passed

	return CheckResult{
		Passed:     passed,
		ConfigPath: configPath,
		Sanity:     sanityResult,
		Regression: regressionResult,
		Behavioral: behavioralResult,
		Duration:   duration,
	}
}

// calculateExitCode determines the exit code based on test results.
func calculateExitCode(result CheckResult) CheckExitCode {
	if result.Passed {
		return 0
	}

	if !result.Sanity.Passed {
		return 3
	}
	return 2
}

// outputResults formats and outputs the check results.
func outputResults(result CheckResult, format string, showDiff bool) error {
	if format == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data)) //nolint:forbidigo
	} else {
		PrintCheckResult(result, showDiff)
	}
	return nil
}

// RunRegressionTests executes the regression baseline against the current router.
func RunRegressionTests(ctx context.Context, litmusConfig *config.LitmusConfig, router *pipeline.Router, tags []string) RegressionResult {
	result := RegressionResult{Passed: true}

	state, err := LoadRegressionState(litmusConfig.RegressionsYamlFilePath())
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "WARN: could not read regression baseline: %v\n", err)
		}
		return result
	}

	baseline := state.Tests
	if len(baseline) == 0 {
		return result
	}

	result.TotalTests = len(baseline)
	baseline = filterByTags(baseline, tags)
	result.Tests = len(baseline)
	executor := pipeline.NewTestExecutor(nil)

	for _, res := range executor.ExecuteAll(ctx, baseline, router) {
		if res.Pass {
			result.PassCount++
		} else {
			result.Passed = false
			result.Failures = append(result.Failures, TestFailure{
				Name:     res.Name,
				Type:     res.Type,
				Labels:   res.Labels,
				Expected: res.Expected,
				Actual:   res.Actual,
			})
		}
	}

	return result
}

// RunBehavioralTests loads and executes all behavioral unit tests.
func RunBehavioralTests(ctx context.Context, litmusConfig *config.LitmusConfig, fragments []*fragment.Fragment, workspaceTests []*types.TestCase, router *pipeline.Router, inhibitRules []amconfig.InhibitRule, tags []string) BehavioralResult {
	result := BehavioralResult{Passed: true}

	var tests []*types.TestCase
	for _, frag := range fragments {
		tests = append(tests, frag.Tests...)
	}
	tests = append(tests, workspaceTests...)

	if len(tests) == 0 {
		return result
	}

	result.TotalTests = len(tests)
	tests = filterByTags(tests, tags)
	result.Tests = len(tests)
	executor := pipeline.NewTestExecutor(inhibitRules)

	for _, test := range tests {
		res := executor.Execute(ctx, test, router)
		if res.Pass {
			result.PassCount++
		} else {
			result.Passed = false
			result.Failures = append(result.Failures, TestFailure{
				Name:  res.Name,
				Type:  res.Type,
				Error: res.Error,
			})
		}
	}

	return result
}

// filterByTags filters test cases to only those with matching tags.
// If tags is empty, all tests are returned. Uses OR semantics: a test
// is included if it has at least one tag in the filter list.
// Whitespace is trimmed from tag names and empty tags are ignored.
func filterByTags(tests []*types.TestCase, tags []string) []*types.TestCase {
	if len(tags) == 0 {
		return tests
	}
	tagSet := make(map[string]struct{})
	for _, t := range tags {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			tagSet[trimmed] = struct{}{}
		}
	}
	if len(tagSet) == 0 {
		return tests
	}
	out := make([]*types.TestCase, 0, len(tests))
	for _, tc := range tests {
		for _, tag := range tc.Tags {
			if _, ok := tagSet[tag]; ok {
				out = append(out, tc)
				break
			}
		}
	}
	return out
}

// PrintCheckResult writes the formatted validation report to stdout.
//
//nolint:forbidigo
func PrintCheckResult(r CheckResult, showDiff bool) {
	fmt.Printf("Litmus Check: %s\n", r.ConfigPath)
	fmt.Println(divider)
	fmt.Println()

	// 1. Sanity
	fmt.Println("1. Sanity (Static Analysis)")
	for _, c := range r.Sanity.Checks {
		printSanityCategory(sanityCheckLabel(c.Name), c.Issues, c.Mode)
	}
	fmt.Println()

	// 2. Regressions
	fmt.Println("2. Regressions (Automated)")
	//nolint:gocritic
	if r.Regression.TotalTests == 0 {
		fmt.Println("   [SKIP]  No baseline found — run 'litmus snapshot' first")
	} else if r.Regression.Tests == 0 {
		fmt.Printf("   [SKIP]  No tests matched filter (0/%d baseline cases)\n", r.Regression.TotalTests)
	} else if r.Regression.Passed {
		fmt.Printf("   [PASS]  %d/%d cases passed\n", r.Regression.Tests, r.Regression.Tests)
	} else {
		fmt.Printf("   [PASS]  %d/%d cases passed\n", r.Regression.PassCount, r.Regression.Tests)
		for _, f := range r.Regression.Failures {
			fmt.Printf("   [FAIL]  %s\n", f.Name)
			fmt.Printf("           - Labels:   %s\n", formatLabels(f.Labels))
			fmt.Printf("           - Expected: %s\n", formatReceivers(f.Expected))
			fmt.Printf("           - Actual:   %s", formatReceivers(f.Actual))
			if missing := missingReceivers(f.Expected, f.Actual); len(missing) > 0 {
				fmt.Printf("  <-- Missing %s", formatMissing(missing))
			}
			fmt.Println()
		}

		if showDiff {
			fmt.Println("\n   Behavioral Delta:")
			// Generate a temporary diff for the failures
			deltas := make([]types.RegressionDelta, 0, len(r.Regression.Failures))
			for _, f := range r.Regression.Failures {
				deltas = append(deltas, types.RegressionDelta{
					Kind:     types.DeltaModified,
					Labels:   f.Labels,
					Expected: f.Expected,
					Actual:   f.Actual,
				})
			}
			PrintDiffReport(&types.RegressionDiff{Deltas: deltas})
		}
	}
	fmt.Println()

	// 3. Behavioral
	fmt.Println("3. Behavioral (BUT)")
	//nolint:gocritic
	if r.Behavioral.TotalTests == 0 {
		fmt.Println("   [SKIP]  No tests found")
	} else if r.Behavioral.Tests == 0 {
		fmt.Printf("   [SKIP]  No tests matched filter (0/%d tests)\n", r.Behavioral.TotalTests)
	} else if r.Behavioral.Passed {
		fmt.Printf("   [PASS]  %d/%d unit tests passed\n", r.Behavioral.Tests, r.Behavioral.Tests)
	} else {
		fmt.Printf("   [PASS]  %d/%d unit tests passed\n", r.Behavioral.PassCount, r.Behavioral.Tests)
		for _, f := range r.Behavioral.Failures {
			fmt.Printf("   [FAIL]  %s\n", f.Name)
			fmt.Printf("           - %s\n", f.Error)
		}
	}
	fmt.Println()

	// Footer
	fmt.Println(divider)
	fmt.Printf("SUMMARY: %s\n", formatSummary(r))
	fmt.Printf("Time: %s | Exit Code: %d\n", formatDuration(r.Duration), r.ExitCode)
}

func printSanityCategory(okMsg string, issues []string, mode string) {
	if len(issues) == 0 {
		fmt.Printf("   [OK]    %s\n", okMsg) //nolint:forbidigo
		return
	}
	isFail := strings.ToLower(mode) == "fail"
	label := "[WARN]"
	if isFail {
		label = "[FAIL]"
	}
	for _, issue := range issues {
		fmt.Printf("   %s  %s\n", label, issue) //nolint:forbidigo
	}
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, labels[k]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func formatReceivers(receivers []string) string {
	if len(receivers) == 0 {
		return "[]"
	}
	return "[" + strings.Join(receivers, ", ") + "]"
}

func missingReceivers(expected, actual []string) []string {
	actualSet := make(map[string]bool, len(actual))
	for _, r := range actual {
		actualSet[r] = true
	}
	var missing []string
	for _, r := range expected {
		if !actualSet[r] {
			missing = append(missing, r)
		}
	}
	return missing
}

func formatMissing(missing []string) string {
	quoted := make([]string, len(missing))
	for i, m := range missing {
		quoted[i] = "'" + m + "'"
	}
	return strings.Join(quoted, ", ")
}

func formatSummary(r CheckResult) string {
	if r.Passed {
		return "PASS"
	}
	var parts []string
	if n := len(r.Regression.Failures); n > 0 {
		parts = append(parts, fmt.Sprintf("%d Regression%s", n, plural(n)))
	}
	var sanityWarnings int
	for _, c := range r.Sanity.Checks {
		sanityWarnings += len(c.Issues)
	}
	if sanityWarnings > 0 {
		parts = append(parts, fmt.Sprintf("%d Sanity Warning%s", sanityWarnings, plural(sanityWarnings)))
	}
	if n := len(r.Behavioral.Failures); n > 0 {
		parts = append(parts, fmt.Sprintf("%d Behavioral Failure%s", n, plural(n)))
	}
	return "FAIL (" + strings.Join(parts, ", ") + ")"
}

// sanityCheckLabel returns the human-readable ok message for a check name.
func sanityCheckLabel(name string) string {
	labels := map[string]string{
		"shadowed_routes":      "No shadowed routes detected",
		"orphan_receivers":     "No orphan receivers",
		"inhibition_cycles":    "No inhibition cycles",
		"policy_violations":    "No policy violations",
		"dead_receivers":       "No dead receivers detected",
		"negative_only_routes": "No negative-only routes detected",
	}
	if l, ok := labels[name]; ok {
		return l
	}
	return fmt.Sprintf("No %s issues", name)
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
