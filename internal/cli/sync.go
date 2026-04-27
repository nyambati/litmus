package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/mimir"
	"gopkg.in/yaml.v3"
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

	amConfig, _, _, err := litmusConfig.LoadAssembledConfig()
	if err != nil {
		return fmt.Errorf("loading alertmanager config: %w", err)
	}

	if amConfig.Route == nil {
		return fmt.Errorf("alertmanager config has no route defined")
	}

	if !skipValidate {
		amCfg, err := config.ToAMConfig(amConfig)
		if err != nil {
			return fmt.Errorf("converting config for validation: %w", err)
		}
		sanity := RunSanityChecks(amCfg, litmusConfig.Sanity)
		if !sanity.Passed {
			fmt.Fprintf(os.Stderr, "Sanity checks failed. Use --skip-validate to bypass.\n")
			return fmt.Errorf("sanity check failures")
		}
	}

	if dryRun {
		return outputAssembledConfig(amConfig, output)
	}

	if err := litmusConfig.Mimir.Validate(); err != nil {
		return err
	}

	templates := loadTemplates(litmusConfig, amConfig.Templates)

	client := mimir.NewClient(litmusConfig.Mimir.Address, litmusConfig.Mimir.TenantID, litmusConfig.Mimir.APIKey)
	payload := mimir.PushPayload{
		Config:    assembleYAML(amConfig),
		Templates: templates,
	}

	if err := client.Push(ctx, payload); err != nil {
		return fmt.Errorf("pushing to mimir: %w", err)
	}

	fmt.Printf("✓ Alertmanager config synced to %s\n", litmusConfig.Mimir.Address) //nolint:forbidigo
	return nil
}

func outputAssembledConfig(amCfg *config.AlertmanagerConfig, output string) error {
	yamlData, err := yaml.Marshal(amCfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if output != "" {
		if err := os.WriteFile(output, yamlData, 0600); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Printf("Config written to %s\n", output) //nolint:forbidigo
		return nil
	}

	fmt.Println(string(yamlData)) //nolint:forbidigo
	return nil
}

func assembleYAML(amCfg *config.AlertmanagerConfig) string {
	yamlData, err := yaml.Marshal(amCfg)
	if err != nil {
		return ""
	}

	return string(yamlData)
}

func loadTemplates(litmusConfig *config.LitmusConfig, templateNames []string) map[string]string {
	templates := make(map[string]string)

	for _, filename := range templateNames {
		filePath := filepath.Join(litmusConfig.TemplatesDir(), filename)

		if filepath.IsAbs(filename) {
			filePath = filename
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: could not read template %q: %v\n", filename, err)
			continue
		}

		key := filepath.Base(filename)
		templates[key] = string(data)
	}

	return templates
}
