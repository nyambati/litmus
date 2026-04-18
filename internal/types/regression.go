package types

// RegressionTest represents a machine-generated "Golden Reality" baseline.
// It is grouped by Outcome (the ordered list of receivers).
type RegressionTest struct {
	Name     string              `msgpack:"name" yaml:"name"`
	Labels   []map[string]string `msgpack:"labels" yaml:"labels"`
	Expected []string            `msgpack:"expect" yaml:"expect"`
	Tags     []string            `msgpack:"tags" yaml:"tags"`
}

// DeltaKind represents the type of change in a regression test.
type DeltaKind string

const (
	// DeltaAdded indicates a new route was discovered.
	DeltaAdded DeltaKind = "added"
	// DeltaRemoved indicates a route no longer exists.
	DeltaRemoved DeltaKind = "removed"
	// DeltaModified indicates a route's outcome has changed.
	DeltaModified DeltaKind = "modified"
)

// RegressionDelta represents a single change in routing behavior.
type RegressionDelta struct {
	Kind     DeltaKind         `json:"kind"`
	Labels   map[string]string `json:"labels"`
	Expected []string          `json:"expected"` // Old outcome for Modified/Removed
	Actual   []string          `json:"actual"`   // New outcome for Modified/Added
}

// RegressionDiff holds the collection of all behavioral changes.
type RegressionDiff struct {
	Deltas []RegressionDelta `json:"deltas"`
}
