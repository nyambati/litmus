package snapshot

import (
	"context"
	"testing"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/stretchr/testify/require"
)

func TestRegressionTestExecutor_Execute_Pass(t *testing.T) {
	executor := NewRegressionTestExecutor()
	tests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route to api-team",
			Labels:   []map[string]string{{"service": "api"}},
			Expected: []string{"api-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.Execute(context.Background(), tests, router)

	require.Len(t, results, 1)
	require.True(t, results[0].Pass)
	require.Empty(t, results[0].Error)
}

func TestRegressionTestExecutor_Execute_Fail_WrongReceivers(t *testing.T) {
	executor := NewRegressionTestExecutor()
	tests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Route to api-team",
			Labels:   []map[string]string{{"service": "api"}},
			Expected: []string{"wrong-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.Execute(context.Background(), tests, router)

	require.Len(t, results, 1)
	require.False(t, results[0].Pass)
	require.NotEmpty(t, results[0].Labels)
	require.Equal(t, []string{"wrong-team"}, results[0].Expected)
	require.Equal(t, []string{"api-team"}, results[0].Actual)
}

func TestRegressionTestExecutor_Execute_Fail_NoLabels(t *testing.T) {
	executor := NewRegressionTestExecutor()
	tests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Empty labels test",
			Labels:   nil,
			Expected: []string{"api-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.Execute(context.Background(), tests, router)

	require.Len(t, results, 1)
	require.False(t, results[0].Pass)
	require.NotEmpty(t, results[0].Error)
}

func TestRegressionTestExecutor_Execute_MultipleTests(t *testing.T) {
	executor := NewRegressionTestExecutor()
	tests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "Pass case",
			Labels:   []map[string]string{{"service": "api"}},
			Expected: []string{"api-team"},
		},
		{
			Type:     "regression",
			Name:     "Fail case",
			Labels:   []map[string]string{{"service": "api"}},
			Expected: []string{"wrong-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.Execute(context.Background(), tests, router)

	require.Len(t, results, 2)
	require.True(t, results[0].Pass)
	require.False(t, results[1].Pass)
}

func TestRegressionTestExecutor_Execute_MultipleLabels(t *testing.T) {
	executor := NewRegressionTestExecutor()
	tests := []*types.TestCase{
		{
			Type: "regression",
			Name: "Multiple label sets all route to same receiver",
			Labels: []map[string]string{
				{"service": "api"},
				{"service": "api", "severity": "critical"},
			},
			Expected: []string{"api-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.Execute(context.Background(), tests, router)

	require.Len(t, results, 1)
	require.True(t, results[0].Pass)
}

func TestRegressionTestExecutor_Execute_ResultType(t *testing.T) {
	executor := NewRegressionTestExecutor()
	tests := []*types.TestCase{
		{
			Type:     "regression",
			Name:     "type preserved in result",
			Labels:   []map[string]string{{"service": "api"}},
			Expected: []string{"api-team"},
		},
	}

	router := pipeline.NewRouter(&amconfig.Route{Receiver: "api-team"})
	results := executor.Execute(context.Background(), tests, router)

	require.Len(t, results, 1)
	require.Equal(t, "regression", results[0].Type)
	require.Equal(t, "type preserved in result", results[0].Name)
}
