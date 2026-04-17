# LIT-6: Snapshot: Regex Expansion & Balanced Coverage
**Summary:** Implement the expansion of alternations with the 5-sample limit.
**Component:** internal/engine/snapshot
**Priority:** High

### Description:
Generate literal label maps from regex matchers. Implement the **Balanced Option Coverage** strategy to ensure 100% branch coverage with at most 5 samples per path.

### Tasks:
*   [ ] Implement literal synthesis from regexes (`api-.*` -> `api-`).
*   [ ] Implement Cartesian product logic for alternations (`(a|b|c)`).
*   [ ] Implement the **Balanced Coverage Governor** to cap at 5 samples per path.
*   [ ] Apply `global_labels` from `litmus.yaml` to each synthesized map.

### Acceptance Criteria:
*   [ ] Regex alternations expand into multiple distinct label maps.
*   [ ] No single route path ever generates more than 5 test samples.
*   [ ] All individual options in a regex are used in at least one map.
