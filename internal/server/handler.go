package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nyambati/litmus/internal/cli"
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/behavioral"
	"github.com/nyambati/litmus/internal/engine/pipeline"
	"github.com/nyambati/litmus/internal/engine/snapshot"
	"github.com/nyambati/litmus/internal/stores"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

// loadAssembled loads the assembled alertmanager config and fragments, writing
// an error response and returning false on failure.
func loadAssembled(c *gin.Context, litmusConfig *config.LitmusConfig) (*amconfig.Config, []*config.Fragment, bool) {
	alertConfig, fragments, _, err := litmusConfig.LoadAssembledConfig()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Loading alertmanager config: %v", err))
		return nil, nil, false
	}
	return alertConfig, fragments, true
}

func getLitmusConfig(c *gin.Context) *config.LitmusConfig {
	val, exists := c.Get(string(LitmusConfigKey))
	if !exists {
		c.String(http.StatusInternalServerError, "Litmus config not found in context")
		c.Abort()
		return nil
	}
	litmusConfig, ok := val.(*config.LitmusConfig)
	if !ok {
		c.String(http.StatusInternalServerError, "Invalid Litmus config type in context")
		c.Abort()
		return nil
	}
	return litmusConfig
}

func configHandler(c *gin.Context) {
	litmusConfig := getLitmusConfig(c)
	if litmusConfig == nil {
		return // getLitmusConfig already handled the error
	}
	resp := ConfigResponse{
		ConfigPath: litmusConfig.FilePath(),
		Ready:      true,
	}
	c.JSON(http.StatusOK, resp)
}

func testsHandler(c *gin.Context) {
	litmusConfig := getLitmusConfig(c)
	if litmusConfig == nil {
		return
	}

	switch c.DefaultQuery("type", "behavioral") {
	case "regression":
		state, err := cli.LoadRegressionState(litmusConfig.RegressionsYamlFilePath())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				c.JSON(http.StatusOK, []types.TestCase{})
				return
			}
			c.String(http.StatusInternalServerError, fmt.Sprintf("Loading regressions: %v", err))
			return
		}
		c.JSON(http.StatusOK, state.Tests)
	default:
		loader := behavioral.NewBehavioralTestLoader()
		tests, err := loader.LoadFromDirectory(litmusConfig.TestsDir())
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Loading tests: %v", err))
			return
		}
		frags, err := config.LoadFragments(litmusConfig.FragmentsPath())
		if err != nil {
			log.Printf("warn: loading fragments for tests: %v", err)
		}
		for _, frag := range frags {
			tests = append(tests, frag.Tests...)
		}
		c.JSON(http.StatusOK, tests)
	}
}

func runTestsHandler(c *gin.Context) {
	litmusConfig := getLitmusConfig(c)
	if litmusConfig == nil {
		return
	}

	alertConfig, fragments, ok := loadAssembled(c, litmusConfig)
	if !ok {
		return
	}

	router := pipeline.NewRouter(alertConfig.Route)
	name := c.Query("name")

	switch c.DefaultQuery("type", "behavioral") {
	case "regression":
		state, err := cli.LoadRegressionState(litmusConfig.RegressionsYamlFilePath())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				c.JSON(http.StatusOK, []*types.TestResult{})
				return
			}
			c.String(http.StatusInternalServerError, fmt.Sprintf("Loading regressions: %v", err))
			return
		}
		tests := state.Tests
		if name != "" {
			found := false
			for i, t := range tests {
				if t.Name == name {
					tests = tests[i : i+1]
					found = true
					break
				}
			}
			if !found {
				c.String(http.StatusNotFound, fmt.Sprintf("Test not found: %s", name))
				return
			}
		}
		c.JSON(http.StatusOK, snapshot.NewRegressionTestExecutor().Execute(context.Background(), tests, router))
	default:
		loader := behavioral.NewBehavioralTestLoader()
		tests, err := loader.LoadFromDirectory(litmusConfig.TestsDir())
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Loading tests: %v", err))
			return
		}
		for _, frag := range fragments {
			tests = append(tests, frag.Tests...)
		}
		executor := behavioral.NewBehavioralTestExecutor(alertConfig.InhibitRules)
		if name != "" {
			for _, test := range tests {
				if test.Name == name {
					c.JSON(http.StatusOK, []*types.TestResult{executor.Execute(context.Background(), test, router)})
					return
				}
			}
			c.String(http.StatusNotFound, fmt.Sprintf("Test not found: %s", name))
			return
		}
		results := make([]*types.TestResult, 0, len(tests))
		for _, test := range tests {
			results = append(results, executor.Execute(context.Background(), test, router))
		}
		c.JSON(http.StatusOK, results)
	}
}

