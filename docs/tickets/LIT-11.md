# LIT-11: Sanity: Inhibition Cycle & Orphan Receiver Detection
**Summary:** Implement cycle detection for inhibition rules and cleanup for unused receivers.
**Component:** internal/engine/sanity
**Priority:** Medium

### Description:
Build a dependency graph of all `inhibit_rules` and detect circular dependencies. Identify receivers that are defined but never used in any route.

### Tasks:
*   [ ] Implement a cycle detection algorithm (e.g., Tarjan's or DFS).
*   [ ] Build the receiver usage map.
*   [ ] Report ERROR for cycles and WARNING for orphans.

### Acceptance Criteria:
*   [ ] Inhibition cycles (e.g., A->B->A) are correctly identified.
*   [ ] Orphaned receivers are flagged for cleanup.
