# Project Organization: Specification (Final)

`litmus` is organized as a modular Go project. To prevent "God Files" and manage complexity, the core engine is split into specialized sub-packages based on their testing concerns.

## 1. Directory Structure

```text
litmus/
├── cmd/
│   └── litmus/                # CLI entry point (main.go)
├── internal/
│   ├── engine/                # Core Testing Logic
│   │   ├── pipeline/          # SHARED: Unified executor (Silencer -> Inhibitor -> Router)
│   │   ├── behavioral/        # BUT: Human-authored test management & assertions
│   │   ├── snapshot/          # REGRESSION: Synthesis and lockfile management
│   │   └── sanity/            # SANITY: Static analysis linter rules
│   │       ├── shadowed.go
│   │       ├── inhibition_overlap.go
│   │       ├── cycles.go
│   │       └── orphans.go
│   ├── stores/                # In-memory data providers for the Pipeline
│   │   ├── silence_store.go   # Implements silence.Silences interface
│   │   └── alert_store.go     # Implements provider.Alerts interface
│   ├── types/                 # Shared data structures (The "Dependency Anchor")
│   │   ├── behavioral.go      # BehavioralTest (BUT) struct
│   │   └── regression.go      # RegressionTest struct
│   └── codec/                 # MessagePack and YAML encoding/decoding
├── pkg/
│   └── litmus/                # Public Go API (for use as a library)
├── docs/                      # Specification and documentation
│   ├── cli/
│   │   ├── configuration.md   # litmus.yaml schema
│   │   ├── design.md          # Command design & Lockfile philosophy
│   │   └── ui_ux.md           # Terminal UI and reporting standards
│   ├── testing/
│   │   ├── regression.md      # Regression synthesis & snapshot logic
│   │   └── behavioral.md      # Behavioral Unit Test (BUT) logic
│   ├── sanity/
│   │   └── static_analysis.md # Linter & sanity rules
│   ├── technical/
│   │   ├── pipeline_runner.md # Unified Pipeline execution logic
│   │   └── snapshot_synthesis.md # Snapshot generation logic
│   ├── archive/               # Historical notes (IDEAS.md)
│   ├── backlog.md             # Future enhancements
│   ├── whitepaper.md          # The vision for deterministic validation
│   └── architecture.d2        # D2 architecture diagram
├── go.mod                     # Go module definition
├── go.sum
└── GEMINI.md                  # Agent Engineering Mandates
```
...
