package cmd

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/templates"
	"github.com/spf13/cobra"
)

// newInitCmd creates the init command.
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "init",
		Short:        "Initialize a new litmus workspace",
		Long:         "Creates .litmus.yaml, tests/ directory, and .gitattributes for a new workspace",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	// Check if litmus.yaml already exists
	if _, err := os.Stat(".litmus.yaml"); err == nil {
		return fmt.Errorf(".litmus.yaml already exists in this directory")
	}

	// Create .litmus.yaml
	if err := os.WriteFile(".litmus.yaml", []byte(templates.LitmusYAML), 0644); err != nil {
		return fmt.Errorf("creating .litmus.yaml: %w", err)
	}

	// Create directories
	dirs := []string{"config", "config/templates", "regressions", "tests"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating %s directory: %w", dir, err)
		}
	}

	// Create tests/README.md
	testReadmePath := filepath.Join("tests", "README.md")
	if err := os.WriteFile(testReadmePath, []byte(templates.ReadmeMD), 0644); err != nil {
		return fmt.Errorf("creating tests/README.md: %w", err)
	}

	// Create .gitattributes
	if err := os.WriteFile(".gitattributes", []byte(templates.GitAtrributes), 0644); err != nil {
		return fmt.Errorf("creating .gitattributes: %w", err)
	}

	fmt.Println("✓ .litmus.yaml created")
	fmt.Println("✓ config/ and regressions/ directories created")
	fmt.Println("✓ tests/ directory created")
	fmt.Println("✓ .gitattributes created")
	fmt.Println("\nWorkspace initialized! Next steps:")
	fmt.Println("1. Update .litmus.yaml with your Alertmanager config path")
	fmt.Println("2. Add your behavioral unit tests to tests/")
	fmt.Println("3. Run 'litmus snapshot' to generate regression baseline")

	return nil
}
