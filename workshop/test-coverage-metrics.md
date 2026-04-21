# Plan: `litmus coverage` — Route Branch Coverage Metrics

## Context
No visibility into which terminal route branches are exercised by existing behavioral tests. Add `litmus coverage` (CLI + UI panel) reporting % of terminal paths covered by at least one test, with a list of uncovered branches to guide test authoring.

---

## How Coverage Is Computed

1. **Denominator** — `RouteWalker.FindTerminalPaths()` (already exists in `internal/engine/snapshot/route_walker.go`) returns every leaf route in the tree. Each `RoutePath` has `Receiver string` and `Matchers []model.LabelSet` (cumulative from root to leaf).

2. **Coverage** — for each behavioral test (`type: unit`), route its `test.Alert.Labels` through `router.Match(labels)` → `[]string`. A `RoutePath` is **covered** when at least one test's matched receivers contains `path.Receiver` AND the test's labels satisfy the path's accumulated matchers (checked via `RouteMatches` per level).

3. **Unique path key** — `fmt.Sprintf("%s#%d", path.Receiver, i)` (receiver + index from `FindTerminalPaths`). Handles duplicate-receiver configs without merging distinct branches.

4. **Description** — compact matcher chain for each path, e.g. `{severity=critical} → {dh_env=production}`, built from `RoutePath.Matchers`.

---

## Output (CLI)

```
Route Coverage: 7/10 (70.0%)

Covered (7):
  ✓ pagerduty           3 tests  ({severity=critical, dh_env=production})
  ✓ slack-infra         1 test   ({team=infra})
  ✓ missing-squad-label 2 tests  ({label_team=""})
  ...

Uncovered (3):
  ✗ opsgenie-call       ({severity=critical} → {dh_env=~prd.*})
  ✗ slack-warnings      ({severity=warning})
  ✗ default             (catch-all)
```

---

## New Files

### `internal/engine/coverage/analyzer.go`
```go
type PathCoverage struct {
    Receiver    string
    Description string   // compact matcher chain
    Tests       []string // test names that cover this path
    Covered     bool
}

type CoverageReport struct {
    Total     int
    Covered   int
    Percent   float64
    Paths     []*PathCoverage
}

func Analyze(
    paths  []*snapshot.RoutePath,
    tests  []*types.TestCase,
    router *pipeline.Router,
) *CoverageReport
```

`Analyze` logic:
- For each test with non-nil `Alert`, call `router.Match(labelSet)` → receiver set
- For each `RoutePath[i]`: covered if `path.Receiver ∈ receiverSet` (optimistic — receiver name match)
- Tracks `test.Name` in `PathCoverage.Tests` for attribution
- Builds `Description` from `RoutePath.Matchers` as `{k=v, ...} → {k=v, ...}`

### `cmd/coverage.go`
```go
func newCoverageCmd() *cobra.Command
// Flags: --format/-f (text|json), default text
// Calls cli.RunCoverage(format)
```

### `internal/cli/coverage.go`
```go
func RunCoverage(format string) error
// Loads litmus config + alertmanager config
// Builds router, walker, loader
// Calls coverage.Analyze(paths, tests, router)
// Prints or marshals result
```

### `ui/src/components/coverage/CoveragePage.tsx`
- Progress bar showing coverage %
- Two collapsible sections: Covered (green) / Uncovered (red)
- Each row: receiver name + matcher chain description + test count badge
- Right-rail stats (passed via `AppLayout stats` prop): Total paths, Covered, Uncovered, %

---

## Files to Modify

### `cmd/root.go`
Register `newCoverageCmd()` alongside existing commands.

### `internal/server/server.go`
Add `GET /api/v1/coverage`:
```go
// Loads config, runs coverage.Analyze, returns CoverageReport as JSON
```

### `ui/src/App.tsx`
Add `<Route path="/coverage" element={<CoveragePage />} />`.

### `ui/src/components/layout/Sidebar.tsx`
Add coverage nav link (after Lab, before or after Regression).

---

## Existing Code Reused (no changes needed)

| Symbol | File | Purpose |
|---|---|---|
| `snapshot.RouteWalker` / `FindTerminalPaths()` | `internal/engine/snapshot/route_walker.go` | Discover terminal paths |
| `snapshot.RoutePath` | same | Path struct with `Receiver` + `Matchers` |
| `pipeline.Router.Match()` | `internal/engine/pipeline/router.go` | Route labels → receivers |
| `behavioral.NewBehavioralTestLoader()` / `LoadFromDirectory()` | `internal/engine/behavioral/loader.go` | Load tests |
| `config.LoadConfig()` / `config.LoadAlertmanagerConfig()` | `internal/config/loader.go` | Config loading |
| `AppLayout` + right-rail `stats` prop pattern | `ui/src/components/layout/AppLayout.tsx` | UI shell |

---

## Backward Compatibility
- No changes to existing commands or their signatures
- New command registered alongside existing ones
- UI panel is additive (new route, new nav link)

---

## Tests
- `TestAnalyze_AllCovered` — all paths covered, expect 100%
- `TestAnalyze_NoCoverage` — no tests, expect 0%
- `TestAnalyze_Partial` — some paths uncovered, correct count and attribution
- `TestAnalyze_DuplicateReceiver` — two paths to same receiver, covered independently
- `TestCoverageCommand_TextOutput` — CLI integration test
- `TestCoverageCommand_JSONOutput` — JSON format check
