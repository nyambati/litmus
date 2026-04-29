package snapshot

import (
	"context"
	"reflect"
	"testing"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
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
		Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
		Tags:   []string{"regression"},
	}

	require.Equal(t, "Route to api-team", tc.Name)
	require.Equal(t, "regression", tc.Type)
	require.Len(t, tc.Expect.Receivers, 1)
	require.Contains(t, tc.Expect.Receivers, "api-team")
}

func TestSnapshotSynthesizer_SkipsNegativeOnlyRouteWithReason(t *testing.T) {
	matcher, err := labels.NewMatcher(labels.MatchNotEqual, "team", "ops")
	require.NoError(t, err)

	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "non-ops",
				Matchers: config.Matchers{matcher},
			},
		},
	}

	runner := pipeline.NewRunner(
		stores.NewSilenceStore([]types.Silence{}),
		stores.NewAlertStore(),
		pipeline.NewRouter(root),
		nil,
	)

	paths := NewRouteWalker(root).FindTerminalPaths()
	synth := NewSnapshotSynthesizer(runner)
	results, err := synth.DiscoverOutcomes(context.Background(), paths)
	require.NoError(t, err)
	require.Empty(t, results, "negative-only routes are not synthesizable and should be skipped")

	diagnosticsMethod := reflect.ValueOf(synth).MethodByName("Diagnostics")
	require.True(t, diagnosticsMethod.IsValid(), "snapshot synthesizer should surface skip diagnostics")

	diagnostics := diagnosticsMethod.Call(nil)
	require.Len(t, diagnostics, 1)
	require.Greater(t, diagnostics[0].Len(), 0, "skip diagnostics should explain why no outcome was synthesized")
}

func TestSnapshotSynthesizer_WarnsWhenNegativeMatchersReduceCoverage(t *testing.T) {
	positive, err := labels.NewMatcher(labels.MatchEqual, "service", "api")
	require.NoError(t, err)
	negative, err := labels.NewMatcher(labels.MatchNotRegexp, "env", "prod|staging")
	require.NoError(t, err)

	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "api-non-prod",
				Matchers: config.Matchers{positive, negative},
			},
		},
	}

	runner := pipeline.NewRunner(
		stores.NewSilenceStore([]types.Silence{}),
		stores.NewAlertStore(),
		pipeline.NewRouter(root),
		nil,
	)

	paths := NewRouteWalker(root).FindTerminalPaths()
	synth := NewSnapshotSynthesizer(runner)
	results, err := synth.DiscoverOutcomes(context.Background(), paths)
	require.NoError(t, err)
	require.NotEmpty(t, results, "positive constraints should still produce a synthesized outcome")

	var childResult *SynthesisResult
	for _, result := range results {
		if len(result.Receivers) == 1 && result.Receivers[0] == "api-non-prod" {
			childResult = result
			break
		}
	}
	require.NotNil(t, childResult, "the partially synthesizable child route should still produce an outcome")

	resultValue := reflect.ValueOf(childResult).Elem()
	warningsField := resultValue.FieldByName("Warnings")
	require.True(t, warningsField.IsValid(), "SynthesisResult should report incomplete coverage warnings")
	require.Greater(t, warningsField.Len(), 0, "partial synthesis should produce a warning when negative matchers were ignored")
}
