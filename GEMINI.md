# Litmus: Agent Engineering Mandates

This document contains absolute procedural mandates for any AI agent contributing to the `litmus` project. These rules take precedence over general defaults.

## 1. Environment & Tools
*   **Runtime:** Go 1.26.2.
*   **CLI Framework:** `spf13/cobra`.
*   **Serialization:** `vmihailenco/msgpack` for regressions, `gopkg.in/yaml.v3` for human configs/tests.
*   **Quality Gating:** `golangci-lint`, `Codacy` (Security/Complexity).
*   **Knowledge Context:** Use **Graphify** for architectural research and contextual queries.

## 2. Coding Standards
*   **Test-First (TDD):** You MUST write Table-Driven Tests in `*_test.go` before implementing any feature logic.
*   **Documentation:** Every exported symbol (Struct, Func, Const) MUST have a GoDoc comment explaining the **Intent**.
*   **Complexity:** Maintain Cyclomatic Complexity < 10. Break down logic into single-responsibility helpers.
*   **Error Handling:** Never return "blind" errors. Wrap with context: `fmt.Errorf("context: %w", err)`.
*   **Naming:** Use explicit, self-explanatory variable names (e.g., `labelFingerprint`). No single-letter variables except in very short loops.
*   **State:** No global state. No `init()` for setup. All dependencies must be injected.

## 3. Security Mandates
*   **Data Protection:** Never log, print, or store sensitive label data (e.g., `api_key`, `secret`) in cleartext.
*   **Scanning:** All PRs must pass Codacy SAST and dependency vulnerability scans.

## 4. Git & Workflow Discipline
*   **No Co-Authors:** The use of `Co-authored-by` in commit messages is strictly forbidden. All work must be attributed to the primary committer.
*   **Gated Commits:** Local commits MUST be gated by `git-prehooks` (LIT-17) that run `go fmt`, `go vet`, and unit tests.
*   **Lockfile Integrity:** Never manually modify `.litmus.mpk` files. Only update them via `litmus snapshot --update`.

## 5. Architectural Alignment
*   **Graph-Awareness:** Before making cross-cutting changes, run `graphify query` to understand the structural dependencies.
*   **Parity:** Always use the shared `internal/engine/pipeline` for any alert execution logic.
*   **Isolation:** Behavioral Unit Tests (BUT) must run in isolated `State Stores`.
*   **The "Star" Dependency:** All internal packages depend on `internal/types`. `types` depends on nothing.

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
