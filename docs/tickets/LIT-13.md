# LIT-13: CLI: Snapshot & Update Workflow
**Summary:** Implement litmus snapshot with safety checks and the --update flag.
**Component:** cmd/litmus
**Priority:** High

### Description:
Implement the `litmus snapshot` command. It should support "Drift Detection" (failing if the current config differs from the `.mpk` baseline) and the `--update` flag to intentional overwrite it.

### Tasks:
*   [ ] Implement `cobra` command for `snapshot`.
*   [ ] Add logic to compare "Generated State" vs. "Baseline State."
*   [ ] Implement the `--update` flag.
*   [ ] Ensure `regressions.litmus.yml` is always generated alongside `.mpk`.

### Acceptance Criteria:
*   [ ] Snapshot fails if any routing behavior has changed since the last update.
*   [ ] `.mpk` and `.yml` are kept in sync.
*   [ ] User must explicitly use `--update` to accept new behavior.
