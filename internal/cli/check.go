package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/types"
	"github.com/nyambati/litmus/internal/workspace"
	amconfig "github.com/prometheus/alertmanager/config"
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

// CheckOptions bundles the flags for a check run.
type CheckOptions struct {
	Format string
	Diff   bool
	Tags   []string
}

// RunCheck loads config, runs all validation stages, prints results, and returns
// the exit code the CLI layer should pass to os.Exit (0 = all passed).
func RunCheck(ctx context.Context, opts *CheckOptions) (CheckExitCode, error) {
	cfg := config.ConfigFromContext(ctx)
	logger := config.LoggerFromContext(ctx)
	start := time.Now()
	ws, err := workspace.Load(cfg, logger)
	if err != nil {
		return 1, err
	}

	amCfg, err := ws.Config()
	if err != nil {
		return 1, fmt.Errorf("failed to load alertmanager config: %w", err)
	}

	router := pipeline.NewRouter(amCfg.Route)

	sanityResult := sanity.Run(buildCheckContext(amCfg, ws, cfg.Policy), cfg.Sanity)
	regressionResult := RunRegressionTests(ctx, ws, router, opts.Tags)
	behavioralResult := RunBehavioralTests(ctx, ws.Tests(), router, amCfg.InhibitRules, opts.Tags)

	result := buildCheckResult(cfg.FilePath(), sanityResult, regressionResult, behavioralResult, time.Since(start))

	code := calculateExitCode(result)
	result.ExitCode = code

	if err := outputResults(result, opts.Format, opts.Diff); err != nil {
		return 1, err
	}

	return code, nil
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
	switch format {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Fprintln(os.Stdout, string(data))
	case "text":
		PrintCheckResult(result, showDiff)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	return nil
}

// RunRegressionTests executes the regression baseline against the current router.
func RunRegressionTests(ctx context.Context, ws *workspace.Workspace, router *pipeline.Router, tags []string) RegressionResult {
	result := RegressionResult{Passed: true}

	if ws.RegressionState == nil {
		return result
	}

	baseline := ws.RegressionState.Tests
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
func RunBehavioralTests(ctx context.Context, tests []*types.TestCase, router *pipeline.Router, inhibitRules []amconfig.InhibitRule, tags []string) BehavioralResult {
	result := BehavioralResult{Passed: true}

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
				Name:     res.Name,
				Type:     res.Type,
				Error:    res.Error,
				Labels:   res.Labels,
				Expected: res.Expected,
				Actual:   res.Actual,
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

// PrintCheckResult writes the formatted validation report to stdout. Color is
// applied automatically when stdout is a terminal (see styler).
func PrintCheckResult(r CheckResult, showDiff bool) {
	st := newStyler(os.Stdout)

	fmt.Fprintln(os.Stdout, st.bold("Litmus Check: "+r.ConfigPath))
	fmt.Fprintln(os.Stdout, divider)
	fmt.Fprintln(os.Stdout)

	// 1. Sanity
	fmt.Fprintln(os.Stdout, st.bold("1. Sanity (Static Analysis)"))
	for _, c := range r.Sanity.Checks {
		printSanityCategory(st, sanityCheckLabel(c.Name), c.Issues, c.Mode)
	}
	fmt.Fprintln(os.Stdout)

	// 2. Regressions
	fmt.Fprintln(os.Stdout, st.bold("2. Regressions (Automated)"))
	//nolint:gocritic
	if r.Regression.TotalTests == 0 {
		fmt.Fprintf(os.Stdout, "   %s  No baseline found — run 'litmus snapshot capture' first\n", st.status("SKIP"))
	} else if r.Regression.Tests == 0 {
		fmt.Fprintf(os.Stdout, "   %s  No tests matched filter (0/%d baseline cases)\n", st.status("SKIP"), r.Regression.TotalTests)
	} else if r.Regression.Passed {
		fmt.Fprintf(os.Stdout, "   %s  %d/%d cases passed\n", st.status("PASS"), r.Regression.Tests, r.Regression.Tests)
	} else {
		fmt.Fprintf(os.Stdout, "   %s  %d/%d cases passed\n", st.status("FAIL"), r.Regression.PassCount, r.Regression.Tests)
		for _, f := range r.Regression.Failures {
			printFailure(st, f)
		}

		if showDiff {
			fmt.Fprintln(os.Stdout, "\n   Behavioral Delta:")
			// Generate a temporary diff for the failures
			deltas := make([]snapshot.RegressionDelta, 0, len(r.Regression.Failures))
			for _, f := range r.Regression.Failures {
				deltas = append(deltas, snapshot.RegressionDelta{
					Kind:     snapshot.DeltaModified,
					Labels:   f.Labels,
					Expected: f.Expected,
					Actual:   f.Actual,
				})
			}
			PrintDiffReport(&snapshot.RegressionDiff{Deltas: deltas})
		}
	}
	fmt.Fprintln(os.Stdout)

	// 3. Behavioral
	fmt.Fprintln(os.Stdout, st.bold("3. Behavioral (Unit Tests)"))
	//nolint:gocritic
	if r.Behavioral.TotalTests == 0 {
		fmt.Fprintf(os.Stdout, "   %s  No tests found\n", st.status("SKIP"))
	} else if r.Behavioral.Tests == 0 {
		fmt.Fprintf(os.Stdout, "   %s  No tests matched filter (0/%d tests)\n", st.status("SKIP"), r.Behavioral.TotalTests)
	} else if r.Behavioral.Passed {
		fmt.Fprintf(os.Stdout, "   %s  %d/%d unit tests passed\n", st.status("PASS"), r.Behavioral.Tests, r.Behavioral.Tests)
	} else {
		fmt.Fprintf(os.Stdout, "   %s  %d/%d unit tests passed\n", st.status("FAIL"), r.Behavioral.PassCount, r.Behavioral.Tests)
		for _, f := range r.Behavioral.Failures {
			printFailure(st, f)
		}
	}
	fmt.Fprintln(os.Stdout)

	// Footer
	fmt.Fprintln(os.Stdout, divider)
	fmt.Fprintf(os.Stdout, "SUMMARY: %s\n", formatSummary(st, r))
	fmt.Fprintf(os.Stdout, "Time: %s | Exit Code: %d\n", formatDuration(r.Duration), r.ExitCode)
}

// printFailure renders a single test failure in a uniform layout shared by both
// regression and behavioral results. Receiver-level failures show an aligned
// Labels/Expected/Actual breakdown with a Missing hint; other failures fall
// back to the raw error reason.
func printFailure(st *styler, f TestFailure) {
	fmt.Fprintf(os.Stdout, "   %s  %s\n", st.status("FAIL"), f.Name)
	if len(f.Labels) > 0 {
		fmt.Fprintf(os.Stdout, "           - Labels:   %s\n", formatLabels(f.Labels))
	}
	if len(f.Expected) > 0 || len(f.Actual) > 0 {
		fmt.Fprintf(os.Stdout, "           - Expected: %s\n", formatReceivers(f.Expected))
		fmt.Fprintf(os.Stdout, "           - Actual:   %s", formatReceivers(f.Actual))
		if missing := missingReceivers(f.Expected, f.Actual); len(missing) > 0 {
			fmt.Fprintf(os.Stdout, "  %s", st.red("<-- Missing "+formatMissing(missing)))
		}
		fmt.Fprintln(os.Stdout)
		return
	}
	if f.Error != "" {
		fmt.Fprintf(os.Stdout, "           - %s\n", f.Error)
	}
}

func printSanityCategory(st *styler, okMsg string, issues []string, mode string) {
	if len(issues) == 0 {
		fmt.Fprintf(os.Stdout, "   %s    %s\n", st.status("OK"), okMsg)
		return
	}
	kind := "WARN"
	if strings.ToLower(mode) == "fail" {
		kind = "FAIL"
	}
	for _, issue := range issues {
		fmt.Fprintf(os.Stdout, "   %s  %s\n", st.status(kind), issue)
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

func formatSummary(st *styler, r CheckResult) string {
	if r.Passed {
		return st.green("PASS")
	}
	var parts []string
	if n := len(r.Regression.Failures); n > 0 {
		parts = append(parts, fmt.Sprintf("%d Regression%s", n, plural(n)))
	}
	var sanityFailures, sanityWarnings int
	for _, c := range r.Sanity.Checks {
		if strings.ToLower(c.Mode) == "fail" {
			sanityFailures += len(c.Issues)
		} else {
			sanityWarnings += len(c.Issues)
		}
	}
	if sanityFailures > 0 {
		parts = append(parts, fmt.Sprintf("%d Sanity Failure%s", sanityFailures, plural(sanityFailures)))
	}
	if sanityWarnings > 0 {
		parts = append(parts, fmt.Sprintf("%d Sanity Warning%s", sanityWarnings, plural(sanityWarnings)))
	}
	if n := len(r.Behavioral.Failures); n > 0 {
		parts = append(parts, fmt.Sprintf("%d Behavioral Failure%s", n, plural(n)))
	}
	return st.red("FAIL") + " (" + strings.Join(parts, ", ") + ")"
}

// sanityCheckLabel returns the human-readable ok message for a check name.
func sanityCheckLabel(name config.SanityCheck) string {
	labels := map[config.SanityCheck]string{
		config.CheckShadowedRoutes:     "No shadowed routes detected",
		config.CheckOrphanReceivers:    "No orphan receivers",
		config.CheckInhibitionCycles:   "No inhibition cycles",
		config.CheckPolicyViolations:   "No policy violations",
		config.CheckDeadRoutes:         "No dead routes detected",
		config.CheckNegativeOnlyRoutes: "No negative-only routes detected",
		config.CheckRequireTests:       "All fragments have required tests",
		config.CheckRequireRegression:  "Regression baseline coverage satisfied",
	}
	if l, ok := labels[name]; ok {
		return l
	}
	return fmt.Sprintf("No %s issues", name)
}

// formatDuration renders a run duration at a sensible resolution: microseconds
// below 1ms, milliseconds below 1s, seconds otherwise. Avoids the unhelpful
// "0.0s" for sub-second runs.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
