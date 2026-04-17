# Behavioral Unit Tests (BUT)

This directory contains human-authored test scenarios for validating
your Alertmanager routing and suppression logic.

## File Format

Tests are defined in YAML format with the following structure:
```yaml
name: "Test scenario description"
tags:
  - "tag1"
  - "tag2"
state:
  active_alerts:
    - labels:
        service: "api"
        severity: "critical"
  silences:
    - labels:
        service: "maintenance"
      comment: "scheduled maintenance"
alert:
  labels:
    service: "api"
    severity: "critical"
expect:
  outcome: "active"
  receivers:
    - "api-team"
```
## Running Tests

```bash
$ litmus check
```
