# LIT-17: Workflow: Git Pre-commit Hooks & Gating
**Summary:** Implement automated developer gating for commits.
**Component:** .git/hooks, Makefile
**Priority:** Medium

### Description:
Set up a mechanism (e.g., a `Makefile` or a dedicated script) to install mandatory `git-prehooks`. These hooks must run `go fmt`, `go vet`, and basic unit tests before allowing a commit to proceed.

### Tasks:
*   [ ] Implement a `pre-commit` script that executes quality checks.
*   [ ] Create a setup task in the `Makefile` to install the hook to `.git/hooks/pre-commit`.
*   [ ] Ensure the hook blocks the commit if any check fails.
*   [ ] Explicitly check for and prevent the use of `Co-authored-by` in commit messages.

### Acceptance Criteria:
*   [ ] Running `git commit` fails if code is not formatted with `go fmt`.
*   [ ] Running `git commit` fails if tests do not pass.
*   [ ] Running `git commit` fails if `Co-authored-by` is found in the commit message.
