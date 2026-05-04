# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project: Litmus — Alertmanager Validator

Litmus validates Alertmanager configurations through two test types:
- **Regression Tests**: Machine-generated golden baselines (stored as `.mpk` binary, mirrored as `.yml`)
- **Behavioral Unit Tests (BUT)**: Human-authored intent scenarios in YAML files (`*-tests.yml` or `tests/` subdirectory)

Core Pipeline: `Silencer → Inhibitor → Router` (unified execution path in `internal/engine/pipeline`)

## Key Data Types (Dependency Anchor)

All code depends on `internal/types`:
- `TestCase`: Unified type for both unit and regression tests (`type` field: `"unit"` or `"regression"`)
- `TestResult`: Unified execution result for both test types
- `SystemState`: Active alerts + silences for suppression testing (unit tests only)
- `AlertSample`: Firing alert with label set
- `Silence`: Maintenance window (labels + comment)
- `BehavioralExpect`: Expected outcome (outcome string + receivers list)
- `RegressionState`: Active baseline ID + test list
- `RegressionDiff` / `RegressionDelta`: Diff between two baselines

Fragment model (`internal/fragment`):
- `Fragment`: Namespace, routes, receivers, inhibit rules, time intervals — assembled into a full Alertmanager config
- `AugmentedFragment`: Fragment + file metadata
- Tests discovered from sibling `*-tests.yml` files and `tests/` subdirs — **never serialized in Fragment YAML**

## Build & Test

**Go version:** 1.26.2

```bash
# Test (table-driven tests in *_test.go)
make test ./...

make lint

# Build CLI
make build

# Formatter & vet
make fmt
make vet ./...
```

## Architecture

```
litmus/
├── main.go                     # Entrypoint
├── cmd/                        # Cobra command wiring
│   ├── root.go                 # Root cmd, PersistentPreRunE (config + logger injection)
│   ├── check.go                # litmus check
│   ├── snapshot.go             # litmus snapshot capture/update
│   ├── diff.go                 # litmus diff
│   ├── inspect.go              # litmus inspect
│   ├── init.go                 # litmus init
│   ├── sync.go                 # litmus sync (push to Mimir)
│   ├── history.go              # litmus history
│   └── server.go               # litmus serve (web UI)
├── internal/
│   ├── cli/                    # Business logic for each command
│   ├── engine/
│   │   ├── pipeline/           # SHARED: Unified Silencer→Inhibitor→Router executor
│   │   ├── matching/           # Receiver matching logic
│   │   ├── sanity/             # Linter: Static analysis checks
│   │   └── snapshot/           # Regression: Synthesis, diff, route walking
│   ├── fragment/               # Fragment model + loader
│   ├── workspace/              # Workspace assembly (fragments → full AM config)
│   ├── stores/                 # In-memory data providers (silence_store, alert_store)
│   ├── types/                  # Dependency anchor: TestCase, TestResult, SystemState, etc.
│   ├── codec/                  # msgpack + YAML serialization
│   ├── config/                 # LitmusConfig parsing + SanityConfig modes
│   ├── mimir/                  # Grafana Mimir API client
│   ├── server/                 # Web UI HTTP server
│   ├── labelmatcher/           # Label matcher helpers
│   ├── fixtures/               # Test fixtures
│   ├── templates/              # litmus.yaml init template
│   └── utils/                  # Shared utilities
├── docs/                       # Specifications & tickets
└── graphify-out/               # Knowledge graph (update after code changes)
```

## CLI Commands

- `litmus init`: Setup workspace + `.gitattributes`
- `litmus snapshot capture [--strict]`: Capture baseline; warns on drift; `--strict` fails on drift (CI use)
- `litmus snapshot update [--strict]`: Accept drift and update baseline
- `litmus check [-f text|json] [-d] [-t tags]`: Run Sanity → Regression → Behavioral; collect all failures; unified report
- `litmus diff`: Show routing changes from baseline
- `litmus inspect`: Human-read `.mpk` files
- `litmus history`: View regression baseline history
- `litmus sync [--dry-run] [--skip-validate] [-o file]`: Validate then push config to Grafana Mimir
- `litmus serve [-p port] [--dev]`: Start web UI server (default port 8080)

## Sanity Checks

Registered in `DefaultRunner` (`internal/engine/sanity/runner.go`). Each is configurable as `warn` or `fail` in `litmus.yaml`:

| Check | Config key |
|---|---|
| `ShadowedRouteDetector` | `shadowed_routes` |
| `OrphanReceiverDetector` | `orphan_receivers` |
| `InhibitionCycleDetector` | `inhibition_cycles` |
| `DeadRouteDetector` | `dead_routes` |
| `NegativeOnlyRouteDetector` | `negative_only_routes` |
| `RegressionChecker` | `require_regression` |
| `RequireTestsChecker` | `require_tests` |
| `EnforceChecker` | `policy_violations` |

Default mode for unconfigured checks: `fail`.

## Configuration (`litmus.yaml`)

```yaml
workspace:
  root: "config"         # Root package dir
  fragments: "fragments" # Fragment discovery pattern (relative to root)
  history: 5             # Baselines to retain

global_labels: {}        # Labels auto-added to every synthesized alert

policy:
  skip_root: []          # Exempt root from: enforce, tests
  require:
    tests: true
    regression: true
  enforce:
    strict: false        # true=AND all matchers, false=OR any matcher
    matchers: []         # Label names required on every route path

sanity:
  orphan_receivers: warn|fail
  dead_routes: warn|fail
  shadowed_routes: warn|fail
  inhibition_cycles: warn|fail
  policy_violations: warn|fail
  negative_only_routes: warn|fail
  require_tests: warn|fail
  require_regression: warn|fail

mimir:
  address: ""
  tenant_id: ""
  api_key: ""
```

## Coding Standards & Mandates

- **TDD required**: Table-driven tests before features
- **GoDoc**: Every exported symbol needs intent comment
- **Complexity**: < 10 cyclomatic. Break into helpers.
- **Errors**: Wrap with context (`fmt.Errorf`)
- **No global state**. No `init()` setup. Inject all deps.
- **Graph-aware**: Run `graphify query` before cross-cutting changes
- **Parity rule**: Use shared pipeline for all alert logic
- **Isolation**: BUT run in isolated State Stores

## graphify

This project has a graphify knowledge graph at `graphify-out/`.

Rules:
- Use engineering standards defined in `docs/engineering/`
- Before answering architecture questions, read `graphify-out/GRAPH_REPORT.md` for god nodes & community structure
- If `graphify-out/wiki/index.md` exists, navigate it instead of raw files
- **After modifying code**, run `graphify update .` to keep graph current (AST-only, no API cost)

## Documentation Map

| Doc | Purpose |
|---|---|
| `docs/architecture.md` | System design & data flow |
| `docs/fragment.md` | Fragment model spec |
| `docs/sanity.md` | Sanity check catalogue |
| `docs/policies.md` | Policy enforcement rules |
| `docs/cli/` | CLI user guide & configuration reference |
| `docs/engineering/standards.md` | Coding standards |
