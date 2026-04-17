package behavioral

import (
	"context"
	"testing"

	"github.com/nyambati/litmus/internal/types"
	"github.com/stretchr/testify/require"
)

func TestBehavioralTestExecutor_Execute_Active(t *testing.T) {
	executor := NewBehavioralTestExecutor()

	test := &types.BehavioralTest{
		Name: "Alert routes to api-team",
		Tags: []string{"routing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences:     []types.Silence{},
		},
		Alert: types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"api-team"},
		},
	}

	result := executor.Execute(context.Background(), test, []string{"api-team"})
	require.True(t, result.Pass)
	require.Equal(t, "", result.Error)
}

func TestBehavioralTestExecutor_Execute_Silenced(t *testing.T) {
	executor := NewBehavioralTestExecutor()

	test := &types.BehavioralTest{
		Name: "Alert is silenced during maintenance",
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
		Alert: types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: types.BehavioralExpect{
			Outcome: "silenced",
		},
	}

	result := executor.Execute(context.Background(), test, []string{"api-team"})
	require.True(t, result.Pass)
}

func TestBehavioralTestExecutor_Execute_Silenced_Mismatch(t *testing.T) {
	executor := NewBehavioralTestExecutor()

	test := &types.BehavioralTest{
		Name: "Alert should be silenced but is not",
		Tags: []string{"silencing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences: []types.Silence{
				{
					Labels: map[string]string{"service": "db"},
				},
			},
		},
		Alert: types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: types.BehavioralExpect{
			Outcome: "silenced",
		},
	}

	result := executor.Execute(context.Background(), test, []string{"api-team"})
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "silenced")
}

func TestBehavioralTestExecutor_Execute_Inhibited(t *testing.T) {
	executor := NewBehavioralTestExecutor()

	test := &types.BehavioralTest{
		Name: "Alert is inhibited by critical alert",
		Tags: []string{"inhibition"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{
				{
					Labels: map[string]string{"service": "api"},
				},
			},
			Silences: []types.Silence{},
		},
		Alert: types.AlertSample{
			Labels: map[string]string{"service": "api", "severity": "warning"},
		},
		Expect: types.BehavioralExpect{
			Outcome: "inhibited",
		},
	}

	result := executor.Execute(context.Background(), test, []string{"api-team"})
	require.True(t, result.Pass)
}

func TestBehavioralTestExecutor_Execute_Receivers_Mismatch(t *testing.T) {
	executor := NewBehavioralTestExecutor()

	test := &types.BehavioralTest{
		Name: "Alert should route to specific receivers",
		Tags: []string{"routing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences:     []types.Silence{},
		},
		Alert: types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: types.BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"wrong-team"},
		},
	}

	result := executor.Execute(context.Background(), test, []string{"api-team"})
	require.False(t, result.Pass)
	require.Contains(t, result.Error, "receivers")
}

func TestBehavioralTestExecutor_Execute_OutcomeOnly(t *testing.T) {
	executor := NewBehavioralTestExecutor()

	test := &types.BehavioralTest{
		Name: "Alert is active",
		Tags: []string{"routing"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{},
			Silences:     []types.Silence{},
		},
		Alert: types.AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: types.BehavioralExpect{
			Outcome: "active",
		},
	}

	result := executor.Execute(context.Background(), test, []string{"api-team"})
	require.True(t, result.Pass)
}
