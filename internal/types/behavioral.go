package types

// BehavioralTest represents a human-authored "Intent" scenario.
type BehavioralTest struct {
	Name   string           `json:"name" msgpack:"name" yaml:"name"`
	Type   string           `json:"type,omitempty" msgpack:"type,omitempty" yaml:"type,omitempty"`
	Tags   []string         `json:"tags,omitempty" msgpack:"tags" yaml:"tags"`
	State  *SystemState     `json:"state,omitempty" msgpack:"state,omitempty" yaml:"state,omitempty"`
	Alert  AlertSample      `json:"alert" msgpack:"alert" yaml:"alert"`
	Expect BehavioralExpect `json:"expect" msgpack:"expect" yaml:"expect"`
}

// SystemState represents the environment for suppression testing.
type SystemState struct {
	ActiveAlerts []AlertSample `json:"active_alerts,omitempty" msgpack:"active_alerts" yaml:"active_alerts"`
	Silences     []Silence     `json:"silences,omitempty" msgpack:"silences" yaml:"silences"`
}

// AlertSample represents a firing alert with its label set.
type AlertSample struct {
	Labels map[string]string `json:"labels" msgpack:"labels" yaml:"labels"`
}

// Silence represents a maintenance window.
type Silence struct {
	Labels  map[string]string `json:"labels" msgpack:"labels" yaml:"labels"`
	Comment string            `json:"comment" msgpack:"comment" yaml:"comment"`
}

// BehavioralExpect defines the expected outcome of the simulation.
type BehavioralExpect struct {
	Outcome   string   `json:"outcome" msgpack:"outcome" yaml:"outcome"`
	Receivers []string `json:"receivers,omitempty" msgpack:"receivers,omitempty" yaml:"receivers,omitempty"`
}
