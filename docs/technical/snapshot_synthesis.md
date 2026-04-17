# Snapshot Synthesis: Technical Specification

Snapshot Synthesis is the logic responsible for transforming a static `alertmanager.yml` into a dynamic set of regression test cases. It lives in `internal/engine/snapshot/`.

## 1. The Synthesis Pipeline
The snapshot generation process is implemented as a sequence of discrete transformation stages.

### Stage 1: Tree Traversal (Path Finding)
Performs a recursive descent of the `Route` tree to identify every "Terminal Node" (a leaf route that leads to a receiver).
*   **Output:** A list of `RoutePath` objects, each representing a single branch of the configuration.

### Stage 2: Label Synthesis (Matcher Extraction)
Extracts positive matchers (`=` and `=~`) from each path and calculates the `global_labels` to create a base set of synthesized labels.

### Stage 3: Regex Expansion (Combinatorial Synthesis)
Expands alternations (`|`) into literal strings and applies the **Balanced Option Coverage** strategy (capping the result at 5 samples per path).

### Stage 4: Outcome Discovery (Dependency on Pipeline)
**The Parity Step.** For every synthesized label set, `snapshot` calls the **`internal/engine/pipeline`** with empty stores.
*   **Logic:** `snapshot` asks the pipeline: "If I fire an alert with these labels, exactly which receivers will be hit?"
*   **Capture:** This empirical result becomes the "Golden Outcome."

### Stage 5: Deduplication & Finalization
Groups all label sets that result in the same "Golden Outcome" into a single, grouped `RegressionTest` case.

## 2. Parity through Shared Dependencies
By using the **Execution Pipeline** during the **Synthesis Pipeline's Stage 4**, `litmus` ensures that the "Reality" it captures in the `regressions.litmus.mpk` is 100% accurate to the Alertmanager engine.

## 3. The Governor: Balanced Coverage Logic
If the Cartesian product of a path's regexes exceeds **5**, the expansion logic switches from "All Combinations" to a **"Balanced Coverage"** set:
1.  Identify all unique options in all regex alternations.
2.  Map each option to at least one synthesized label map.
3.  Ensure that 100% of logical branches are exercised with $O(N)$ maps instead of $O(N^x)$.
