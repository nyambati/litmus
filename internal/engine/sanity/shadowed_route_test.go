package sanity

import (
	"regexp"
	"strings"
	"testing"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/require"
)

// helpers -------------------------------------------------------------------

func mustMatcher(t *testing.T, typ labels.MatchType, name, value string) *labels.Matcher {
	t.Helper()
	m, err := labels.NewMatcher(typ, name, value)
	require.NoError(t, err)
	return m
}

func mustRegexp(pattern string) config.Regexp {
	return config.Regexp{Regexp: regexp.MustCompile("^(?:" + pattern + ")$")}
}

func detectIssues(root *config.Route) []string {
	return NewShadowedRouteDetector(root).Detect()
}

// Table-driven core cases ---------------------------------------------------

func TestShadowedRouteDetector(t *testing.T) {
	tests := []struct {
		name      string
		root      *config.Route
		wantCount int
		wantIn    []string // receiver names expected somewhere in issues
		wantNotIn []string // receiver names that must NOT appear in issues
	}{
		// ── continue: true guard ────────────────────────────────────────────
		{
			name: "continue_true_never_shadows",
			// parent.continue=true → Alertmanager evaluates child regardless of match
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "broad", Match: map[string]string{"service": "api"}, Continue: true},
				{Receiver: "specific", Match: map[string]string{"service": "api", "env": "prod"}},
			}},
			wantCount: 0,
		},
		{
			name: "continue_false_explicit_shadows",
			// explicit continue:false (same as default) must shadow
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "broad", Match: map[string]string{"service": "api"}, Continue: false},
				{Receiver: "specific", Match: map[string]string{"service": "api", "env": "prod"}},
			}},
			wantCount: 1,
			wantIn:    []string{"specific"},
		},
		{
			// A(continue:true) → B(continue:false) → C
			// B may shadow C; A never shadows because it continues
			name: "continue_chain_middle_shadows_not_first",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "A", Match: map[string]string{"service": "api"}, Continue: true},
				{Receiver: "B", Match: map[string]string{"service": "api"}, Continue: false},
				{Receiver: "C", Match: map[string]string{"service": "api", "env": "prod"}},
			}},
			wantCount: 1,
			wantIn:    []string{"C"},
			wantNotIn: []string{"A"},
		},

		// ── catch-all routes ────────────────────────────────────────────────
		{
			name: "catch_all_parent_shadows_specific_child",
			// no matchers = match everything → shadows any child
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "catch-all", Continue: false},
				{Receiver: "specific", Match: map[string]string{"service": "api"}},
			}},
			wantCount: 1,
			wantIn:    []string{"specific"},
		},
		{
			name: "catch_all_child_not_shadowed_by_specific_parent",
			// specific parent only handles service=api; catch-all child handles everything else
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "specific", Match: map[string]string{"service": "api"}, Continue: false},
				{Receiver: "catch-all"},
			}},
			wantCount: 0,
		},
		{
			name: "duplicate_catch_all_shadowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "first-catch-all", Continue: false},
				{Receiver: "second-catch-all"},
			}},
			wantCount: 1,
			wantIn:    []string{"second-catch-all"},
		},

		// ── identical / duplicate matchers ──────────────────────────────────
		{
			name: "identical_matchers_shadowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "first", Match: map[string]string{"service": "api"}},
				{Receiver: "duplicate", Match: map[string]string{"service": "api"}},
			}},
			wantCount: 1,
			wantIn:    []string{"duplicate"},
		},

		// ── subset / superset direction ──────────────────────────────────────
		{
			name: "broader_parent_shadows_specific_child",
			// parent fewer constraints → matches superset → shadows child
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "broad", Match: map[string]string{"service": "api"}},
				{Receiver: "specific", Match: map[string]string{"service": "api", "env": "prod"}},
			}},
			wantCount: 1,
			wantIn:    []string{"specific"},
		},
		{
			name: "specific_parent_does_not_shadow_broader_child",
			// parent more constraints → matches subset → child still handles the rest
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "specific", Match: map[string]string{"service": "api", "env": "prod"}},
				{Receiver: "broad", Match: map[string]string{"service": "api"}},
			}},
			wantCount: 0,
		},

		// ── disjoint / partial-overlap labels ───────────────────────────────
		{
			name: "different_values_same_label_not_shadowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "api", Match: map[string]string{"service": "api"}},
				{Receiver: "db", Match: map[string]string{"service": "db"}},
			}},
			wantCount: 0,
		},
		{
			name: "disjoint_label_keys_not_shadowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "by-service", Match: map[string]string{"service": "api"}},
				{Receiver: "by-env", Match: map[string]string{"env": "prod"}},
			}},
			wantCount: 0,
		},
		{
			name: "partial_label_overlap_not_shadowed",
			// parent={service,env}, child={service,severity} — service matches but env≠severity
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "parent", Match: map[string]string{"service": "api", "env": "prod"}},
				{Receiver: "child", Match: map[string]string{"service": "api", "severity": "critical"}},
			}},
			wantCount: 0,
		},

		// ── multiple siblings ────────────────────────────────────────────────
		{
			name: "one_broad_parent_shadows_two_specific_children",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "broad", Match: map[string]string{"service": "api"}},
				{Receiver: "child-a", Match: map[string]string{"service": "api", "env": "prod"}},
				{Receiver: "child-b", Match: map[string]string{"service": "api", "severity": "critical"}},
			}},
			wantCount: 2,
			wantIn:    []string{"child-a", "child-b"},
		},
		{
			name: "middle_sibling_shadows_last_first_does_not",
			// first is specific (env=prod, service=api); middle is broader (service=api);
			// last matches middle → shadowed by middle, not first
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "first", Match: map[string]string{"service": "api", "env": "prod"}},
				{Receiver: "middle", Match: map[string]string{"service": "api"}},
				{Receiver: "last", Match: map[string]string{"service": "api", "severity": "critical"}},
			}},
			wantCount: 1,
			wantIn:    []string{"last"},
			wantNotIn: []string{"middle"},
		},
		{
			name: "no_shadowing_among_three_distinct_routes",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "svc-api", Match: map[string]string{"service": "api"}},
				{Receiver: "svc-db", Match: map[string]string{"service": "db"}},
				{Receiver: "svc-cache", Match: map[string]string{"service": "cache"}},
			}},
			wantCount: 0,
		},

		// ── negative matchers in child ───────────────────────────────────────
		{
			// child negative matchers make child MORE restrictive; parent (broader) still shadows
			name: "negative_in_child_shadowed_by_broader_parent",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "broad", Match: map[string]string{"service": "api"}, Continue: false},
				{
					Receiver: "narrow",
					Match:    map[string]string{"service": "api"},
					// negative matcher restricts child further, but parent still broader
				},
			}},
			wantCount: 1,
			wantIn:    []string{"narrow"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := detectIssues(tt.root)
			require.Len(t, issues, tt.wantCount, "issues: %v", issues)
			for _, want := range tt.wantIn {
				require.True(t, isShadowedVictim(issues, want),
					"expected %q as shadowed victim in issues %v", want, issues)
			}
			for _, notWant := range tt.wantNotIn {
				require.False(t, isShadowedVictim(issues, notWant),
					"unexpected %q as shadowed victim in issues %v", notWant, issues)
			}
		})
	}
}

