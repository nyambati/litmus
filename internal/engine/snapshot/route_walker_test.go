package snapshot

import (
	"regexp"
	"testing"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestRouteWalker_FindTerminalPaths(t *testing.T) {
	tests := []struct {
		name         string
		route        *config.Route
		wantPathCount int
		wantReceivers []string
	}{
		{
			name: "single terminal route",
			route: &config.Route{
				Receiver: "default",
				Routes:   []*config.Route{},
			},
			wantPathCount: 1,
			wantReceivers: []string{"default"},
		},
		{
			name: "two child routes",
			route: &config.Route{
				Receiver: "root",
				Routes: []*config.Route{
					{
						Receiver: "api",
						Routes:   []*config.Route{},
						Match: map[string]string{
							"service": "api",
						},
					},
					{
						Receiver: "db",
						Routes:   []*config.Route{},
						Match: map[string]string{
							"service": "db",
						},
					},
				},
			},
			wantPathCount: 2,
			wantReceivers: []string{"api", "db"},
		},
		{
			name: "nested routes",
			route: &config.Route{
				Receiver: "root",
				Routes: []*config.Route{
					{
						Receiver: "prod",
						Match: map[string]string{
							"env": "prod",
						},
						Routes: []*config.Route{
							{
								Receiver: "prod-critical",
								Match: map[string]string{
									"severity": "critical",
								},
								Routes: []*config.Route{},
							},
						},
					},
				},
			},
			wantPathCount: 1,
			wantReceivers: []string{"prod-critical"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walker := NewRouteWalker(tt.route)
			paths := walker.FindTerminalPaths()

			require.Len(t, paths, tt.wantPathCount)
			for _, path := range paths {
				require.Contains(t, tt.wantReceivers, path.Receiver)
			}
		})
	}
}

func TestRouteWalker_MatchersCapture(t *testing.T) {
	rePattern := regexp.MustCompile("^(?:apac-compliance-team)$")

	matcherEq, err := labels.NewMatcher(labels.MatchEqual, "dh_env", "production")
	require.NoError(t, err)

	tests := []struct {
		name          string
		route         *config.Route
		wantLabelKey  model.LabelName
		wantLabelVal  model.LabelValue
	}{
		{
			name: "match_re matchers captured",
			route: &config.Route{
				Receiver: "apac",
				MatchRE:  config.MatchRegexps{"local_team": config.Regexp{Regexp: rePattern}},
			},
			wantLabelKey: "local_team",
			wantLabelVal: model.LabelValue(rePattern.String()),
		},
		{
			name: "modern matchers captured",
			route: &config.Route{
				Receiver: "prod",
				Matchers: config.Matchers{matcherEq},
			},
			wantLabelKey: "dh_env",
			wantLabelVal: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walker := NewRouteWalker(tt.route)
			paths := walker.FindTerminalPaths()

			require.Len(t, paths, 1)
			merged := make(model.LabelSet)
			for _, ls := range paths[0].Matchers {
				for k, v := range ls {
					merged[k] = v
				}
			}
			require.Equal(t, tt.wantLabelVal, merged[tt.wantLabelKey])
		})
	}
}

func TestRoutePath_Matchers(t *testing.T) {
	path := &RoutePath{
		Receiver: "critical",
		Matchers: []model.LabelSet{
			{"severity": "critical"},
			{"env": "prod"},
		},
	}

	// All matchers in path should be satisfied for alert to reach this route
	require.Len(t, path.Matchers, 2)
	require.Equal(t, "critical", path.Receiver)
}
