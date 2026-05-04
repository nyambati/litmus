package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/mimir"
	"github.com/nyambati/litmus/internal/workspace"
	"github.com/sirupsen/logrus"
)

func RunSync(cfg *config.LitmusConfig, logger logrus.FieldLogger, address, tenantID, apiKey string, skipValidate, dryRun bool, output string) error {
	ctx := context.Background()

	mimirCfg := cfg.Mimir // local copy; does not mutate caller
	if address != "" {
		mimirCfg.Address = address
	}
	if tenantID != "" {
		mimirCfg.TenantID = tenantID
	}
	if apiKey != "" {
		mimirCfg.APIKey = apiKey
	}

	ws, err := workspace.Load(cfg, logger)
	if err != nil {
		return err
	}

	amConfig, err := ws.AMConfig()
	if err != nil {
		return fmt.Errorf("failed to load alertmanager config: %w", err)
	}

	if amConfig.Route == nil {
		return fmt.Errorf("alertmanager config has no route defined")
	}

	if !skipValidate {
		sanityResult := sanity.Run(buildCheckContext(amConfig, ws, cfg.Policy), cfg.Sanity)
		if !sanityResult.Passed {
			fmt.Fprintf(os.Stderr, "Sanity checks failed. Use --skip-validate to bypass.\n")
			return fmt.Errorf("sanity check failures")
		}
	}

	if dryRun {
		return printYAML(ws.ConfigString(), output)
	}

	if err := mimirCfg.Validate(); err != nil {
		return err
	}

	templates, err := loadTemplates(cfg, amConfig.Templates)
	if err != nil {
		return err
	}

	client := mimir.NewClient(&mimirCfg)

	payload := mimir.PushPayload{
		Config:    ws.ConfigString(),
		Templates: templates,
	}

	if err := client.Push(ctx, payload); err != nil {
		return fmt.Errorf("pushing to mimir: %w", err)
	}

	fmt.Printf("✓ Alertmanager config synced to %s\n", mimirCfg.Address) //nolint:forbidigo
	return nil
}

func printYAML(amCfg string, output string) error {
	data := []byte(amCfg)

	if output != "" {
		if err := os.WriteFile(output, data, 0o600); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Config written to %s\n", output)
		return nil
	}

	fmt.Fprintln(os.Stdout, string(amCfg))
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
