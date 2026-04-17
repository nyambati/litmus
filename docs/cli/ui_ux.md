# CLI UI/UX: Terminal Reporting Specification

The `litmus` CLI is designed for high-signal output, providing engineers with clear, actionable feedback during both local development and automated CI/CD runs.

## 1. The "System Health Report" (`litmus check`)
Instead of sequential logs, `litmus check` produces a structured, three-pillar report summarizing the state of the configuration.

### Visual Format:
```text
Litmus Check: alertmanager.yml
--------------------------------------------------

1. Sanity (Static Analysis)
   [OK]   No shadowed routes detected
   [WARN] Orphan receiver 'legacy-pager' found (L:142)
   [OK]   No inhibition cycles

2. Regressions (Automated)
   [PASS] 420/422 cases passed
   [FAIL] Regression: Route Path [0] -> [db-team]
          - Labels: {service: mysql, env: prod}
          - Expected: [pd-db, audit-log]
          - Actual:   [pd-db]  <-- Missing 'audit-log'

3. Behavioral (BUT)
   [PASS] 12/12 unit tests passed

--------------------------------------------------
SUMMARY: FAIL (2 Regressions, 1 Sanity Warning)
Time: 1.2s | Exit Code: 2
```

## 2. Route Tree Visualization (`litmus show`)
For interactive debugging, `litmus show` renders the deterministic "Path of Least Resistance" an alert takes through the tree.

### Visual Format:
```text
$ litmus show service=mysql env=prod

Labels: {service: "mysql", env: "prod"}
Routing Path:
└─ route [root] (default)
   └─ route [1] (service=~"mysql|postgres")
      ├─ [MATCH] env="prod" -> receiver: "pd-db"
      └─ [MATCH] continue=true -> receiver: "audit-log"

Outcome: [pd-db, audit-log]
```

## 3. Semantic Coloring & Formatting
*   **Green (`\033[32m`)**: Successes, valid logic, and pass marks (`[OK]`, `[PASS]`).
*   **Yellow (`\033[33m`)**: Non-breaking warnings (`[WARN]`).
*   **Red (`\033[31m`)**: Test failures, critical logic errors, and system errors (`[FAIL]`, `[ERROR]`).
*   **Bold/Cyan**: Highlights for labels, receiver names, and specific line numbers.

## 4. Progressive Feedback
For long-running operations (e.g., `snapshot` on massive trees), `litmus` uses progressive UI elements:
*   **Spinners**: Used during the initial `alertmanager.yml` parsing and tree traversal.
*   **Progress Bars**: Used during the multi-stage `[Synthesis -> Expansion -> Execution]` cycle.
    *   `Snapshotting: [###########---------] 60% (142/300)`

## 5. CI/CD vs. Human Modes
*   **Human Mode (Default)**: Rich Unicode characters (checkmarks, box-drawing), ANSI colors, and interactive progress bars.
*   **CI/CD Mode (`--format text`)**: Plain ASCII output, no animations, and optimized for log readability.
*   **Structured Mode (`--format json|junit`)**: Machine-readable output for integration with external reporting dashboards.
*   **GitHub Integration**: Automated line-level annotations for failures when running in a GitHub Actions environment.

## 6. The "Side-by-Side" Failure Diff
When a regression or unit test fails, `litmus` provides a clear "Expect vs. Reality" diff:
```text
[-] Expected: [slack-ops, audit-trail]
[+] Actual:   [slack-ops, dev-null]
                           ^^^^^^^^
```
