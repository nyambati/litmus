package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	Passed           bool     `json:"passed"`
	ShadowedIssues   []string `json:"shadowed_issues,omitempty"`
	OrphanIssues     []string `json:"orphan_issues,omitempty"`
	InhibitionIssues []string `json:"inhibition_issues,omitempty"`
}

// RegressionFailure holds structured detail for a single regression failure.
type RegressionFailure struct {
	Name     string            `json:"name"`
	Labels   map[string]string `json:"labels"`
	Expected []string          `json:"expected"`
	Actual   []string          `json:"actual"`
}

// RegressionResult holds regression test results.
type RegressionResult struct {
	Passed    bool                `json:"passed"`
	Tests     int                 `json:"tests"`
	PassCount int                 `json:"pass_count"`
	Failures  []RegressionFailure `json:"failures,omitempty"`
}

// BehavioralFailure holds structured detail for a single behavioral test failure.
type BehavioralFailure struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

// BehavioralResult holds behavioral test results.
type BehavioralResult struct {
	Passed    bool                `json:"passed"`
	Tests     int                 `json:"tests"`
	PassCount int                 `json:"pass_count"`
	Failures  []BehavioralFailure `json:"failures,omitempty"`
}

// RunCheck loads config, runs all validation stages, prints results, and returns
// the exit code the CLI layer should pass to os.Exit (0 = all passed).
func RunCheck(format string, showDiff bool) (CheckExitCode, error) {
	start := time.Now()
	ctx := context.Background()

	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return 1, fmt.Errorf("loading litmus config: %w", err)
	}

	alertConfigPath := filepath.Join(litmusConfig.Config.Directory, litmusConfig.Config.File)
	alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
	if err != nil {
		return 1, fmt.Errorf("loading alertmanager config: %w", err)
	}

	if alertConfig.Route == nil {
		return 1, fmt.Errorf("alertmanager config has no route defined")
	}

	router := pipeline.NewRouter(alertConfig.Route)

	sanityResult := RunSanityChecks(alertConfig)
	regressionResult := RunRegressionTests(ctx, litmusConfig, router)
	behavioralResult := RunBehavioralTests(ctx, litmusConfig, router, alertConfig.InhibitRules)

	passed := sanityResult.Passed && regressionResult.Passed && behavioralResult.Passed

	var code CheckExitCode
	if !passed {
		if !sanityResult.Passed {
			code = 3
		} else {
			code = 2
		}
	}

	result := CheckResult{
		Passed:     passed,
		ConfigPath: alertConfigPath,
		Sanity:     sanityResult,
		Regression: regressionResult,
		Behavioral: behavioralResult,
		Duration:   time.Since(start),
		ExitCode:   code,
	}

	if format == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return 1, fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		PrintCheckResult(result, showDiff)
	}

	return code, nil
}

// RunSanityChecks runs all static analysis linters, returning per-category results.
func RunSanityChecks(alertConfig *amconfig.Config) SanityResult {
	result := SanityResult{Passed: true}

	shadowed := sanity.NewShadowedRouteDetector(alertConfig.Route)
	result.ShadowedIssues = shadowed.Detect()

	receiversMap := make(map[string]*amconfig.Receiver)
	for i := range alertConfig.Receivers {
		receiversMap[alertConfig.Receivers[i].Name] = &alertConfig.Receivers[i]
	}
	orphan := sanity.NewOrphanReceiverDetector(alertConfig.Route, receiversMap)
	result.OrphanIssues = orphan.DetectOrphans()

	var rules []*amconfig.InhibitRule
	for i := range alertConfig.InhibitRules {
		rules = append(rules, &alertConfig.InhibitRules[i])
	}
	inhibition := sanity.NewInhibitionCycleDetector(rules)
	result.InhibitionIssues = inhibition.DetectCycles()

	if len(result.ShadowedIssues)+len(result.OrphanIssues)+len(result.InhibitionIssues) > 0 {
		result.Passed = false
	}

	return result
}

// RunRegressionTests executes the regression baseline against the current router.
func RunRegressionTests(ctx context.Context, litmusConfig *config.LitmusConfig, router *pipeline.Router) RegressionResult {
	result := RegressionResult{Passed: true}

	baselinePath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.mpk")
	baseline, err := LoadBaseline(baselinePath)
	if err != nil {
		return result
	}

	result.Tests = len(baseline)
	executor := snapshot.NewRegressionTestExecutor()

	for _, res := range executor.Execute(ctx, baseline, router) {
		if res.Pass {
			result.PassCount++
		} else {
			result.Passed = false
			result.Failures = append(result.Failures, RegressionFailure{
				Name:     res.Name,
				Labels:   res.Labels,
				Expected: res.Expected,
				Actual:   res.Actual,
			})
		}
	}

	return result
}

// RunBehavioralTests loads and executes all behavioral unit tests.
func RunBehavioralTests(ctx context.Context, litmusConfig *config.LitmusConfig, router *pipeline.Router, inhibitRules []amconfig.InhibitRule) BehavioralResult {
	result := BehavioralResult{Passed: true}

	loader := behavioral.NewBehavioralTestLoader()
	tests, err := loader.LoadFromDirectory(litmusConfig.Tests.Directory)
	if err != nil {
		return result
	}

	result.Tests = len(tests)
	executor := behavioral.NewBehavioralTestExecutor(inhibitRules)

	for _, test := range tests {
		res := executor.Execute(ctx, test, router)
		if res.Pass {
			result.PassCount++
		} else {
			result.Passed = false
			result.Failures = append(result.Failures, BehavioralFailure{
				Name:  res.Name,
				Error: res.Error,
			})
		}
	}

	return result
}

// PrintCheckResult writes the formatted validation report to stdout.
func PrintCheckResult(r CheckResult, showDiff bool) {
	fmt.Printf("Litmus Check: %s\n", r.ConfigPath)
	fmt.Println(divider)
	fmt.Println()

	// 1. Sanity
	fmt.Println("1. Sanity (Static Analysis)")
	printSanityCategory("No shadowed routes detected", r.Sanity.ShadowedIssues)
	printSanityCategory("No orphan receivers", r.Sanity.OrphanIssues)
	printSanityCategory("No inhibition cycles", r.Sanity.InhibitionIssues)
	fmt.Println()

	// 2. Regressions
	fmt.Println("2. Regressions (Automated)")
	if r.Regression.Tests == 0 {
		fmt.Println("   [SKIP]  No baseline found — run 'litmus snapshot' first")
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
			var deltas []types.RegressionDelta
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
	if r.Behavioral.Tests == 0 {
		fmt.Println("   [SKIP]  No tests found")
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

func printSanityCategory(okMsg string, issues []string) {
	if len(issues) == 0 {
		fmt.Printf("   [OK]    %s\n", okMsg)
		return
	}
	for _, issue := range issues {
		fmt.Printf("   [WARN]  %s\n", issue)
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
	sanityWarnings := len(r.Sanity.ShadowedIssues) + len(r.Sanity.OrphanIssues) + len(r.Sanity.InhibitionIssues)
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
