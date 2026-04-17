package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/templates"
)

// RunInit creates the litmus workspace skeleton in the current directory.
func RunInit() error {
	if _, err := os.Stat(".litmus.yaml"); err == nil {
		return fmt.Errorf(".litmus.yaml already exists in this directory")
	}

	if err := os.WriteFile(".litmus.yaml", []byte(templates.LitmusYAML), 0644); err != nil {
		return fmt.Errorf("creating .litmus.yaml: %w", err)
	}

	dirs := []string{"config", "config/templates", "regressions", "tests"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating %s directory: %w", dir, err)
		}
	}

	if err := os.WriteFile(filepath.Join("tests", "README.md"), []byte(templates.ReadmeMD), 0644); err != nil {
		return fmt.Errorf("creating tests/README.md: %w", err)
	}

	if err := os.WriteFile(".gitattributes", []byte(templates.GitAttributes), 0644); err != nil {
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
