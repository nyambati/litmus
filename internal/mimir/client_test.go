package mimir

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientPush(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		tenantID     string
		apiKey       string
		checkHeaders func(*testing.T, *http.Request)
		shouldErr    bool
		errContains  string
	}{
		{
			name:       "201 success",
			statusCode: http.StatusCreated,
			tenantID:   "anonymous",
			apiKey:     "secret",
			checkHeaders: func(t *testing.T, r *http.Request) {
				t.Helper()
				if got := r.Header.Get("X-Scope-OrgID"); got != "anonymous" {
					t.Errorf("X-Scope-OrgID: got %q, want %q", got, "anonymous")
				}
				if got := r.Header.Get("Authorization"); got != "Bearer secret" {
					t.Errorf("Authorization: got %q, want %q", got, "Bearer secret")
				}
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type: got %q, want %q", got, "application/json")
				}
			},
			shouldErr: false,
		},
		{
			name:       "201 success with empty api key",
			statusCode: http.StatusCreated,
			tenantID:   "anonymous",
			apiKey:     "",
			checkHeaders: func(t *testing.T, r *http.Request) {
				t.Helper()
				if got := r.Header.Get("X-Scope-OrgID"); got != "anonymous" {
					t.Errorf("X-Scope-OrgID: got %q, want %q", got, "anonymous")
				}
				if got := r.Header.Get("Authorization"); got != "" {
					t.Errorf("Authorization: got %q, want empty", got)
				}
			},
			shouldErr: false,
		},
		{
			name:       "400 bad request",
			statusCode: http.StatusBadRequest,
			tenantID:   "anonymous",
			apiKey:     "secret",
			checkHeaders: func(t *testing.T, r *http.Request) {
				t.Helper()
			},
			shouldErr:   true,
			errContains: "400",
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			tenantID:   "anonymous",
			apiKey:     "secret",
			checkHeaders: func(t *testing.T, r *http.Request) {
				t.Helper()
			},
			shouldErr:   true,
			errContains: "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/alerts" {
					t.Errorf("URL path: got %q, want /api/v1/alerts", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Method: got %s, want POST", r.Method)
				}
				tt.checkHeaders(t, r)
				w.WriteHeader(tt.statusCode)
				if tt.statusCode != http.StatusCreated {
					_, _ = w.Write([]byte("error message"))
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, tt.tenantID, tt.apiKey)
			payload := PushPayload{
				Config:    "global:\n  resolve_timeout: 5m\n",
				Templates: map[string]string{"slack.tpl": "{{ .GroupLabels }}"},
			}

			err := client.Push(context.Background(), payload)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errContains != "" && !contains(t, err.Error(), tt.errContains) {
					t.Errorf("error: got %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func contains(t *testing.T, s, substr string) bool {
	t.Helper()
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