// containsAny returns true if substr appears in any issue string.
func containsAny(issues []string, substr string) bool {
	for _, s := range issues {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// isShadowedVictim returns true if receiver appears as the shadowed (victim) route
// in any issue — i.e. after "Route to" and before "is shadowed by".
func isShadowedVictim(issues []string, receiver string) bool {
	needle := `"` + receiver + `" is shadowed`
	return containsAny(issues, needle)
}

// Negative matcher mutual-exclusion -----------------------------------------

func TestShadowedRouteDetector_NegativeMatcher(t *testing.T) {
	t.Run("parent_neg_regex_child_pos_regex_same_value_not_shadowed", func(t *testing.T) {
		// parent type!~"job-x", child type=~"job-x" → mutually exclusive
		team := mustMatcher(t, labels.MatchRegexp, "label_team", "team-alpha")
		typeExclude := mustMatcher(t, labels.MatchNotRegexp, "type", "job-x")
		typeInclude := mustMatcher(t, labels.MatchRegexp, "type", "job-x")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "no-job-x", Matchers: config.Matchers{team, typeExclude}},
			{Receiver: "job-x-only", Matchers: config.Matchers{team, typeInclude}},
		}}
		require.Len(t, detectIssues(root), 0)
	})

	t.Run("parent_neg_equal_child_pos_equal_same_value_not_shadowed", func(t *testing.T) {
		// parent env!="prod", child env="prod" → mutually exclusive
		svc := mustMatcher(t, labels.MatchEqual, "service", "api")
		envExclude := mustMatcher(t, labels.MatchNotEqual, "env", "prod")
		envInclude := mustMatcher(t, labels.MatchEqual, "env", "prod")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "non-prod", Matchers: config.Matchers{svc, envExclude}},
			{Receiver: "prod-only", Matchers: config.Matchers{svc, envInclude}},
		}}
		require.Len(t, detectIssues(root), 0)
	})

	t.Run("parent_neg_different_value_still_shadows", func(t *testing.T) {
		// parent type!="other" (different value) + service=api → does not conflict with child type="prod"
		// parent positive {service:api} ⊆ child positive {service:api, type:prod} → shadowed
		svc := mustMatcher(t, labels.MatchEqual, "service", "api")
		typeExcludeOther := mustMatcher(t, labels.MatchNotEqual, "type", "other") // different value
		typeIncludeProd := mustMatcher(t, labels.MatchEqual, "type", "prod")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "parent", Matchers: config.Matchers{svc, typeExcludeOther}},
			{Receiver: "child", Matchers: config.Matchers{svc, typeIncludeProd}},
		}}
		// parent positive={service:api}, child positive={service:api, type:prod}
		// parent positive ⊆ child positive → shadowed; no mutual exclusion (different values)
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, isShadowedVictim(issues, "child"))
	})

	t.Run("parent_neg_on_label_absent_from_child_positive_still_shadows", func(t *testing.T) {
		// parent: service=api (pos), dh_app!~"dashboard" (neg on label child doesn't have)
		// child: service=api (pos), severity=critical (pos)
		// mutual exclusion check: parent neg label "dh_app" not in child positive → no conflict
		// parent positive {service} ⊆ child positive {service, severity} → shadowed
		svc := mustMatcher(t, labels.MatchEqual, "service", "api")
		appExclude := mustMatcher(t, labels.MatchNotRegexp, "dh_app", "dashboard")
		severity := mustMatcher(t, labels.MatchEqual, "severity", "critical")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "parent", Matchers: config.Matchers{svc, appExclude}},
			{Receiver: "child", Matchers: config.Matchers{svc, severity}},
		}}
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, isShadowedVictim(issues, "child"))
	})

	t.Run("child_only_negative_matchers_shadowed_by_catch_all", func(t *testing.T) {
		// child has only negative matchers → child positive set is empty
		// catch-all parent (no matchers) positive set is also empty → parent ⊆ child → shadowed
		typeExclude := mustMatcher(t, labels.MatchNotEqual, "type", "noise")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "catch-all", Continue: false},
			{Receiver: "neg-only-child", Matchers: config.Matchers{typeExclude}},
		}}
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, isShadowedVictim(issues, "neg-only-child"))
	})
}

