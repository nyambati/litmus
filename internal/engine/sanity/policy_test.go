package sanity

import (
	"testing"

	litconfig "github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/assert"
)

// enforce builds an EnforceConfig for use in test policies.
func enforce(strict bool, matchers ...string) litconfig.EnforceConfig {
	return litconfig.EnforceConfig{Strict: strict, Matchers: matchers}
}

func TestPolicyChecker_RequireTests(t *testing.T) {
	tests := []struct {
		name       string
		fragments  []*litconfig.Fragment
		wantIssues int
	}{
		{
			name:       "fragment with tests — no violation",
			fragments:  []*litconfig.Fragment{{Name: "db", Tests: []*types.TestCase{{Name: "t1"}}}},
			wantIssues: 0,
		},
		{
			name:       "fragment with no tests — violation",
			fragments:  []*litconfig.Fragment{{Name: "db"}},
			wantIssues: 1,
		},
		{
			name:       "root fragment checked — no tests is a violation",
			fragments:  []*litconfig.Fragment{{Name: "root"}},
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
			assert.Len(t, checker.Check(tt.fragments), tt.wantIssues)
		})
	}
}

func TestPolicyChecker_EnforceMatchers_Flat(t *testing.T) {
	tests := []struct {
		name       string
		policy     litconfig.PolicyConfig
		fragments  []*litconfig.Fragment
		wantIssues int
	}{
		{
			name:   "route has required label in Match — no violation",
			policy: litconfig.PolicyConfig{Enforce: enforce(true, "team")},
			fragments: []*litconfig.Fragment{{
				Name:   "db",
				Routes: []*amconfig.Route{{Receiver: "slack", Match: map[string]string{"team": "db"}}},
			}},
			wantIssues: 0,
		},
		{
			name:   "route missing required label — violation",
			policy: litconfig.PolicyConfig{Enforce: enforce(true, "team")},
			fragments: []*litconfig.Fragment{{
				Name:   "db",
				Routes: []*amconfig.Route{{Receiver: "slack", Match: map[string]string{"severity": "critical"}}},
			}},
			wantIssues: 1,
		},
		{
			name:   "route has required label in modern Matchers — no violation",
			policy: litconfig.PolicyConfig{Enforce: enforce(true, "team")},
			fragments: []*litconfig.Fragment{{
				Name:   "db",
				Routes: []*amconfig.Route{{Receiver: "slack", Matchers: mustMatchers(t, "team", "db")}},
			}},
			wantIssues: 0,
		},
		{
			name:       "fragment with no routes — no violation",
			policy:     litconfig.PolicyConfig{Enforce: enforce(true, "team")},
			fragments:  []*litconfig.Fragment{{Name: "db"}},
			wantIssues: 0,
		},
		{
			name:   "root child route missing required matcher — violation",
			policy: litconfig.PolicyConfig{Enforce: enforce(true, "team")},
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
			assert.Len(t, checker.Check(tt.fragments), tt.wantIssues)
		})
	}
}

func TestPolicyChecker_StrictMode_Inheritance(t *testing.T) {
	// strict=true (AND): all labels must be present collectively across the path.

	t.Run("parent covers one label, children complete the set — no violations", func(t *testing.T) {
		// parent[label_team] → child1[severity], child2[severity]
		// child1 union={label_team,severity} → covered; child2 same → covered
		// all children clean → parent cleared → no violations
		root := &amconfig.Route{
			Receiver: "root",
			Routes: []*amconfig.Route{
				{
					Receiver: "parent",
					Match:    map[string]string{"label_team": "payments"},
					Routes: []*amconfig.Route{
						{Receiver: "critical", Match: map[string]string{"severity": "critical"}},
						{Receiver: "warning", Match: map[string]string{"severity": "warning"}},
					},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: root.Routes}})
		assert.Empty(t, issues)
	})

	t.Run("parent covers one label, one child completes, one does not — violation on gap", func(t *testing.T) {
		// parent[label_team] → child1[severity] (clean), child2[] (missing severity → violation)
		root := &amconfig.Route{
			Receiver: "root",
			Routes: []*amconfig.Route{
				{
					Receiver: "parent",
					Match:    map[string]string{"label_team": "payments"},
					Routes: []*amconfig.Route{
						{Receiver: "critical", Match: map[string]string{"severity": "critical"}},
						{Receiver: "catch-all"},
					},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: root.Routes}})
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "catch-all")
	})

	t.Run("parent empty, children each cover one label — violations on both", func(t *testing.T) {
		// parent[] → child1[label_team] (missing severity), child2[severity] (missing label_team)
		root := &amconfig.Route{
			Receiver: "root",
			Routes: []*amconfig.Route{
				{
					Receiver: "parent",
					Routes: []*amconfig.Route{
						{Receiver: "team-only", Match: map[string]string{"label_team": "payments"}},
						{Receiver: "sev-only", Match: map[string]string{"severity": "critical"}},
					},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: root.Routes}})
		assert.Len(t, issues, 2)
	})

	t.Run("grandparent starts coverage, parent completes — descendants exempted", func(t *testing.T) {
		// grandparent[label_team] → parent[severity] (union complete) → child[] (skipped)
		routes := []*amconfig.Route{
			{
				Receiver: "grandparent",
				Match:    map[string]string{"label_team": "payments"},
				Routes: []*amconfig.Route{
					{
						Receiver: "parent",
						Match:    map[string]string{"severity": "critical"},
						Routes: []*amconfig.Route{
							{Receiver: "child"},
						},
					},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: routes}})
		assert.Empty(t, issues)
	})

	t.Run("parent fully covers all labels — children exempted entirely", func(t *testing.T) {
		routes := []*amconfig.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"label_team": "payments", "severity": "critical"},
				Routes: []*amconfig.Route{
					{Receiver: "child-a"},
					{Receiver: "child-b"},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: routes}})
		assert.Empty(t, issues)
	})

	t.Run("leaf route with no matchers anywhere — violation at leaf", func(t *testing.T) {
		routes := []*amconfig.Route{{Receiver: "orphan"}}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: routes}})
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "orphan")
	})
}

