package sanity

import (
	"testing"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/require"
)

// helpers -------------------------------------------------------------------

func detectDeadRoutes(root *config.Route) []string {
	return NewDeadReceiverDetector(root).Detect()
}

// isDeadReceiver returns true if receiver appears as the dead route in any issue.
func isDeadReceiver(issues []string, receiver string) bool {
	return containsAny(issues, `Route to "`+receiver+`" can never`)
}

// Table-driven core cases ---------------------------------------------------

func TestDeadReceiverDetector(t *testing.T) {
	// Pre-built matchers usable inside the test table (mustMatcher requires *testing.T).
	svcApi := mustMatcher(t, labels.MatchEqual, "service", "api")
	svcDb := mustMatcher(t, labels.MatchEqual, "service", "db")
	svcApiNeg := mustMatcher(t, labels.MatchNotEqual, "service", "api")
	envProd := mustMatcher(t, labels.MatchEqual, "env", "prod")

	tests := []struct {
		name      string
		root      *config.Route
		wantCount int
		wantIn    []string // receivers expected in issues
		wantNotIn []string // receivers that must NOT appear
	}{
		// ── no contradiction ────────────────────────────────────────────────
		{
			name:      "no_matchers_anywhere",
			root:      &config.Route{Receiver: "root", Routes: []*config.Route{{Receiver: "team-a"}}},
			wantCount: 0,
		},
		{
			name:      "root_only_no_children",
			root:      &config.Route{Receiver: "root"},
			wantCount: 0,
		},
		{
			name: "child_adds_different_label",
			// parent: service=api, child: env=prod — different label keys, compatible
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "parent",
					Match:    map[string]string{"service": "api"},
					Routes:   []*config.Route{{Receiver: "child", Match: map[string]string{"env": "prod"}}},
				},
			}},
			wantCount: 0,
		},
		{
			name: "child_repeats_same_label_same_value",
			// parent: service=api, child: service=api — redundant but not contradictory
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "parent",
					Match:    map[string]string{"service": "api"},
					Routes:   []*config.Route{{Receiver: "child", Match: map[string]string{"service": "api"}}},
				},
			}},
			wantCount: 0,
		},
		{
			name: "child_narrows_with_additional_label",
			// parent: service=api, child: service=api + env=prod — more specific, still reachable
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "parent",
					Match:    map[string]string{"service": "api"},
					Routes: []*config.Route{
						{Receiver: "child", Match: map[string]string{"service": "api", "env": "prod"}},
					},
				},
			}},
			wantCount: 0,
		},
		{
			name: "sibling_routes_no_inheritance",
			// siblings don't inherit each other's matchers — only ancestor chain matters
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "api-team", Match: map[string]string{"service": "api"}},
				{Receiver: "db-team", Match: map[string]string{"service": "db"}},
			}},
			wantCount: 0,
		},

		// ── basic dead routes ────────────────────────────────────────────────
		{
			name: "child_contradicts_parent_same_label",
			// parent: service=api, child: service=db — service can't be both → dead
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "api-parent",
					Match:    map[string]string{"service": "api"},
					Routes:   []*config.Route{{Receiver: "db-team", Match: map[string]string{"service": "db"}}},
				},
			}},
			wantCount: 1,
			wantIn:    []string{"db-team"},
			wantNotIn: []string{"api-parent"},
		},
		{
			name: "grandchild_contradicts_grandparent",
			// grandparent: service=api, parent: env=prod (compatible), grandchild: service=db → dead
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "grandparent",
					Match:    map[string]string{"service": "api"},
					Routes: []*config.Route{
						{
							Receiver: "parent",
							Match:    map[string]string{"env": "prod"},
							Routes:   []*config.Route{{Receiver: "grandchild", Match: map[string]string{"service": "db"}}},
						},
					},
				},
			}},
			wantCount: 1,
			wantIn:    []string{"grandchild"},
			wantNotIn: []string{"parent", "grandparent"},
		},
		{
			name: "self_contradiction_pos_neg_same_label_value",
			// single route: service=api AND service!=api — impossible to satisfy
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "impossible", Matchers: config.Matchers{svcApi, svcApiNeg}},
			}},
			wantCount: 1,
			wantIn:    []string{"impossible"},
		},
		{
			name: "self_contradiction_two_pos_same_label_different_values",
			// Matchers list allows duplicate label — service=api AND service=db on same route
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "conflict", Matchers: config.Matchers{svcApi, svcDb}},
			}},
			wantCount: 1,
			wantIn:    []string{"conflict"},
		},
		{
			name: "inherited_neg_conflicts_child_positive",
			// parent: service!=api, child: service=api → contradiction
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "no-api",
					Matchers: config.Matchers{svcApiNeg},
					Routes:   []*config.Route{{Receiver: "api-team", Match: map[string]string{"service": "api"}}},
				},
			}},
			wantCount: 1,
			wantIn:    []string{"api-team"},
			wantNotIn: []string{"no-api"},
		},
		{
			name: "root_matcher_conflicts_with_child",
			// root itself has service=api; child requires service=db → dead
			root: &config.Route{
				Receiver: "root",
				Match:    map[string]string{"service": "api"},
				Routes:   []*config.Route{{Receiver: "db-team", Match: map[string]string{"service": "db"}}},
			},
			wantCount: 1,
			wantIn:    []string{"db-team"},
			wantNotIn: []string{"root"}, // root is never reported — it has no ancestors
		},

		// ── report-and-stop: dead route blocks descendant reporting ──────────
		{
			name: "dead_child_stops_recursion_grandchild_not_reported",
			// dead-child contradicts parent; grandchild would also be dead, but only dead-child reported
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "parent",
					Match:    map[string]string{"service": "api"},
					Routes: []*config.Route{
						{
							Receiver: "dead-child",
							Match:    map[string]string{"service": "db"},
							Routes:   []*config.Route{{Receiver: "grandchild", Matchers: config.Matchers{envProd}}},
						},
					},
				},
			}},
			wantCount: 1,
			wantIn:    []string{"dead-child"},
			wantNotIn: []string{"grandchild"},
		},

		// ── multiple independent dead routes ─────────────────────────────────
		{
			name: "two_dead_routes_in_separate_subtrees",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "api-parent",
					Match:    map[string]string{"service": "api"},
					Routes:   []*config.Route{{Receiver: "dead-under-api", Match: map[string]string{"service": "db"}}},
				},
				{
					Receiver: "web-parent",
					Match:    map[string]string{"service": "web"},
					Routes:   []*config.Route{{Receiver: "dead-under-web", Match: map[string]string{"service": "db"}}},
				},
			}},
			wantCount: 2,
			wantIn:    []string{"dead-under-api", "dead-under-web"},
		},
		{
			name: "two_dead_siblings_under_same_parent",
			// parent: service=api; both children contradict it
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "api-parent",
					Match:    map[string]string{"service": "api"},
					Routes: []*config.Route{
						{Receiver: "dead-a", Match: map[string]string{"service": "db"}},
						{Receiver: "dead-b", Match: map[string]string{"service": "web"}},
					},
				},
			}},
			wantCount: 2,
			wantIn:    []string{"dead-a", "dead-b"},
			wantNotIn: []string{"api-parent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := detectDeadRoutes(tt.root)
			require.Len(t, issues, tt.wantCount, "issues: %v", issues)
			for _, want := range tt.wantIn {
				require.True(t, isDeadReceiver(issues, want),
					"expected %q as dead receiver in issues %v", want, issues)
			}
			for _, notWant := range tt.wantNotIn {
				require.False(t, isDeadReceiver(issues, notWant),
					"unexpected %q as dead receiver in issues %v", notWant, issues)
			}
		})
	}
}