func evaluateHandler(c *gin.Context) {
	litmusConfig := getLitmusConfig(c)
	if litmusConfig == nil {
		return
	}
	var req struct {
		Labels map[string]string `json:"labels"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body")
		return
	}

	// Reload config on every request for "live" feel
	alertConfig, _, ok := loadAssembled(c, litmusConfig)
	if !ok {
		return
	}

	labelSet := model.LabelSet{}
	for k, v := range req.Labels {
		labelSet[model.LabelName(k)] = model.LabelValue(v)
	}

	if alertConfig.Route == nil {
		c.String(http.StatusInternalServerError, "Alertmanager config has no root route")
		return
	}

	router := pipeline.NewRouter(alertConfig.Route)
	receivers := router.Match(labelSet)

	path := traceRoute(alertConfig.Route, labelSet)

	resp := EvalResponse{
		Labels:    req.Labels,
		Receivers: receivers,
		Path:      path,
	}

	c.JSON(http.StatusOK, resp)
}

func suggestHandler(c *gin.Context) {
	litmusConfig := getLitmusConfig(c)
	if litmusConfig == nil {
		return
	}
	alertConfig, fragments, ok := loadAssembled(c, litmusConfig)
	if !ok {
		return
	}

	labelMap := make(map[string]map[string]struct{})
	addSuggestion := func(k, v string) {
		if _, ok := labelMap[k]; !ok {
			labelMap[k] = make(map[string]struct{})
		}
		if v != "" {
			labelMap[k][v] = struct{}{}
		}
	}

	var walkRoute func(*amconfig.Route)
	walkRoute = func(route *amconfig.Route) {
		if route == nil {
			return
		}
		for k, v := range route.Match {
			addSuggestion(k, v)
		}
		for k, v := range route.MatchRE {
			addSuggestion(k, v.String())
		}
		for _, m := range route.Matchers {
			addSuggestion(m.Name, m.Value)
		}
		for _, child := range route.Routes {
			walkRoute(child)
		}
	}
	walkRoute(alertConfig.Route)

	loader := behavioral.NewBehavioralTestLoader()
	tests, err := loader.LoadFromDirectory(litmusConfig.TestsDir())
	if err != nil {
		log.Printf("warning: failed to load tests for suggestions: %v", err)
	}
	for _, frag := range fragments {
		tests = append(tests, frag.Tests...)
	}
	for _, test := range tests {
		if test.Alert == nil {
			continue
		}
		for k, v := range test.Alert.Labels {
			addSuggestion(k, v)
		}
	}

	type suggestResponse struct {
		Labels []string            `json:"labels"`
		Values map[string][]string `json:"values"`
	}

	resp := suggestResponse{
		Labels: make([]string, 0, len(labelMap)),
		Values: make(map[string][]string),
	}

	expander := snapshot.NewRegexExpander()
	for k, vSet := range labelMap {
		resp.Labels = append(resp.Labels, k)
		seen := make(map[string]struct{})
		var values []string
		for v := range vSet {
			for _, part := range expander.ExpandAlternations(v) {
				if _, dup := seen[part]; !dup {
					seen[part] = struct{}{}
					values = append(values, part)
				}
			}
		}
		if len(values) > 0 {
			resp.Values[k] = values
		}
	}

	c.JSON(http.StatusOK, resp)
}

func generateRegressionsHandler(c *gin.Context) {
	litmusConfig := getLitmusConfig(c)
	if litmusConfig == nil {
		return
	}
	_, err := os.Stat(litmusConfig.RegressionsYamlFilePath())

	update := c.Query("update") == "true" || errors.Is(err, os.ErrNotExist)

	if err := cli.RunSnapshot(update, false); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Snapshot failed: %v", err))
		return
	}
	c.String(http.StatusOK, "OK")
}

func diffHandler(c *gin.Context) {
	litmusConfig := getLitmusConfig(c)
	if litmusConfig == nil {
		return
	}
	alertConfig, _, ok := loadAssembled(c, litmusConfig)
	if !ok {
		return
	}

	if alertConfig.Route == nil {
		c.String(http.StatusInternalServerError, "Alertmanager config has no root route")
		return
	}

	router := pipeline.NewRouter(alertConfig.Route)
	runner := pipeline.NewRunner(stores.NewSilenceStore(nil), stores.NewAlertStore(), router, nil)
	walker := snapshot.NewRouteWalker(alertConfig.Route)
	paths := walker.FindTerminalPaths()

	synthesizer := snapshot.NewSnapshotSynthesizer(runner)
	outcomes, err := synthesizer.DiscoverOutcomes(context.Background(), paths)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Synthesis failed: %v", err))
		return
	}
	currentTests := cli.BuildRegressionTests(outcomes, litmusConfig.GlobalLabels)

	resp := diffResponse{
		Total:   len(currentTests),
		Results: make([]*deltaResult, 0),
	}

	state, err := cli.LoadRegressionState(litmusConfig.RegressionsYamlFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			resp.Total = 0
			c.JSON(http.StatusOK, resp)
			return
		}
		c.String(http.StatusInternalServerError, fmt.Sprintf("Loading baseline: %v", err))
		return
	}

	baseline := state.Tests
	if len(baseline) == 0 {
		resp.Total = 0
		c.JSON(http.StatusOK, resp)
		return
	}

	diff := snapshot.ComputeDiff(baseline, currentTests)

	driftedMap := make(map[string]bool)

	for _, delta := range diff.Deltas {
		dr := &deltaResult{
			Name:     fmt.Sprintf("%s route", delta.Kind),
			Pass:     false,
			Kind:     string(delta.Kind),
			Labels:   delta.Labels,
			Expected: delta.Expected,
			Actual:   delta.Actual,
		}

		if delta.Kind != types.DeltaRemoved {
			labelSet := make(model.LabelSet)
			for k, v := range delta.Labels {
				labelSet[model.LabelName(k)] = model.LabelValue(v)
			}
			dr.RoutePath = flattenMatchedPath(traceRoute(alertConfig.Route, labelSet))

			driftedMap[snapshot.LabelKey(delta.Labels)] = true

			if delta.Kind == types.DeltaModified {
				for _, expectedReceiver := range delta.Expected {
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
		}

		resp.Results = append(resp.Results, dr)
		resp.Drifted++
	}

	for _, test := range currentTests {
		if len(test.Labels) == 0 {
			continue
		}
		key := snapshot.LabelKey(test.Labels[0])
		if !driftedMap[key] {
			resp.Results = append(resp.Results, &deltaResult{
				Name:   test.Name,
				Pass:   true,
				Kind:   "passing",
				Labels: test.Labels[0],
				Actual: test.Expect.Receivers,
			})
		}
	}

	resp.Passed = resp.Total - resp.Drifted

	c.JSON(http.StatusOK, resp)
}

func healthHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func serveStatic(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.Next()
		return
	}

	if staticFS == nil {
		c.String(http.StatusNotFound, "UI filesystem not initialized")
		return
	}

	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Check if the file exists in the filesystem
	if _, err := fs.Stat(staticFS, path); err != nil {
		if !strings.Contains(path, ".") {
			c.Request.URL.Path = "/"
		}
	}

	http.FileServer(http.FS(staticFS)).ServeHTTP(c.Writer, c.Request)
}
