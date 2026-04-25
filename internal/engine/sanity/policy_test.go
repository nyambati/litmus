package sanity

import (
	"testing"

	litconfig "github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/assert"
)

func TestPolicyChecker_RequireTests(t *testing.T) {
	tests := []struct {
		name       string
		fragments  []*litconfig.Fragment
		wantIssues int
	}{
		{
			name: "fragment with tests — no violation",
			fragments: []*litconfig.Fragment{
				{Name: "db", Tests: []*types.TestCase{{Name: "t1"}}},
			},
			wantIssues: 0,
		},
		{
			name: "fragment with no tests — violation",
			fragments: []*litconfig.Fragment{
				{Name: "db"},
			},
			wantIssues: 1,
		},
		{
			name: "root fragment checked — no tests is a violation",
			fragments: []*litconfig.Fragment{
				{Name: "root"},
			},
			wantIssues: 1,
		},
		{
			name: "multiple fragments — only violators reported",
			fragments: []*litconfig.Fragment{
				{Name: "db"},
				{Name: "net", Tests: []*types.TestCase{{Name: "t1"}}},
				{Name: "api"},
			},
			wantIssues: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewPolicyChecker(litconfig.PolicyConfig{RequireTests: true})
			issues := checker.Check(tt.fragments)
			assert.Len(t, issues, tt.wantIssues)
		})
	}
}

func TestPolicyChecker_EnforceMatchers(t *testing.T) {
	tests := []struct {
		name       string
		policy     litconfig.PolicyConfig
		fragments  []*litconfig.Fragment
		wantIssues int
	}{
		{
			name:   "route has required label in Match — no violation",
			policy: litconfig.PolicyConfig{EnforceMatchers: []string{"team"}},
			fragments: []*litconfig.Fragment{{
				Name:   "db",
				Routes: []*amconfig.Route{{Receiver: "slack", Match: map[string]string{"team": "db"}}},
			}},
			wantIssues: 0,
		},
		{
			name:   "route missing required label — violation",
			policy: litconfig.PolicyConfig{EnforceMatchers: []string{"team"}},
			fragments: []*litconfig.Fragment{{
				Name:   "db",
				Routes: []*amconfig.Route{{Receiver: "slack", Match: map[string]string{"severity": "critical"}}},
			}},
			wantIssues: 1,
		},
		{
			name:   "OR semantics — any required label satisfies policy",
			policy: litconfig.PolicyConfig{EnforceMatchers: []string{"team", "service"}},
			fragments: []*litconfig.Fragment{{
				Name:   "db",
				Routes: []*amconfig.Route{{Receiver: "slack", Match: map[string]string{"service": "mysql"}}},
			}},
			wantIssues: 0,
		},
		{
			name:   "route has required label in modern Matchers — no violation",
			policy: litconfig.PolicyConfig{EnforceMatchers: []string{"team"}},
			fragments: []*litconfig.Fragment{{
				Name: "db",
				Routes: []*amconfig.Route{{
					Receiver: "slack",
					Matchers: mustMatchers(t, "team", "db"),
				}},
			}},
			wantIssues: 0,
		},
		{
			name:   "fragment with no routes — no violation",
			policy: litconfig.PolicyConfig{EnforceMatchers: []string{"team"}},
			fragments: []*litconfig.Fragment{
				{Name: "db"},
			},
			wantIssues: 0,
		},
		{
			name:   "root child route missing required matcher — violation",
			policy: litconfig.PolicyConfig{EnforceMatchers: []string{"team"}},
			fragments: []*litconfig.Fragment{{
				Name:   "root",
				Routes: []*amconfig.Route{{Receiver: "slack", Match: map[string]string{"severity": "critical"}}},
			}},
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewPolicyChecker(tt.policy)
			issues := checker.Check(tt.fragments)
			assert.Len(t, issues, tt.wantIssues)
		})
	}
}

func TestPolicyChecker_NoPolicy_ReturnsNil(t *testing.T) {
	checker := NewPolicyChecker(litconfig.PolicyConfig{})
	fragments := []*litconfig.Fragment{{Name: "db"}}
	assert.Nil(t, checker.Check(fragments))
}

func TestPolicyChecker_Root_RequireTests(t *testing.T) {
	policy := litconfig.PolicyConfig{RequireTests: true}

	t.Run("root with no tests fails", func(t *testing.T) {
		checker := NewPolicyChecker(policy)
		issues := checker.Check([]*litconfig.Fragment{{Name: "root"}})
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "root")
	})

	t.Run("root with tests passes", func(t *testing.T) {
		checker := NewPolicyChecker(policy)
		issues := checker.Check([]*litconfig.Fragment{
			{Name: "root", Tests: []*types.TestCase{{Name: "base test"}}},
		})
		assert.Empty(t, issues)
	})
}

func TestPolicyChecker_Root_EnforceMatchers(t *testing.T) {
	// Root's catch-all route is excluded from rootFrag.Routes (it's base.Route, not
	// base.Route.Routes). Only child routes are evaluated.

	t.Run("child route without required matcher fails", func(t *testing.T) {
		checker := NewPolicyChecker(litconfig.PolicyConfig{EnforceMatchers: []string{"team"}})
		issues := checker.Check([]*litconfig.Fragment{{
			Name:   "root",
			Routes: []*amconfig.Route{{Receiver: "pagerduty", Match: map[string]string{"severity": "critical"}}},
		}})
		assert.Len(t, issues, 1)
	})

	t.Run("child route with required matcher passes", func(t *testing.T) {
		checker := NewPolicyChecker(litconfig.PolicyConfig{EnforceMatchers: []string{"team"}})
		issues := checker.Check([]*litconfig.Fragment{{
			Name:   "root",
			Routes: []*amconfig.Route{{Receiver: "pagerduty", Match: map[string]string{"team": "platform"}}},
		}})
		assert.Empty(t, issues)
	})

	t.Run("no child routes — no violation", func(t *testing.T) {
		checker := NewPolicyChecker(litconfig.PolicyConfig{EnforceMatchers: []string{"team"}})
		issues := checker.Check([]*litconfig.Fragment{{Name: "root"}})
		assert.Empty(t, issues)
	})
}

func TestPolicyChecker_SkipRoot(t *testing.T) {
	policy := litconfig.PolicyConfig{
		RequireTests: true,
		SkipRoot:     true,
	}
	checker := NewPolicyChecker(policy)

	// Root has no tests — would violate require_tests, but SkipRoot suppresses it.
	// The db fragment has tests so it must pass cleanly.
	issues := checker.Check([]*litconfig.Fragment{
		{Name: "root"},
		{Name: "db", Tests: []*types.TestCase{{Name: "t1"}}},
	})
	assert.Empty(t, issues, "root require_tests violation must be suppressed when SkipRoot=true")
}

func mustMatchers(t *testing.T, name, value string) amconfig.Matchers {
	t.Helper()
	m, err := labels.NewMatcher(labels.MatchEqual, name, value)
	if err != nil {
		t.Fatalf("creating matcher: %v", err)
	}
	return amconfig.Matchers{m}
}
