# LIT-10: Sanity: Shadowed Route Detection
**Summary:** Implement the static analysis rule for unreachable routes.
**Component:** internal/engine/sanity
**Priority:** Medium

### Description:
Traverse the routing tree to detect "shadowed" routes (those that are unreachable because a parent or sibling route captures their traffic).

### Tasks:
*   [ ] Implement the `IsSubset(child, parent)` matcher logic.
*   [ ] Traverse the tree and identify nodes that can never be triggered.
*   [ ] Report ERRORs for shadowed routes without `continue: true`.

### Acceptance Criteria:
*   [ ] All shadowed routes are identified.
*   [ ] Clear, actionable error messages explain *why* the route is unreachable.
