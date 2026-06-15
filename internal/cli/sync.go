package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/mimir"
	"github.com/nyambati/litmus/internal/workspace"
)

type SyncOptions struct {
	Address  string
	TenantID string
	APIKey   string
	DryRun   bool
	Output   string
}

func RunSync(ctx context.Context, options SyncOptions) error {
	cfg := config.ConfigFromContext(ctx)
	logger := config.LoggerFromContext(ctx)
	mimirCfg := cfg.Mimir // local copy; does not mutate caller
	if options.Address != "" {
		mimirCfg.Address = options.Address
	}
	if options.TenantID != "" {
		mimirCfg.TenantID = options.TenantID
	}
	if options.APIKey != "" {
		mimirCfg.APIKey = options.APIKey
	}

	ws, err := workspace.Load(cfg, logger)
	if err != nil {
		return err
	}

	amConfig, err := ws.Config()
	if err != nil {
		return fmt.Errorf("failed to load alertmanager config: %w", err)
	}

	sanityResult := sanity.Run(buildCheckContext(amConfig, ws, cfg.Policy), cfg.Sanity)
	if !sanityResult.Passed {
		fmt.Fprintf(os.Stderr, "Sanity checks failed; fix the reported issues before syncing.\n")
		return fmt.Errorf("sanity check failures")
	}

	if options.DryRun {
		return printYAML(ws.ConfigString(), options.Output)
	}

	if err := mimirCfg.Validate(); err != nil {
		return err
	}

	templates, err := ws.Templates()
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

	fmt.Fprintf(os.Stdout, "✓ Alertmanager config synced to %s\n", mimirCfg.Address)
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
