# Litmus: Deterministic Alertmanager Validation

[![Go Version](https://img.shields.io/badge/go-1.26-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)]()

**Stop hoping your alerts work. Start proving they do.**

Litmus is a testing and validation framework for Prometheus Alertmanager configurations. It brings software engineering rigor to alert routing through automated regression snapshots, behavioral unit tests, and static analysis.

---

## Features

✅ **Regression Testing** — Detect unintended routing changes  
✅ **Behavioral Tests** — Verify alert logic under real-world conditions  
✅ **Static Analysis** — Catch shadowed routes and circular inhibitions  
✅ **CI/CD Ready** — Fast validation, exit codes, clear reporting  
✅ **Production Parity** — Uses official Alertmanager libraries  

---

## Quick Start

```bash
# Initialize workspace
litmus init

# Create regression baseline
litmus snapshot capture

# Validate configuration
litmus check

# See what changed
litmus diff
```

---

## Documentation

- **[User Guide](docs/cli/user_guide.md)** — How to use Litmus
- **[Configuration](docs/cli/configuration.md)** — Config schema
- **[Whitepaper](docs/whitepaper.md)** — Vision and motivation
- **[Architecture](docs/architecture.md)** — Design philosophy
- **[Full Docs Index](docs/INDEX.md)** — Complete documentation map

---

<!--## Installation

### Homebrew (macOS/Linux)

```bash
brew install nyambati/tap/litmus
```-->

### Script (curl)

```bash
curl -sSL https://raw.githubusercontent.com/nyambati/litmus/main/scripts/install.sh | sh
```

### Docker

```bash
docker run ghcr.io/nyambati/litmus:latest litmus check
```

---

## Example Workflow

```bash
# 1. Initialize
$ litmus init
Created .litmus.yaml, tests/, .gitattributes

# 2. Create baseline
$ litmus snapshot capture
✓ Generated baseline: regressions/regressions.litmus.yml
✓ Archived: regressions/<timestamp>.mpk

# 3. Write a test
$ cat > tests/critical-alert.yml << 'EOF'
- name: "Critical alerts reach on-call"
  system_state:
    active_alerts: []
  alert:
    labels:
      severity: critical
      team: database
  expect:
    receivers: [database-oncall]
EOF

# 4. Validate
$ litmus check
Litmus Check: alertmanager.yaml
--------------------------------------------------

1. Sanity (Static Analysis)
   [OK]    No shadowed routes detected

2. Regressions (Automated)
   [PASS]  42/42 cases passed

3. Behavioral (Unit Tests)
   [PASS]  1/1 unit tests passed
```

---

## Design Philosophy

### Parity with Production
Uses official Prometheus Alertmanager libraries. If a test passes in Litmus, it behaves identically in production.

### Three Layers of Validation
- **Regression** — Catch accidental routing changes
- **Behavioral** — Verify intent under specific conditions
- **Sanity** — Find dead code and logical errors

### Shift Left
Catch alert routing errors during development, not in production.

---

## Use Cases

### "Did I break anything?"
```bash
litmus diff    # See exactly what changed
litmus check   # Validate the change
```

### "How do I know my alerts work?"
Write behavioral tests. Litmus verifies routing and silencing logic.

### "Can I safely refactor my config?"
Use `litmus snapshot update` to accept changes, then test with `litmus check`.

---

## Project Status

**Version:** 0.2.0-alpha  
**Status:** Active Development

See [Backlog](docs/backlog.md) for planned features.

---

## Contributing

Contributions welcome! Open an issue or pull request on GitHub.

---

## Support

- **Questions?** Check [FAQ in User Guide](docs/cli/user_guide.md#troubleshooting)
- **Found a bug?** [File an issue](https://github.com/nyambati/litmus/issues)
- **Ideas?** See [Backlog](docs/backlog.md) or start a discussion

---

## License

MIT License. See LICENSE file.

---

## Made with ❤️

By the Litmus team. Part of the observability ecosystem.