// Negative matcher edge cases -----------------------------------------------

func TestDeadReceiverDetector_NegativeMatchers(t *testing.T) {
	t.Run("parent_neg_child_pos_same_label_same_value_dead", func(t *testing.T) {
		// parent: service!=api (excludes api), child: service=api (requires api) → dead
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "no-api",
				Matchers: config.Matchers{mustMatcher(t, labels.MatchNotEqual, "service", "api")},
				Routes:   []*config.Route{{Receiver: "api-only", Match: map[string]string{"service": "api"}}},
			},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 1)
		require.True(t, isDeadReceiver(issues, "api-only"))
	})

	t.Run("parent_neg_child_pos_different_value_reachable", func(t *testing.T) {
		// parent: service!=api (excludes api), child: service=db (requires db, not api) → reachable
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "no-api",
				Matchers: config.Matchers{mustMatcher(t, labels.MatchNotEqual, "service", "api")},
				Routes:   []*config.Route{{Receiver: "db-team", Match: map[string]string{"service": "db"}}},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})

	t.Run("child_negative_on_unrelated_label_reachable", func(t *testing.T) {
		// parent: service=api, child adds env!=prod (negative on a different label) → reachable
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "api-parent",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{
						Receiver: "non-prod-api",
						Matchers: config.Matchers{mustMatcher(t, labels.MatchNotEqual, "env", "prod")},
					},
				},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})

	t.Run("pos_neg_same_label_different_value_reachable", func(t *testing.T) {
		// service=api (positive) AND service!=db (negative, different value) — no contradiction
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "api-not-db",
				Matchers: config.Matchers{
					mustMatcher(t, labels.MatchEqual, "service", "api"),
					mustMatcher(t, labels.MatchNotEqual, "service", "db"),
				},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})
}

