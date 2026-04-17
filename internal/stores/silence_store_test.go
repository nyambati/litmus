package stores

import (
	"context"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"litmus/internal/types"
)

func TestSilenceStore_Mutes(t *testing.T) {
	tests := []struct {
		name     string
		silences []types.Silence
		labels   model.LabelSet
		wantMute bool
	}{
		{
			name:     "empty silence list",
			silences: []types.Silence{},
			labels:   model.LabelSet{"service": "api"},
			wantMute: false,
		},
		{
			name: "exact label match mutes",
			silences: []types.Silence{
				{
					Labels:  map[string]string{"service": "api"},
					Comment: "api maintenance",
				},
			},
			labels:   model.LabelSet{"service": "api"},
			wantMute: true,
		},
		{
			name: "partial silence matches all labels",
			silences: []types.Silence{
				{
					Labels:  map[string]string{"service": "api"},
					Comment: "api maintenance",
				},
			},
			labels:   model.LabelSet{"service": "api", "env": "prod"},
			wantMute: true,
		},
		{
			name: "silence doesn't match different value",
			silences: []types.Silence{
				{
					Labels:  map[string]string{"service": "api"},
					Comment: "api maintenance",
				},
			},
			labels:   model.LabelSet{"service": "db"},
			wantMute: false,
		},
		{
			name: "multiple silences, one matches",
			silences: []types.Silence{
				{
					Labels:  map[string]string{"service": "api"},
					Comment: "api maintenance",
				},
				{
					Labels:  map[string]string{"env": "staging"},
					Comment: "staging maintenance",
				},
			},
			labels:   model.LabelSet{"service": "api", "env": "prod"},
			wantMute: true,
		},
		{
			name: "no silence matches",
			silences: []types.Silence{
				{
					Labels:  map[string]string{"service": "api"},
					Comment: "api maintenance",
				},
				{
					Labels:  map[string]string{"env": "staging"},
					Comment: "staging maintenance",
				},
			},
			labels:   model.LabelSet{"service": "db", "env": "prod"},
			wantMute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewSilenceStore(tt.silences)
			got := store.Mutes(context.Background(), tt.labels)
			require.Equal(t, tt.wantMute, got)
		})
	}
}

func TestSilenceStore_Reset(t *testing.T) {
	initial := []types.Silence{
		{
			Labels:  map[string]string{"service": "api"},
			Comment: "api maintenance",
		},
	}

	store := NewSilenceStore(initial)
	require.True(t, store.Mutes(context.Background(), model.LabelSet{"service": "api"}))

	store.Reset([]types.Silence{})
	require.False(t, store.Mutes(context.Background(), model.LabelSet{"service": "api"}))
}
