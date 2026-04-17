# Litmus Configuration: Specification

The `.litmus.yaml` file defines the project-level settings for the `litmus` workspace. It acts as the bridge between your test suite and your Alertmanager configuration.

## 1. Schema Definition

```yaml
# Path to the Alertmanager configuration file to be tested
config_file: "alertmanager.yml"

# Global Test Labels: These labels are automatically added to 
# EVERY synthesized alert during the 'snapshot' process.
# This ensures that alerts have enough context to pass through 
# the routing tree realistically (e.g., matching mandatory severity levels).
global_labels:
  severity: "warning"
  cluster: "production"

# Regression Settings
regression:
  # The maximum number of label combinations to generate per route path
  # before switching to Balanced Option Coverage.
  max_samples: 5
  # Path to the MessagePack baseline
  baseline_path: "regressions.litmus.mpk"

# Behavioral Unit Test (BUT) Settings
tests:
  # Directory where human-authored .yml tests are stored
  directory: "tests/"
```

## 2. Global Labels
By defining `global_labels`, you provide a "baseline alert" that `litmus` uses as a template for all regression synthesis. This is critical because many routing trees require a minimum set of labels (like `env` or `severity`) to be present before any sub-routing can occur.

## 3. Initialization
Running `litmus init` will generate a default version of this file.
