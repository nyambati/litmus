package cli

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// RunInspect loads a msgpack baseline and prints it as YAML or JSON.
func RunInspect(filePath, format string) error {
	tests, err := LoadBaseline(filePath)
	if err != nil {
		return fmt.Errorf("loading baseline: %w", err)
	}

	if format == "json" {
		data, err := json.MarshalIndent(tests, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		data, err := yaml.Marshal(tests)
		if err != nil {
			return fmt.Errorf("marshaling YAML: %w", err)
		}
		fmt.Println(string(data))
	}

	return nil
}
