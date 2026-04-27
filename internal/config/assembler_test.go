package config

import (
	"testing"

	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssembler(t *testing.T) {
	tests := []struct {
		name      string
		base      *AlertmanagerConfig
		fragments []*Fragment
		wantErr   bool
		validate  func(t *testing.T, assembled *AlertmanagerConfig)
	}{
		{
			name: "Namespacing and Basic Merging",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:      "db-team",
					Namespace: "db",
					Receivers: []Receiver{{Name: "critical"}},
					Routes:    []*amconfig.Route{{Receiver: "critical"}},
				},
			},
			validate: func(t *testing.T, assembled *AlertmanagerConfig) {
				t.Helper()
				assert.Equal(t, "db-critical", assembled.Receivers[0].Name)
				assert.Equal(t, "db-critical", assembled.Route.Routes[0].Receiver)
			},
		},
		{
			name: "No Group — Flat Merge to Root",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:   "db-team",
					Routes: []*amconfig.Route{{Receiver: "db-receiver"}},
				},
			},
			validate: func(t *testing.T, assembled *AlertmanagerConfig) {
				t.Helper()
				require.Len(t, assembled.Route.Routes, 1)
				assert.Equal(t, "db-receiver", assembled.Route.Routes[0].Receiver)
			},
		},
		{
			name: "Group — Synthetic Parent Created",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:   "db-team",
					Group:  &FragmentGroup{Match: map[string]string{"scope": "teams"}},
					Routes: []*amconfig.Route{{Receiver: "db-receiver"}},
				},
			},
			validate: func(t *testing.T, assembled *AlertmanagerConfig) {
				t.Helper()
				require.Len(t, assembled.Route.Routes, 1)
				parent := assembled.Route.Routes[0]
				assert.Equal(t, map[string]string{"scope": "teams"}, parent.Match)
				assert.Equal(t, "default", parent.Receiver, "inherits root receiver")
				require.Len(t, parent.Routes, 1)
				assert.Equal(t, "db-receiver", parent.Routes[0].Receiver)
			},
		},
		{
			name: "Group Receiver Explicit",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name: "db-team",
					Group: &FragmentGroup{
						Match:    map[string]string{"scope": "teams"},
						Receiver: "teams-fallback",
					},
					Routes: []*amconfig.Route{{Receiver: "db-receiver"}},
				},
			},
			validate: func(t *testing.T, assembled *AlertmanagerConfig) {
				t.Helper()
				assert.Equal(t, "teams-fallback", assembled.Route.Routes[0].Receiver)
			},
		},
		{
			name: "Namespace prefixes Group.Receiver and child route receivers",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:      "payments",
					Namespace: "payments",
					Group: &FragmentGroup{
						Match:    map[string]string{"label_team": "payments"},
						Receiver: "fallback",
					},
					Routes: []*amconfig.Route{
						{Receiver: "critical"},
						{Receiver: "warning"},
					},
					Receivers: []Receiver{
						{Name: "fallback"},
						{Name: "critical"},
						{Name: "warning"},
					},
				},
			},
			validate: func(t *testing.T, assembled *AlertmanagerConfig) {
				t.Helper()
				require.Len(t, assembled.Route.Routes, 1)
				parent := assembled.Route.Routes[0]
				assert.Equal(t, "payments-fallback", parent.Receiver)
				require.Len(t, parent.Routes, 2)
				assert.Equal(t, "payments-critical", parent.Routes[0].Receiver)
				assert.Equal(t, "payments-warning", parent.Routes[1].Receiver)
				names := make([]string, len(assembled.Receivers))
				for i, r := range assembled.Receivers {
					names[i] = r.Name
				}
				assert.ElementsMatch(t, []string{"payments-fallback", "payments-critical", "payments-warning"}, names)
			},
		},
		{
			name: "Two Fragments Same Group — Co-located",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:   "db-team",
					Group:  &FragmentGroup{Match: map[string]string{"scope": "teams"}},
					Routes: []*amconfig.Route{{Receiver: "db-receiver"}},
				},
				{
					Name:   "net-team",
					Group:  &FragmentGroup{Match: map[string]string{"scope": "teams"}},
					Routes: []*amconfig.Route{{Receiver: "net-receiver"}},
				},
			},
			validate: func(t *testing.T, assembled *AlertmanagerConfig) {
				t.Helper()
				require.Len(t, assembled.Route.Routes, 1, "single synthetic parent for same group")
				parent := assembled.Route.Routes[0]
				require.Len(t, parent.Routes, 2)
			},
		},
		{
			name: "Group Receiver Conflict — Error",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:   "db-team",
					Group:  &FragmentGroup{Match: map[string]string{"scope": "teams"}, Receiver: "fallback-a"},
					Routes: []*amconfig.Route{{Receiver: "db-receiver"}},
				},
				{
					Name:   "net-team",
					Group:  &FragmentGroup{Match: map[string]string{"scope": "teams"}, Receiver: "fallback-b"},
					Routes: []*amconfig.Route{{Receiver: "net-receiver"}},
				},
			},
			wantErr: true,
		},
		{
			name: "Merge Inhibit Rules",
			base: &AlertmanagerConfig{
				Route: &amconfig.Route{Receiver: "default"},
				InhibitRules: []amconfig.InhibitRule{
					{SourceMatch: map[string]string{"global": "rule"}},
				},
			},
			fragments: []*Fragment{
				{
					Name: "db-team",
					InhibitRules: []amconfig.InhibitRule{
						{SourceMatch: map[string]string{"team": "db"}},
					},
				},
			},
			validate: func(t *testing.T, assembled *AlertmanagerConfig) {
				t.Helper()
				assert.Len(t, assembled.InhibitRules, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assembler := NewAssembler(tt.base)
			assembled, err := assembler.Assemble(tt.fragments)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, assembled)
			}
		})
	}
}
