package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/nyambati/litmus/internal/codec"
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/behavioral"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

var staticFS fs.FS

// SetStaticFS registers the embedded UI filesystem before the server starts.
func SetStaticFS(f fs.FS) {
	staticFS = f
}

// RouteNode represents a node in the route tree with evaluation details.
type RouteNode struct {
	Receiver       string       `json:"receiver,omitempty"`
	Match          []string     `json:"match,omitempty"`
	Matched        bool         `json:"matched"`
	Continue       bool         `json:"continue"`
	GroupBy        []string     `json:"group_by,omitempty"`
	GroupWait      string       `json:"group_wait,omitempty"`
	GroupInterval  string       `json:"group_interval,omitempty"`
	RepeatInterval string       `json:"repeat_interval,omitempty"`
	Children       []*RouteNode `json:"children,omitempty"`
}

// EvalResponse is the verbose result of an alert evaluation.
type EvalResponse struct {
	Labels      map[string]string `json:"labels"`
	Receivers   []string          `json:"receivers"`
	Path        *RouteNode        `json:"path"`
	Suppression *SuppressionInfo  `json:"suppression,omitempty"`
}

// SuppressionInfo holds details about why an alert might be silenced or inhibited.
type SuppressionInfo struct {
	Inhibited bool   `json:"inhibited"`
	Silenced  bool   `json:"silenced"`
	Reason    string `json:"reason,omitempty"`
}

// ConfigResponse returns metadata about the current workspace.
type ConfigResponse struct {
	ConfigPath string `json:"config_path"`
	Ready      bool   `json:"ready"`
}

