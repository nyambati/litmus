# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0-alpha] - 2026-05-04

### Added
- Interactive Web UI for route exploration and testing.
- Static analysis sanity checks (shadowed routes, inhibition cycles, orphaned receivers).
- Regression testing with MessagePack-encoded baselines.
- Behavioral unit tests with system state simulation.
- Support for Alertmanager configuration fragments.
- Mimir/Prometheus integration for syncing live alerts/silences.
- Multi-stage Dockerfile and GoReleaser Docker support.
- Pre-release build check workflow for GitHub Actions.

### Fixed
- Compilation error in `cmd` package due to undefined version.
- Empty `LICENSE` file.
- Broken documentation links in `README.md` and `docs/INDEX.md`.
- Build-time version injection via GoReleaser.

## [0.1.0] - 2026-04-17
- Initial internal release.
