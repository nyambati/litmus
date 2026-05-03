package types

// RegressionState holds the current active baseline ID and its tests.
type RegressionState struct {
	ID    string      `yaml:"id"`
	Tests []*TestCase `yaml:"tests"`
}

// RegressionDiff holds the comparison result between two regression baselines.
type RegressionDiff struct {
	Deltas []RegressionDelta `json:"deltas"`
}

// RegressionDelta represents a single change between two regression baselines.
type RegressionDelta struct {
	Kind     DeltaKind         `json:"kind"`
	Labels   map[string]string `json:"labels"`
	Expected []string          `json:"expected,omitempty"`
	Actual   []string          `json:"actual,omitempty"`
}

// DeltaKind represents the kind of change in a regression delta.
type DeltaKind string

const (
	DeltaAdded    DeltaKind = "added"
	DeltaModified DeltaKind = "modified"
	DeltaRemoved  DeltaKind = "removed"
)
