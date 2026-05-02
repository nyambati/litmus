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
	amconfig "github.com/prometheus/alertmanager/config"
)

func RunSync(cfg *config.LitmusConfig, address, tenantID, apiKey string, skipValidate, dryRun bool, output string) error {
	ctx := context.Background()

	if address != "" {
		cfg.Mimir.Address = address
	}
	if tenantID != "" {
		cfg.Mimir.TenantID = tenantID
	}
	if apiKey != "" {
		cfg.Mimir.APIKey = apiKey
	}

	ws, err := workspace.Load(cfg.Workspace.Root)
	if err != nil {
		return err
	}

	amConfig, err := ws.Config()
	if err != nil {
		return fmt.Errorf("failed to load alertmanager config: %w", err)
	}

	if amConfig.Route == nil {
		return fmt.Errorf("alertmanager config has no route defined")
	}

	// // For syncing, we might want the serializable version
	// assembled, err := workspace.Load(cfg.Workspace.Root)
	// if err != nil {
	// 	return fmt.Errorf("loading base alertmanager config for sync: %w", err)
	// }
	// TODO: we should probably have a way to get the fully assembled workspace.AlertmanagerConfig
	// For now, let's use what we have. If LoadAlertmanagerConfigYAML only loads base, we might need more.
	// But according to user, we are using them directly in CLI.

	if !skipValidate {
		receiversMap := make(map[string]*amconfig.Receiver)
		for i := range amConfig.Receivers {
			receiversMap[amConfig.Receivers[i].Name] = &amConfig.Receivers[i]
		}
		rules := make([]*amconfig.InhibitRule, 0, len(amConfig.InhibitRules))
		for i := range amConfig.InhibitRules {
			rules = append(rules, &amConfig.InhibitRules[i])
		}
		ctx := sanity.CheckContext{
			Route:     amConfig.Route,
			Receivers: receiversMap,
			Rules:     rules,
			Policy:    cfg.Policy,
		}
		sanityResult := sanity.Run(ctx, cfg.Sanity)
		if !sanityResult.Passed {
			fmt.Fprintf(os.Stderr, "Sanity checks failed. Use --skip-validate to bypass.\n")
			return fmt.Errorf("sanity check failures")
		}
	}

	if dryRun {
		return printYAML(ws.ConfigString(), output)
	}

	if err := cfg.Mimir.Validate(); err != nil {
		return err
	}

	templates, err := loadTemplates(cfg, amConfig.Templates)
	if err != nil {
		return err
	}

	client := mimir.NewClient(&cfg.Mimir)

	payload := mimir.PushPayload{
		Config:    ws.ConfigString(),
		Templates: templates,
	}

	if err := client.Push(ctx, payload); err != nil {
		return fmt.Errorf("pushing to mimir: %w", err)
	}

	fmt.Printf("✓ Alertmanager config synced to %s\n", cfg.Mimir.Address) //nolint:forbidigo
	return nil
}

func printYAML(amCfg string, output string) error {
	data := []byte(amCfg)

	if output != "" {
		if err := os.WriteFile(output, data, 0600); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Config written to %s\n", output)
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
