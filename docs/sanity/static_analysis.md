# Static Analysis (Sanity): Specification

Static Analysis in `litmus` provides a "Linter" for Alertmanager configurations. It identifies logical contradictions and human errors that are syntactically valid but behaviorally broken.

## 1. Core Analysis Rules

### A. Shadowed Routes

Route B is unreachable (shadowed) when all three conditions hold:

1. Route A appears **before** B in the route tree (sibling or ancestor)
2. Route A has **`continue: false`** (the default) — if `continue: true`, Alertmanager evaluates B regardless of whether A matched
3. Route A's positive matchers are a **subset** of Route B's positive matchers — A is broader (fewer constraints → larger match set → catches every alert B would catch)

#### Algorithm

The detector walks the route tree collecting every root-to-leaf path as a `sanityPath` with cumulative typed matchers. For each path pair `(B, A)` where A precedes B:

```
isShadowed(child=B, parent=A):
  1. if A.continue == true  → return false   (A passes through, B is still evaluated)
  2. build childPos  = { name → value | m ∈ B.matchers, m.type is positive }
  3. build parentPos = { name → value | m ∈ A.matchers, m.type is positive }
  4. for each (name, value) in parentPos:
       if name not in childPos OR childPos[name] ≠ value → return false
       (A has a positive constraint not present in B → A is more specific, not broader)
  5. for each (name, value) in A.negative_matchers:
       if childPos[name] == value → return false
       (A excludes the exact label+value B requires → routes are mutually exclusive)
  6. return true  (A is broader and blocks B)
```

#### Matcher polarity

All three Alertmanager matcher formats are understood:

| Format | Example | Polarity |
|--------|---------|----------|
| `match` (deprecated) | `label_team: "security"` | always positive |
| `match_re` (deprecated) | `match_re: {env: "prod.*"}` | always positive |
| `matchers` (modern) | `- dh_app!~"dashboard"` | operator determines polarity: `=`/`=~` positive, `!=`/`!~` negative |

Negative matchers narrow a route's match set — `type!~"metaflow-k8s"` does not intercept alerts requiring `type=~"metaflow-k8s"`; they are mutually exclusive. Negative matchers are excluded from the subset check and used only for the mutual-exclusion guard (step 5).

#### Scope and limitations

- Shadowing analysis runs only in the sanity check layer. Regression test synthesis (`snapshot`) uses a separate route walker (`RouteWalker`) that captures plain label values for alert generation; negative-matcher semantics are not applied there.
- Matcher values are compared as **raw strings** — regex patterns are not evaluated or expanded. Two routes with semantically equivalent but differently-written patterns (e.g. `"a|b"` vs `"b|a"`) are treated as distinct. This means some mutual-exclusion cases involving regex supersets may not be detected (false negatives are preferred over false positives).
- `active_time_intervals` are not considered — a route gated by a time interval may still be reported if its matchers shadow another route outside that interval.

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
