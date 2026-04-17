# LIT-14: CLI: Validation Runner (litmus check)
**Summary:** Implement the primary check command for CI/CD pipelines.
**Component:** cmd/litmus
**Priority:** Critical

### Description:
Implement the `litmus check` command. It should run the Sanity Linter, the Regression Suite, and the Behavioral Unit Tests in sequence.

### Tasks:
*   [ ] Implement `cobra` command for `check`.
*   [ ] Orchestrate all three engines.
*   [ ] Implement descriptive output and exit codes (0 for success, 2 for test fail, 3 for sanity error).
*   [ ] Support different formats (`text`, `json`).

### Acceptance Criteria:
*   [ ] All three verification types are executed.
*   [ ] Failures are clearly reported with specific details.
*   [ ] Command is suitable for use in a GitHub Action or other CI environment.
