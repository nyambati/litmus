package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/mimir"
)

// RunSync validates the alertmanager config and pushes it to Mimir.
// Flags override config file values if non-empty.
func RunSync(address, tenantID, apiKey string, skipValidate, dryRun bool) error {
	ctx := context.Background()

	// Load .litmus.yaml
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	// Override with flags if provided
	if address != "" {
		litmusConfig.Mimir.Address = address
	}
	if tenantID != "" {
		litmusConfig.Mimir.TenantID = tenantID
	}
	if apiKey != "" {
		litmusConfig.Mimir.APIKey = apiKey
	}

	// Validate mimir config
	if err := litmusConfig.Mimir.Validate(); err != nil {
		return err
	}

	// Load and assemble config (validates against full assembled routing tree).
	// rawYAML is the base alertmanager.yml — that is what gets pushed to Mimir.
	alertConfig, _, rawYAML, err := litmusConfig.LoadAssembledConfig()
	if err != nil {
		return fmt.Errorf("loading alertmanager config: %w", err)
	}

	if alertConfig.Route == nil {
		return fmt.Errorf("alertmanager config has no route defined")
	}

	// Run sanity checks unless skipped
	if !skipValidate {
		sanity := RunSanityChecks(alertConfig)
		if !sanity.Passed {
			fmt.Fprintf(os.Stderr, "Sanity checks failed. Use --skip-validate to bypass.\n")
			return fmt.Errorf("sanity check failures")
		}
	}

	// Load templates from disk
	templates := make(map[string]string)

	for _, filename := range alertConfig.Templates {
		// Support both absolute and relative paths
		filePath := filepath.Join(litmusConfig.TemplatesDir(), filename)

		if filepath.IsAbs(filename) {
			filePath = filename
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading template %q: %w", filename, err)
		}

		// Use base filename as map key (matches Mimir requirement)
		key := filepath.Base(filename)
		templates[key] = string(data)
	}

	// Dry run: skip actual push
	if dryRun {
		fmt.Println("Dry run: validation passed, skipping push") //nolint:forbidigo
		return nil
	}

	// Push to Mimir
	client := mimir.NewClient(litmusConfig.Mimir.Address, litmusConfig.Mimir.TenantID, litmusConfig.Mimir.APIKey)
	payload := mimir.PushPayload{
		Config:    rawYAML,
		Templates: templates,
	}

	if err := client.Push(ctx, payload); err != nil {
		return fmt.Errorf("pushing to mimir: %w", err)
	}

	fmt.Printf("✓ Alertmanager config synced to %s\n", litmusConfig.Mimir.Address) //nolint:forbidigo
	return nil
}
