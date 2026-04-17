package snapshot

import (
	"testing"

	"github.com/prometheus/alertmanager/config"
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
