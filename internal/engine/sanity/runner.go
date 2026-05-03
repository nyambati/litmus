package sanity

import (
	"github.com/nyambati/litmus/internal/config"
)

// CheckResult holds the outcome of a single sanity check run.
type CheckResult struct {
	Name   string
	Issues []string
	Mode   config.SanityMode
}

// CheckEntry holds the serialisable result for a single check.
type CheckEntry struct {
	Name   string   `json:"name"`
	Mode   string   `json:"mode"`
	Issues []string `json:"issues,omitempty"`
}

// Result holds the aggregated results of all static analysis checks.
type Result struct {
	Passed bool         `json:"passed"`
	Checks []CheckEntry `json:"checks,omitempty"`
}

// Run executes all registered checks, aggregates pass/fail, and returns a Result.
func Run(ctx CheckContext, cfg config.SanityConfig) Result {
	runner := DefaultRunner(cfg)
	out := Result{Passed: true}
	for _, r := range runner.Run(ctx) {
		out.Checks = append(out.Checks, CheckEntry{
			Name:   r.Name,
			Mode:   string(r.Mode),
			Issues: r.Issues,
		})
		if r.Mode.IsFail() && len(r.Issues) > 0 {
			out.Passed = false
		}
	}
	return out
}

// Runner executes a set of registered checks and looks up their mode from config.
type Runner struct {
	checks []Check
	modeFn func(string) config.SanityMode
}

// NewRunner creates a runner with the given mode-lookup function and checks.
func NewRunner(modeFn func(string) config.SanityMode, checks ...Check) *Runner {
	return &Runner{checks: checks, modeFn: modeFn}
}

// Run executes all registered checks against ctx and returns one result per check.
func (r *Runner) Run(ctx CheckContext) []CheckResult {
	results := make([]CheckResult, 0, len(r.checks))
	for _, c := range r.checks {
		results = append(results, CheckResult{
			Name:   c.Name(),
			Issues: c.Run(ctx),
			Mode:   r.modeFn(c.Name()),
		})
	}
	return results
}

// DefaultRunner wires up all built-in sanity checks with modes from cfg.
func DefaultRunner(cfg config.SanityConfig) *Runner {
	return NewRunner(cfg.ModeFor,
		&ShadowedRouteDetector{},
		&OrphanReceiverDetector{},
		&InhibitionCycleDetector{},
		&DeadReceiverDetector{},
		&NegativeOnlyRouteDetector{},
		&PolicyChecker{},
	)
}
