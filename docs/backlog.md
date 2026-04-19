# Backlog: Future Enhancements

This document tracks high-value features and research areas for `litmus` that are planned for future releases.

## Completed

- ✅ **`litmus diff` command** — Compare current configuration against regression baseline (v0.2.0)
- ✅ **JSON Output Format** — `litmus check --format json` emits machine-readable results; `litmus inspect --format json` for baselines

## 1. Negative Matcher Synthesis Paradox
*   **The Problem:** Routes defined *only* with `!=` or `!~` matchers have no positive labels to synthesize, making automated regression testing difficult for these branches.
*   **The Goal:** Implement a **"Non-Matching Seed"** generator. This would require an engine that analyzes a regex and produces a string guaranteed to *fail* that regex (thereby passing the negative route). 

## 2. The Time Dimension (BUT)
*   **The Problem:** `active_time_intervals` can cause tests to pass on weekdays but fail on weekends, introducing "flaky" CI based on the system clock.
*   **The Goal:** Implement full **Time Simulation**. This requires mocking the internal clock used by the Alertmanager pipeline so that unit tests can assert behavior at specific timestamps (e.g., "Saturday at 2 AM") in a deterministic way.

## 3. Template Fragility
*   **The Problem:** A route may be syntactically correct, but the destination receiver's template (e.g., Slack or Email) might require labels (like `owner` or `service_id`) that are not guaranteed to exist on alerts reaching that path.
*   **The Goal:** Implement **Template-Aware Sanity Checks** that parse the `templates/` and verify that the labels matched in a route path satisfy the requirements of the associated receiver.

## 4. Modular Configuration Support
*   **The Problem:** Large organizations struggle with massive, monolithic `alertmanager.yml` files that cause Git merge conflicts and make team-based ownership difficult.
*   **The Goal:** Support for **Configuration Fragments**. Teams define their own routes and receivers in isolated files (e.g., `teams/database.yml`). 
*   **The Mechanism:** Litmus would "assemble" these fragments into a virtual routing tree for validation. This enables **Cross-Impact Detection**, where Litmus can warn a team if their local change has accidentally shadowed or inhibited another team's alerts in the global tree.

## ~~5. JSON Output Format~~ ✅ Done
*   Implemented as `litmus check --format json` and `litmus inspect --format json`. See `docs/cli/user_guide.md` for schema.

## 6. Baseline History & Rollback
*   **The Problem:** If a baseline update is reverted, users must re-run `litmus snapshot --update` to restore it.
*   **The Goal:** Support for baseline versioning, allowing rollback to previous baselines without re-synthesis.
*   **Mechanism:** Keep timestamped baselines in `.litmus/history/` with a pointer to the current baseline.

## 7. Watch Mode
*   **The Problem:** During development, users must manually re-run `litmus check` after each change.
*   **The Goal:** Implement `litmus check --watch` to automatically re-validate on file changes.
*   **Benefits:** Fast feedback loop, similar to `go test -watch` or `jest --watch`.

## 8. Label Cardinality Analysis
*   **The Problem:** Routes with high-cardinality labels (e.g., `service_name` with 1000+ values) can cause synthesis to timeout or consume excessive memory.
*   **The Goal:** Add `litmus analyze` command to report label cardinality and identify problematic routes.
*   **Recommendation:** Suggest route refactoring (use regex instead of direct matches).

## 9. Integration Tests for Popular Notification Channels
*   **The Problem:** Templates can fail at runtime due to missing labels, but this only shows up in production.
*   **The Goal:** Add built-in tests for common receivers (Slack, PagerDuty, Webhook, Email) to validate that routes provide required labels.

## 10. Alertmanager Versions Matrix
*   **The Problem:** Alertmanager behavior changes across versions (e.g., new matcher syntax in 0.24).
*   **The Goal:** Allow specifying target Alertmanager version in `litmus.yaml` and warn about version-specific features.

## 11. Alert Grouping Assertions
*   **The Problem:** `group_by` behavior is untested — routes may silently mis-group alerts, causing notification storms or missed batching.
*   **The Goal:** Extend behavioral unit tests to assert that alerts with specific labels are grouped correctly under a given route.

## 12. Inhibition Simulation in Unit Tests
*   **The Problem:** Inhibition rules can only be verified end-to-end today; no unit test syntax exists to assert "alert A suppresses alert B."
*   **The Goal:** Add an `inhibits:` assertion block to behavioral tests so users can write isolated inhibition scenarios without a full pipeline run.

## 13. Timing Assertions (group_wait / group_interval / repeat_interval)
*   **The Problem:** Notification timing semantics (`group_wait`, `group_interval`, `repeat_interval`) are invisible to current tests.
*   **The Goal:** Allow unit tests to assert the effective timing values on a matched route path, catching silent timing regressions.

## 14. Test Tagging and Selective Runs
*   **The Problem:** Running all tests on every change is slow in large configs; users need a way to run only critical paths.
*   **The Goal:** Support `litmus check --tags=critical` to filter which behavioral/regression tests are executed, using the existing `tags:` field.

## 15. Named Snapshots (Multi-Environment Baselines)
*   **The Problem:** A single baseline can't represent multiple environments (staging, prod, canary).
*   **The Goal:** Support named snapshots via `litmus snapshot --name=prod`, storing baselines as `regressions.prod.mpk` and enabling `litmus check --baseline=prod`.

## 16. Route Tree Visualizer (UI)
*   **The Problem:** The full Alertmanager route hierarchy is hard to reason about as plain YAML.
*   **The Goal:** Add an interactive collapsible tree view in the UI showing the full route hierarchy with matcher details, receiver labels, and which behavioral tests cover each branch.

## 17. Test Coverage Metric
*   **The Problem:** No visibility into which route branches are exercised by existing behavioral tests.
*   **The Goal:** Add a `litmus coverage` command (and UI panel) reporting % of terminal route paths covered by at least one behavioral test, with a list of uncovered branches.

## 18. Config Diff View (UI)
*   **The Problem:** When the alertmanager config changes, it's hard to see what routing impact the change has.
*   **The Goal:** Side-by-side before/after diff in the UI showing which routes were added, removed, or modified, alongside the regression delta.

## 19. Dead Receiver Detection
*   **The Problem:** The current orphan check detects receivers not referenced by any route. The inverse — routes that are unreachable due to being shadowed earlier — is a separate class of bug.
*   **The Goal:** Extend sanity checks to flag routes that can never be reached given the matchers of their ancestors (complements the existing shadowed-route detector).

## 20. Git Hook / CI Integration Mode
*   **The Problem:** Manual `litmus check` runs are easy to forget before merging config changes.
*   **The Goal:** Add `litmus install-hook` to register a pre-commit/pre-push git hook, and emit GitHub Actions annotation format (`::error file=...`) for inline PR feedback.
