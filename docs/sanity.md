# Sanity Checks

Litmus runs static analysis on your Alertmanager configuration before executing any tests. Each check can be set to `fail` (default) or `warn` in `.litmus.yaml`.

```yaml
sanity:
  orphan_receivers: warn
  dead_routes: fail
  shadowed_routes: fail
  inhibition_cycles: fail
  negative_only_routes: warn
  policy_violations: fail
  require_tests: fail
  require_regression: fail
```

---

## Implemented Checks

### `orphan_receivers`

**Definition:** A receiver is defined in the `receivers:` block but never referenced by any route in the routing tree.

**What it catches:** Dead configuration. The receiver can never be invoked, which usually means either a route was deleted without removing the receiver, or the receiver was added but the corresponding route was never written.

**Configuration:**

```yaml
sanity:
  orphan_receivers: warn | fail   # default: fail
```

---

### `dead_routes`

**Definition:** A route's own matchers contradict its ancestor matchers on the same label, making the route permanently unreachable regardless of any alert that arrives.

**What it catches:** Vertical unreachability. An ancestor route requires `env=prod` and a descendant requires `env=staging` — no alert can satisfy both, so the descendant is a dead branch. Distinct from `shadowed_routes`, which detects horizontal shadowing between siblings. Both positive/positive conflicts (two different exact values) and positive/negative conflicts (a positive value negated by a parent) are detected, including regex variants.

**Configuration:**

```yaml
sanity:
  dead_routes: warn | fail   # default: fail
```

---

### `shadowed_routes`

**Definition:** A route is unreachable because an earlier sibling route with broader (or equal) matchers and `continue: false` will always match first.

**What it catches:** Horizontal unreachability. Route A has matchers `{env=prod}` and Route B (defined after A) has matchers `{env=prod, team=infra}`. Every alert that matches B also matches A, and A has `continue: false`, so B is never evaluated. The detector checks that parent's positive matchers are a subset of the child's and that no parent negative matcher excludes the child's required labels.

**Configuration:**

```yaml
sanity:
  shadowed_routes: warn | fail   # default: fail
```

---

### `inhibition_cycles`

**Definition:** Two or more inhibition rules form a directed cycle where rule A suppresses alerts matching rule B's source, and rule B suppresses alerts matching rule A's source.

**What it catches:** Mutual suppression loops. In a cycle, which alert fires depends on arrival order rather than intent, leading to unpredictable suppression behavior. The detector builds a directed graph of inhibition rules and runs DFS cycle detection.

**Configuration:**

```yaml
sanity:
  inhibition_cycles: warn | fail   # default: fail
```

---

### `negative_only_routes`

**Definition:** A route's matcher set contains only negative matchers (`!=` or `!~`) with no positive matchers.

**What it catches:** Implicit over-broad matching. A route with only `{env!="staging"}` matches every alert that does not have `env=staging` — including alerts with no `env` label at all. This is rarely the intended behavior and makes routing logic hard to reason about. The check uses `route.Matchers` only; legacy `match:` / `match_re:` maps are not flagged since they cannot express negative-only patterns.

**Configuration:**

```yaml
sanity:
  negative_only_routes: warn | fail   # default: fail
```

---

### `policy_violations`

Two sub-checks share this key. Both are controlled by the `policy:` block, not `sanity:`.

#### Enforce matchers (`enforce_matchers`)

**Definition:** Leaf routes in a fragment do not match any of the required label names defined in `policy.enforce.matchers`.

**What it catches:** Routes that bypass org-wide routing standards. For example, a policy requiring every leaf route to filter on `team` or `service` prevents catch-all routes that send all alerts to one receiver. In `strict` mode every required label must appear; in non-strict mode at least one must appear.

**Configuration:**

```yaml
policy:
  enforce:
    strict: true             # true = all required, false = any one required
    matchers:
      - team
      - service
  skip_root:
    - enforce                # skip the root fragment from enforce checks
sanity:
  policy_violations: warn | fail   # default: fail
```

---

### `require_tests`

**Definition:** A fragment has fewer behavioral unit tests than routes defined in it.

