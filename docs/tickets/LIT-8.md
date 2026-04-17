# LIT-8: Behavioral: BUT Test Loader & State Injection
**Summary:** Implement the loading of .yml unit tests and state setup.
**Component:** internal/engine/behavioral
**Priority:** Medium

### Description:
Load human-authored Behavioral Unit Tests from the `tests/` directory. For each test, initialize the `SilenceStore` and `AlertStore` with the data provided in the `state` block.

### Tasks:
*   [ ] Implement the YAML test loader for `internal/types.BehavioralTest`.
*   [ ] Implement the mapping logic from `BehavioralTest.State` to `stores`.

### Acceptance Criteria:
*   [ ] All `.yml` files in the tests directory are discovered and parsed.
*   [ ] Each test is executed in a completely isolated environment.
