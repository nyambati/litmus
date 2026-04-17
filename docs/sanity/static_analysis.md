# Static Analysis (Sanity): Specification

Static Analysis in `litmus` provides a "Linter" for Alertmanager configurations. It identifies logical contradictions and human errors that are syntactically valid but behaviorally broken.

## 1. Core Analysis Rules

### A. Shadowed Routes
If Route A is above Route B, and B is a logical subset of A, and A does not `continue`, Route B is unreachable.

### B. Global Inhibition Overlap (Cross-Inhibition)
*   **The Check:** For every regression test case, `litmus` checks if any *other* existing inhibition rule in the system could accidentally suppress that alert.
*   **Impact:** Critical. Prevents new inhibition rules from silencing unrelated critical alerts.

### C. Version Drift
*   **The Check:** Compares the compiled-in Alertmanager library version against the `version` or features used in the `alertmanager.yml`.
*   **Impact:** Warning. Ensures `litmus` doesn't misinterpret config fields it is too old to understand.

### D. Inhibition Cycles
Detects circular dependencies in `inhibit_rules` (e.g., A inhibits B, B inhibits A).

### E. Orphan Receivers
Identifies receivers that are defined but never referenced by any route.

### F. Mutually Exclusive Matchers
Detects routes that can never be reached because their matcher path contains logical contradictions (e.g., `env=prod` and `env=dev` on the same branch).

## 2. Severity Levels
*   **ERROR:** Shadowed routes, Inhibition cycles, Cross-Inhibition overlap, Unreachable routes. (Fails `litmus check`).
*   **WARNING:** Orphan receivers, Version drift, Redundant matchers.
