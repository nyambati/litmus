# LIT-1: Project Initialization & Go Module Setup
**Summary:** Initialize the litmus Go project with Go 1.26.2.
**Component:** /cmd, /internal, /pkg
**Priority:** High

### Description:
Initialize a new Go module (`go mod init litmus`) using Go version 1.26.2. Create the basic directory tree as defined in `docs/project_organization.md`.

### Tasks:
*   [ ] Initialize Go module with `go 1.26.2`.
*   [ ] Create `/cmd/litmus` for the entry point.
*   [ ] Create `/internal/engine`, `/internal/stores`, `/internal/types`, `/internal/codec`.
*   [ ] Add a basic `main.go` that prints the version.

### Acceptance Criteria:
*   [ ] Project builds with `go build ./cmd/litmus`.
*   [ ] `go.mod` specifies `go 1.26.2`.
*   [ ] All base directories are present.
