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
	"github.com/nyambati/litmus/internal/engine/behavioral"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
)

const divider = "--------------------------------------------------"

// CheckExitCode signals which exit code the caller should use.
type CheckExitCode int

// CheckResult holds results from a full validation run.
type CheckResult struct {
	Passed     bool             `json:"passed"`
	ConfigPath string           `json:"config_path"`
	Sanity     SanityResult     `json:"sanity"`
	Regression RegressionResult `json:"regression"`
	Behavioral BehavioralResult `json:"behavioral"`
	Duration   time.Duration    `json:"duration_ns"`
	ExitCode   CheckExitCode    `json:"exit_code"`
}

// SanityResult holds per-category static analysis results.
type SanityResult struct {
	Passed             bool     `json:"passed"`
	ShadowedIssues     []string `json:"shadowed_issues,omitempty"`
	OrphanIssues       []string `json:"orphan_issues,omitempty"`
	InhibitionIssues   []string `json:"inhibition_issues,omitempty"`
	PolicyIssues       []string `json:"policy_issues,omitempty"`
	DeadReceiverIssues []string `json:"dead_receiver_issues,omitempty"`
	ShadowedMode       string   `json:"shadowed_mode,omitempty"`
	OrphanMode         string   `json:"orphan_mode,omitempty"`
	InhibitionMode     string   `json:"inhibition_mode,omitempty"`
	PolicyMode         string   `json:"policy_mode,omitempty"`
	DeadReceiverMode   string   `json:"dead_receiver_mode,omitempty"`
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
func RunCheck(format string, showDiff bool, tags []string) (CheckExitCode, error) {
	start := time.Now()

	litmusConfig, router, fragments, amCfg, err := loadConfigAndRouter()
	if err != nil {
		return 1, err
	}

	sanityResult := runSanityChecks(litmusConfig, fragments, amCfg)
	regressionResult := runRegressionTests(litmusConfig, router, tags)
	behavioralResult := runBehavioralTests(amCfg.InhibitRules, litmusConfig, fragments, router, tags)

	result := buildCheckResult(litmusConfig.FilePath(), sanityResult, regressionResult, behavioralResult, time.Since(start))

	code := calculateExitCode(result)
	result.ExitCode = code

	if err := outputResults(result, format, showDiff); err != nil {
		return 1, err
	}

	return code, nil
}

// loadConfigAndRouter loads the litmus config, assembles alertmanager config, and creates a router.
func loadConfigAndRouter() (*config.LitmusConfig, *pipeline.Router, []*config.Fragment, *amconfig.Config, error) {
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("loading litmus config: %w", err)
	}

	_, fragments, amCfg, err := litmusConfig.LoadAssembledConfig()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("loading assembled alertmanager config: %w", err)
	}

	if amCfg.Route == nil {
		return nil, nil, nil, nil, fmt.Errorf("alertmanager config has no route defined")
	}

	router := pipeline.NewRouter(amCfg.Route)
	return litmusConfig, router, fragments, amCfg, nil
}

// runSanityChecks executes all sanity checks including policy enforcement.
func runSanityChecks(litmusConfig *config.LitmusConfig, fragments []*config.Fragment, amCfg *amconfig.Config) SanityResult {
	sanityResult := RunSanityChecks(amCfg, litmusConfig.Sanity)
	checker := sanity.NewPolicyChecker(litmusConfig.Policy)
	sanityResult.PolicyIssues = checker.Check(fragments)

	policyMode := string(litmusConfig.Sanity.PolicyViolations)
	sanityResult.PolicyMode = policyMode
	if litmusConfig.Sanity.PolicyViolations.IsFail() && len(sanityResult.PolicyIssues) > 0 {
		sanityResult.Passed = false
	}

	return sanityResult
}

// runRegressionTests executes regression tests against the router.
func runRegressionTests(litmusConfig *config.LitmusConfig, router *pipeline.Router, tags []string) RegressionResult {
	ctx := context.Background()
	return RunRegressionTests(ctx, litmusConfig, router, tags)
}

// runBehavioralTests executes behavioral tests against the router and inhibit rules.
func runBehavioralTests(inhibitRules []amconfig.InhibitRule, litmusConfig *config.LitmusConfig, fragments []*config.Fragment, router *pipeline.Router, tags []string) BehavioralResult {
	ctx := context.Background()
	return RunBehavioralTests(ctx, litmusConfig, fragments, router, inhibitRules, tags)
}

