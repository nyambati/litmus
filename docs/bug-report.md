# Litmus Code Review — Bug Report

**Date:** 2026-04-17  
**Scope:** All non-test Go source files under `internal/` and `cmd/`  
**Principles applied:** correctness, security, dead code elimination, complexity

---

## 🔴 Bugs (broken behavior)

### BUG-01 — Slice aliasing corrupts sibling route matchers
**Files:** `internal/engine/sanity/route_inspector.go:39`, `internal/engine/snapshot/route_walker.go:33`

Both recursive walkers do `current := inherited` then `append(current, ...)`. When `append` does not exceed `inherited`'s capacity, it writes into the shared backing array. Sibling branches at the same tree depth then receive a polluted `inherited` slice with matchers from a previous sibling.

`sanity/route_inspector.go:39`:
```go
current := inherited  // shares backing array
for k, v := range route.Match {
    current = append(current, ...)  // may overwrite sibling's data
```

`snapshot/route_walker.go:33`:
```go
currentMatchers := matchers  // same bug
```

**Fix:** Copy before appending.
```go
current := append([]sanityMatcher{}, inherited...)
// or
currentMatchers := append([]model.LabelSet{}, matchers...)
```

---

### BUG-02 — Alertmanager config loaded with raw `yaml.Unmarshal`, bypasses validation
**File:** `internal/config/loader.go:74`

`amconfig.Config` fields implement custom `UnmarshalYAML` hooks that validate matchers, apply defaults (e.g. `GroupWait`, `GroupInterval`, `RepeatInterval`), and compile regexes. Using raw `yaml.Unmarshal` skips all of this. Routes can be silently misconfigured — compiled matchers are nil, time values are zero.

```go
// current — wrong
if err := yaml.Unmarshal(data, &cfg); err != nil { ...

// fix — use alertmanager's own loader
cfg, err := amconfig.Load(string(data))
```

---

### BUG-03 — `alertIterator.Close()` panics on double-call
**File:** `internal/stores/alert_store.go:128-130`

```go
func (ai *alertIterator) Close() {
    close(ai.done)  // panics if called twice
}
```

The pipeline calls `iter.Close()` early on inhibition match, then the caller might call it again. Any future refactor that calls `Close()` twice will panic at runtime. No `sync.Once` protection.

**Fix:**
```go
func (ai *alertIterator) Close() {
    ai.once.Do(func() { close(ai.done) })
}
```

---

### BUG-04 — `hasChanges` compares only test names, not content
**File:** `internal/cli/snapshot.go:99-114`

Drift detection returns `false` (no drift) when test names match, regardless of whether expected receivers or label sets changed. A route can be rewired to a completely different receiver with no drift reported as long as the route name is unchanged.

```go
// current — wrong
if t.Name == e.Name {
    found = true
```

**Fix:** Compare `Expected` receivers and `Labels` in addition to `Name`.

---

### BUG-05 — `outcomeKey` does not sort receivers — non-deterministic dedup
**File:** `internal/engine/snapshot/snapshot.go:86-96`

The dedup key is built by joining `receivers` in iteration order. Since routes may return receivers in different orders across runs, identical receiver sets produce different keys, creating duplicate regression tests or missing expected deduplication. The comment acknowledges this: *"in production would sort"* — but it wasn't fixed.

**Fix:** Sort `receivers` before building the key.

---

### BUG-06 — `alertStore.Put()` error silently discarded in behavioral executor
**File:** `internal/engine/behavioral/executor.go:55`

```go
alertStore.Put(alertmgrAlert)  // error return ignored
```

If `Put` fails, the active alert is silently absent from the store. The inhibition check then finds no active alerts and returns `active` instead of `inhibited`, producing a false-pass on the test.

**Fix:** Return the error up to `Execute`.

---

## 🟡 Risks (works today, breaks under stress or future change)

### RISK-01 — `os.Exit` inside `internal/cli.RunCheck`
**File:** `internal/cli/check.go:81-87`

```go
if !result.Passed {
    if !sanityResult.Passed {
        os.Exit(3)
    }
    os.Exit(2)
}
```

`os.Exit` inside a library function (`internal/cli`) terminates the process without running deferred cleanup and cannot be tested — any test that exercises a failing check path kills the test binary. Exit codes belong at the `cmd` layer.

**Fix:** Return a typed error (e.g., `ExitCodeError{Code: 2}`) and call `os.Exit` in `cmd/check.go`.

---

### RISK-02 — Empty silence matches all alerts
**File:** `internal/stores/silence_store.go:33-39`

`silenceMatches` iterates over `silence.Labels`. If the map is empty (e.g., from a malformed YAML file with `silences: [{}]`), the loop body never executes and the function returns `true`, silencing every alert.

