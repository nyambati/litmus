package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTestCase_JSON_UnitType(t *testing.T) {
	tc := &TestCase{
		Name: "alert routes to api-team",
		Type: "unit",
		Tags: []string{"routing"},
		Alert: &AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"api-team"},
		},
	}

	data, err := json.Marshal(tc)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	require.Equal(t, "unit", out["type"])
	require.Equal(t, "alert routes to api-team", out["name"])
	require.NotNil(t, out["alert"])
	require.NotNil(t, out["expect"])

	// regression fields must be absent
	require.Nil(t, out["labels"])
	require.Nil(t, out["expected"])
}

func TestTestCase_JSON_RegressionType(t *testing.T) {
	tc := &TestCase{
		Name: "route to api-team",
		Type: "regression",
		Tags: []string{"regression"},
		Labels: []map[string]string{
			{"service": "api", "severity": "critical"},
		},
		Expected: []string{"api-team"},
	}

	data, err := json.Marshal(tc)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	require.Equal(t, "regression", out["type"])
	require.NotNil(t, out["labels"])
	require.NotNil(t, out["expected"])

	// unit fields must be absent
	require.Nil(t, out["alert"])
	require.Nil(t, out["expect"])
	require.Nil(t, out["state"])
}

func TestTestCase_JSON_RoundTrip_Unit(t *testing.T) {
	original := &TestCase{
		Name: "alert silenced during maintenance",
		Type: "unit",
		Tags: []string{"silencing"},
		State: &SystemState{
			Silences: []Silence{
				{Labels: map[string]string{"service": "api"}, Comment: "maintenance"},
			},
		},
		Alert: &AlertSample{
			Labels: map[string]string{"service": "api"},
		},
		Expect: &BehavioralExpect{
			Outcome: "silenced",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded TestCase
	require.NoError(t, json.Unmarshal(data, &decoded))

	require.Equal(t, original.Name, decoded.Name)
	require.Equal(t, original.Type, decoded.Type)
	require.Equal(t, original.Tags, decoded.Tags)
	require.NotNil(t, decoded.State)
	require.Len(t, decoded.State.Silences, 1)
	require.NotNil(t, decoded.Alert)
	require.Equal(t, original.Alert.Labels, decoded.Alert.Labels)
	require.NotNil(t, decoded.Expect)
	require.Equal(t, original.Expect.Outcome, decoded.Expect.Outcome)
}

func TestTestCase_JSON_RoundTrip_Regression(t *testing.T) {
	original := &TestCase{
		Name: "route to db-team",
		Type: "regression",
		Labels: []map[string]string{
			{"service": "db"},
			{"service": "db", "severity": "critical"},
		},
		Expected: []string{"db-team", "ops-team"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded TestCase
	require.NoError(t, json.Unmarshal(data, &decoded))

	require.Equal(t, original.Name, decoded.Name)
	require.Equal(t, original.Type, decoded.Type)
	require.Equal(t, original.Labels, decoded.Labels)
	require.Equal(t, original.Expected, decoded.Expected)
}

func TestTestCase_YAML_RoundTrip_Unit(t *testing.T) {
	original := &TestCase{
		Name: "alert routes to team",
		Type: "unit",
		Tags: []string{"routing"},
		Alert: &AlertSample{
			Labels: map[string]string{"team": "backend"},
		},
		Expect: &BehavioralExpect{
			Outcome:   "active",
			Receivers: []string{"backend-team"},
		},
	}

	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	var decoded TestCase
	require.NoError(t, yaml.Unmarshal(data, &decoded))

	require.Equal(t, original.Name, decoded.Name)
	require.Equal(t, original.Type, decoded.Type)
	require.NotNil(t, decoded.Alert)
	require.Equal(t, original.Alert.Labels, decoded.Alert.Labels)
	require.NotNil(t, decoded.Expect)
	require.Equal(t, original.Expect.Receivers, decoded.Expect.Receivers)
}

func TestTestCase_YAML_RoundTrip_Regression(t *testing.T) {
	original := &TestCase{
		Name: "route to api-team",
		Type: "regression",
		Labels: []map[string]string{
			{"service": "api"},
		},
		Expected: []string{"api-team"},
		Tags:     []string{"regression"},
	}

	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	var decoded TestCase
	require.NoError(t, yaml.Unmarshal(data, &decoded))

	require.Equal(t, original.Name, decoded.Name)
	require.Equal(t, original.Type, decoded.Type)
	require.Equal(t, original.Labels, decoded.Labels)
	require.Equal(t, original.Expected, decoded.Expected)
	require.Equal(t, original.Tags, decoded.Tags)
}

func TestTestResult_JSON(t *testing.T) {
	tests := []struct {
		name   string
		result TestResult
		want   map[string]any
	}{
		{
			name: "passing unit result",
			result: TestResult{
				Name: "alert routes to api-team",
				Type: "unit",
				Pass: true,
			},
			want: map[string]any{"name": "alert routes to api-team", "type": "unit", "pass": true},
		},
		{
			name: "failing regression result",
			result: TestResult{
				Name:     "route to api-team",
				Type:     "regression",
				Pass:     false,
				Labels:   map[string]string{"service": "api"},
				Expected: []string{"api-team"},
				Actual:   []string{"wrong-team"},
			},
			want: map[string]any{"name": "route to api-team", "type": "regression", "pass": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			require.NoError(t, err)

			var out map[string]any
			require.NoError(t, json.Unmarshal(data, &out))

			for k, v := range tt.want {
				require.Equal(t, v, out[k])
			}
		})
	}
}

func TestTestResult_OmitEmpty(t *testing.T) {
	result := TestResult{
		Name: "passing test",
		Type: "unit",
		Pass: true,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))

	// error, labels, expected, actual must be absent when empty
	require.Nil(t, out["error"])
	require.Nil(t, out["labels"])
	require.Nil(t, out["expected"])
	require.Nil(t, out["actual"])
}
