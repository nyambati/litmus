# Project Organization: Specification (Final)

`litmus` is organized as a modular Go project. To prevent "God Files" and manage complexity, the core engine is split into specialized sub-packages based on their testing concerns.

## 1. Directory Structure

```text
litmus/
├── main.go                    # CLI entry point
├── cmd/
│   ├── root.go                # Root command & registration
│   ├── init.go
│   ├── snapshot.go
│   ├── check.go
│   ├── diff.go
│   ├── inspect.go
│   ├── show.go
│   └── sync.go
├── internal/
│   ├── engine/                # Core Testing Logic
│   │   ├── pipeline/          # SHARED: Unified executor (Silencer → Inhibitor → Router)
│   │   ├── behavioral/        # BUT: Human-authored test management & assertions
│   │   ├── snapshot/          # REGRESSION: Synthesis and lockfile management
│   │   ├── sanity/            # SANITY: Static analysis linter rules
│   │   └── matching/          # Receiver matching utilities
│   ├── stores/                # In-memory data providers for the Pipeline
│   │   ├── silence_store.go   # Implements silence.Silences interface
│   │   └── alert_store.go     # Implements provider.Alerts interface
│   ├── cli/                   # CLI business logic
│   │   ├── check.go
│   │   ├── snapshot.go
│   │   ├── diff.go
│   │   └── sync.go
│   ├── config/                # Configuration loading & env expansion
│   ├── mimir/                 # Grafana Mimir API client
│   ├── types/                 # Shared data structures (Dependency Anchor)
│   │   ├── behavioral.go
│   │   └── regression.go
│   └── codec/                 # MessagePack and YAML serialization
├── docs/                      # User documentation
│   ├── cli/
│   │   ├── configuration.md   # litmus.yaml schema
│   │   └── user_guide.md      # How to use each command
│   ├── README.md              # Quick start
│   ├── INDEX.md               # Documentation index
│   ├── whitepaper.md          # Vision and motivation
│   ├── architecture.md        # Design philosophy
│   ├── backlog.md             # Future enhancements
│   ├── project_organization.md # Project structure
│   └── engineering_standards.md # Coding standards
├── go.mod                     # Go module definition
├── go.sum
├── Makefile                   # Build targets
├── CLAUDE.md                  # Claude Code instructions
└── .gitattributes             # Git configuration for binary diffs
```
...
