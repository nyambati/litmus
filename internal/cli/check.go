package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/behavioral"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	amconfig "github.com/prometheus/alertmanager/config"
)

// CheckResult holds results from a validation run.
type CheckResult struct {
	Passed     bool             `json:"passed"`
	Sanity     SanityResult     `json:"sanity"`
	Regression RegressionResult `json:"regression"`
	Behavioral BehavioralResult `json:"behavioral"`
	Summary    string           `json:"summary"`
}

// SanityResult holds sanity linter results.
type SanityResult struct {
	Passed bool     `json:"passed"`
	Issues []string `json:"issues"`
}

// RegressionResult holds regression test results.
type RegressionResult struct {
	Passed   bool     `json:"passed"`
	Tests    int      `json:"tests"`
	Failures []string `json:"failures,omitempty"`
}

// BehavioralResult holds behavioral test results.
type BehavioralResult struct {
	Passed   bool     `json:"passed"`
	Tests    int      `json:"tests"`
	Failures []string `json:"failures,omitempty"`
}

// CheckExitCode signals which exit code the caller should use.
// Non-zero only when validation fails.
type CheckExitCode int

// RunCheck loads config, runs all validation stages, prints results, and returns
// the exit code the CLI layer should pass to os.Exit (0 = all passed).
func RunCheck(format string) (CheckExitCode, error) {
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

	router := pipeline.NewRouter(alertConfig.Route)

	sanityResult := RunSanityChecks(alertConfig)
	regressionResult := RunRegressionTests(ctx, litmusConfig, router)
	behavioralResult := RunBehavioralTests(ctx, litmusConfig, router, alertConfig.InhibitRules)

	result := CheckResult{
		Passed:     sanityResult.Passed && regressionResult.Passed && behavioralResult.Passed,
		Sanity:     sanityResult,
		Regression: regressionResult,
		Behavioral: behavioralResult,
	}

	if format == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return 1, fmt.Errorf("marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		PrintCheckResult(result)
	}

	if !result.Passed {
		if !sanityResult.Passed {
			return 3, nil
		}
		return 2, nil
	}

	return 0, nil
}

// RunSanityChecks runs all static analysis linters against the alertmanager config.
func RunSanityChecks(alertConfig *amconfig.Config) SanityResult {
	result := SanityResult{Passed: true}

	shadowed := sanity.NewShadowedRouteDetector(alertConfig.Route)
	result.Issues = append(result.Issues, shadowed.Detect()...)

	receiversMap := make(map[string]*amconfig.Receiver)
	for i := range alertConfig.Receivers {
		receiversMap[alertConfig.Receivers[i].Name] = &alertConfig.Receivers[i]
	}
	orphan := sanity.NewOrphanReceiverDetector(alertConfig.Route, receiversMap)
	result.Issues = append(result.Issues, orphan.DetectOrphans()...)

	var rules []*amconfig.InhibitRule
	for i := range alertConfig.InhibitRules {
		rules = append(rules, &alertConfig.InhibitRules[i])
	}
	inhibition := sanity.NewInhibitionCycleDetector(rules)
	result.Issues = append(result.Issues, inhibition.DetectCycles()...)

	if len(result.Issues) > 0 {
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
		result.Tests = 0
		return result
	}

	result.Tests = len(baseline)
	executor := snapshot.NewRegressionTestExecutor()
	execResults := executor.Execute(ctx, baseline, router)

	for _, res := range execResults {
		if !res.Pass {
			result.Passed = false
			result.Failures = append(result.Failures, fmt.Sprintf("%s: %s", res.Name, res.Error))
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
		result.Tests = 0
		return result
	}

	result.Tests = len(tests)
	executor := behavioral.NewBehavioralTestExecutor(inhibitRules)

	for _, test := range tests {
		res := executor.Execute(ctx, test, router)
		if !res.Pass {
			result.Passed = false
			result.Failures = append(result.Failures, fmt.Sprintf("%s: %s", res.Name, res.Error))
		}
	}

	return result
}

// PrintCheckResult writes a human-readable validation report to stdout.
func PrintCheckResult(result CheckResult) {
	fmt.Println("=== Litmus Validation Report ===")

	fmt.Println("Sanity Checks:")
	if result.Sanity.Passed {
		fmt.Println("  ✓ Passed")
	} else {
		fmt.Println("  ✗ Failed")
		for _, issue := range result.Sanity.Issues {
			fmt.Printf("    - %s\n", issue)
		}
	}

	fmt.Println("\nRegression Tests:")
	if result.Regression.Passed {
		fmt.Printf("  ✓ %d tests passed\n", result.Regression.Tests)
	} else {
		fmt.Printf("  ✗ %d test failures\n", len(result.Regression.Failures))
		for _, fail := range result.Regression.Failures {
			fmt.Printf("    - %s\n", fail)
		}
	}

	fmt.Println("\nBehavioral Tests:")
	if result.Behavioral.Passed {
		fmt.Printf("  ✓ %d tests passed\n", result.Behavioral.Tests)
	} else {
		fmt.Printf("  ✗ %d test failures\n", len(result.Behavioral.Failures))
		for _, fail := range result.Behavioral.Failures {
			fmt.Printf("    - %s\n", fail)
		}
	}

	fmt.Println()
	if result.Summary != "" {
		fmt.Println(result.Summary)
	}
	if result.Passed {
		fmt.Println("✓ All checks passed")
	} else {
		fmt.Println("✗ Some checks failed")
	}
}