// Matcher format coverage ---------------------------------------------------

func TestDeadReceiverDetector_MatcherFormats(t *testing.T) {
	t.Run("legacy_match_map_conflict", func(t *testing.T) {
		// Both parent and child use legacy match: map format
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"service": "api"},
				Routes:   []*config.Route{{Receiver: "child", Match: map[string]string{"service": "db"}}},
			},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 1)
		require.True(t, isDeadReceiver(issues, "child"))
	})

	t.Run("modern_matchers_equal_conflict", func(t *testing.T) {
		// Both use modern matchers: with MatchEqual
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Matchers: config.Matchers{mustMatcher(t, labels.MatchEqual, "service", "api")},
				Routes: []*config.Route{
					{
						Receiver: "child",
						Matchers: config.Matchers{mustMatcher(t, labels.MatchEqual, "service", "db")},
					},
				},
			},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 1)
		require.True(t, isDeadReceiver(issues, "child"))
	})

	t.Run("mixed_legacy_and_modern_conflict", func(t *testing.T) {
		// parent uses legacy match:, child uses modern matchers: — still detected
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{
						Receiver: "child",
						Matchers: config.Matchers{mustMatcher(t, labels.MatchEqual, "service", "db")},
					},
				},
			},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 1)
		require.True(t, isDeadReceiver(issues, "child"))
	})

	t.Run("parent_regex_child_exact_not_flagged", func(t *testing.T) {
		// parent: service=~"api" (regex) — cannot statically prove conflict with child service=db
		// regex matchers are skipped to avoid false positives
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				MatchRE:  config.MatchRegexps{"service": mustRegexp("api")},
				Routes:   []*config.Route{{Receiver: "child", Match: map[string]string{"service": "db"}}},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})

	t.Run("parent_exact_child_regex_not_flagged", func(t *testing.T) {
		// parent: service=api (exact), child: service=~"db|cache" (regex)
		// child's regex matcher skipped — no false positive
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{
						Receiver: "child",
						Matchers: config.Matchers{mustMatcher(t, labels.MatchRegexp, "service", "db|cache")},
					},
				},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})

	t.Run("both_regex_same_label_not_flagged", func(t *testing.T) {
		// parent: service=~"api", child: service=~"db" — both regex, both skipped
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Matchers: config.Matchers{mustMatcher(t, labels.MatchRegexp, "service", "api")},
				Routes: []*config.Route{
					{
						Receiver: "child",
						Matchers: config.Matchers{mustMatcher(t, labels.MatchRegexp, "service", "db")},
					},
				},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})

	t.Run("neg_regex_not_flagged_against_pos_exact", func(t *testing.T) {
		// parent: service!~"api" (neg regex), child: service=api (pos exact)
		// neg regex skipped — no false positive (can't statically evaluate regex)
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Matchers: config.Matchers{mustMatcher(t, labels.MatchNotRegexp, "service", "api")},
				Routes:   []*config.Route{{Receiver: "child", Match: map[string]string{"service": "api"}}},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})
}

