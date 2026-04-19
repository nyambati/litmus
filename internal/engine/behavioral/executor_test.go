package behavioral

import (
	"context"
	"testing"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestBehavioralTestExecutor_Execute_Active(t *testing.T) {
	executor := NewBehavioralTestExecutor(nil)

	test := &types.TestCase{
		Name: "Alert routes to api-team",
		Type: "unit",
		Tags: []string{"routing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences:     []types.Silence{},
		},
		Alert: &types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"api-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
	require.Equal(t, "", result.Error)
}

func TestBehavioralTestExecutor_Execute_Silenced(t *testing.T) {
	executor := NewBehavioralTestExecutor(nil)

	test := &types.TestCase{
		Name: "Alert is silenced during maintenance",
		Type: "unit",
		Tags: []string{"silencing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences: []types.Silence{
				{
					Labels:  map[string]string{"service": "api"},
					Comment: "scheduled maintenance",
				},
			},
		},
		Alert: &types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &types.BehavioralExpect{
			Outcome: "silenced",
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestBehavioralTestExecutor_Execute_Silenced_Mismatch(t *testing.T) {
	executor := NewBehavioralTestExecutor(nil)

	test := &types.TestCase{
		Name: "Alert should be silenced but is not",
		Type: "unit",
		Tags: []string{"silencing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences: []types.Silence{
				{
					Labels: map[string]string{"service": "db"},
				},
			},
		},
		Alert: &types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &types.BehavioralExpect{
			Outcome: "silenced",
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "silenced")
}

func TestBehavioralTestExecutor_Execute_Inhibited(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "api"},
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
	}
	executor := NewBehavioralTestExecutor(rules)

	test := &types.TestCase{
		Name: "Alert is inhibited by critical alert",
		Type: "unit",
		Tags: []string{"inhibition"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{
				{
					Labels: map[string]string{"service": "api"},
				},
			},
			Silences: []types.Silence{},
		},
		Alert: &types.AlertSample{
			Labels: map[string]string{"service": "api", "severity": "warning"},
		},
		Expect: &types.BehavioralExpect{
			Outcome: "inhibited",
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestBehavioralTestExecutor_Execute_Receivers_Mismatch(t *testing.T) {
	executor := NewBehavioralTestExecutor(nil)

	test := &types.TestCase{
		Name: "Alert should route to specific receivers",
		Type: "unit",
		Tags: []string{"routing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences:     []types.Silence{},
		},
		Alert: &types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"wrong-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "receivers")
}

func TestBehavioralTestExecutor_Execute_OutcomeOnly(t *testing.T) {
	executor := NewBehavioralTestExecutor(nil)

	test := &types.TestCase{
		Name: "Alert is active",
		Type: "unit",
		Tags: []string{"routing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences:     []types.Silence{},
		},
		Alert: &types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &types.BehavioralExpect{
			Outcome: "active",
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestBehavioralTestExecutor_Execute_NilState(t *testing.T) {
	executor := NewBehavioralTestExecutor(nil)

	test := &types.TestCase{
		Name: "Alert with no state defaults to empty env",
		Type: "unit",
		Alert: &types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"api-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)
	require.True(t, result.Pass)
}

func TestBehavioralTestExecutor_Execute_ResultType(t *testing.T) {
	executor := NewBehavioralTestExecutor(nil)

	test := &types.TestCase{
		Name:  "type preserved in result",
		Type:  "unit",
		Alert: &types.AlertSample{Labels: map[string]string{"service": "api"}},
		Expect: &types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"api-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	result := executor.Execute(context.Background(), test, router)

	require.Equal(t, "unit", result.Type)
	require.Equal(t, test.Name, result.Name)
}
