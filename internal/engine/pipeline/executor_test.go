package pipeline

import (
	"context"
	"testing"

	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

// --- nil / guard cases ---

func TestTestExecutor_Execute_NilTest(t *testing.T) {
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "default"})
	result := executor.Execute(context.Background(), nil, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "nil")
}

func TestTestExecutor_Execute_NilRouter(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{Name: "t", Type: "unit", Alert: &types.AlertSample{Labels: map[string]string{}}}
	result := executor.Execute(context.Background(), test, nil)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "router")
	require.Equal(t, "t", result.Name)
}

func TestTestExecutor_ExecuteAll_EmptySlice(t *testing.T) {
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "default"})
	results := executor.ExecuteAll(context.Background(), nil, router)
	require.Empty(t, results)
}

func TestTestExecutor_ExecuteAll_SkipsNilElements(t *testing.T) {
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	tests := []*types.TestCase{
		nil,
		{
			Type:   "regression",
			Name:   "real test",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Receivers: []string{"api-team"}},
		},
		nil,
	}
	results := executor.ExecuteAll(context.Background(), tests, router)
	require.Len(t, results, 1, "nil elements must be skipped, not turned into error results")
	require.Equal(t, "real test", results[0].Name)
}

// --- dispatch ---

func TestTestExecutor_UnknownType_DispatchesToUnit(t *testing.T) {
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	// unknown type has no alert → falls to unit path → returns unit-style error
	test := &types.TestCase{Name: "t", Type: "unknown"}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "alert")
}

func TestTestExecutor_Unit_NilAlert(t *testing.T) {
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	test := &types.TestCase{Name: "t", Type: "unit", Alert: nil, Expect: &types.BehavioralExpect{Outcome: "active"}}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "alert")
}

func TestTestExecutor_Unit_NilExpect(t *testing.T) {
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	test := &types.TestCase{
		Name:   "t",
		Type:   "unit",
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: nil,
	}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "expect")
}

func TestTestExecutor_Regression_NilExpect(t *testing.T) {
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	test := &types.TestCase{
		Type:   "regression",
		Name:   "t",
		Labels: []map[string]string{{"service": "api"}},
		Expect: nil,
	}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "expect")
}

// --- match semantics: SubsetMatch (unit) vs ExactMatch (regression) ---

func TestTestExecutor_Unit_SubsetMatch_ActualIsSupersetOfExpected_Passes(t *testing.T) {
	// Unit uses SubsetMatch: actual ["pagerduty","slack"] contains expected ["slack"] → PASS.
	// Two children with continue:true on the first produces multiple receivers.
	router := NewRouter(&amconfig.Route{
		Receiver: "default",
		Routes: []*amconfig.Route{
			{Receiver: "pagerduty", Match: map[string]string{"severity": "critical"}, Continue: true},
			{Receiver: "slack", Match: map[string]string{"severity": "critical"}},
		},
	})
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Type:  "unit",
		Alert: &types.AlertSample{Labels: map[string]string{"severity": "critical"}},
		Expect: &types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"slack"}, // subset of actual ["pagerduty","slack"]
		},
	}
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass, "unit SubsetMatch: expected subset of actual must pass")
}

func TestTestExecutor_Regression_ExactMatch_ActualIsSupersetOfExpected_Fails(t *testing.T) {
	// Regression uses ExactMatch: actual ["pagerduty","slack"] ≠ expected ["slack"] → FAIL.
	// Same router as SubsetMatch test above.
	router := NewRouter(&amconfig.Route{
		Receiver: "default",
		Routes: []*amconfig.Route{
			{Receiver: "pagerduty", Match: map[string]string{"severity": "critical"}, Continue: true},
			{Receiver: "slack", Match: map[string]string{"severity": "critical"}},
		},
	})
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Type:   "regression",
		Labels: []map[string]string{{"severity": "critical"}},
		Expect: &types.BehavioralExpect{
			Receivers: []string{"slack"}, // exact match required but actual is ["pagerduty","slack"]
		},
	}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass, "regression ExactMatch: actual superset of expected must fail")
}