**What it catches:** Under-tested routing logic. When `policy.require.tests` is true, `count(tests) >= count(routes)` must hold for every fragment. Routes are counted recursively across the full route tree. Fragments with no routes are skipped. The issue message shows both counts so the gap is immediately visible.

**Configuration:**

```yaml
policy:
  require:
    tests: true
  skip_root:
    - tests                  # skip the root fragment from test requirement
sanity:
  require_tests: warn | fail   # default: fail
```

---

### `require_regression`

**Definition:** No committed regression baseline exists for the workspace.

**What it catches:** Missing snapshot coverage at the system level. When `policy.require.regression` is true, a captured baseline (`.mpk`) must exist before `litmus check` passes. This enforces that routing behavior has been snapshotted at least once, so future changes are caught by regression comparison. Run `litmus snapshot capture` to satisfy this check.

**Configuration:**

```yaml
policy:
  require:
    regression: true
sanity:
  require_regression: warn | fail   # default: fail
```

---

## Planned Checks (Not Yet Implemented)

### `wildcard_route`

**Definition:** A non-root route has no matchers or has `.*` on every label, making it an unintended catch-all that shadows all sibling routes defined after it.

**What it catches:** A child route with zero matchers matches every alert unconditionally. When `continue: false` (the default), all later siblings starve. Differs from `negative_only_routes` in that this route has no matchers at all rather than only negative ones.

---

### `duplicate_sibling_matchers`

**Definition:** Two sibling routes under the same parent have an identical matcher set.

**What it catches:** The second route can never be reached. Unlike `shadowed_routes` (which detects subset relationships), this targets the exact-duplicate case that can appear when routes are copy-pasted during config refactoring.

---

### `default_receiver_missing`

**Definition:** The name given as the root route's receiver does not exist in the `receivers:` list.

**What it catches:** Every alert that does not match any child route is sent to the default receiver. If that receiver name is misspelled or was deleted, Alertmanager will fail to load the config — but this check surfaces the error earlier, in litmus, with a clearer message.

---

### `self_inhibiting_rule`

**Definition:** An inhibition rule's source matchers and target matchers can both be satisfied by the same alert (the matcher sets are not mutually exclusive).

**What it catches:** A rule where `source_matchers: [{alertname=Watchdog}]` and `target_matchers: [{alertname=Watchdog}]` causes an alert to suppress itself. This is usually a copy-paste error and results in silent alert suppression.

---

### `duplicate_inhibition_rules`

**Definition:** Two inhibition rules have identical source matchers, target matchers, and `equal` labels.

**What it catches:** Redundant config that increases cognitive load without effect. Usually left over after a refactor.

---

### `equal_labels_not_in_matchers`

**Definition:** A label name listed in an inhibition rule's `equal:` field does not appear in either the `source_matchers` or `target_matchers` of that rule.

**What it catches:** The `equal:` constraint requires both source and target alerts to carry the same value for those labels. If the label is not constrained by either matcher set, alerts without that label at all will still match, making the equality constraint a no-op and the rule broader than intended.

---

### `undefined_time_interval`

**Definition:** A route references a name in `active_time_intervals` or `mute_time_intervals` that is not defined in the top-level `time_intervals:` block.

**What it catches:** Alertmanager silently ignores undefined interval names, meaning the intended muting or activation window is never applied. Alerts route as if no time restriction exists.

---

### `overlapping_mute_intervals`

**Definition:** Two entries in `mute_time_intervals` define time windows that completely overlap.

**What it catches:** Redundant interval definitions that add noise without effect. The duplicate interval can mask the original intent and makes the config harder to maintain.

---

### `timing_paradox`

**Definition:** A route's effective `group_wait` is greater than its effective `repeat_interval`.

**What it catches:** The first notification fires only after `group_wait` elapses. If `group_wait > repeat_interval`, the repeat window has already passed before the first notification is sent, which can cause alerts to loop or never fire a second notification depending on Alertmanager version. This is almost always a misconfiguration.

---

### `zero_group_interval`

**Definition:** A route's effective `group_interval` is zero (or unset in a context where it resolves to zero).

**What it catches:** `group_interval=0` causes Alertmanager to re-send a notification for every alert update with no debounce, which can produce notification storms during flapping or mass-firing incidents.
