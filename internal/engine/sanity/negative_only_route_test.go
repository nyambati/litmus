package sanity

import (
	"testing"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestNegativeOnlyRouteDetector(t *testing.T) {
	negativeTeam := mustMatcher(t, labels.MatchNotEqual, "team", "ops")
	negativeEnv := mustMatcher(t, labels.MatchNotRegexp, "env", "prod|staging")
	positiveTeam := mustMatcher(t, labels.MatchEqual, "team", "platform")
	positiveEnv := mustMatcher(t, labels.MatchRegexp, "env", "prod|staging")

	tests := []struct {
		name      string
		root      *config.Route
		wantCount int
		wantIn    []string
		wantNotIn []string
	}{
		{
			name:      "nil root has no issues",
			root:      nil,
			wantCount: 0,
		},
		{
			name:      "catch-all route with no matchers is allowed",
			root:      &config.Route{Receiver: "root", Routes: []*config.Route{{Receiver: "catch-all"}}},
			wantCount: 0,
		},
		{
			name:      "route with only nil modern matchers is allowed",
			root:      &config.Route{Receiver: "root", Routes: []*config.Route{{Receiver: "nil-matchers", Matchers: config.Matchers{nil}}}},
			wantCount: 0,
		},
		{
			name:      "root route with only negative matcher is flagged",
			root:      &config.Route{Receiver: "root", Matchers: config.Matchers{negativeTeam}},
			wantCount: 1,
			wantIn:    []string{"root", `team!="ops"`},
		},
		{
			name: "legacy positive match is allowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "api", Match: map[string]string{"service": "api"}},
			}},
			wantCount: 0,
		},
		{
			name: "legacy regex positive match is allowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "api", MatchRE: map[string]config.Regexp{"service": mustRegexp("api|web")}},
			}},
			wantCount: 0,
		},
		{
			name: "route with only negative matcher is flagged",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "non-ops", Matchers: config.Matchers{negativeTeam}},
			}},
			wantCount: 1,
			wantIn:    []string{"non-ops", `team!="ops"`},
		},
		{
			name: "route with multiple negative matchers is flagged once",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "negative-only", Matchers: config.Matchers{negativeTeam, negativeEnv}},
			}},
			wantCount: 1,
			wantIn:    []string{"negative-only", `team!="ops"`, `env!~"prod|staging"`},
		},
		{
			name: "route with one positive and one negative matcher is allowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "platform-non-prod", Matchers: config.Matchers{positiveTeam, negativeEnv}},
			}},
			wantCount: 0,
		},
		{
			name: "route with positive regex and negative matcher is allowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "prodish-non-ops", Matchers: config.Matchers{positiveEnv, negativeTeam}},
			}},
			wantCount: 0,
		},
		{
			name: "legacy positive match plus modern negative matcher is allowed",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "api-non-ops", Match: map[string]string{"service": "api"}, Matchers: config.Matchers{negativeTeam}},
			}},
			wantCount: 0,
		},
		{
			name: "nested negative-only route is flagged",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "platform",
					Matchers: config.Matchers{positiveTeam},
					Routes: []*config.Route{
						{Receiver: "platform-non-prod", Matchers: config.Matchers{negativeEnv}},
					},
				},
			}},
			wantCount: 1,
			wantIn:    []string{"platform-non-prod", `env!~"prod|staging"`},
			wantNotIn: []string{"platform\""},
		},
		{
			name: "inherited positive matcher does not hide negative-only child",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{
					Receiver: "api",
					Match:    map[string]string{"service": "api"},
					Routes: []*config.Route{
						{Receiver: "api-non-ops", Matchers: config.Matchers{negativeTeam}},
					},
				},
			}},
			wantCount: 1,
			wantIn:    []string{"api-non-ops", `team!="ops"`},
		},
		{
			name: "multiple negative-only routes are all reported",
			root: &config.Route{Receiver: "root", Routes: []*config.Route{
				{Receiver: "non-ops", Matchers: config.Matchers{negativeTeam}},
				{Receiver: "non-prod", Matchers: config.Matchers{negativeEnv}},
			}},
			wantCount: 2,
			wantIn:    []string{"non-ops", "non-prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := NewNegativeOnlyRouteDetector(tt.root).Detect()
			require.Len(t, issues, tt.wantCount, "issues: %v", issues)
			for _, want := range tt.wantIn {
				require.True(t, containsAny(issues, want), "expected %q in issues %v", want, issues)
			}
			for _, notWant := range tt.wantNotIn {
				require.False(t, containsAny(issues, notWant), "did not expect %q in issues %v", notWant, issues)
			}
		})
	}
}
