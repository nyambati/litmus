package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	defaultLitmusYAML = `# Litmus configuration
# Path to the Alertmanager configuration file to be tested
config_file: "alertmanager.yml"

# Global Test Labels: These labels are automatically added to
# EVERY synthesized alert during the 'snapshot' process.
global_labels:
  severity: "warning"
  cluster: "production"

# Regression Settings
regression:
  # Maximum label combinations per route before balanced coverage
  max_samples: 5
  # Path to the MessagePack baseline
  baseline_path: "regressions.litmus.mpk"

# Behavioral Unit Test (BUT) Settings
tests:
  # Directory where human-authored .yml tests are stored
  directory: "tests/"
`

	testReadme = `# Behavioral Unit Tests (BUT)

This directory contains human-authored test scenarios for validating
your Alertmanager routing and suppression logic.

## File Format

Tests are defined in YAML format with the following structure:

    name: "Test scenario description"
    tags:
      - "tag1"
      - "tag2"

    state:
      active_alerts:
        - labels:
            service: "api"
            severity: "critical"
      silences:
        - labels:
            service: "maintenance"
          comment: "scheduled maintenance"

    alert:
      labels:
        service: "api"
        severity: "critical"

    expect:
      outcome: "active"
      receivers:
        - "api-team"

## Running Tests

    litmus check
`

	gitattributes = `# Litmus MessagePack files
*.mpk binary diff=litmus

[diff "litmus"]
	textconv = litmus inspect
`
)

// newInitCmd creates the init command.
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new litmus workspace",
		Long:  "Creates litmus.yaml, tests/ directory, and .gitattributes for a new workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	// Check if litmus.yaml already exists
	if _, err := os.Stat("litmus.yaml"); err == nil {
		return fmt.Errorf("litmus.yaml already exists in this directory")
	}

	// Create litmus.yaml
	if err := os.WriteFile("litmus.yaml", []byte(defaultLitmusYAML), 0644); err != nil {
		return fmt.Errorf("creating litmus.yaml: %w", err)
	}

	// Create tests directory
	if err := os.MkdirAll("tests", 0755); err != nil {
		return fmt.Errorf("creating tests directory: %w", err)
	}

	// Create tests/README.md
	testReadmePath := filepath.Join("tests", "README.md")
	if err := os.WriteFile(testReadmePath, []byte(testReadme), 0644); err != nil {
		return fmt.Errorf("creating tests/README.md: %w", err)
	}

	// Create .gitattributes
	if err := os.WriteFile(".gitattributes", []byte(gitattributes), 0644); err != nil {
		return fmt.Errorf("creating .gitattributes: %w", err)
	}

	fmt.Println("✓ litmus.yaml created")
	fmt.Println("✓ tests/ directory created")
	fmt.Println("✓ .gitattributes created")
	fmt.Println("\nWorkspace initialized! Next steps:")
	fmt.Println("1. Update litmus.yaml with your Alertmanager config path")
	fmt.Println("2. Add your behavioral unit tests to tests/")
	fmt.Println("3. Run 'litmus snapshot' to generate regression baseline")

	return nil
}