**Fix:** Return `false` when `silence.Labels` is empty.

---

### RISK-03 — `equalLabelsMatch` with no `equal` labels silently matches all pairs
**File:** `internal/engine/pipeline/pipeline.go:91-97`

An inhibit rule without an `equal` field produces an empty `model.LabelNames` slice. `equalLabelsMatch` loops over it and returns `true` unconditionally — any source alert inhibits any target alert matching the matchers, regardless of team/environment. This may be intended per the Alertmanager spec but is a subtle foot-gun.

**Fix:** Document explicitly or add a guard if cross-team inhibition is undesirable.

---

### RISK-04 — `regexp.MustCompile` called on every `ExpandAlternations` invocation
**File:** `internal/engine/snapshot/regex_expansion.go:19,25,36,56`

Four `regexp.MustCompile` calls are inside the function body. For large route trees with many regex matchers, this recompiles the same patterns thousands of times.

**Fix:** Promote to package-level `var`.

---

### RISK-05 — Swallowed marshal errors in `inspect.go` and `snapshot.go`
**Files:** `internal/cli/inspect.go:17,20`, `internal/cli/snapshot.go:68`

```go
data, _ := json.MarshalIndent(tests, "", "  ")
data, _ := yaml.Marshal(tests)
ymlData, _ := yaml.Marshal(regTests)
```

Errors silently produce empty/partial output with exit code 0. Users see no output and no error.

**Fix:** Check and return errors.

---

## 🔵 Dead Code

### DEAD-01 — `AlertStore.Subscribe` never called in production
**File:** `internal/stores/alert_store.go:52-62`

`Subscribe` is only referenced in `alert_store_test.go`. No production code path uses it. Remove or mark with a `// test-only` comment if retention is intentional.

---

### DEAD-02 — Duplicate `alerts` and `pending` maps in `AlertStore` hold identical data
**File:** `internal/stores/alert_store.go:13-37`

`Put` writes every alert into both `alerts` and `pending`. `GetPending` returns `pending`, `Subscribe` returns `alerts`. Both maps always contain the same set. The distinction has no semantic meaning in this implementation.

**Fix:** Remove `pending`, make `GetPending` call `Subscribe` (or collapse into one map).

---

### DEAD-03 — `SilenceStore.Reset` is never called
**File:** `internal/stores/silence_store.go:43-45`

No call site exists outside tests. Dead export.

---

### DEAD-04 — Early-return guard in `minimalCoveringSet` is unreachable
**File:** `internal/engine/snapshot/regex_expansion.go:143-145`

```go
full := lcg.cartesianProduct(matchers)
if len(full) <= lcg.maxCombinations {
    return full  // unreachable: GenerateCovering only calls this when totalCombos > max
}
```

`GenerateCovering` only calls `minimalCoveringSet` after confirming `totalCombos > max`. The guard is dead.

---

### DEAD-05 — `AlertStore.Get` never called in production
**File:** `internal/stores/alert_store.go:41-49`

Only used in `alert_store_test.go`. No production caller exists.

---

## 🔵 Unnecessary Complexity

### COMPLEX-01 — `LoadFromFile` returns a slice for a single-test format
**File:** `internal/engine/behavioral/loader.go:21-33`

Each behavioral YAML file contains exactly one test (the struct is not a list). Returning `[]*types.BehavioralTest` with a single element just to satisfy `LoadFromDirectory`'s append loop is misleading. Either the format should support multiple tests per file, or `LoadFromFile` should return `*types.BehavioralTest`.

---

### COMPLEX-02 — `expandEnv` in config only processes Mimir fields
**File:** `internal/config/loader.go:105-113`

The `{{ env "VAR" }}` expansion is applied only to `Mimir.Address`, `Mimir.TenantID`, and `Mimir.APIKey`. If someone adds `{{ env "VAR" }}` to another field (e.g., `Config.Directory`), it silently goes unexpanded. Either expand all string fields or remove the mechanism and use plain `os.ExpandEnv`.

---

### COMPLEX-03 — Typo in exported symbol `GitAtrributes`
**File:** `internal/templates/templates.go:13`

```go
GitAtrributes string  // missing second 't'
```

Public API typo. Fix now before this is depended on further. Rename to `GitAttributes`.

---

## Summary

| Severity | Count |
|---|---|
| 🔴 Bugs | 6 |
| 🟡 Risks | 5 |
| 🔵 Dead code | 5 |
| 🔵 Complexity | 3 |

**Highest priority:** BUG-01 (slice aliasing) and BUG-02 (wrong alertmanager YAML loader) — both affect correctness of every test and sanity check run.
