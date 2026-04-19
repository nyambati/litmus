package snapshot

import (
	"context"
	"testing"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestSnapshotSynthesizer_DiscoverOutcomes(t *testing.T) {
	// Create pipeline runner
	silenceStore := stores.NewSilenceStore([]types.Silence{})
	alertStore := stores.NewAlertStore()
	router := pipeline.NewRouter(&config.Route{Receiver: "api-team"})
	runner := pipeline.NewRunner(silenceStore, alertStore, router, nil)

	// Create test route paths
	paths := []*RoutePath{
		{
			Receiver: "api-team",
			Matchers: []model.LabelSet{
				{"service": "api"},
			},
		},
		{
			Receiver: "db-team",
			Matchers: []model.LabelSet{
				{"service": "db"},
			},
		},
	}

	synth := NewSnapshotSynthesizer(runner)
	results, err := synth.DiscoverOutcomes(context.Background(), paths)
	require.NoError(t, err)

	// Should synthesize outcomes from paths
	require.GreaterOrEqual(t, len(results), 1)

	// Each result should have receivers and labels
	for _, result := range results {
		require.NotEmpty(t, result.Receivers)
		require.NotEmpty(t, result.Labels)
		require.Contains(t, result.Receivers, "api-team")
	}
}

func TestTestCase_Regression_Roundtrip(t *testing.T) {
	tc := &types.TestCase{
		Type: "regression",
		Name: "Route to api-team",
		Labels: []map[string]string{
			{"service": "api", "severity": "critical"},
		},
		Expected: []string{"api-team"},
		Tags:     []string{"regression"},
	}

	require.Equal(t, "Route to api-team", tc.Name)
	require.Equal(t, "regression", tc.Type)
	require.Len(t, tc.Expected, 1)
	require.Contains(t, tc.Expected, "api-team")
}
