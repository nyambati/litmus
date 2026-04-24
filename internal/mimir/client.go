package mimir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PushPayload contains the alertmanager config and template files to push to Mimir.
type PushPayload struct {
	Config    string            // raw expanded alertmanager YAML
	Templates map[string]string // filename → content
}

// Client is the HTTP client for Mimir API calls.
type Client struct {
	address    string
	tenantID   string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Mimir API client.
func NewClient(address, tenantID, apiKey string) *Client {
	return &Client{
		address:    address,
		tenantID:   tenantID,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// Push sends the alertmanager config and templates to Mimir /api/v1/alerts.
// Returns error on non-201 response.
func (c *Client) Push(ctx context.Context, payload PushPayload) error {
	body := map[string]interface{}{
		"alertmanager_config": payload.Config,
		"template_files":      payload.Templates,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	url := c.address + "/api/v1/alerts"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Scope-OrgID", c.tenantID)
	if c.apiKey != "" {
		req.SetBasicAuth(c.tenantID, c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pushing to mimir: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("mimir API returned %d (failed to read response: %w)", resp.StatusCode, err)
		}
		return fmt.Errorf("mimir API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
