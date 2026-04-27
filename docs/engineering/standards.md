# Engineering Standards: The Litmus Manifesto

This document defines the core engineering principles for the `litmus` project. Adherence to these standards ensures a reliable, maintainable, and idiomatic Go codebase.

## 1. Leverage the Ecosystem (Don't Reinvent)
Before implementing a complex algorithm or utility, research the Prometheus and CNCF ecosystem for existing, battle-tested packages.
*   **Rule:** Favor community-supported libraries for "hard problems" (e.g., regex expansion, graph cycle detection, MessagePack encoding).
*   **Goal:** Focus our engineering effort on the unique value of `litmus` rather than re-solving solved problems.

## 2. Test-First Development (TDD)
We are building a testing tool; our own code must be the gold standard of reliability.
*   **Rule:** Every feature or bug fix must start with a failing **Table-Driven Test** in the `*_test.go` file.
*   **Rule:** Maintain high branch coverage for the core `internal/engine` packages.

## 3. Idiomatic Error Handling
Errors are first-class citizens in Go.
*   **Rule:** Never return "blind" errors. Always wrap errors with semantic context: `fmt.Errorf("context: %w", err)`.
*   **Rule:** Use custom error types or sentinel errors for logic-based failures (e.g., `ErrShadowedRoute`) to allow the CLI to provide specific exit codes and advice.
*   **Structure Errors:** Use structured errors when callers need to programmatically distinguish error conditions. Prefer sentinel errors or error types over string matching.
*   **Error Wrapping:** 
    - Use `%v` for simple annotation or creating new errors
    - Use `%w` to preserve original error for programmatic inspection
    - Place `%w` at end of error string: `fmt.Errorf("context: %w", err)`
    - Exception: When wrapping sentinel errors, place `%w` at beginning for immediate category identification
*   **Adding Information:** Avoid redundant information; add meaningful context without duplicating what the underlying error already provides.
*   **Logging:** Don't log errors you return; let callers handle logging to avoid duplication. Log messages should clearly express what went wrong and help diagnose the problem. Be careful with PII in logs. Use `log.Error` sparingly as it's more expensive than lower levels.

## 4. Documentation as Code (GoDoc)
The code should be its own manual.
*   **Rule:** Every exported function, struct, and constant MUST have a GoDoc comment.
*   **Rule:** Focus comments on the **Intent** (the "Why") and edge cases, rather than simply restating the function signature.
*   **Parameters and Configuration:** Document only non-obvious or error-prone fields/parameters. Explain why they are interesting or what pitfalls to avoid. Don't repeat obvious information already clear from parameter names.
*   **Contexts:** Don't document implied behavior (context cancellation → ctx.Err()). Document special expectations about context lifetime, lineage, or attached values. Document when function returns error other than ctx.Err() on cancellation. Document other mechanisms that may interrupt function (e.g., Stop method).
*   **Concurrency:** Document when unclear if operation is read-only or mutating. Document when synchronization is provided by the API. Don't document obvious concurrency safety (read-only operations safe, mutating operations not safe).

## 5. Interface Segregation
Design internal components to depend on the smallest possible set of behaviors.
*   **Rule:** Use small, focused interfaces (e.g., `Muter`, `Router`) rather than concrete types from the Alertmanager library.
*   **Goal:** Enables the "State Store" architecture for fast, in-memory testing without the overhead of real databases.

## 6. Explicit Over Implicit
*   **Variable Naming:** Use self-explanatory names. Avoid single-letter variables except for very short loops. (e.g., `pathIndex` vs `i`, `labelSet` vs `l`).
*   **No Global State:** Avoid `init()` functions for setup. All dependencies (config, stores, loggers) must be explicitly passed into constructors.
*   **Runtime Version:** The project targets **Go 1.26.2** to leverage the latest performance and language features.
*   **Function Design:**
    - Functions and variables should be self-explanatory
    - Functions should have single responsibility only
    - Functions should be small and concise

## 7. Zero-Value Utility & Defensive Logic
*   **Safe Defaults:** Structs should be designed so that their "Zero Value" is safe to use or provides a sensible default (e.g., a `PipelineRunner` with a `nil` store defaults to stateless routing).
*   **Parser Empathy:** Provide human-friendly error messages for YAML/MsgPack parsing issues, including file paths and line numbers where possible.

## 8. Maintain Small Code Complexity
We prioritize readable, maintainable logic over "clever" or highly optimized but obscure code.
*   **Rule:** Keep **Cyclomatic Complexity** low (targeted < 10 per function).
*   **Rule:** Break down complex functions into smaller, single-responsibility helpers.
*   **Measurement:** Use **Codacy** to monitor and gate PRs on complexity trends.

## 9. Security-First Development
As a tool that handles infrastructure configuration, security is paramount.
*   **Rule:** Never log or store potentially sensitive label data (e.g., `api_key`, `secret`) in cleartext.
*   **Rule:** Implement automated dependency scanning to detect vulnerable packages.
*   **Rule:** Use **Codacy** for automated static application security testing (SAST) on every PR.

## 10. Automated Release Discipline (GoReleaser)
We treat the release process as a first-class engineering concern.
*   **Rule:** No manual binary releases. All production artifacts must be generated by **GoReleaser** through GitHub Actions.
*   **Rule:** All releases must follow Semantic Versioning (SemVer).
*   **Rule:** The release must always include the `.gitattributes` helper to ensure the MessagePack lockfiles are readable in Git environments.

## 11. Go Coding Practices Based on Google Style Guide

### Naming Conventions
#### Function and Method Names
- **Avoid repetition**: Don't repeat package names, receiver types, parameter names, or return value types when unnecessary
  - Bad: `func ParseYAMLConfig(input string) (*Config, error)` in package `yamlconfig`
  - Good: `func Parse(input string) (*Config, error)` in package `yamlconfig`
- **Function names**: Use noun-like names for functions that return something; verb-like names for functions that do something
- **Test doubles**: Name stubs by behavior (e.g., `AlwaysCharges`, `AlwaysDeclines`); for multiple types, use explicit names like `StubService`
- **Variables in tests**: Prefix test double names for clarity (e.g., `spyCC` instead of `cc`)

### Package Organization
#### Package Names
- Avoid util/helper/common: Choose package names related to what the package provides
- Package name should be clear from callsite usage (e.g., `spannertest.NewDatabaseFromFile(...)`)

#### Package Size
- Group related types in same package when implementation is tightly coupled
- Split into different packages when conceptually distinct
- Standard library demonstrates good scoping and layering

#### File Organization
- No "one type, one file" convention
- Group related code by file
- Avoid extremely large files (many thousands of lines) or many tiny files
- Files should be focused enough that maintainers can locate code easily

### Imports
#### Proto Imports
- Use descriptive names with `pb` or `grpc` suffix (e.g., `foopb`, `foogrpc`)
- Prefer whole words; avoid ambiguity
- When in doubt, use proto package name up to `_go` with `pb` suffix

#### Import Ordering
- Follow Go Style Decisions for import grouping

### Shadowing
#### Stomping vs Shadowing
- **Stomping**: Reusing variable with `:=` in same scope - OK when original value no longer needed
- **Shadowing**: Creating new variable with `:=` in inner scope - Avoid unless intentional; prefer new names for clarity
- Be careful with standard package names: Avoid variables named after standard packages (e.g., `url`) as it makes those packages inaccessible

#### Package Names
- Avoid names requiring import renaming or causing shadowing of otherwise good variable names at client side