func TestTestExecutor_Unit_SubsetMatch_ExpectedNotFullyCovered_Fails(t *testing.T) {
	// Unit SubsetMatch: expected ["api-team","pagerduty"] but actual only ["api-team"] → FAIL.
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	test := &types.TestCase{
		Type:  "unit",
		Alert: &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"api-team", "pagerduty"},
		},
	}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
}

// --- state isolation ---

func TestTestExecutor_Regression_IgnoresState(t *testing.T) {
	// Regression test with populated State: silences and active alerts must have no effect.
	executor := NewTestExecutor(nil)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	test := &types.TestCase{
		Type: "regression",
		Name: "state must be ignored",
		// State is set but regression must not use it.
		State: &types.SystemState{
			Silences: []types.Silence{{Labels: map[string]string{"service": "api"}}},
		},
		Labels: []map[string]string{{"service": "api"}},
		Expect: &types.BehavioralExpect{Receivers: []string{"api-team"}},
	}
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass, "regression must not apply state silences")
}

func TestTestExecutor_Regression_IgnoresInhibitRules(t *testing.T) {
	// Executor has inhibit rules, but regression execution must not apply them.
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "api"},
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
	}
	executor := NewTestExecutor(rules)
	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	test := &types.TestCase{
		Type:   "regression",
		Name:   "inhibit rules must not apply",
		Labels: []map[string]string{{"service": "api", "severity": "warning"}},
		Expect: &types.BehavioralExpect{Receivers: []string{"api-team"}},
	}
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass, "regression must not apply executor inhibit rules")
}

// --- multi-label failure attribution ---

func TestTestExecutor_Regression_MultiLabel_FirstFailsReported(t *testing.T) {
	// Route: service=api → api-team, service=db → db-team.
	router := NewRouter(&amconfig.Route{
		Receiver: "default",
		Routes: []*amconfig.Route{
			{Receiver: "api-team", Match: map[string]string{"service": "api"}},
			{Receiver: "db-team", Match: map[string]string{"service": "db"}},
		},
	})
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Type: "regression",
		Name: "first failing label set is reported",
		Labels: []map[string]string{
			{"service": "api"}, // routes to api-team — PASSES
			{"service": "db"},  // routes to db-team — FAILS (expected api-team)
			{"service": "api"}, // would pass but never reached
		},
		Expect: &types.BehavioralExpect{Receivers: []string{"api-team"}},
	}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Equal(t, map[string]string{"service": "db"}, result.Labels, "must report the first failing label set")
	require.Equal(t, []string{"db-team"}, result.Actual)
}

func TestTestExecutor_Regression_MultiLabel_LastFailsReported(t *testing.T) {
	router := NewRouter(&amconfig.Route{
		Receiver: "default",
		Routes: []*amconfig.Route{
			{Receiver: "api-team", Match: map[string]string{"service": "api"}},
			{Receiver: "db-team", Match: map[string]string{"service": "db"}},
		},
	})
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Type: "regression",
		Name: "last label set fails",
		Labels: []map[string]string{
			{"service": "api"}, // PASSES
			{"service": "db"},  // FAILS
		},
		Expect: &types.BehavioralExpect{Receivers: []string{"api-team"}},
	}
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Equal(t, map[string]string{"service": "db"}, result.Labels)
}

// --- unit (behavioral) tests ---

func TestTestExecutor_Unit_Active(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Name: "Alert routes to api-team",
		Type: "unit",
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences:     []types.Silence{},
		},
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
	require.Empty(t, result.Error)
}

func TestTestExecutor_Unit_Silenced(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Name: "Alert is silenced during maintenance",
		Type: "unit",
		State: &types.SystemState{
			Silences: []types.Silence{{Labels: map[string]string{"service": "api"}, Comment: "maintenance"}},
		},
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{Outcome: "silenced"},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestTestExecutor_Unit_Silenced_Mismatch(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Name: "Alert should be silenced but is not",
		Type: "unit",
		State: &types.SystemState{
			Silences: []types.Silence{{Labels: map[string]string{"service": "db"}}},
		},
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{Outcome: "silenced"},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "silenced")
}

