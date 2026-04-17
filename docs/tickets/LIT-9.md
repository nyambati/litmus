# LIT-9: Behavioral: Outcome Assertion & Reporting
**Summary:** Execute BUT tests through the pipeline and verify assertions.
**Component:** internal/engine/behavioral
**Priority:** Medium

### Description:
Execute each Behavioral Unit Test through the `PipelineRunner` and compare the actual result against the `expect` block. Report success or failure.

### Tasks:
*   [ ] Integrate `pipeline.Execute()` with BUT state.
*   [ ] Verify the `outcome` (silenced, inhibited, active).
*   [ ] Verify the `receivers` list if the outcome is active.
*   [ ] Generate a test result report for the CLI.

### Acceptance Criteria:
*   [ ] Unit tests correctly identify if an alert was silenced or inhibited.
*   [ ] Destination receivers are verified for unsuppressed alerts.
*   [ ] Clear, descriptive errors are provided for test failures.
