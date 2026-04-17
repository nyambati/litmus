# LIT-3: In-Memory State Stores
**Summary:** Implement Silence and Alert stores satisfying Alertmanager interfaces.
**Component:** internal/stores
**Priority:** High

### Description:
Create in-memory implementations of the `silence.Silences` and `provider.Alerts` interfaces from the Prometheus Alertmanager library. These will be used by the `PipelineRunner` to simulate state.

### Tasks:
*   [ ] Implement `SilenceStore` in `internal/stores/silence_store.go`.
*   [ ] Implement `AlertStore` in `internal/stores/alert_store.go`.
*   [ ] Ensure stores are isolated and can be re-initialized per test case.

### Acceptance Criteria:
*   [ ] `SilenceStore` correctly mutes labels that match its internal silence list.
*   [ ] `AlertStore` correctly returns "firing" alerts for the inhibitor to query.
*   [ ] Both stores satisfy the official Alertmanager interfaces.
