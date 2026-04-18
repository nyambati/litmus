# Litmus CLI User Guide

Learn how to use Litmus to validate your Alertmanager configuration.

---

## Overview

Litmus provides six commands for managing alert configuration validation:

```
litmus init       Initialize workspace
litmus snapshot   Create/update regression baseline
litmus check      Validate configuration (for CI/CD)
litmus diff       Show changes from baseline
litmus inspect    Read binary regression file (for auditing)
litmus sync       Push validated config to Grafana Mimir
```

---

## Getting Started

### Step 1: Initialize

```bash
litmus init
```

This creates:
- **`litmus.yaml`** — Configuration file for Litmus
- **`tests/`** — Directory for behavioral unit tests
- **`.gitattributes`** — Git configuration for binary diffs

### Step 2: Create Baseline

Before Litmus can validate, it needs a baseline to compare against:

```bash
litmus snapshot
```

This generates:
- **`regressions.litmus.mpk`** — Binary regression baseline (git-tracked, protected from hand-edits)
- **`regressions.litmus.yml`** — Human-readable YAML mirror for auditing

**Commit these files to version control.** They represent the "ground truth" of your configuration.

### Step 3: Write Tests (Optional)

Create behavioral tests to verify your routing logic under specific conditions:

```bash
mkdir -p tests
cat > tests/critical-alert.yml << 'EOF'
- name: "Critical database alert reaches on-call"
  system_state:
    active_alerts: []
    silences: []
  alert:
    labels:
      severity: critical
      team: database
  expect:
    receivers: [database-oncall]
EOF
```

### Step 4: Run Validation

```bash
litmus check
```

This runs three validation engines:
1. **Static Analysis** — Checks for shadowed routes, orphaned receivers, circular inhibitions
2. **Regression Tests** — Verifies configuration matches baseline
3. **Behavioral Tests** — Runs your custom test scenarios

Output:
```
Litmus Check: alertmanager.yaml
--------------------------------------------------

1. Sanity (Static Analysis)
   [OK]    No shadowed routes detected
   [OK]    No orphan receivers
   [OK]    No inhibition cycles

2. Regressions (Automated)
   [PASS]  139/139 cases passed

3. Behavioral (Unit Tests)
   [PASS]  5/5 unit tests passed
```

---

## Common Workflows

### Modifying Your Configuration

```bash
# Edit your alertmanager.yaml
vi alertmanager.yaml

# See what changed
litmus diff

# Validate the changes
litmus check

# If intentional, update baseline
litmus snapshot --update

# Commit the changes
git add alertmanager.yaml regressions.litmus.mpk regressions.litmus.yml
git commit -m "refactor: consolidate database routes"
```

### Debugging a Test Failure

```bash
# See what changed
litmus diff

# Output:
# [+] ADDED: Route to [new-receiver]
# [-] REMOVED: Route to [old-receiver]
# [!] MODIFIED: Route behavior changed

# Read the regression baseline
litmus inspect regressions.litmus.mpk

# Review git diff for the changes
git diff alertmanager.yaml
```

### Verifying a Bug Fix

```bash
# After fixing a bug in alertmanager.yaml

# Check if regression tests now pass
litmus check

# If routes changed (intentional fix), update baseline
litmus snapshot --update

# Add test case to prevent regression
cat >> tests/bug-fix.yml << 'EOF'
- name: "Bug fix: team X alerts route correctly"
  system_state: {}
  alert:
    labels:
      team: X
      severity: critical
  expect:
    receivers: [team-x-oncall]
EOF

# Verify new test passes
litmus check
```

---

## Command Reference

### `litmus init`

Initialize a Litmus workspace.

```bash
litmus init
```

**Creates:**
- `litmus.yaml` — Configuration file
- `tests/` — Test directory
- `.gitattributes` — Git configuration

**Options:** None

---

### `litmus snapshot`

Capture current configuration as baseline.

```bash
litmus snapshot
litmus snapshot --update   # Force update (overwrite existing baseline)
litmus snapshot -u         # Short form
```

**Output:**
- `regressions.litmus.mpk` — Binary baseline
- `regressions.litmus.yml` — YAML mirror

**Behavior:**
- If no baseline exists: Creates one
- If baseline exists and config matches: Success (no changes)
- If baseline exists and config differs: **Fails with drift warning**
  - Use `--update` to intentionally accept the new baseline
  - Use `diff` to see what changed

**Exit Codes:**
- `0` — Baseline created/updated
- `1` — Drift detected (use `--update` to accept)

---

### `litmus check`

Validate configuration against baseline and tests.

```bash
litmus check
```

**Runs:**
1. Static analysis (sanity checks)
2. Regression validation (baseline comparison)
3. Behavioral tests (custom scenarios)

**Output:** System health report (see example above)

**Exit Codes:**
- `0` — All validation passed
- `2` — Regression or behavioral test failure
- `3` — Static analysis error

**For CI/CD:** Add to your pipeline:
```yaml
# .github/workflows/validate-alerts.yml
- name: Validate Alertmanager
  run: litmus check
```

---

### `litmus diff`

Compare current configuration against baseline.

```bash
litmus diff
```

**Output:**
```
[+] ADDED: Route to [new-team-receiver]
    Labels: {team: new-team, severity: critical}
    Receivers: [new-team-receiver]

[-] REMOVED: Route to [old-team-receiver]
    Labels: {team: old-team, severity: warning}
    Receivers: [old-team-receiver]

[!] MODIFIED: Behavior for Route
    Labels: {team: existing-team, severity: info}
    Expected: [receiver-a]
    Actual: [receiver-b]
```

**Use when:**
- You want to see what changed
- Reviewing a pull request
- Debugging a test failure

---

