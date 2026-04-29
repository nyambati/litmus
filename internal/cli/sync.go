package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/mimir"
)

func RunSync(address, tenantID, apiKey string, skipValidate, dryRun bool, output string) error {
	ctx := context.Background()

	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	if address != "" {
		litmusConfig.Mimir.Address = address
	}
	if tenantID != "" {
		litmusConfig.Mimir.TenantID = tenantID
	}
	if apiKey != "" {
		litmusConfig.Mimir.APIKey = apiKey
	}

	assembled, _, amConfig, err := litmusConfig.LoadAssembledConfig()
	if err != nil {
		return fmt.Errorf("loading alertmanager config: %w", err)
	}

	if amConfig.Route == nil {
		return fmt.Errorf("alertmanager config has no route defined")
	}

	if !skipValidate {
		sanity := RunSanityChecks(amConfig, litmusConfig.Sanity)
		if !sanity.Passed {
			fmt.Fprintf(os.Stderr, "Sanity checks failed. Use --skip-validate to bypass.\n")
			return fmt.Errorf("sanity check failures")
		}
	}

	if dryRun {
		return outputAssembledConfig(assembled, output)
	}

	if err := litmusConfig.Mimir.Validate(); err != nil {
		return err
	}

	templates, err := loadTemplates(litmusConfig, amConfig.Templates)
	if err != nil {
		return err
	}

	client := mimir.NewClient(&litmusConfig.Mimir)
	configPayload, err := assembled.String()
	if err != nil {
		return fmt.Errorf("converting assembled config to string: %w", err)
	}

	payload := mimir.PushPayload{
		Config:    configPayload,
		Templates: templates,
	}

	if err := client.Push(ctx, payload); err != nil {
		return fmt.Errorf("pushing to mimir: %w", err)
	}

	fmt.Printf("✓ Alertmanager config synced to %s\n", litmusConfig.Mimir.Address) //nolint:forbidigo
	return nil
}

func outputAssembledConfig(amCfg *config.AlertmanagerConfig, output string) error {
	data, err := amCfg.MarshalIndent(2)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if output != "" {
		if err := os.WriteFile(output, data.Bytes(), 0600); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Config written to %s\n", output)
	}

	fmt.Fprintln(os.Stdout, data.String())
	return nil
}

func loadTemplates(litmusConfig *config.LitmusConfig, templateNames []string) (map[string]string, error) {
	templates := make(map[string]string)

	for _, filename := range templateNames {
		filePath := filepath.Join(litmusConfig.TemplatesDir(), filename)

		if filepath.IsAbs(filename) {
			filePath = filename
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading template %q: %w", filename, err)
		}

		key := filepath.Base(filename)
		templates[key] = string(data)
	}

	return templates, nil
}
