# Interactive UI Guide

Litmus includes a web-based UI for interactive testing and exploration. This guide provides a detailed overview of its features.

---

## Starting the Server

To get started, run the `litmus serve` command from your project root:

```bash
litmus serve
```

By default, the server will be available at `http://localhost:8080`.

---

## Features

The interactive UI is composed of three main pages, accessible from the sidebar:

1.  **Route Explorer**
2.  **Test Lab**
3.  **Regression**

### 1. Route Explorer

The Route Explorer is an interactive tool for debugging and understanding your Alertmanager routing tree.

**How it works:**
1.  **Enter Labels**: In the "Alert Labels" section, provide a set of key-value pairs that represent a Prometheus alert. The UI provides autocompletion for known label names and values.
2.  **Evaluate**: Click the "Run" button.
3.  **Visualize**: Litmus will process the alert through your entire Alertmanager configuration and render a complete, top-to-bottom visualization of the routing path.

The visualization shows:
- Which route nodes the alert matched.
- The final set of receivers that will be notified.
- Whether the alert was silenced by any inhibition rules.

This feature is invaluable for answering the question: "Where would this alert go?"

### 2. Test Lab

The Test Lab provides a user-friendly interface for running your behavioral tests.

**How it works:**
1.  **View Tests**: The page automatically discovers and displays all behavioral tests (`.yml` files) from your `tests/` directory.
2.  **Run All Tests**: Click the "Run All" button to execute the entire test suite.
3.  **Run Single Test**: Click the "Run" button on an individual test card to run only that test.

For each test, the UI will display:
- **[PASS]** or **[FAIL]** status.
- The expected receivers.
- The actual receivers.

This provides a fast feedback loop for writing and debugging behavioral tests.

### 3. Regression Testing

The Regression page allows you to manage snapshot-based regression testing. It helps you catch unintended changes to your routing logic.

**How it works:**

#### Generating a Snapshot
1.  Navigate to the "Regression" page.
2.  Click the **Generate Snapshot** button.
3.  Litmus will create a `regressions.litmus.mpk` file (and a YAML mirror) that captures the "ground truth" of your current routing configuration.

#### Detecting Regressions
1.  After making changes to your `alertmanager.yaml`, navigate to the "Regression" page.
2.  Click the **Run Snapshot** button.
3.  Litmus will compare the behavior of your new configuration against the last known snapshot.

The UI will display a rich diff view highlighting any changes:
- **New Routes**: Routing paths that exist in the new configuration but not in the snapshot.
- **Removed Routes**: Routing paths that existed in the snapshot but are no longer present.
- **Modified Routes**: Paths where the set of receivers has changed.

This makes it easy to spot and review the full impact of any configuration change. To accept the changes, you can regenerate the snapshot.
