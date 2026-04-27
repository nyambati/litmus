package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nyambati/litmus/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupFragmentWorkspace creates a minimal workspace with a base alertmanager
// config and a single fragment that mounts a db-critical route under scope=teams.
func setupFragmentWorkspace(t *testing.T) *config.LitmusConfig {
	t.Helper()
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })
	require.NoError(t, os.Chdir(tmpDir))

	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  fragments: "fragments/*"
  history: 3
global_labels: {}
`), 0600))

	require.NoError(t, os.MkdirAll("config/tests", 0755))
	require.NoError(t, os.WriteFile("config/alertmanager.yml", []byte(`
global:
  resolve_timeout: 5m
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600))

	require.NoError(t, os.MkdirAll("config/fragments", 0755))
	require.NoError(t, os.WriteFile("config/fragments/db.yml", []byte(`
name: "db-team"
group:
  match:
    scope: "teams"
routes:
  - receiver: "db-critical"
    match:
      service: "mysql"
receivers:
  - name: "db-critical"
`), 0600))
	require.NoError(t, os.WriteFile("config/fragments/db-tests.yml", []byte(`
- name: "mysql routes to db-critical"
  type: "unit"
  alert:
    labels:
      scope: "teams"
      service: "mysql"
  expect:
    outcome: "active"
    receivers:
      - "db-critical"
`), 0600))

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	return cfg
}

func setupRootBehavioralWorkspace(t *testing.T) *config.LitmusConfig {
	t.Helper()
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })
	require.NoError(t, os.Chdir(tmpDir))

	require.NoError(t, os.WriteFile(".litmus.yaml", []byte(`
workspace:
  root: "config"
  fragments: "fragments/*"
  history: 3
global_labels: {}
`), 0600))

	require.NoError(t, os.MkdirAll("config/tests", 0755))
	require.NoError(t, os.MkdirAll("config/fragments", 0755))
	require.NoError(t, os.WriteFile("config/alertmanager.yml", []byte(`
route:
  receiver: 'default'
receivers:
  - name: 'default'
`), 0600))
	require.NoError(t, os.WriteFile("config/tests/root.yml", []byte(`
name: "root test"
type: "unit"
alert:
  labels:
    service: "api"
expect:
  outcome: "active"
  receivers:
    - "default"
`), 0600))

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	return cfg
}

func newTestCtx(cfg *config.LitmusConfig, req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(string(LitmusConfigKey), cfg)
	c.Request = req
	return c, w
}

func TestEvaluateHandler_UsesAssembledConfig(t *testing.T) {
	cfg := setupFragmentWorkspace(t)

	body := `{"labels":{"scope":"teams","service":"mysql"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	c, w := newTestCtx(cfg, req)
	evaluateHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp EvalResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp.Receivers, "db-critical",
		"evaluate must route through assembled fragment routes")
}

func TestSuggestHandler_IncludesFragmentRouteLabels(t *testing.T) {
	cfg := setupFragmentWorkspace(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/label_values", nil)
	c, w := newTestCtx(cfg, req)
	suggestHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Labels []string            `json:"labels"`
		Values map[string][]string `json:"values"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp.Labels, "service",
		"suggest must include label keys from assembled fragment routes")
}

func TestTestsHandler_IncludesFragmentTests(t *testing.T) {
	cfg := setupFragmentWorkspace(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tests?type=behavioral", nil)
	c, w := newTestCtx(cfg, req)
	testsHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	var tests []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tests))

	names := make([]string, 0, len(tests))
	for _, tc := range tests {
		if n, ok := tc["name"].(string); ok {
			names = append(names, n)
		}
	}
	assert.Contains(t, names, "mysql routes to db-critical",
		"tests endpoint must include tests from fragments")
}

func TestConfigHandler_ExposesWorkspaceAndFragmentCount(t *testing.T) {
	cfg := setupFragmentWorkspace(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	c, w := newTestCtx(cfg, req)
	configHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp ConfigResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Ready)
	assert.NotEmpty(t, resp.ConfigPath)
	assert.Equal(t, "config", resp.Workspace.Root)
	assert.Equal(t, "fragments/*", resp.Workspace.Fragments)
	assert.Equal(t, 2, resp.FragmentCount, "root + db-team fragment")
}

func TestFragmentsHandler_ListsFragmentsWithMetadata(t *testing.T) {
	cfg := setupFragmentWorkspace(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/fragments", nil)
	c, w := newTestCtx(cfg, req)
	fragmentsHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	var frags []FragmentInfo
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &frags))

	require.Len(t, frags, 2, "root + db-team")

	var dbFrag *FragmentInfo
	for i := range frags {
		if frags[i].Name == "db-team" {
			dbFrag = &frags[i]
		}
	}
	require.NotNil(t, dbFrag, "db-team fragment must be present")
	assert.Equal(t, 1, dbFrag.Routes)
	assert.Equal(t, 1, dbFrag.Receivers)
	assert.Equal(t, 1, dbFrag.Tests)
	require.NotNil(t, dbFrag.Group)
	assert.Equal(t, map[string]string{"scope": "teams"}, dbFrag.Group.Match)
}

func TestRunTestsHandler_DoesNotDuplicateRootTests(t *testing.T) {
	cfg := setupRootBehavioralWorkspace(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tests/run?type=behavioral", nil)
	c, w := newTestCtx(cfg, req)
	runTestsHandler(c)

	require.Equal(t, http.StatusOK, w.Code)
	var results []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &results))
	require.Len(t, results, 1)
}
