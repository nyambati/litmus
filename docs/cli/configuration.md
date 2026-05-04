# Litmus Configuration Design

## Overview

Litmus uses a structured configuration with sensible defaults to manage Alertmanager configurations and Mimir integration.

## Default Folder Structure

```
.
├── .litmus.yaml          # litmus configuration
├── config/               # workspace root (workspace.root)
│   ├── alertmanager.yml  # alertmanager configuration
│   ├── templates/
│   │   ├── slack.tpl
│   │   └── email.tpl
│   ├── fragments/        # team fragment packages
│   ├── tests/            # behavioral unit tests
│   └── regressions/      # regression baselines
│       ├── regressions.litmus.yml      # active baseline pointer
│       └── 20260503-120000.000000.mpk  # timestamped snapshot
```

## .litmus.yaml Schema

```yaml
workspace:
  root: "config"          # root package directory (alertmanager config, fragments, tests, regressions)
  fragments: "fragments" # discovery pattern for team fragments (relative to root)
  history: 5              # number of regression snapshots to keep

global_labels:            # labels added to every synthesized alert during snapshot
  env: prod

policy:
  require:
    tests: true           # count(tests) >= count(routes) per fragment
    regression: true      # a regression baseline must exist
  skip_root:              # exempt root from specific checks
    - tests
    - enforce
  enforce:
    strict: true          # true = AND (all labels), false = OR (any one label)
    matchers:
      - team
      - severity

sanity:
  orphan_receivers: fail
  dead_routes: fail
  shadowed_routes: fail
  inhibition_cycles: fail
  policy_violations: fail
  negative_only_routes: fail

mimir:
  address: ""             # Grafana Mimir URL
  tenant_id: ""
  api_key: ""             # or use LITMUS_MIMIR_API_KEY env var
```

### Field Reference

| Field | Default | Description |
|-------|---------|-------------|
| `workspace.root` | `config` | Root package directory |
| `workspace.fragments` | `fragments/*` | Fragment discovery pattern (relative to root) |
| `workspace.history` | `5` | Snapshots to retain in history |
| `global_labels` | `{}` | Labels injected into every synthesized alert |
| `policy.require.tests` | `false` | Enforce `count(tests) >= count(routes)` per fragment |
| `policy.require.regression` | `false` | Enforce a committed regression baseline |
| `policy.skip_root` | `[]` | Exempt root from `tests` and/or `enforce` checks |
| `policy.enforce.strict` | `true` | AND mode: all matchers required; `false` = OR mode |
| `policy.enforce.matchers` | `[]` | Label names every route path must accumulate |
| `mimir.address` | `""` | Grafana Mimir push URL |
| `mimir.tenant_id` | `""` | Mimir tenant ID |
| `mimir.api_key` | `""` | Mimir API key |

### Baseline Files

| File | Description |
|------|-------------|
| `regressions/regressions.litmus.yml` | Active baseline state — contains `id` and `tests` |
| `regressions/<timestamp>.mpk` | Timestamped binary snapshot (e.g. `20260503-120000.000000.mpk`) |

### Mimir Configuration

Mimir credentials support environment variable substitution using `env(VAR)` syntax:

```yaml
mimir:
  address: "https://mimir.example.com"
  tenant_id: "anonymous"
  api_key: "env(MIMIR_API_KEY)"
```

When `env(VAR)` is encountered, litmus replaces it with the value of the environment variable.

## Configuration Loading

### Precedence Order

1. **CLI Flags** (highest priority)
2. **Environment Variables** (`LITMUS_*` prefix)
3. **`.litmus.yaml`** (project config)
4. **Defaults** (lowest priority)

### Environment Variables

| Variable | Config Key | Description |
|----------|------------|-------------|
| `LITMUS_WORKSPACE_ROOT` | `workspace.root` | Root package directory |
| `LITMUS_WORKSPACE_FRAGMENTS` | `workspace.fragments` | Fragment discovery pattern |
| `LITMUS_WORKSPACE_HISTORY` | `workspace.history` | Snapshots to retain |
| `LITMUS_MIMIR_ADDRESS` | `mimir.address` | Mimir URL |
| `LITMUS_MIMIR_TENANT_ID` | `mimir.tenant_id` | Tenant ID |
| `LITMUS_MIMIR_API_KEY` | `mimir.api_key` | API key |

## Template File Handling

### Supported Extensions

Litmus reads template files with extensions: `.tpl`, `.tmpl`

### Template Resolution

1. Parse alertmanager config's `templates:` field
2. Resolve each template against the templates directory
3. **Error** if referenced template not found
4. **Ignore** extra template files not referenced in config

Example:
```
config/alertmanager.yml:
  templates:
    - slack.tpl
    - email.tpl

config/templates/:
  ├── slack.tpl    ✓ uploaded
  ├── email.tpl    ✓ uploaded
  └── unused.tmpl  ✗ ignored
```

## Security & Access Control

### Web UI (Serve)
The `litmus serve` command is designed for local development and secure team environments:
- **Binding:** By default, the server binds to `localhost` (or `127.0.0.1`), making it inaccessible from outside the local machine.
- **CORS Policy:** The API enforces a restricted Cross-Origin Resource Sharing (CORS) policy. It only permits requests from `http://localhost:*` and `http://127.0.0.1:*`. This prevents malicious websites from interacting with the Litmus API if a user has the server running.
- **Authentication:** As a local-first CLI tool, `litmus serve` does not currently implement authentication. It should only be exposed on public interfaces using a secure reverse proxy with appropriate auth headers.
