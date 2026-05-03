package cli

import (
	"testing"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	"github.com/nyambati/litmus/internal/workspace"
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

func TestRunSanityChecks_NegativeOnlyRoutesMode(t *testing.T) {
	matcher, err := labels.NewMatcher(labels.MatchNotEqual, "team", "ops")
	require.NoError(t, err)

	amCfg := &amconfig.Config{
		Route: &amconfig.Route{Receiver: "default", Routes: []*amconfig.Route{
			{Receiver: "non-ops", Matchers: amconfig.Matchers{matcher}},
		}},
		Receivers: []amconfig.Receiver{{Name: "default"}, {Name: "non-ops"}},
	}

	receiversMap := map[string]*amconfig.Receiver{
		"default": {Name: "default"},
		"non-ops": {Name: "non-ops"},
	}
	ctx := sanity.CheckContext{
		Route:     amCfg.Route,
		Receivers: receiversMap,
	}

	findCheck := func(result sanity.Result, name config.SanityCheck) *sanity.CheckEntry {
		for i := range result.Checks {
			if result.Checks[i].Name == name {
				return &result.Checks[i]
			}
		}
		return nil
	}

	t.Run("fail mode fails sanity", func(t *testing.T) {
		result := sanity.Run(ctx, config.SanityConfig{
			OrphanReceivers:    config.SanityModeFail,
			DeadReceivers:      config.SanityModeFail,
			ShadowedRoutes:     config.SanityModeFail,
			InhibitionCycles:   config.SanityModeFail,
			NegativeOnlyRoutes: config.SanityModeFail,
		})

		require.False(t, result.Passed)
		entry := findCheck(result, "negative_only_routes")
		require.NotNil(t, entry)
		require.Len(t, entry.Issues, 1)
		require.Equal(t, string(config.SanityModeFail), entry.Mode)
	})

	t.Run("warn mode reports without failing sanity", func(t *testing.T) {
		result := sanity.Run(ctx, config.SanityConfig{
			OrphanReceivers:    config.SanityModeFail,
			DeadReceivers:      config.SanityModeFail,
			ShadowedRoutes:     config.SanityModeFail,
			InhibitionCycles:   config.SanityModeFail,
			NegativeOnlyRoutes: config.SanityModeWarn,
		})

		require.True(t, result.Passed)
		entry := findCheck(result, "negative_only_routes")
		require.NotNil(t, entry)
		require.Len(t, entry.Issues, 1)
		require.Equal(t, string(config.SanityModeWarn), entry.Mode)
	})
}

func TestBuildCheckContext_MapsReceiversAndRules(t *testing.T) {
	route := &amconfig.Route{Receiver: "default"}
	amCfg := &amconfig.Config{
		Route: route,
		Receivers: []amconfig.Receiver{
			{Name: "default"},
			{Name: "critical"},
		},
		InhibitRules: []amconfig.InhibitRule{
			{},
		},
	}
	frags := []*fragment.Fragment{{Namespace: "db"}}
	policy := config.PolicyConfig{Require: config.RequireConfig{Tests: true}}
	ws := &workspace.Workspace{Fragments: frags}

	ctx := buildCheckContext(amCfg, ws, policy)

	require.Equal(t, route, ctx.Route)
	require.Len(t, ctx.Receivers, 2)
	require.Contains(t, ctx.Receivers, "default")
	require.Contains(t, ctx.Receivers, "critical")
	// Receivers map holds pointers into the slice — verify identity.
	require.Equal(t, &amCfg.Receivers[0], ctx.Receivers["default"])
	require.Equal(t, &amCfg.Receivers[1], ctx.Receivers["critical"])
	require.Len(t, ctx.Rules, 1)
	require.Equal(t, &amCfg.InhibitRules[0], ctx.Rules[0])
	require.Equal(t, frags, ctx.Fragments)
	require.Equal(t, policy, ctx.Policy)
}

func TestBuildCheckContext_EmptyReceiversAndRules(t *testing.T) {
	amCfg := &amconfig.Config{Route: &amconfig.Route{Receiver: "default"}}
	ws := &workspace.Workspace{}

	ctx := buildCheckContext(amCfg, ws, config.PolicyConfig{})

	require.Empty(t, ctx.Receivers)
	require.Empty(t, ctx.Rules)
	require.Nil(t, ctx.Fragments)
}

func TestRunBehavioralTests_NoLitmusConfigParameter(t *testing.T) {
	// Verifies that RunBehavioralTests does NOT take a *config.LitmusConfig argument.
	// The parameter was present but never used; this test enforces the cleaned-up signature.
	result := RunBehavioralTests(
		t.Context(),
		[]*types.TestCase{{Name: "t1", Type: "unit"}},
		nil,                         // router — nil is fine for empty tests after tag filter
		nil,                         // inhibitRules
		[]string{"nonexistent-tag"}, // filter that matches nothing
	)
	require.True(t, result.Passed, "no tests run after tag filter should be a pass")
	require.Equal(t, 0, result.Tests)
}

func TestPrintCheckResult_FailedRegressionShowsFAILPrefix(t *testing.T) {
	result := CheckResult{
		Regression: RegressionResult{
			Passed:     false,
			TotalTests: 2,
			Tests:      2,
			PassCount:  1,
			Failures: []TestFailure{
				{Name: "route-to-db", Type: "regression", Error: "expected [db-team], got [default]"},
			},
		},
	}

	out := captureStdout(t, func() { PrintCheckResult(result, false) })

	require.Contains(t, out, "[FAIL]  1/2 cases passed",
		"failed regression summary line must use [FAIL] prefix")
	require.NotContains(t, out, "[PASS]  1/2 cases passed",
		"failed regression summary must not use [PASS] prefix")
}

func TestPrintCheckResult_FailedBehavioralShowsFAILPrefix(t *testing.T) {
	result := CheckResult{
		Behavioral: BehavioralResult{
			Passed:     false,
			TotalTests: 3,
			Tests:      3,
			PassCount:  2,
			Failures: []TestFailure{
				{Name: "mysql-routes-to-db", Type: "unit", Error: "expected [db-team], got [default]"},
			},
		},
	}

	out := captureStdout(t, func() { PrintCheckResult(result, false) })

	require.Contains(t, out, "[FAIL]  2/3 unit tests passed",
		"failed behavioral summary line must use [FAIL] prefix")
	require.NotContains(t, out, "[PASS]  2/3 unit tests passed",
		"failed behavioral summary must not use [PASS] prefix")
}