### `litmus inspect`

Read a binary regression file as human-readable YAML.

```bash
litmus inspect regressions.litmus.mpk

# Output entire file
litmus inspect regressions.litmus.mpk | less

# Compare with git
git show HEAD:regressions.litmus.yml | diff - <(litmus inspect regressions.litmus.mpk)
```

**Use when:**
- Auditing regression changes
- Troubleshooting binary file issues
- Integrating with Git diffs

**Git Integration:**
```bash
# Add to .git/config or local shell setup
git config diff.litmus_inspect.textconv "litmus inspect"

# Now `git diff` automatically handles .mpk files
git diff regressions.litmus.mpk
```

---

### `litmus sync`

Validate and push alertmanager configuration to Grafana Mimir.

```bash
litmus sync
litmus sync --dry-run                              # Validate without pushing
litmus sync --address https://mimir.example.com   # Override config
litmus sync --skip-validate                       # Skip sanity checks
```

**Flags:**
- `--address` — Mimir API address (overrides config)
- `--tenant-id` — Mimir tenant ID (overrides config)
- `--api-key` — Mimir API key (overrides config)
- `--dry-run` — Validate only, do not push
- `--skip-validate` — Skip sanity checks before push

**Configuration:**
Mimir credentials in `.litmus.yaml`:
```yaml
mimir:
  address: "https://mimir.example.com"
  tenant_id: "anonymous"
  api_key: "env(MIMIR_API_KEY)"
```

Or via environment variables:
```bash
LITMUS_MIMIR_ADDRESS=https://mimir.example.com \
LITMUS_MIMIR_TENANT_ID=anonymous \
LITMUS_MIMIR_API_KEY=secret-key \
litmus sync
```

**Output:**
```
✓ Alertmanager config synced to https://mimir.example.com
```

**Use when:**
- Pushing validated configs to Mimir
- CI/CD deployment pipelines
- Testing Mimir integration

**Exit Codes:**
- `0` — Config synced successfully
- `1` — Error (config, validation, or push failure)

---

## Configuration

See [`docs/cli/configuration.md`](configuration.md) for the full `litmus.yaml` schema.

### Minimal Example

```yaml
config:
  directory: config
  file: alertmanager.yml
  templates: templates/

global_labels:
  cluster: production

regression:
  directory: regressions/
  max_samples: 5

tests:
  directory: tests/
```

### Options

- **`config.directory`** (default: `config`) — Directory containing alertmanager config
- **`config.file`** (default: `alertmanager.yml`) — Config filename
- **`config.templates`** (default: `templates/`) — Templates subdirectory
- **`global_labels`** (optional) — Labels added to all synthesized alerts
- **`regression.directory`** (default: `regressions/`) — Baseline directory
- **`regression.max_samples`** (default: 5) — Limit label combinations per route
- **`tests.directory`** (default: `tests/`) — Behavioral test directory

---

## Behavioral Tests

Write test cases in `tests/` directory (YAML format).

```yaml
# tests/critical-alerts.yml
- name: "Critical alerts reach on-call"
  system_state:
    active_alerts: []
    silences: []
  alert:
    labels:
      severity: critical
      team: database
  expect:
    receivers: [database-oncall]

- name: "Silenced alerts don't reach receivers"
  system_state:
    silences:
      - labels: { severity: warning }
  alert:
    labels:
      severity: warning
  expect:
    receivers: []
    silenced: true
```

**Fields:**
- **`name`** — Test description
- **`system_state`** — Active alerts and silences (mocked conditions)
- **`alert`** — Alert labels to route
- **`expect`** — Expected outcome (receivers, silenced, etc.)

See `docs/cli/configuration.md` for full behavioral test schema.

---

## Exit Codes

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Proceed |
| 1 | Error (generic) | Check logs |
| 2 | Test failure | Fix config or update baseline |
| 3 | Static analysis error | Fix configuration issues |

---

## Troubleshooting

### "Drift detected in routing behavior"

Your Alertmanager configuration has changed since the last baseline.

```bash
# See what changed
litmus diff

# Review the changes
git diff alertmanager.yaml

# Accept if intentional
litmus snapshot --update

# Reject if accidental
git checkout alertmanager.yaml
```

### "No behavioral changes detected" (but I made changes)

The regression baseline matches your current config. This is normal if:
- Your changes don't affect routing (e.g., adding a comment)
- Your changes only affect downstream receivers (not routing logic)

Use `litmus diff` to confirm what changed.

### Test failures with "not found"

Your `alertmanager.yaml` or test files may have errors.

```bash
# Validate alertmanager.yaml syntax
amtool config routes

# Check test file syntax
cat tests/your-test.yml
```

---

## Best Practices

1. **Commit baselines to git** — Track regression changes alongside config changes
2. **Review diffs** — Always use `litmus diff` before `litmus snapshot --update`
3. **Write tests for critical paths** — Behavioral tests catch intent errors
4. **Run in CI/CD** — Catch configuration regressions before production
5. **Update baseline intentionally** — Only use `--update` when you understand the change

---

## Integration Examples

### GitHub Actions

```yaml
name: Validate Alertmanager Config

on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go install github.com/nyambati/litmus/cmd/litmus@latest
      - run: litmus check
```

### GitLab CI

```yaml
validate_alerts:
  image: golang:1.26
  script:
    - go install github.com/nyambati/litmus/cmd/litmus@latest
    - litmus check
  only:
    - merge_requests
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
litmus check || exit 1
```

---

## Next Steps

- Read the **Whitepaper** (`docs/whitepaper.md`) for design philosophy
- Review **Backlog** (`docs/backlog.md`) for upcoming features
- Check **Configuration** (`docs/cli/configuration.md`) for full schema