func TestTestExecutor_Unit_Inhibited(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "api"},
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
	}
	executor := NewTestExecutor(rules)
	test := &types.TestCase{
		Name: "Alert is inhibited by critical alert",
		Type: "unit",
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{{Labels: map[string]string{"service": "api"}}},
		},
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api", "severity": "warning"}},
		Expect: &types.BehavioralExpect{Outcome: "inhibited"},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestTestExecutor_Unit_ReceiverMismatch(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Name:   "Alert should route to specific receivers",
		Type:   "unit",
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"wrong-team"}},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "receivers")
}

func TestTestExecutor_Unit_OutcomeOnly(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Name:   "Alert is active",
		Type:   "unit",
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{Outcome: "active"},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestTestExecutor_Unit_NilState(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Name:   "Alert with no state defaults to empty env",
		Type:   "unit",
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestTestExecutor_Unit_TypePreserved(t *testing.T) {
	executor := NewTestExecutor(nil)
	test := &types.TestCase{
		Name:   "type preserved in result",
		Type:   "unit",
		Alert:  &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.Equal(t, "unit", result.Type)
	require.Equal(t, test.Name, result.Name)
}

// --- regression tests ---

func TestTestExecutor_Regression_Pass(t *testing.T) {
	executor := NewTestExecutor(nil)
	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "Route to api-team",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
		},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.ExecuteAll(context.Background(), tests, router)
	require.Len(t, results, 1)
	require.True(t, results[0].Pass)
	require.Empty(t, results[0].Error)
}

func TestTestExecutor_Regression_Fail_WrongReceivers(t *testing.T) {
	executor := NewTestExecutor(nil)
	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "Route to api-team",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"wrong-team"}},
		},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.ExecuteAll(context.Background(), tests, router)
	require.Len(t, results, 1)
	require.False(t, results[0].Pass)
	require.NotEmpty(t, results[0].Labels)
	require.Equal(t, []string{"wrong-team"}, results[0].Expected)
	require.Equal(t, []string{"api-team"}, results[0].Actual)
}

func TestTestExecutor_Regression_Fail_NoLabels(t *testing.T) {
	executor := NewTestExecutor(nil)
	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "Empty labels test",
			Labels: nil,
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
		},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.ExecuteAll(context.Background(), tests, router)
	require.Len(t, results, 1)
	require.False(t, results[0].Pass)
	require.NotEmpty(t, results[0].Error)
}

func TestTestExecutor_Regression_MultipleTests(t *testing.T) {
	executor := NewTestExecutor(nil)
	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "Pass case",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
		},
		{
			Type:   "regression",
			Name:   "Fail case",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"wrong-team"}},
		},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.ExecuteAll(context.Background(), tests, router)
	require.Len(t, results, 2)
	require.True(t, results[0].Pass)
	require.False(t, results[1].Pass)
}

func TestTestExecutor_Regression_MultipleLabels(t *testing.T) {
	executor := NewTestExecutor(nil)
	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "Multiple label sets all route to same receiver",
			Labels: []map[string]string{{"service": "api"}, {"service": "api", "severity": "critical"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
		},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.ExecuteAll(context.Background(), tests, router)
	require.Len(t, results, 1)
	require.True(t, results[0].Pass)
}

func TestTestExecutor_Regression_TypePreserved(t *testing.T) {
	executor := NewTestExecutor(nil)
	tests := []*types.TestCase{
		{
			Type:   "regression",
			Name:   "type preserved in result",
			Labels: []map[string]string{{"service": "api"}},
			Expect: &types.BehavioralExpect{Outcome: "active", Receivers: []string{"api-team"}},
		},
	}

	router := NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.ExecuteAll(context.Background(), tests, router)
	require.Len(t, results, 1)
	require.Equal(t, "regression", results[0].Type)
	require.Equal(t, "type preserved in result", results[0].Name)
}
