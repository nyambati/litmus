package types

// BehavioralTest represents a human-authored "Intent" scenario.
type BehavioralTest struct {
	Name   string           `msgpack:"name" yaml:"name"`
	Tags   []string         `msgpack:"tags" yaml:"tags"`
	State  *SystemState     `msgpack:"state,omitempty" yaml:"state,omitempty"`
	Alert  AlertSample      `msgpack:"alert" yaml:"alert"`
	Expect BehavioralExpect `msgpack:"expect" yaml:"expect"`
}

// SystemState represents the environment for suppression testing.
type SystemState struct {
	ActiveAlerts []AlertSample `msgpack:"active_alerts" yaml:"active_alerts"`
	Silences     []Silence     `msgpack:"silences" yaml:"silences"`
}

// AlertSample represents a firing alert with its label set.
type AlertSample struct {
	Labels map[string]string `msgpack:"labels" yaml:"labels"`
}

// Silence represents a maintenance window.
type Silence struct {
	Labels  map[string]string `msgpack:"labels" yaml:"labels"`
	Comment string            `msgpack:"comment" yaml:"comment"`
}

// BehavioralExpect defines the expected outcome of the simulation.
type BehavioralExpect struct {
	Outcome   string   `msgpack:"outcome" yaml:"outcome"`
	Receivers []string `msgpack:"receivers,omitempty" yaml:"receivers,omitempty"`
}
