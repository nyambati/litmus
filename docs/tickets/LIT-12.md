# LIT-12: CLI: Project Initialization (litmus init)
**Summary:** Implement the bootstrapping command and default configuration.
**Component:** cmd/litmus
**Priority:** Low

### Description:
Implement the `litmus init` command. It should generate the default `litmus.yaml` configuration and the `tests/` directory structure.

### Tasks:
*   [ ] Implement `cobra` command for `init`.
*   [ ] Write default `litmus.yaml` based on `docs/cli/configuration.md`.
*   [ ] Create initial `tests/README.md`.
*   [ ] Generate `.gitattributes` for `.mpk` diffing.

### Acceptance Criteria:
*   [ ] Command creates a consistent, ready-to-use workspace.
*   [ ] Existing files are not overwritten by mistake.
