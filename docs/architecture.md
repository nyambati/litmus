# Architecture & Design Philosophy

This document explains the design decisions behind Litmus.

---

## Core Philosophy

### 1. Parity with Production

Litmus uses the official Prometheus Alertmanager libraries. If a test passes in Litmus, it behaves identically in production.

**Why:** No simulation gaps. Tests validate the actual code path, not a mock.

**How:** Litmus orchestrates the real Alertmanager routing engine (Silencer → Inhibitor → Router) with in-memory data stores.

---

### 2. Three Layers of Validation

Alerting validation requires multiple perspectives:

#### **Regression Layer** (Automated Safety Net)
Capture the current behavior as a baseline. Any future change that alters routing is flagged.

**Goal:** Prevent accidental regressions  
**Method:** Synthesize all possible routing paths, store as lockfile  
**When it catches issues:** Config changes with unintended routing effects

#### **Behavioral Layer** (Intent Verification)
Write tests that express your routing intent under specific conditions.

**Goal:** Verify logic matches intent  
**Method:** Simulate system states (active alerts, silences) and assert outcomes  
**When it catches issues:** Logic errors that don't show as regressions (e.g., "this team should NOT get these alerts")

#### **Sanity Layer** (Static Analysis)
Audit the configuration for logical impossibilities before execution.

**Goal:** Catch obvious errors early  
**Method:** Structural analysis (dead routes, orphaned receivers, cycles)  
**When it catches issues:** Config problems that won't fail at runtime (e.g., unused receivers)

---

### 3. Lockfile Pattern

The regression baseline (`.mpk` file) is the single source of truth.

**Properties:**
- **Binary format** — Protected from accidental hand-edits
- **Version controlled** — Tracked alongside config changes
- **Human-readable mirror** — `.yml` file for auditing and Git diffs
- **Deterministic** — Same config always produces same baseline

**When you update:**
1. You modify `alertmanager.yaml`
2. `litmus check` detects drift
3. You review changes with `litmus diff`
4. You intentionally accept changes with `litmus snapshot --update`
5. Both `.mpk` and `.yml` are updated and committed

This ensures baseline updates are **deliberate**, not accidental.

---

### 4. Shift Left

Validate alert routing *before* it reaches production.

**Traditional workflow:**
```
Config Change → Deploy to Production → Alert Incident → Post-Mortem
```

**Litmus workflow:**
```
Config Change → litmus check (in CI/CD) → Tests Pass → Deploy to Production
```

**Benefit:** Catch routing errors during development, not in production.

---

## Design Choices

### Why Go?

- **Parity:** Use official Alertmanager libraries (written in Go)
- **Performance:** Fast validation (thousands of paths in < 2 seconds)
- **Deployment:** Single binary, no runtime dependencies

### Why MessagePack for Baselines?

- **Integrity:** Binary format prevents accidental edits
- **Efficiency:** Compact representation for large configs
- **Audit Trail:** YAML mirror provides human-readable history

### Why In-Memory Stores?

Instead of:
- Real databases (PostgreSQL, etc.)
- Mock interfaces

Litmus uses:
- **In-memory data structures** (maps, slices)
- **Official Alertmanager code path** (no mocks)

**Benefit:** Speed + Parity. Validation is fast enough for CI/CD loops.

### Why Three Test Types (Sanity, Regression, Behavioral)?

Different error types require different detection methods:

| Error Type | Detection |
|-----------|-----------|
| "I have dead code" | Static analysis (sanity) |
| "I broke routing by accident" | Regression testing |
| "My logic doesn't match intent" | Behavioral tests |

Using all three provides defense in depth.

---

## Non-Goals