// Matcher format coverage ---------------------------------------------------

func TestShadowedRouteDetector_MatcherFormats(t *testing.T) {
	t.Run("match_re_same_pattern_shadowed", func(t *testing.T) {
		re := mustRegexp("api|gateway")
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "broad", MatchRE: config.MatchRegexps{"service": re}},
			{Receiver: "specific", MatchRE: config.MatchRegexps{"service": re, "env": mustRegexp("prod")}},
		}}
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, containsAny(issues, "specific"))
	})

	t.Run("match_re_different_keys_not_shadowed", func(t *testing.T) {
		// periodicity=weekly has no overlap with local_team — different label keys
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "weekly",
				Match:    map[string]string{"periodicity": "weekly"},
				Continue: false,
			},
			{
				Receiver: "regional",
				MatchRE:  config.MatchRegexps{"local_team": mustRegexp("team-gamma")},
				Continue: true,
			},
		}}
		require.Len(t, detectIssues(root), 0)
	})

	t.Run("modern_matchers_positive_equal_shadowed", func(t *testing.T) {
		svc := mustMatcher(t, labels.MatchEqual, "service", "api")
		svcEnv := mustMatcher(t, labels.MatchEqual, "env", "prod")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "broad", Matchers: config.Matchers{svc}},
			{Receiver: "specific", Matchers: config.Matchers{svc, svcEnv}},
		}}
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, isShadowedVictim(issues, "specific"))
	})

	t.Run("modern_matchers_positive_regex_shadowed", func(t *testing.T) {
		svc := mustMatcher(t, labels.MatchRegexp, "service", "api|gateway")
		svcEnv := mustMatcher(t, labels.MatchRegexp, "env", "prod.*")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "broad", Matchers: config.Matchers{svc}},
			{Receiver: "specific", Matchers: config.Matchers{svc, svcEnv}},
		}}
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, isShadowedVictim(issues, "specific"))
	})

	t.Run("mixed_match_and_matchers_shadowed", func(t *testing.T) {
		// parent uses old match:, child uses modern matchers: — still detected
		severity := mustMatcher(t, labels.MatchEqual, "severity", "critical")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "broad", Match: map[string]string{"service": "api"}},
			{Receiver: "specific", Matchers: config.Matchers{
				mustMatcher(t, labels.MatchEqual, "service", "api"),
				severity,
			}},
		}}
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, isShadowedVictim(issues, "specific"))
	})

	t.Run("mixed_match_re_and_matchers_not_shadowed_different_keys", func(t *testing.T) {
		env := mustMatcher(t, labels.MatchRegexp, "env", "prod.*")

		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "by-team",
				MatchRE:  config.MatchRegexps{"label_team": mustRegexp("team-beta")},
			},
			{
				Receiver: "by-env",
				Matchers: config.Matchers{env},
			},
		}}
		require.Len(t, detectIssues(root), 0)
	})
}

