# Regression Testing: Snapshot & Synthesis Specification

The goal of regression testing in `litmus` is to capture the ground truth of a running configuration and ensure that any future changes to the routing tree do not silently alter which receivers get notified.

## 1. The Snapshotting Algorithm
The `litmus snapshot` command performs a deep traversal of the Alertmanager routing tree to generate a "Lockfile" of behaviors.

### A. Terminal Node Exploration
For every "Terminal Node" (a leaf in the route tree that assigns a receiver), the algorithm starts a path exploration, tracking the cumulative matchers from the root.

### B. Regex Expansion & Balanced Option Coverage
To ensure full logical coverage without "Combinatorial Explosion":
*   **Expansion Rule:** `service=~"(api|db)"` is expanded into its discrete possibilities.
*   **Balanced Option Coverage:** 
    *   If total possible combinations for a path are $\le 5$, `litmus` generates the full Cartesian product.
    *   If combinations exceed **5**, `litmus` switches to a **Covering Set** strategy: it generates the minimum number of explicit label maps (max 5) required to ensure every individual option in every regex alternation is exercised at least once.
*   **Deterministic Sampling (Non-Alternation):**
    *   **Anchored Prefixes:** `^api-.*` resolves to `api-`.
    *   **Character Classes:** `[a-z]` to `a`.
    *   **Wildcards:** `.*` to `litmus_match`.

### C. Global Label Injection
During synthesis, `litmus` automatically merges the `global_labels` defined in `litmus.yaml` into every generated label map. This ensures the synthesized alert is "thick" enough to satisfy top-level routing requirements.

### D. Outcome-Based Grouping (Deduplication)
Label sets that produce the **exact same global outcome** (the ordered list of receivers triggered across the entire tree) are collected into a single test case to keep the baseline clean.

## 2. Test Case Signature
```yaml
name: "Regression: Route Path [0] -> [db-team]"
labels:
  - service: "mysql", env: "prod", severity: "warning" # severity from global_labels
  - service: "postgres", env: "prod", severity: "warning"
expect:
  - "db-team"
tags:
  - "regression"
```

## 3. Storage
Regression baselines are stored in two formats:
1.  **`regressions.litmus.mpk`**: The binary, protected Source of Truth.
2.  **`regressions.litmus.yml`**: A human-readable audit mirror.
