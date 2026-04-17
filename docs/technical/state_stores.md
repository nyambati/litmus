# State Store Implementation: Technical Specification

To simulate Silences and Inhibitions in a single CLI run without the overhead of a real Alertmanager instance, `litmus` must implement lightweight, in-memory versions of Alertmanager's data providers.

## 1. The Silence Store Store
The `Silencer` in Alertmanager depends on a data store to query active silences.

*   **Official Interface:** `github.com/prometheus/alertmanager/silence.Silences`
*   **Store Implementation:** A simple Go struct holding an array of `Silence` objects derived from the BUT test case's `state.silences` block.
*   **Key Behavior:** 
    *   **Stateless Execution:** The store is re-initialized for every test case to ensure absolute isolation.
    *   **Time Neutrality:** The store will assume all silences provided in the test case are currently within their active time range, eliminating "flaky" tests due to clock drift in CI.

## 2. The Alert Store Store (for Inhibition)
The `Inhibitor` requires access to currently firing alerts to see if a "Target" alert matches a "Source" alert.

*   **Official Interface:** `github.com/prometheus/alertmanager/provider.Alerts`
*   **Store Implementation:** An in-memory map of alerts keyed by their label fingerprint.
*   **Key Behavior:**
    *   **Population:** The store is pre-filled with the alerts from the BUT `state.active_alerts` block.
    *   **Querying:** It satisfies the `GetActive` method, returning the list of source alerts for the `Inhibitor` to evaluate.

## 3. The `PipelineRunner` Orchestrator
The `PipelineRunner` ties the stores together with the official Alertmanager logic to simulate a full alert processing cycle.

```go
type PipelineRunner struct {
    Inhibitor *inhibit.Inhibitor
    Silencer  *silence.Silencer
    Router    *dispatch.Route
}

func (p *PipelineRunner) Execute(testCase BehavioralTest) Outcome {
    // 1. Initial State: The alert enters the pipeline
    alertLabels := testCase.Alert.Labels

    // 2. Silence Phase
    if p.Silencer.Mutes(alertLabels) {
        return Outcome{ Status: "silenced" }
    }

    // 3. Inhibition Phase
    if p.Inhibitor.Mutes(alertLabels) {
        return Outcome{ Status: "inhibited" }
    }

    // 4. Routing Phase
    // If not suppressed, find the destination(s)
    receivers := p.Router.Match(alertLabels)
    return Outcome{ 
        Status: "active",
        Receivers: receivers,
    }
}
```

## 4. Advantages of the Store Approach
*   **Speed:** No disk I/O, no network overhead, and no database locks.
*   **Parity:** By using the official `Mutes()` and `Match()` methods from Alertmanager's own packages, `litmus` ensures its results exactly mirror a real production instance.
*   **Determinism:** Stores remove non-deterministic factors like real-time clocks and concurrent state changes.