// Nested route inheritance --------------------------------------------------

func TestDeadReceiverDetector_NestedRoutes(t *testing.T) {
	t.Run("three_level_contradiction_at_leaf", func(t *testing.T) {
		// root → A (service=api) → B (env=prod) → C (service=db) — contradiction at C
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "A",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{
						Receiver: "B",
						Match:    map[string]string{"env": "prod"},
						Routes:   []*config.Route{{Receiver: "C", Match: map[string]string{"service": "db"}}},
					},
				},
			},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 1)
		require.True(t, isDeadReceiver(issues, "C"))
		require.False(t, isDeadReceiver(issues, "A"))
		require.False(t, isDeadReceiver(issues, "B"))
	})

	t.Run("two_separate_trees_independent_contradictions", func(t *testing.T) {
		// api-tree and db-tree each have a dead child — reported independently
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "api-parent",
				Match:    map[string]string{"service": "api"},
				Routes:   []*config.Route{{Receiver: "api-dead", Match: map[string]string{"service": "db"}}},
			},
			{
				Receiver: "db-parent",
				Match:    map[string]string{"service": "db"},
				Routes:   []*config.Route{{Receiver: "db-dead", Match: map[string]string{"service": "api"}}},
			},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 2)
		require.True(t, isDeadReceiver(issues, "api-dead"))
		require.True(t, isDeadReceiver(issues, "db-dead"))
	})

	t.Run("healthy_subtree_alongside_dead_subtree", func(t *testing.T) {
		// api-parent → dead-child (dead); db-parent → db-child (healthy)
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "api-parent",
				Match:    map[string]string{"service": "api"},
				Routes:   []*config.Route{{Receiver: "dead-child", Match: map[string]string{"service": "db"}}},
			},
			{
				Receiver: "db-parent",
				Match:    map[string]string{"service": "db"},
				Routes:   []*config.Route{{Receiver: "db-child", Match: map[string]string{"severity": "critical"}}},
			},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 1)
		require.True(t, isDeadReceiver(issues, "dead-child"))
		require.False(t, isDeadReceiver(issues, "db-child"))
	})
}

// Edge cases ----------------------------------------------------------------

func TestDeadReceiverDetector_EdgeCases(t *testing.T) {
	t.Run("nil_root", func(t *testing.T) {
		require.Len(t, detectDeadRoutes(nil), 0)
	})

	t.Run("root_with_no_children", func(t *testing.T) {
		require.Len(t, detectDeadRoutes(&config.Route{Receiver: "root"}), 0)
	})

	t.Run("root_never_reported_even_with_no_children", func(t *testing.T) {
		// Root is always reachable — it has no ancestors to contradict it
		root := &config.Route{Receiver: "root", Match: map[string]string{"service": "api"}}
		issues := detectDeadRoutes(root)
		require.False(t, isDeadReceiver(issues, "root"))
	})

	t.Run("contradiction_only_on_second_pos_matcher_with_same_label", func(t *testing.T) {
		// Covers the case where the Matchers slice has two MatchEqual entries for same label
		svcApi := mustMatcher(t, labels.MatchEqual, "service", "api")
		svcWeb := mustMatcher(t, labels.MatchEqual, "service", "web")
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "team", Matchers: config.Matchers{svcApi, svcWeb}},
		}}
		issues := detectDeadRoutes(root)
		require.Len(t, issues, 1)
		require.True(t, isDeadReceiver(issues, "team"))
	})

	t.Run("empty_matchers_on_child_not_dead", func(t *testing.T) {
		// Child has no matchers of its own — inherits parent's only, which are fine
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"service": "api"},
				Routes:   []*config.Route{{Receiver: "child"}},
			},
		}}
		require.Len(t, detectDeadRoutes(root), 0)
	})
}
