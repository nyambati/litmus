package cli

import (
	"context"
	"os"
	"testing"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestFilterByTags_NoTags(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
	}

	result := filterByTags(tests, []string{})
	require.Equal(t, tests, result)
}

func TestFilterByTags_SingleTag(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
		{Name: "test3", Tags: []string{"critical", "smoke"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 2)
	require.Equal(t, "test1", result[0].Name)
	require.Equal(t, "test3", result[1].Name)
}

func TestFilterByTags_MultipleTagsOr(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
		{Name: "test3", Tags: []string{"critical", "smoke"}},
		{Name: "test4", Tags: []string{"e2e"}},
	}

	result := filterByTags(tests, []string{"critical", "smoke"})
	require.Len(t, result, 3)
	require.Equal(t, "test1", result[0].Name)
	require.Equal(t, "test2", result[1].Name)
	require.Equal(t, "test3", result[2].Name)
}

func TestFilterByTags_NoMatches(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
	}

	result := filterByTags(tests, []string{"nonexistent"})
	require.Len(t, result, 0)
}

func TestFilterByTags_NoTagsOnTest(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{}},
		{Name: "test2", Tags: []string{"critical"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 1)
	require.Equal(t, "test2", result[0].Name)
}

func TestFilterByTags_NilTagsOnTest(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: nil},
		{Name: "test2", Tags: []string{"critical"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 1)
	require.Equal(t, "test2", result[0].Name)
}

func TestRunBehavioralTests_DoesNotDuplicateRootTests(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })
	require.NoError(t, os.Chdir(tmpDir))

	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  fragments: "fragments/*"
`), 0600))
	require.NoError(t, os.MkdirAll("config/tests", 0755))
	require.NoError(t, os.WriteFile("config/alertmanager.yml", []byte(`
route:
  receiver: "default"
receivers:
  - name: "default"
`), 0600))
	require.NoError(t, os.WriteFile("config/tests/root.yml", []byte(`
name: "root behavioral test"
alert:
  labels:
    service: "api"
expect:
  outcome: "active"
  receivers:
    - "default"
`), 0600))

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	_, fragments, amCfg, err := cfg.LoadAssembledConfig()
	require.NoError(t, err)

	result := RunBehavioralTests(
		context.Background(),
		cfg,
		fragments,
		pipeline.NewRouter(amCfg.Route),
		amCfg.InhibitRules,
		nil,
	)

	require.Equal(t, 1, result.TotalTests)
	require.Equal(t, 1, result.Tests)
	require.Equal(t, 1, result.PassCount)
	require.Empty(t, result.Failures)
}

func TestRunSanityChecks_NegativeOnlyRoutesMode(t *testing.T) {
	matcher, err := labels.NewMatcher(labels.MatchNotEqual, "team", "ops")
	require.NoError(t, err)

	amCfg := &amconfig.Config{
		Route: &amconfig.Route{Receiver: "default", Routes: []*amconfig.Route{
			{Receiver: "non-ops", Matchers: amconfig.Matchers{matcher}},
		}},
		Receivers: []amconfig.Receiver{{Name: "default"}, {Name: "non-ops"}},
	}

	t.Run("fail mode fails sanity", func(t *testing.T) {
		result := RunSanityChecks(amCfg, config.SanityConfig{
			OrphanReceivers:    config.SanityModeFail,
			DeadReceivers:      config.SanityModeFail,
			ShadowedRoutes:     config.SanityModeFail,
			InhibitionCycles:   config.SanityModeFail,
			NegativeOnlyRoutes: config.SanityModeFail,
		})

		require.False(t, result.Passed)
		require.Len(t, result.NegativeOnlyRouteIssues, 1)
		require.Equal(t, string(config.SanityModeFail), result.NegativeOnlyRouteMode)
	})

	t.Run("warn mode reports without failing sanity", func(t *testing.T) {
		result := RunSanityChecks(amCfg, config.SanityConfig{
			OrphanReceivers:    config.SanityModeFail,
			DeadReceivers:      config.SanityModeFail,
			ShadowedRoutes:     config.SanityModeFail,
			InhibitionCycles:   config.SanityModeFail,
			NegativeOnlyRoutes: config.SanityModeWarn,
		})

		require.True(t, result.Passed)
		require.Len(t, result.NegativeOnlyRouteIssues, 1)
		require.Equal(t, string(config.SanityModeWarn), result.NegativeOnlyRouteMode)
	})
}
