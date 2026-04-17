package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/behavioral"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/spf13/cobra"
)

// CheckResult holds results from validation run.
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

// newCheckCmd creates the check command.
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "check",
		Short:        "Validate alertmanager configuration",
		Long:         "Runs sanity linter, regression tests, and behavioral unit tests",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			return runCheck(format)
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format: text or json")
	return cmd
}

func runCheck(format string) error {
	ctx := context.Background()

	// Load configs
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	alertConfigPath := filepath.Join(litmusConfig.Config.Directory, litmusConfig.Config.File)
	alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
	if err != nil {
		return fmt.Errorf("loading alertmanager config: %w", err)
	}

	// Initialize Router
	router := pipeline.NewRouter(alertConfig.Route)

	// Run sanity checks
	sanityResult := runSanityChecks(alertConfig)

	// Run regression tests
	regressionResult := runRegressionTests(ctx, litmusConfig, router)

	// Run behavioral tests
	behavioralResult := runBehavioralTests(ctx, litmusConfig, router)

	// Compile results
	result := CheckResult{
		Passed:     sanityResult.Passed && regressionResult.Passed && behavioralResult.Passed,
		Sanity:     sanityResult,
		Regression: regressionResult,
		Behavioral: behavioralResult,
	}

	// Format output
	if format == "json" {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		printTextOutput(result)
	}

	// Exit with appropriate code
	if !result.Passed {
		if !sanityResult.Passed {
			os.Exit(3)
		}
		os.Exit(2)
	}

	return nil
}

func runSanityChecks(alertConfig *amconfig.Config) SanityResult {
	result := SanityResult{Passed: true}

	// Shadowed routes
	shadowed := sanity.NewShadowedRouteDetector(alertConfig.Route)
	result.Issues = append(result.Issues, shadowed.Detect()...)

	// Orphan receivers - convert slice to map
	receiversMap := make(map[string]*amconfig.Receiver)
	for i := range alertConfig.Receivers {
		receiversMap[alertConfig.Receivers[i].Name] = &alertConfig.Receivers[i]
	}
	orphan := sanity.NewOrphanReceiverDetector(alertConfig.Route, receiversMap)
	result.Issues = append(result.Issues, orphan.DetectOrphans()...)

	// Inhibition cycles - convert slice to pointers
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

func runRegressionTests(ctx context.Context, litmusConfig *config.LitmusConfig, router *pipeline.Router) RegressionResult {
	result := RegressionResult{Passed: true}

	baselinePath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.mpk")
	// For now, just report no tests run if baseline doesn't exist
	baseline, err := loadBaseline(baselinePath)
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

func runBehavioralTests(ctx context.Context, litmusConfig *config.LitmusConfig, router *pipeline.Router) BehavioralResult {
	result := BehavioralResult{Passed: true}

	// Load test files
	loader := behavioral.NewBehavioralTestLoader()
	tests, err := loader.LoadFromDirectory(litmusConfig.Tests.Directory)
	if err != nil {
		result.Tests = 0
		return result
	}

	result.Tests = len(tests)
	executor := behavioral.NewBehavioralTestExecutor()

	for _, test := range tests {
		res := executor.Execute(ctx, test, router)
		if !res.Pass {
			result.Passed = false
			result.Failures = append(result.Failures, fmt.Sprintf("%s: %s", res.Name, res.Error))
		}
	}

	return result
}

func printTextOutput(result CheckResult) {
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
