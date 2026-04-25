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
		base      *amconfig.Config
		fragments []*Fragment
		wantErr   bool
		validate  func(t *testing.T, assembled *amconfig.Config)
	}{
		{
			name: "Namespacing and Basic Merging",
			base: &amconfig.Config{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:      "db-team",
					Namespace: "db",
					Receivers: []amconfig.Receiver{{Name: "critical"}},
					Routes:    []*amconfig.Route{{Receiver: "critical"}},
				},
			},
			validate: func(t *testing.T, assembled *amconfig.Config) {
				t.Helper()
				assert.Equal(t, "db-critical", assembled.Receivers[0].Name)
				assert.Equal(t, "db-critical", assembled.Route.Routes[0].Receiver)
			},
		},
		{
			name: "Hierarchical Mounting",
			base: &amconfig.Config{
				Route: &amconfig.Route{
					Receiver: "default",
					Routes: []*amconfig.Route{
						{
							Receiver: "platform-fallback",
							Match:    map[string]string{"scope": "teams"},
						},
					},
				},
			},
			fragments: []*Fragment{
				{
					Name:       "db-team",
					MountPoint: map[string]string{"scope": "teams"},
					Routes:     []*amconfig.Route{{Receiver: "db-receiver"}},
				},
			},
			validate: func(t *testing.T, assembled *amconfig.Config) {
				t.Helper()
				teamsRoute := assembled.Route.Routes[0]
				assert.Equal(t, "platform-fallback", teamsRoute.Receiver)
				assert.Len(t, teamsRoute.Routes, 1)
				assert.Equal(t, "db-receiver", teamsRoute.Routes[0].Receiver)
			},
		},
		{
			name: "Mount Point Not Found",
			base: &amconfig.Config{
				Route: &amconfig.Route{Receiver: "default"},
			},
			fragments: []*Fragment{
				{
					Name:       "db-team",
					MountPoint: map[string]string{"nonexistent": "label"},
					Routes:     []*amconfig.Route{{Receiver: "db-receiver"}},
				},
			},
			wantErr: true,
		},
		{
			name: "Merge Inhibit Rules",
			base: &amconfig.Config{
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
			validate: func(t *testing.T, assembled *amconfig.Config) {
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