// RunUIServer starts the Litmus UI backend.
//
//nolint:gocyclo
func RunUIServer(port int, dev bool) error {
	litmusConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading litmus config: %w", err)
	}

	alertConfigPath := filepath.Join(litmusConfig.Config.Directory, litmusConfig.Config.File)

	mux := http.NewServeMux()

	// CORS Middleware for development
	withCORS := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	// API Endpoints
	mux.HandleFunc("/api/v1/config", withCORS(func(w http.ResponseWriter, r *http.Request) {
		resp := ConfigResponse{
			ConfigPath: alertConfigPath,
			Ready:      true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/api/v1/tests", withCORS(func(w http.ResponseWriter, r *http.Request) {
		loader := behavioral.NewBehavioralTestLoader()
		tests, err := loader.LoadFromDirectory(litmusConfig.Tests.Directory)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading tests: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tests)
	}))

	mux.HandleFunc("/api/v1/tests/run", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading alertmanager config: %v", err), http.StatusInternalServerError)
			return
		}

		loader := behavioral.NewBehavioralTestLoader()
		tests, err := loader.LoadFromDirectory(litmusConfig.Tests.Directory)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading tests: %v", err), http.StatusInternalServerError)
			return
		}

		router := pipeline.NewRouter(alertConfig.Route)
		executor := behavioral.NewBehavioralTestExecutor(alertConfig.InhibitRules)

		// If ?name= is provided, run only that test
		if name := r.URL.Query().Get("name"); name != "" {
			for _, test := range tests {
				if test.Name == name {
					result := executor.Execute(context.Background(), test, router)
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]*types.TestResult{result})
					return
				}
			}
			http.Error(w, fmt.Sprintf("Test not found: %s", name), http.StatusNotFound)
			return
		}

		results := make([]*types.TestResult, 0, len(tests))
		for _, test := range tests {
			results = append(results, executor.Execute(context.Background(), test, router))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(results)
	}))

	mux.HandleFunc("/api/v1/evaluate", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Labels map[string]string `json:"labels"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Reload config on every request for "live" feel
		alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading alertmanager config: %v", err), http.StatusInternalServerError)
			return
		}

		labelSet := model.LabelSet{}
		for k, v := range req.Labels {
			labelSet[model.LabelName(k)] = model.LabelValue(v)
		}

		router := pipeline.NewRouter(alertConfig.Route)
		receivers := router.Match(labelSet)

		path := traceRoute(alertConfig.Route, labelSet)

		resp := EvalResponse{
			Labels:    req.Labels,
			Receivers: receivers,
			Path:      path,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	regressionYAMLPath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.yml")

	mux.HandleFunc("/api/v1/regressions", withCORS(func(w http.ResponseWriter, r *http.Request) {
		tests, err := loadBaselineYAML(regressionYAMLPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading regressions: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tests)
	}))

	mux.HandleFunc("/api/v1/regressions/run", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading alertmanager config: %v", err), http.StatusInternalServerError)
			return
		}

		tests, err := loadBaselineYAML(regressionYAMLPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading regressions: %v", err), http.StatusInternalServerError)
			return
		}

		// If ?name= is provided, run only that test
		if name := r.URL.Query().Get("name"); name != "" {
			found := false
			for i, t := range tests {
				if t.Name == name {
					tests = tests[i : i+1]
					found = true
					break
				}
			}
			if !found {
				http.Error(w, fmt.Sprintf("Test not found: %s", name), http.StatusNotFound)
				return
			}
		}

		router := pipeline.NewRouter(alertConfig.Route)
		executor := snapshot.NewRegressionTestExecutor()
		raw := executor.Execute(context.Background(), tests, router)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(raw)
	}))

	regressionMpkPath := filepath.Join(litmusConfig.Regression.Directory, "regressions.litmus.mpk")

	mux.HandleFunc("/api/v1/diff", withCORS(func(w http.ResponseWriter, r *http.Request) {
		alertConfig, err := config.LoadAlertmanagerConfig(alertConfigPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading alertmanager config: %v", err), http.StatusInternalServerError)
			return
		}

		// mpk is the true baseline; yml reflects the current alertmanager state
		baseline, err := loadBaseline(regressionMpkPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Loading baseline: %v", err), http.StatusInternalServerError)
			return
		}

		router := pipeline.NewRouter(alertConfig.Route)
		executor := snapshot.NewRegressionTestExecutor()
		raw := executor.Execute(context.Background(), baseline, router)

		type deltaResult struct {
			Name       string            `json:"name"`
			Pass       bool              `json:"pass"`
			Error      string            `json:"error,omitempty"`
			Labels     map[string]string `json:"labels,omitempty"`
			Expected   []string          `json:"expected,omitempty"`
			Actual     []string          `json:"actual,omitempty"`
			RoutePath  []*RouteNode      `json:"route_path,omitempty"`
			WhyDrifted []RouteDrift      `json:"why_drifted,omitempty"`
		}

		type diffResponse struct {
			Total   int            `json:"total"`
			Passed  int            `json:"passed"`
			Drifted int            `json:"drifted"`
			Results []*deltaResult `json:"results"`
		}

		resp := diffResponse{
			Total:   len(raw),
			Results: make([]*deltaResult, 0, len(raw)),
		}

		for _, res := range raw {
			if res.Pass {
				resp.Passed++
			} else {
				resp.Drifted++
			}

			dr := &deltaResult{
				Name:     res.Name,
				Pass:     res.Pass,
				Error:    res.Error,
				Labels:   res.Labels,
				Expected: res.Expected,
				Actual:   res.Actual,
			}

			// For drifted tests, trace current route and identify why expected routes no longer match
			if !res.Pass && res.Labels != nil {
				labelSet := make(model.LabelSet)
				for k, v := range res.Labels {
					labelSet[model.LabelName(k)] = model.LabelValue(v)
				}

				dr.RoutePath = flattenMatchedPath(traceRoute(alertConfig.Route, labelSet))

				for _, expectedReceiver := range res.Expected {
					routes := findRoutesByReceiver(alertConfig.Route, expectedReceiver)
					if len(routes) == 0 {
						dr.WhyDrifted = append(dr.WhyDrifted, RouteDrift{
							Receiver: expectedReceiver,
							Found:    false,
						})
						continue
					}
					for _, route := range routes {
						mismatches := identifyMatcherFailures(route, labelSet)
						if len(mismatches) > 0 {
							dr.WhyDrifted = append(dr.WhyDrifted, RouteDrift{
								Receiver:   expectedReceiver,
								Found:      true,
								Mismatches: mismatches,
							})
						}
					}
				}
			}

			resp.Results = append(resp.Results, dr)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Serve embedded UI in production mode
	if !dev && staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Static assets have a file extension — let the file server handle them.
			// Everything else (SPA routes including "/") gets index.html.
			if path.Ext(r.URL.Path) != "" {
				fileServer.ServeHTTP(w, r)
				return
			}
			f, err := staticFS.Open("index.html")
			if err != nil {
				http.Error(w, "UI not found", http.StatusNotFound)
				return
			}
			defer f.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, "index.html", time.Time{}, f.(io.ReadSeeker))
		})
	}

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost%s", addr)
	log.Printf("Litmus UI running at %s", url)

	if !dev {
		go func() {
			time.Sleep(150 * time.Millisecond)
			openBrowser(url)
		}()
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}

func traceRoute(route *amconfig.Route, labels model.LabelSet) *RouteNode {
	node := &RouteNode{
		Receiver: route.Receiver,
		Continue: route.Continue,
		Matched:  routeMatches(route, labels),
	}

	// Capture Grouping and Timing
	for _, l := range route.GroupBy {
		node.GroupBy = append(node.GroupBy, string(l))
	}
	if route.GroupWait != nil {
		node.GroupWait = route.GroupWait.String()
	}
	if route.GroupInterval != nil {
		node.GroupInterval = route.GroupInterval.String()
	}
	if route.RepeatInterval != nil {
		node.RepeatInterval = route.RepeatInterval.String()
	}

	// Capture matchers for UI display
	for k, v := range route.Match {
		node.Match = append(node.Match, fmt.Sprintf("%s=%q", k, v))
	}
	for k, re := range route.MatchRE {
		node.Match = append(node.Match, fmt.Sprintf("%s=~%q", k, re.String()))
	}
	for _, m := range route.Matchers {
		node.Match = append(node.Match, m.String())
	}

	if !node.Matched {
		return node
	}

	for _, child := range route.Routes {
		childNode := traceRoute(child, labels)
		if childNode.Matched {
			node.Children = append(node.Children, childNode)
			if !childNode.Continue {
				break
			}
		}
	}

	return node
}

// flattenMatchedPath walks a RouteNode tree and returns each matched node as a flat list.
func flattenMatchedPath(node *RouteNode) []*RouteNode {
	if node == nil {
		return nil
	}
	var steps []*RouteNode
	if node.Matched {
		steps = append(steps, &RouteNode{Receiver: node.Receiver, Match: node.Match})
	}
	for _, child := range node.Children {
		steps = append(steps, flattenMatchedPath(child)...)
	}
	return steps
}

// MatcherMismatch describes a single matcher in a route that no longer matches the label set.
type MatcherMismatch struct {
	Label    string `json:"label"`
	Required string `json:"required"` // value/pattern the current config expects
	Actual   string `json:"actual"`   // value present in the baseline label set
}

// RouteDrift describes why a specific expected receiver no longer matches.
type RouteDrift struct {
	Receiver   string            `json:"receiver"`
	Found      bool              `json:"found"`
	Mismatches []MatcherMismatch `json:"mismatches,omitempty"`
}

// findRoutesByReceiver walks the route tree and returns every route whose receiver equals the target.
func findRoutesByReceiver(route *amconfig.Route, receiver string) []*amconfig.Route {
	var found []*amconfig.Route
	if route.Receiver == receiver {
		found = append(found, route)
	}
	for _, child := range route.Routes {
		found = append(found, findRoutesByReceiver(child, receiver)...)
	}
	return found
}

// identifyMatcherFailures returns matchers on the given route that do not match the label set.
func identifyMatcherFailures(route *amconfig.Route, labels model.LabelSet) []MatcherMismatch {
	var out []MatcherMismatch

	for k, v := range route.Match {
		actual := string(labels[model.LabelName(k)])
		if actual != v {
			out = append(out, MatcherMismatch{Label: k, Required: v, Actual: actual})
		}
	}

	for k, re := range route.MatchRE {
		actual := string(labels[model.LabelName(k)])
		if !re.MatchString(actual) {
			out = append(out, MatcherMismatch{Label: k, Required: "~" + re.String(), Actual: actual})
		}
	}

	for _, m := range route.Matchers {
		actual := string(labels[model.LabelName(m.Name)])
		if !m.Matches(actual) {
			out = append(out, MatcherMismatch{Label: m.Name, Required: m.Value, Actual: actual})
		}
	}

	return out
}

func routeMatches(route *amconfig.Route, labels model.LabelSet) bool {
	for k, v := range route.Match {
		if string(labels[model.LabelName(k)]) != v {
			return false
		}
	}
	for k, re := range route.MatchRE {
		val := string(labels[model.LabelName(k)])
		if !re.MatchString(val) {
			return false
		}
	}
	for _, m := range route.Matchers {
		if !m.Matches(string(labels[model.LabelName(m.Name)])) {
			return false
		}
	}
	return true
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// loadBaseline reads a msgpack regression baseline from disk.
func loadBaseline(path string) ([]*types.TestCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var tests []*types.TestCase
	if err := codec.DecodeMsgPack(file, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}

// loadBaselineYAML reads a YAML regression baseline from disk.
func loadBaselineYAML(path string) ([]*types.TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tests []*types.TestCase
	if err := yaml.Unmarshal(data, &tests); err != nil {
		return nil, err
	}
	return tests, nil
}