func TestPolicyChecker_NonStrictMode_Inheritance(t *testing.T) {
	// strict=false (OR): at least one label from the list must be present in the accumulated path.

	t.Run("parent has one required label — entire branch satisfied", func(t *testing.T) {
		// parent[label_team] satisfies OR → children skipped regardless of their matchers
		routes := []*amconfig.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"label_team": "payments"},
				Routes: []*amconfig.Route{
					{Receiver: "child-a"},
					{Receiver: "child-b"},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(false, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: routes}})
		assert.Empty(t, issues)
	})

	t.Run("parent empty, children each cover one required label — all branches satisfied", func(t *testing.T) {
		routes := []*amconfig.Route{
			{
				Receiver: "parent",
				Routes: []*amconfig.Route{
					{Receiver: "team-child", Match: map[string]string{"label_team": "payments"}},
					{Receiver: "sev-child", Match: map[string]string{"severity": "critical"}},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(false, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: routes}})
		assert.Empty(t, issues)
	})

	t.Run("nothing anywhere — violation at leaf", func(t *testing.T) {
		routes := []*amconfig.Route{
			{
				Receiver: "parent",
				Routes: []*amconfig.Route{
					{Receiver: "child"},
				},
			},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(false, "label_team", "severity")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "payments", Routes: routes}})
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "child")
	})

	t.Run("OR with single required label — any match satisfies", func(t *testing.T) {
		routes := []*amconfig.Route{
			{Receiver: "slack", Match: map[string]string{"service": "mysql"}},
		}
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(false, "team", "service")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "db", Routes: routes}})
		assert.Empty(t, issues)
	})
}

func TestPolicyChecker_NoPolicy_ReturnsNil(t *testing.T) {
	checker := NewPolicyChecker(litconfig.PolicyConfig{})
	assert.Nil(t, checker.Check([]*litconfig.Fragment{{Name: "db"}}))
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
	t.Run("child route without required matcher fails", func(t *testing.T) {
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "team")})
		issues := checker.Check([]*litconfig.Fragment{{
			Name:   "root",
			Routes: []*amconfig.Route{{Receiver: "pagerduty", Match: map[string]string{"severity": "critical"}}},
		}})
		assert.Len(t, issues, 1)
	})

	t.Run("child route with required matcher passes", func(t *testing.T) {
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "team")})
		issues := checker.Check([]*litconfig.Fragment{{
			Name:   "root",
			Routes: []*amconfig.Route{{Receiver: "pagerduty", Match: map[string]string{"team": "platform"}}},
		}})
		assert.Empty(t, issues)
	})

	t.Run("no child routes — no violation", func(t *testing.T) {
		checker := NewPolicyChecker(litconfig.PolicyConfig{Enforce: enforce(true, "team")})
		issues := checker.Check([]*litconfig.Fragment{{Name: "root"}})
		assert.Empty(t, issues)
	})
}

func TestPolicyChecker_SkipRoot(t *testing.T) {
	policy := litconfig.PolicyConfig{RequireTests: true, SkipRoot: true}
	checker := NewPolicyChecker(policy)
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
