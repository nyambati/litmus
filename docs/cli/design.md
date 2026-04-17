# CLI Command Design: Specification (Final)

The `litmus` CLI is designed to be CI/CD friendly, providing clear feedback, deterministic outcomes, and a sharp distinction between human-authored intent and machine-generated reality.

## 1. Primary Commands

### A. `litmus init`
Initializes a new `litmus` workspace in the current directory.
*   **Actions:**
    *   Creates a `litmus.yaml` configuration file.
    *   Creates a `tests/` directory for Behavioral Unit Tests (BUT).
    *   Generates a `.gitattributes` file for binary diffing.

### B. `litmus snapshot`
The engine for regression generation. It captures the "Ground Truth" of the current Alertmanager configuration.
*   **Output Files:** 
    *   `regressions.litmus.mpk`: The binary, protected Source of Truth.
    *   `regressions.litmus.yml`: A human-readable mirror for auditing and Git diffs.
*   **Safety Logic:**
    *   **Drift Detection:** If an `.mpk` exists and the *current* config would produce a different outcome, `snapshot` will **fail**. 
    *   **Intentional Update:** Use the `--update, -u` flag to overwrite the existing baseline.

### C. `litmus check`
The main validation command for CI/CD.
*   **Workflow:**
    1.  **Static Analysis:** Runs the "Sanity" linter.
    2.  **Regression Validation:** Loads the `.mpk` baseline and ensures parity.
    3.  **Behavioral Validation:** Executes all human-authored tests in `tests/`.
*   **The Report Aggregator:**
    *   Instead of exiting on the first failure, `litmus check` collects results from **all three engines**.
    *   It prints a unified **System Health Report** as defined in `docs/cli/ui_ux.md`.
*   **Exit Codes:**
    *   `0`: Success.
    *   `2`: Test failure (Regression or BUT deviation).
    *   `3`: Static Analysis error.

### D. `litmus inspect`
A utility to make the binary regression files human-readable.
*   **Usage:** `litmus inspect regressions.litmus.mpk`
*   **Purpose:** Auditing and Git diffing.

### E. `litmus show`
A debugging utility to visualize the routing path for a set of labels.
*   **Visualization:** Renders the **Path of Least Resistance** tree as defined in `docs/cli/ui_ux.md`.

## 2. The "Lockfile" Philosophy
`litmus` treats the `.mpk` file as a **protected lockfile**:
*   **Integrity:** The binary format ensures that the "Golden Reality" isn't accidentally modified by hand.
*   **Synchronization:** The `.yml` file is automatically updated whenever the `.mpk` is updated, providing a clear audit trail in Git history.

## 3. Git Integration
To support version control, `litmus` provides a `.gitattributes` configuration:
```text
*.litmus.mpk diff=litmus_inspect
```
Users can then configure git to use `litmus inspect` for diffs:
`git config diff.litmus_inspect.textconv "litmus inspect"`
