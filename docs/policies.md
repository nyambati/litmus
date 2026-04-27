# Policies

Litmus policies let you enforce workspace-wide rules on all fragment packages. Policies run automatically during `litmus check` and surface issues in the **Sanity** section of the report.

Policies are optional. Add a `policy:` block to `.litmus.yaml` to enable them.

---

## Configuration

```yaml
policy:
  require_tests: true          # every fragment must have at least one test
  skip_root: true              # exclude the root package from all policy checks

  enforce:
    strict: true               # AND mode (default) — all labels must be covered
    matchers:
      - label_team
      - severity
```

| Field | Type | Default | Description |
|---|---|---|---|
| `require_tests` | bool | `false` | Fragment must have at least one behavioral test |
| `skip_root` | bool | `false` | Exempt the root package from all checks |
| `enforce.matchers` | list | `[]` | Label names that must appear somewhere in each route path |
| `enforce.strict` | bool | `true` | `true` = AND mode, `false` = OR mode (see below) |

---

## `require_tests`

When `true`, every fragment that has routes must also have at least one behavioral test. Fragments with no routes are exempt.

```yaml
policy:
  require_tests: true
```

**Pass:** fragment has a `*-tests.yml` sibling or a `tests/` subdirectory with at least one test case.

**Fail:**
```
fragment "payments" has no tests (policy: require_tests=true)
```

---

## `enforce_matchers`

Enforces that certain label matchers appear somewhere along every routing path in a fragment. This ensures alerts are properly discriminated before reaching a receiver — for example, that every route path carries a team label and a severity label so alerts are always attributable and actionable.

### How path coverage works

Labels accumulate as you descend the route tree. A route inherits all matchers from its ancestors. Litmus checks the **union** of a route's own matchers and all its ancestor matchers — not just the matchers on the route itself.

A routing path is **covered** when its accumulated union contains the required labels. Once a path is covered, Litmus stops descending (descendants are implicitly covered by inheritance).

A violation is reported at the **deepest route** in a path that remains uncovered — the route where coverage could have been completed but was not.

### Strict mode (AND — default)

`enforce.strict: true`

**All** labels in `enforce.matchers` must be present collectively across the route path. Every label must appear at least once somewhere between the root of the fragment and the route being evaluated.

**Use this when** you need every alert path to carry a full set of classification labels. Partial coverage is not acceptable.

```yaml
enforce:
  strict: true
  matchers: [label_team, severity]
```

**Traces:**

```
# Parent covers one label, children complete the set → CLEAN
parent      [label_team]
  child1    [severity]        union={label_team,severity} → covered
  child2    [severity]        union={label_team,severity} → covered
→ no violations
```

```
# Parent covers one label, one child completes, one doesn't → violation on the gap
parent      [label_team]
  child1    [severity]        union={label_team,severity} → covered
  child2    []                union={label_team}          → missing severity → VIOLATION
→ violation on child2
```

```
# Parent empty, children each cover a different label → neither path is complete
parent      []
  child1    [label_team]      union={label_team}   → missing severity → VIOLATION
  child2    [severity]        union={severity}      → missing label_team → VIOLATION
→ violations on child1 and child2
```

```
# Grandparent starts coverage, parent completes it → descendants inherit, all clean
grandparent [label_team]
  parent    [severity]        union={label_team,severity} → covered → skip children
    child   []                (skipped — parent already covered the path)
→ no violations
```

### Non-strict mode (OR)

`enforce.strict: false`

**At least one** label from `enforce.matchers` must appear somewhere in the route path. A single matching label satisfies the entire branch.

**Use this when** your matchers are alternatives — e.g. `[label_team, service]` means a route should be discriminated by either a team label or a service label, not necessarily both.

```yaml
enforce:
  strict: false
  matchers: [label_team, severity]
```

**Traces:**

```
# Parent has one required label → satisfies OR → entire branch clean
parent      [label_team]
  child1    [severity]        (skipped — parent already satisfied)
  child2    []                (skipped — parent already satisfied)
→ no violations
```

```
# Parent empty, children each have one required label → each branch satisfied
parent      []
  child1    [label_team]      union={label_team} → has ≥1 required label → covered
  child2    [severity]        union={severity}   → has ≥1 required label → covered
→ no violations (all branches satisfied via children)
```

```
# Nothing anywhere → violation at the leaf
parent      []
  child     []                union={} → no required label → VIOLATION
→ violation on child
```

### Strict vs non-strict at a glance

| Scenario | strict (AND) | non-strict (OR) |
|---|---|---|
| Parent has one label, children complete the set | No violations | No violations |
| Parent has one label, some children don't complete | Violation on gap | No violations |
| Parent empty, children each cover a different label | Violations on both | No violations |
| Parent empty, children each cover any one label | Violations (one label not enough) | No violations |
| Nothing anywhere | Violation | Violation |
| Parent covers everything | No violations | No violations |

### `skip_root`

Root package routes are often catch-all or structural routes that intentionally lack team or severity matchers. Set `skip_root: true` to exclude the root package from enforce_matchers checks entirely.

```yaml
policy:
  skip_root: true
  enforce:
    strict: true
    matchers: [label_team, severity]
```

---

## Full example

```yaml
policy:
  require_tests: true
  skip_root: true
  enforce:
    strict: true
    matchers:
      - label_team
      - severity
```

With this configuration:
- Every fragment (except root) must have at least one test.
- Every routing path in every fragment must carry both `label_team` and `severity` somewhere in its ancestor chain before reaching a receiver.
- The root package is exempt from both checks.
