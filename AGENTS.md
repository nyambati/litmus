# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project: Litmus — Alertmanager Validator

Litmus validates Alertmanager configurations through two test types:
- **Regression Tests**: Machine-generated golden baselines (stored in `regressions.litmus.mpk` binary, mirrored as `.yml`)
- **Behavioral Unit Tests (BUT)**: Human-authored intent scenarios in YAML files

Core Pipeline: `Silencer → Inhibitor → Router` (unified execution path in `internal/engine/pipeline`)

## Key Data Types (Dependency Anchor)

All code depends on `internal/types`:
- `RegressionTest`: Machine-generated outcome baseline (Name, Labels, Expected receivers, Tags)
- `BehavioralTest`: Human test scenario with SystemState + alert + expect clause
- `SystemState`: Active alerts + silences for suppression testing
- `AlertSample`: Firing alert with label set
- `Silence`: Maintenance window (labels + comment)

## Build & Test

**Go version:** 1.26.2

```bash
# Test (table-driven tests in *_test.go)
make test ./...
.

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
├── cmd/litmus/main.go          # CLI entry (v0.1.0-alpha, stub)
├── internal/
│   ├── engine/
│   │   ├── pipeline/           # SHARED: Unified Silencer→Inhibitor→Router executor
│   │   ├── behavioral/         # BUT: Test mgmt & assertions
│   │   ├── snapshot/           # Regression: Synthesis & lockfile
│   │   └── sanity/             # Linter: Static analysis rules
│   ├── stores/                 # In-memory data providers (silence_store, alert_store)
│   ├── types/                  # Dependency anchor: RegressionTest, BehavioralTest, etc.
│   └── codec/                  # msgpack + YAML serialization
├── docs/                       # Specifications & tickets
└── graphify-out/               # Knowledge graph (update after code changes)
```

## CLI Commands (From Design Spec)

- `litmus init`: Setup workspace + `tests/` dir + `.gitattributes`
- `litmus snapshot [--update]`: Capture baseline → `regressions.litmus.mpk` + `.yml` mirror (drift detection on existing)
- `litmus check`: Validate (Sanity → Regression → Behavioral), collect all failures, unified report
- `litmus inspect`: Human-read `.mpk` files
- `litmus show`: Visualize routing path for labels

## Coding Standards & Mandates

Read `GEMINI.md` for absolute procedural rules. Highlights:
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