// Nested / cumulative matchers ----------------------------------------------

func TestShadowedRouteDetector_NestedRoutes(t *testing.T) {
	t.Run("nested_child_not_shadowed_by_own_parent_tree", func(t *testing.T) {
		// single leaf route — no sibling to compare against → 0 issues
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{Receiver: "child", Match: map[string]string{"severity": "critical"}},
				},
			},
		}}
		require.Len(t, detectIssues(root), 0)
	})

	t.Run("sibling_before_nested_parent_shadows_leaf", func(t *testing.T) {
		// sibling A (service=api) comes before nested-parent B (service=api) + child C (severity=critical)
		// cumulative path for C = [{service:api},{severity:critical}]
		// A's path = [{service:api}]
		// A positive {service:api} ⊆ C positive {service:api, severity:critical} → A shadows C
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{Receiver: "sibling-A", Match: map[string]string{"service": "api"}},
			{
				Receiver: "nested-parent-B",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{Receiver: "leaf-C", Match: map[string]string{"severity": "critical"}},
				},
			},
		}}
		issues := detectIssues(root)
		require.Len(t, issues, 1)
		require.True(t, isShadowedVictim(issues, "leaf-C"))
	})

	t.Run("two_nested_trees_no_cross_shadow", func(t *testing.T) {
		// api-tree and db-tree — different service values, no shadowing
		root := &config.Route{Receiver: "root", Routes: []*config.Route{
			{
				Receiver: "api-parent",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{Receiver: "api-critical", Match: map[string]string{"severity": "critical"}},
				},
			},
			{
				Receiver: "db-parent",
				Match:    map[string]string{"service": "db"},
				Routes: []*config.Route{
					{Receiver: "db-critical", Match: map[string]string{"severity": "critical"}},
				},
			},
		}}
		require.Len(t, detectIssues(root), 0)
	})
}

// Regression cases: patterns that previously produced false positives -----------

