package types

// RegressionTest represents a machine-generated "Golden Reality" baseline.
// It is grouped by Outcome (the ordered list of receivers).
type RegressionTest struct {
	Name     string              `msgpack:"name" yaml:"name"`
	Labels   []map[string]string `msgpack:"labels" yaml:"labels"`
	Expected []string            `msgpack:"expect" yaml:"expect"`
	Tags     []string            `msgpack:"tags" yaml:"tags"`
}
