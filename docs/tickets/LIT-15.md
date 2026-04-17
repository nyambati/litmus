# LIT-15: CLI: Baseline Inspection (litmus inspect)
**Summary:** Implement a utility to make binary regressions human-readable.
**Component:** cmd/litmus
**Priority:** Low

### Description:
Implement the `litmus inspect` command. It should load a binary `.mpk` file and output its content as human-readable YAML or JSON for auditing and Git diffs.

### Tasks:
*   [ ] Implement `cobra` command for `inspect`.
*   [ ] Integrate with the `internal/codec` package.

### Acceptance Criteria:
*   [ ] Binary `.mpk` files can be inspected without the `.yml` mirror.
*   [ ] Output format matches the baseline audit file for consistency.
