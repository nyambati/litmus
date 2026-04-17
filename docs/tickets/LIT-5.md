# LIT-5: Snapshot: Tree Traversal & Path Discovery
**Summary:** Implement the logic to discover all terminal routes in alertmanager.yml.
**Component:** internal/engine/snapshot
**Priority:** High

### Description:
Traverse the recursive `Route` tree to find every leaf node (a route leading to a receiver). Extract the cumulative list of matchers for each path.

### Tasks:
*   [ ] Implement the recursive walker for `alertmanager/config.Route`.
*   [ ] Capture all matchers (positive and negative) for every branch.
*   [ ] Store each unique path as a `RoutePath` internal struct.

### Acceptance Criteria:
*   [ ] All terminal receivers are identified.
*   [ ] `RoutePath` contains the complete matcher history from root to leaf.
