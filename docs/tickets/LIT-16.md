# LIT-16: CI/CD: GitHub Actions, GoReleaser, & Codacy Setup
**Summary:** Automate testing, linting, security, and releases.
**Component:** .github, .goreleaser.yaml, codacy
**Priority:** Medium

### Description:
Set up the CI/CD pipeline using GitHub Actions to automate our testing and linting. Use **GoReleaser** for automated multi-platform binary releases. Integrate **Codacy** for automated quality, complexity, and security (SAST) gating.

### Tasks:
*   [ ] Create `.github/workflows/ci.yml` for linting (`golangci-lint`) and testing on every PR.
*   [ ] Integrate **Codacy Analysis** to report code quality and security metrics.
*   [ ] Create `.github/workflows/release.yml` to trigger GoReleaser on new tags.
*   [ ] Initialize `.goreleaser.yaml` with build targets for Darwin, Linux, and Windows.
*   [ ] Configure GoReleaser to package the binary and the `.gitattributes` helper.

### Acceptance Criteria:
*   [ ] GitHub Actions correctly run `go test ./...` on PRs.
*   [ ] Codacy successfully analyzes PRs for **Security (SAST)** and **Cyclomatic Complexity**.
*   [ ] `goreleaser release --snapshot` produces a valid multi-platform build locally.