// buildCheckResult assembles the final check result from all test results.
func buildCheckResult(configPath string, sanityResult SanityResult, regressionResult RegressionResult, behavioralResult BehavioralResult, duration time.Duration) CheckResult {
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

// RunSanityChecks runs all static analysis linters, returning per-category results.
func RunSanityChecks(alertConfig *amconfig.Config, sanityConfig config.SanityConfig) SanityResult {
	result := SanityResult{Passed: true}

	shadowed := sanity.NewShadowedRouteDetector(alertConfig.Route)
	result.ShadowedIssues = shadowed.Detect()
	result.ShadowedMode = string(sanityConfig.ShadowedRoutes)

	receiversMap := make(map[string]*amconfig.Receiver)
	for i := range alertConfig.Receivers {
		receiversMap[alertConfig.Receivers[i].Name] = &alertConfig.Receivers[i]
	}
	orphan := sanity.NewOrphanReceiverDetector(alertConfig.Route, receiversMap)
	result.OrphanIssues = orphan.DetectOrphans()
	result.OrphanMode = string(sanityConfig.OrphanReceivers)

	rules := make([]*amconfig.InhibitRule, 0, len(alertConfig.InhibitRules))
	for i := range alertConfig.InhibitRules {
		rules = append(rules, &alertConfig.InhibitRules[i])
	}
	inhibition := sanity.NewInhibitionCycleDetector(rules)
	result.InhibitionIssues = inhibition.DetectCycles()
	result.InhibitionMode = string(sanityConfig.InhibitionCycles)

	dead := sanity.NewDeadReceiverDetector(alertConfig.Route)
	result.DeadReceiverIssues = dead.Detect()
	result.DeadReceiverMode = string(sanityConfig.DeadReceivers)

	if sanityConfig.ShadowedRoutes.IsFail() && len(result.ShadowedIssues) > 0 {
		result.Passed = false
	}
	if sanityConfig.OrphanReceivers.IsFail() && len(result.OrphanIssues) > 0 {
		result.Passed = false
	}
	if sanityConfig.InhibitionCycles.IsFail() && len(result.InhibitionIssues) > 0 {
		result.Passed = false
	}
	if sanityConfig.DeadReceivers.IsFail() && len(result.DeadReceiverIssues) > 0 {
		result.Passed = false
	}

	return result
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
	executor := snapshot.NewRegressionTestExecutor()

	for _, res := range executor.Execute(ctx, baseline, router) {
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
func RunBehavioralTests(ctx context.Context, litmusConfig *config.LitmusConfig, fragments []*config.Fragment, router *pipeline.Router, inhibitRules []amconfig.InhibitRule, tags []string) BehavioralResult {
	result := BehavioralResult{Passed: true}

	loader := behavioral.NewBehavioralTestLoader()
	tests, err := loader.LoadFromDirectory(litmusConfig.TestsDir())
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "WARN: could not load behavioral tests from root: %v\n", err)
		}
	}

	for _, frag := range fragments {
		tests = append(tests, frag.Tests...)
	}

	if len(tests) == 0 {
		return result
	}

	result.TotalTests = len(tests)
	tests = filterByTags(tests, tags)
	result.Tests = len(tests)
	executor := behavioral.NewBehavioralTestExecutor(inhibitRules)

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
	printSanityCategory("No shadowed routes detected", r.Sanity.ShadowedIssues, r.Sanity.ShadowedMode)
	printSanityCategory("No orphan receivers", r.Sanity.OrphanIssues, r.Sanity.OrphanMode)
	printSanityCategory("No inhibition cycles", r.Sanity.InhibitionIssues, r.Sanity.InhibitionMode)
	printSanityCategory("No policy violations", r.Sanity.PolicyIssues, r.Sanity.PolicyMode)
	printSanityCategory("No dead receivers detected", r.Sanity.DeadReceiverIssues, r.Sanity.DeadReceiverMode)
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
	sanityWarnings := len(r.Sanity.ShadowedIssues) +
		len(r.Sanity.OrphanIssues) + len(r.Sanity.InhibitionIssues) +
		len(r.Sanity.PolicyIssues) + len(r.Sanity.DeadReceiverIssues)
	if sanityWarnings > 0 {
		parts = append(parts, fmt.Sprintf("%d Sanity Warning%s", sanityWarnings, plural(sanityWarnings)))
	}
	if n := len(r.Behavioral.Failures); n > 0 {
		parts = append(parts, fmt.Sprintf("%d Behavioral Failure%s", n, plural(n)))
	}
	return "FAIL (" + strings.Join(parts, ", ") + ")"
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
