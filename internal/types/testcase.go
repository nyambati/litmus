package types

// TestCase is the unified representation for both unit and regression tests.
// Set Type to "unit" for human-authored behavioral scenarios,
// "regression" for machine-generated routing baselines.
type TestCase struct {
	Name string   `json:"name"            yaml:"name"            msgpack:"name"`
	Type string   `json:"type"            yaml:"type"            msgpack:"type"`
	Tags []string `json:"tags,omitempty"  yaml:"tags,omitempty"  msgpack:"tags,omitempty"`

	// unit-only fields
	State  *SystemState      `json:"state,omitempty"  yaml:"state,omitempty"  msgpack:"state,omitempty"`
	Alert  *AlertSample      `json:"alert,omitempty"  yaml:"alert,omitempty"  msgpack:"alert,omitempty"`
	Expect *BehavioralExpect `json:"expect,omitempty" yaml:"expect,omitempty" msgpack:"expect,omitempty"`

	// regression-only fields
	Labels []map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" msgpack:"labels,omitempty"`
}

// TestResult is the unified execution result for both unit and regression tests.
type TestResult struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Pass     bool              `json:"pass"`
	Error    string            `json:"error,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	Expected []string          `json:"expected,omitempty"`
	Actual   []string          `json:"actual,omitempty"`
}
