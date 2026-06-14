package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nyambati/litmus/internal/engine/snapshot"
	"gopkg.in/yaml.v3"
)

// InspectOptions bundles the inputs for an inspect run.
type InspectOptions struct {
	Path   string
	Format string
}

// RunInspect loads a msgpack baseline and prints it as YAML or JSON.
func RunInspect(opts *InspectOptions) error {
	tests, err := snapshot.LoadBaseline(opts.Path)
	if err != nil {
		return fmt.Errorf("loading baseline: %w", err)
	}

	switch opts.Format {
	case "json":
		data, err := json.MarshalIndent(tests, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Fprintln(os.Stdout, string(data))
	case "yaml":
		data, err := yaml.Marshal(tests)
		if err != nil {
			return fmt.Errorf("marshaling YAML: %w", err)
		}
		fmt.Fprintln(os.Stdout, string(data))
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}

	return nil
}
