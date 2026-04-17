# LIT-4: Shared Pipeline Runner (The Orchestrator)
**Summary:** Implement the unified Execute() logic using Alertmanager libraries.
**Component:** internal/engine/pipeline
**Priority:** Critical

### Description:
Create the `PipelineRunner` that joins the Silencer, Inhibitor, and Router. This is the "Shared Core" used for both regression synthesis and validation.

### Tasks:
*   [ ] Implement `Pipeline` struct and `NewRunner` constructor.
*   [ ] Implement `Execute(labels, silenceStore, alertStore) Outcome`.
*   [ ] Ensure the flow is: Silencing -> Inhibition -> Routing.

### Acceptance Criteria:
*   [ ] Pipeline correctly returns "silenced" if the silence store matches.
*   [ ] Pipeline correctly returns "inhibited" if the alert store matches an inhibit rule.
*   [ ] Pipeline correctly returns the list of receivers from the router if not suppressed.
*   [ ] Logic is deterministic and identical for both test types.
