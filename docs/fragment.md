# Modular Config: Package-First Architecture

## Overview

Litmus uses a "Package-First" architecture where the root configuration and all team-level
fragments are treated as self-contained units. A single `workspace` block in `.litmus.yaml`
defines the entire layout. Global state (regressions, history) lives in the Root Package only.

## 1. Package Types

| Component | Location | Root Package | Fragment Package |
| :--- | :--- | :---: | :---: |
| **Config** | `alertmanager.yml` / `fragment.yml` | ✅ | ✅ |
| **Tests** | `tests/` or `*-tests.yml` | ✅ | ✅ |
| **Regressions** | `regressions/` | ✅ | ❌ |
| **Fragments** | `fragments/` | ✅ (hosts them) | ❌ |

## 2. Configuration

```yaml
workspace:
  root: "config"           # Root Package directory
  fragments: "fragments/*" # Glob pattern — relative to root by default
  history: 5               # Regression baselines to keep

# Policy is optional. Enforced during 'litmus check'.
# Applies to root and all fragments unless skip_root: true.
# See docs/policies.md for full details on strict vs non-strict mode.
# policy:
#   require_tests: true
#   skip_root: false
#   enforce:
#     strict: true     # AND — all labels must be present in the accumulated path
#     matchers:
#       - team
#       - service
```

The `fragments` value is a glob. Beyond the default, useful forms include:
- `fragments/team-*` — only load fragments whose folder starts with `team-`
- `../../platform/fragments/*` — cross-repo fragment discovery
- `fragments/global.yml` — load exactly one file

## 3. Directory Structure

```text
config/                        # ROOT PACKAGE
├── alertmanager.yml           # Base routing skeleton (catch-all + global routes)
├── tests/                     # Root behavioral tests
├── regressions/               # Global regression baseline & history
│   └── regressions.litmus.yml
└── fragments/                 # Team fragment packages
    ├── database/              # Folder package (all YAMLs merged at load time)
    │   ├── receivers.yml
    │   ├── routes.yml
    │   ├── routes-tests.yml   # Sibling test file — matched by *-tests.yml pattern
    │   └── tests/             # Tests subdirectory — all YAMLs loaded as tests
    │       └── unit.yml
    └── networking.yml         # Single-file package
        # networking-tests.yml  ← sibling test file would live here
```

## 4. Fragment Format

A fragment file (or any YAML in a folder package) may define:

```yaml
# db-team.yml
name: "db-team"              # Optional; defaults to filename/folder name
namespace: "db"              # Prefixes all receiver names: db-critical, db-warning, etc.
group:                       # Optional. Creates a synthetic parent route for this fragment's routes.
  match:                     # Exact-match labels on the synthetic parent route
    scope: "teams"
  receiver: "teams-fallback" # Optional; defaults to root route's receiver

routes:
  - receiver: "critical"
    matchers:
      - service = mysql

receivers:
  - name: "critical"
    slack_configs: [...]

inhibit_rules: []
```

Tests are always in separate files — never embedded in the fragment definition:

```yaml
# db-team-tests.yml  (sibling file, auto-discovered)
- name: "mysql routes to db-critical"
  alert:
    labels:
      scope: "teams"
      service: "mysql"
  expect:
    outcome: active
    receivers:
      - db-critical        # namespace applied automatically in test runner
```

## 5. Virtual Assembly

Assembly is in-memory only — nothing is written to disk. The assembled config is used for
all validation and testing; the raw base YAML is what gets pushed to Mimir via `litmus sync`.

```
Discovery
  Root Package  ──► load alertmanager.yml + tests/
  Fragments     ──► load each fragment (config + tests)

Assembly
  1. Receiver namespacing   fragment namespace + "-" + receiver name
  2. Route grouping         fragments with same group.match share one synthetic parent route
                            fragments with no group are merged flat into root
  3. Inhibit rule merge     fragment rules appended to base rules

Execution (all against assembled config)
  Sanity checks     shadowed routes, orphan receivers, inhibition cycles, policy
  Behavioral tests  root tests/ + all fragment tests
  Regression tests  route-walk synthesis against regression baseline
```

### Namespace prefixing

All receiver names and route receiver references within a fragment are prefixed with
`<namespace>-`. A fragment with `namespace: db` and receiver `critical` becomes `db-critical`
in the assembled config. No double-prefixing: the assembler checks for the prefix before
applying it.

### Group routing

`group` is optional. When present, the assembler creates a single synthetic parent route
with `group.match` as its exact-match label matchers and appends all fragment routes as
children of that parent. `group.receiver` sets the parent's receiver — omit to inherit
the root receiver.

Two fragments with identical `group.match` labels share one synthetic parent (their routes
are co-located under it). Two fragments with the same match but different `group.receiver`
values → assembly error.

A fragment with no `group` has its routes merged flat into root — no synthetic parent created.

## 6. Policy Enforcement

Policy rules run during `litmus check` as part of the sanity stage.

| Rule | Applies to | Behaviour |
| :--- | :--- | :--- |
| Rule | Applies to | Behaviour |
| `require_tests` | Root + all fragments | Each package must have ≥1 test |
| `enforce.matchers` | All routes (recursive) | Every route path must accumulate the required labels |
| `skip_root` | Root package | When `true`, root is exempt from all policy checks |

`enforce.matchers` checks the **accumulated** label names from a route and all its ancestors, not just the route's own matchers. A route that inherits a required label from a parent is not flagged. See [docs/policies.md](policies.md) for strict vs non-strict mode and full traces.

## 7. `litmus init` Scaffold

`litmus init` creates the full workspace skeleton:

```text
.litmus.yaml
config/
├── tests/
├── regressions/
├── templates/
└── fragments/
```
