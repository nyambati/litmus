# LIT-7: Snapshot: Outcome Discovery & Lockfile Generation
**Summary:** Use the Pipeline Runner to verify outcomes and save the baseline.
**Component:** internal/engine/snapshot
**Priority:** High

### Description:
Run all synthesized labels through the `PipelineRunner` to discover their "Golden Outcome" (the ordered receiver list). Group identical outcomes and serialize the final test suite to MsgPack/YAML.

### Tasks:
*   [ ] Integrate `pipeline.Execute()` into the snapshot workflow.
*   [ ] Group label maps by their receiver outcomes into `RegressionTest` structs.
*   [ ] Serialize the final suite to `regressions.litmus.mpk` and `regressions.litmus.yml`.

### Acceptance Criteria:
*   [ ] Snapshots are accurate to the Alertmanager routing logic.
*   [ ] The baseline file is correctly deduplicated by outcome.
*   [ ] Both binary and YAML versions are generated consistently.