Litmus is **not**:
- A configuration linter (that's for basic syntax errors)
- A monitoring system (doesn't integrate with Prometheus)
- A template validator (doesn't verify receiver template variables)
- A multi-tenant solution (single workspace per directory)

---

## Comparison to Other Tools

### vs. `amtool` (Alertmanager CLI)

`amtool` provides:
- Manual testing of specific label sets
- Config syntax validation

Litmus provides:
- Automated synthesis of all routing paths
- Behavioral test scenarios
- Regression detection
- Static analysis

**Use together:** `amtool` for one-off debugging, Litmus for CI/CD validation.

### vs. OPA/Rego Policy Engine

OPA is a general-purpose policy language. You could write Rego policies to validate Alertmanager configs.

**Litmus advantages:**
- Purpose-built for Alertmanager routing
- No policy language to learn
- Real parity with Alertmanager (uses official libraries)
- Behavioral scenarios (system state simulation)

---

## Extensibility

### Current Scope

- Validation: Check, Sanity, Regression, Behavioral
- Inspection: Inspect, Diff
- Initialization: Init

### Future Scope (See Backlog)

- **Configuration fragments:** Multi-team ownership
- **Template validation:** Verify receiver templates
- **Time simulation:** Test time-based routing
- **Negative matcher synthesis:** Handle `!=` and `!~` matchers

---

## Principles

### Principle 1: Fail Fast
- Compilation/parsing errors stop execution immediately
- No silent skips or partial failures

### Principle 2: Determinism
- Same config always produces same results
- No randomness or time-dependent behavior (except intentional time tests)

### Principle 3: Transparency
- Diff output shows exactly what changed
- Test failures explain what went wrong
- Config updates are explicit (not implicit)

### Principle 4: Integration
- Works in CI/CD pipelines
- Exit codes follow conventions (0 = pass, 1-3 = specific failures)
- Plain text output (no JSON cruft unless necessary)

### Principle 5: User Empowerment
- Users control baseline updates (no auto-sync)
- Behavioral tests express intent, not just mechanics
- Clear error messages guide fixes

---

## Data Flow

### Snapshot (Regression Baseline Creation)

```
alertmanager.yaml
    ↓
[Parse Config]
    ↓
[Synthesize Routes]
    ├─ Expand regex matchers
    ├─ Combine label sets
    └─ Generate alert combinations
    ↓
[Execute Pipeline]
    ├─ Route each alert
    ├─ Check silences
    ├─ Apply inhibition
    └─ Collect outcomes
    ↓
[Create Regression Tests]
    └─ One test per unique routing outcome
    ↓
regressions.litmus.mpk + .yml
```

### Check (Validation)

```
alertmanager.yaml
    ↓
[Sanity: Static Analysis]
    ├─ Check for shadowed routes
    ├─ Check for orphaned receivers
    └─ Check for circular inhibitions
    ↓
[Regression: Load Baseline]
    ├─ Load regressions.litmus.mpk
    ├─ Synthesize current config
    └─ Compare outcomes
    ↓
[Behavioral: Load and Run Tests]
    ├─ Load tests from tests/
    ├─ Execute each test scenario
    └─ Assert outcomes
    ↓
[Report Results]
    ├─ Aggregate all failures
    └─ Print unified report
```

---

## Testing Strategy

Litmus itself has three test layers:

1. **Unit tests** — Core logic (routing, synthesis, parsing)
2. **Integration tests** — End-to-end workflows (CLI commands)
3. **Regression tests** — Litmus validates itself using Litmus

---

## Performance

**Target:** Validate 1,000+ routing paths in < 2 seconds

**Achieved by:**
- In-memory state stores (no I/O)
- Single-pass synthesis (no repeated traversals)
- Efficient label combination generation
- Bounded iteration (configurable `max_samples`)

---

## Security Considerations

Litmus processes untrusted Alertmanager configurations:
- **No code execution** from config (YAML is data only)
- **No external calls** (isolated validation)
- **No secrets management** (doesn't handle API keys)

**Sensitive data handling:** Litmus doesn't redact labels or values. If your config contains secrets in labels, they'll appear in diffs.

---

## Version Compatibility

Litmus tracks Prometheus Alertmanager versions. Target support: **3 latest minor releases**.

Example: If Alertmanager is at 0.25.x, Litmus supports 0.23, 0.24, 0.25.

---

## Next Steps

- **Get Started:** See `docs/cli/user_guide.md`
- **Whitepaper:** Read `docs/whitepaper.md` for motivation
- **Backlog:** See `docs/backlog.md` for future features