func TestShadowedRouteDetector_RegressionCases(t *testing.T) {
	t.Run("continue_true_sibling_does_not_shadow_next_route", func(t *testing.T) {
		// broad-notify has continue:true → cannot shadow narrow-notify below it
		broadNotify := &config.Route{
			Receiver: "broad-notify",
			Matchers: config.Matchers{
				mustMatcher(t, labels.MatchEqual, "severity", "critical"),
				mustMatcher(t, labels.MatchRegexp, "env", "prd.*|prod.*"),
				mustMatcher(t, labels.MatchNotRegexp, "label_team", "team-a|team-b"),
			},
			Continue: true,
		}
		narrowNotify := &config.Route{
			Receiver: "narrow-notify",
			Matchers: config.Matchers{
				mustMatcher(t, labels.MatchEqual, "severity", "critical"),
				mustMatcher(t, labels.MatchRegexp, "env", "prd.*|prod.*"),
				mustMatcher(t, labels.MatchRegexp, "label_team", "team-a|team-c"),
			},
		}
		root := &config.Route{Receiver: "root", Routes: []*config.Route{broadNotify, narrowNotify}}
		require.Len(t, detectIssues(root), 0)
	})

	t.Run("mutually_exclusive_app_label_not_shadowed", func(t *testing.T) {
		// route-iss: app!~"(app-x|app-y)" — excludes apps route-dd requires
		// route-dd: app=~"(app-x|app-y)"  — requires what route-iss excludes
		// continue:true on route-iss also prevents shadowing
		routeIss := &config.Route{
			Receiver: "team-iss",
			Matchers: config.Matchers{
				mustMatcher(t, labels.MatchEqual, "label_team", "team-one"),
				mustMatcher(t, labels.MatchRegexp, "env", "prod|prod.*"),
				mustMatcher(t, labels.MatchNotRegexp, "app", "(app-x|app-y|app-z)"),
			},
			Continue: true,
		}
		routeDd := &config.Route{
			Receiver: "team-dd",
			Matchers: config.Matchers{
				mustMatcher(t, labels.MatchEqual, "label_team", "team-one"),
				mustMatcher(t, labels.MatchRegexp, "env", "prod|prod.*"),
				mustMatcher(t, labels.MatchRegexp, "app", "(app-x|app-y)"),
				mustMatcher(t, labels.MatchNotRegexp, "periodicity", "(daily|weekly)"),
			},
			Continue: true,
		}
		root := &config.Route{Receiver: "root", Routes: []*config.Route{routeIss, routeDd}}
		require.Len(t, detectIssues(root), 0)
	})

	t.Run("broad_continue_true_does_not_shadow_specific_sibling", func(t *testing.T) {
		// team-base has continue:true → team-specific reachable despite same base label
		baseRoute := &config.Route{
			Receiver: "team-base",
			Match:    map[string]string{"label_team": "team-two"},
			Continue: true,
		}
		specificRoute := &config.Route{
			Receiver: "team-specific",
			Match:    map[string]string{"label_team": "team-two", "source": "external"},
		}
		root := &config.Route{Receiver: "root", Routes: []*config.Route{baseRoute, specificRoute}}
		require.Len(t, detectIssues(root), 0)
	})

	t.Run("broad_continue_true_with_neg_matcher_does_not_shadow_specific", func(t *testing.T) {
		// broad route has continue:true → specific routes reachable
		broad := &config.Route{
			Receiver: "team-three-broad",
			Matchers: config.Matchers{
				mustMatcher(t, labels.MatchEqual, "label_team", "team-three"),
				mustMatcher(t, labels.MatchNotRegexp, "app", "app-a|app-b"),
			},
			Continue: true,
		}
		specific := &config.Route{
			Receiver: "team-three-specific",
			Matchers: config.Matchers{
				mustMatcher(t, labels.MatchEqual, "label_team", "team-three"),
				mustMatcher(t, labels.MatchRegexp, "alertname", "AlertX|AlertY"),
				mustMatcher(t, labels.MatchNotRegexp, "app", "app-a|app-b"),
			},
			Continue: true,
		}
		root := &config.Route{Receiver: "root", Routes: []*config.Route{broad, specific}}
		require.Len(t, detectIssues(root), 0)
	})
}
