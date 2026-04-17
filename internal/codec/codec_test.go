package codec

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/nyambati/litmus/internal/types"
)

func TestRegressionTestRoundTrip(t *testing.T) {
	test := types.RegressionTest{
		Name: "Test Regression",
		Labels: []map[string]string{
			{"service": "mysql", "env": "prod"},
		},
		Expected: []string{"receiver-1"},
		Tags:     []string{"regression"},
	}

	t.Run("YAML", func(t *testing.T) {
		var buf bytes.Buffer
		if err := EncodeYAML(&buf, test); err != nil {
			t.Fatalf("Failed to encode YAML: %v", err)
		}

		var decoded types.RegressionTest
		if err := DecodeYAML(&buf, &decoded); err != nil {
			t.Fatalf("Failed to decode YAML: %v", err)
		}

		if !reflect.DeepEqual(test, decoded) {
			t.Errorf("Mismatch after YAML round-trip.\nExpected: %+v\nGot:      %+v", test, decoded)
		}
	})

	t.Run("MsgPack", func(t *testing.T) {
		var buf bytes.Buffer
		if err := EncodeMsgPack(&buf, test); err != nil {
			t.Fatalf("Failed to encode MsgPack: %v", err)
		}

		var decoded types.RegressionTest
		if err := DecodeMsgPack(&buf, &decoded); err != nil {
			t.Fatalf("Failed to decode MsgPack: %v", err)
		}

		if !reflect.DeepEqual(test, decoded) {
			t.Errorf("Mismatch after MsgPack round-trip.\nExpected: %+v\nGot:      %+v", test, decoded)
		}
	})
}

func TestBehavioralTestRoundTrip(t *testing.T) {
	test := types.BehavioralTest{
		Name: "Test Behavioral",
		Tags: []string{"unit"},
		State: &types.SystemState{
			ActiveAlerts: []types.AlertSample{
				{Labels: map[string]string{"alertname": "NodeDown"}},
			},
			Silences: []types.Silence{}, // Ensure initialized
		},
		Alert: types.AlertSample{
			Labels: map[string]string{"alertname": "HighLatency"},
		},
		Expect: types.BehavioralExpect{
			Outcome: "inhibited",
		},
	}

	t.Run("YAML", func(t *testing.T) {
		var buf bytes.Buffer
		if err := EncodeYAML(&buf, test); err != nil {
			t.Fatalf("Failed to encode YAML: %v", err)
		}

		var decoded types.BehavioralTest
		if err := DecodeYAML(&buf, &decoded); err != nil {
			t.Fatalf("Failed to decode YAML: %v", err)
		}

		// Use a manual comparison for pointers if reflect.DeepEqual is problematic
		if test.Name != decoded.Name || !reflect.DeepEqual(test.Tags, decoded.Tags) {
			t.Errorf("Top level mismatch.\nExpected: %+v\nGot:      %+v", test, decoded)
		}

		if (test.State == nil) != (decoded.State == nil) {
			t.Errorf("State nil mismatch.")
		} else if test.State != nil {
			if !reflect.DeepEqual(*test.State, *decoded.State) {
				t.Errorf("State value mismatch.\nExpected: %+v\nGot:      %+v", *test.State, *decoded.State)
			}
		}

		if !reflect.DeepEqual(test.Alert, decoded.Alert) || !reflect.DeepEqual(test.Expect, decoded.Expect) {
			t.Errorf("Alert/Expect mismatch.")
		}
	})
}
