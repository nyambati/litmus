# Litmus: Deterministic Alertmanager Validation

**Stop hoping your alerts work. Start proving they do.**

Litmus is a testing and validation framework for Prometheus Alertmanager configurations. It brings software engineering rigor to alert routing through automated regression snapshots, behavioral unit tests, and static analysis.

---

## What is Litmus?

Alertmanager configurations are frequently complex, with deeply nested routes, regex matchers, and global inhibition rules. Minor configuration errors often result in **silent failures**:
- Critical alerts misrouted to the wrong team
- Alerts shadowed by overly-broad parent routes
- Alerts accidentally silenced by inhibition rules

Litmus prevents these failures by validating your Alertmanager configuration *before* it reaches production.

---

## Three Layers of Validation

### 1. **Regression Testing** (Automated)
Litmus synthesizes a "ground truth" baseline of your current configuration, capturing where every possible alert will be routed. Any future change that alters this behavior is immediately flagged.

```bash
litmus snapshot        # Create/update regression baseline
litmus check           # Validate configuration matches baseline
litmus diff            # See what changed
```

### 2. **Behavioral Tests** (Human-Authored)
Define scenarios that verify your alert routing logic under real-world conditions: active outages, maintenance windows, multi-team escalations.

```yaml
# tests/critical_alert.yml
- name: "Critical database alert reaches on-call team"
  system_state:
    active_alerts:
      - labels: { severity: critical, team: database }
  alert:
    labels: { severity: critical, team: database }
  expect:
    receivers: [database-oncall]
    not_silenced: true
```

### 3. **Static Analysis** (Sanity Checks)
Before any execution, Litmus audits your configuration for logical errors:
- Shadowed routes (dead code)
- Orphaned receivers (unused)
- Circular inhibition rules

```bash
litmus check           # Includes static analysis
```

---

## Interactive Testing UI

For a more interactive experience, Litmus provides a web UI that brings the validation process into your browser.

```bash
litmus serve
```

The UI provides three main features:

### 1. **Route Explorer**
An interactive tool where you can input a set of alert labels and receive a complete, visual trace of how that alert would be routed through the Alertmanager configuration tree. This allows for easy debugging of routing rules.

### 2. **Behavioral Testing Lab**
A framework that allows you to see and run your YAML-based behavioral tests. Each test defines an input alert and the expected outcome (e.g., which receivers should be notified). These tests can be executed from the UI to validate the configuration's behavior.

### 3. **Snapshot-based Regression Testing**
A system to prevent unintended changes to routing logic. From the UI, you can generate a "snapshot" which captures the expected outcome for all possible routing paths. Subsequently, the system can compare the current configuration's behavior against this snapshot to detect any "drift" or regressions. The UI provides a rich diff viewer to see what has changed.

For a full guide, see the **[Interactive UI Guide](ui_guide.md)**.

---

## Quick Start

### Initialize a Workspace

```bash
litmus init
```

Creates:
- `litmus.yaml` — Configuration file
- `tests/` — Directory for behavioral unit tests
- `.gitattributes` — Git support for binary diffs

### Create a Baseline

```bash
litmus snapshot
```

Generates:
- `regressions.litmus.mpk` — Binary regression baseline (protected from accidental edits)
- `regressions.litmus.yml` — Human-readable mirror for Git diffs

### Validate Your Config

```bash
litmus check
```

Runs three validation engines:
1. Static analysis (sanity checks)
2. Regression validation (baseline comparison)
3. Behavioral tests (intent verification)

Exit codes:
- `0` — All tests pass
- `2` — Regression or behavioral test failure
- `3` — Static analysis error

---

## CLI Commands

| Command | Purpose |
|---------|---------|
| `litmus init` | Initialize workspace |
| `litmus snapshot [--update]` | Create/update regression baseline |
| `litmus check` | Validate configuration (CI/CD) |
| `litmus diff` | Show changes from baseline |
| `litmus inspect` | Read binary regression file |
| `litmus sync` | Push validated config to Grafana Mimir |
| `litmus server` | Launch interactive web UI |

---

## Configuration

Create `.litmus.yaml` in your project root:

```yaml
config:
  directory: config           # Alertmanager config directory
  file: alertmanager.yml      # Config filename
  templates: templates/       # Templates directory

# Global labels added to all synthesized alerts
global_labels:
  cluster: production
  environment: prod

# Regression settings
regression:
  directory: regressions/     # Baseline directory
  max_samples: 5              # Max label combinations per route

# Behavioral test settings
tests:
  directory: tests/           # Test directory

# Mimir sync configuration (optional)
mimir:
  address: "https://mimir.example.com"
  tenant_id: "anonymous"
  api_key: "env(MIMIR_API_KEY)"
```

For full schema, see `docs/cli/configuration.md`.

---

## Design Philosophy

### Parity with Production
Litmus uses the official Prometheus Alertmanager libraries. If a test passes in Litmus, it behaves identically in production—no simulation gaps.

### Shift Left on Alert Testing
Validate alert logic *before* deployment, not in post-mortems. Litmus integrates into your CI/CD pipeline.

### Machine-Managed State
The regression baseline (`regressions.litmus.mpk`) is binary and protected from accidental hand-edits. A YAML mirror provides audit trails.

### Performance First
Litmus validates thousands of routing paths in under 2 seconds using in-memory state stores, making it suitable for fast CI/CD loops.

---

## Typical Workflow

```
1. Initialize workspace:
   litmus init

2. Create initial baseline:
   litmus snapshot

3. Write behavioral tests:
   # Create tests/critical-alerts.yml
   # Create tests/maintenance-window.yml

4. Add to CI/CD:
   litmus check  # Runs all three validation engines

5. Make configuration changes:
   # Edit alertmanager.yaml

6. Validate changes:
   litmus check              # Fails if regression detected
   litmus diff              # See what changed
   litmus snapshot --update # Accept changes if intentional
```

---

## Use Cases

### "Did I break anything?"
```bash
litmus diff     # See exactly what changed
litmus check    # Run full validation suite
```

### "How do I know my alerts work?"
Write behavioral tests in `tests/`. Litmus verifies your routing and silencing logic.

### "Can I safely refactor my config?"
Use `litmus snapshot --update` to accept new baseline, then test with `litmus check`.

### "What's the impact of this change?"
Use `litmus diff` to see which alerts are rerouted, which teams are affected.

---

## Next Steps

- **Read the Whitepaper:** [`docs/whitepaper.md`](whitepaper.md) — Vision and motivation
- **Design Philosophy:** [`docs/architecture.md`](architecture.md) — Architecture and design choices
- **Backlog:** [`docs/backlog.md`](backlog.md) — Future features
- **Configuration:** [`docs/cli/configuration.md`](cli/configuration.md) — Full config schema

---

## Support

- **Bugs:** File an issue on GitHub
- **Questions:** Check existing issues or start a discussion
- **Contributing:** See CONTRIBUTING.md (coming soon)

---

*Made with ❤️ by the Litmus team. Part of the observability ecosystem.*
