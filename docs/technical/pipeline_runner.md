# Pipeline Runner: Technical Specification (Final)

The `PipelineRunner` is the central orchestration engine of `litmus`. It provides a deterministic, state-aware simulation of the alert processing lifecycle by joining internal **State Stores** with the official Alertmanager libraries.

## 1. The Unified Pipeline Architecture
To ensure absolute parity between Regression and Behavioral Unit Tests (BUT), `litmus` uses a shared execution pipeline located in `internal/engine/pipeline/`.

### The "Driver" Pattern
The pipeline is a "dumb" executor. It is "driven" by two different engines:
*   **Regression Engine:** Drives the pipeline with synthesized labels and empty stores.
*   **Behavioral Engine:** Drives the pipeline with human-authored alerts and pre-populated stores.

## 2. Execution Stages (The "Simulator")
The pipeline executes in three discrete stages to mirror the Alertmanager process:

1.  **Silencing Stage**: `github.com/prometheus/alertmanager/silence`.
2.  **Inhibition Stage**: `github.com/prometheus/alertmanager/inhibit`.
3.  **Routing Stage**: `github.com/prometheus/alertmanager/dispatch`.

## 3. The Execution Flow

```go
func (p *Pipeline) Execute(alertLabels labels.Labels) Outcome {
    // 1. Initial State: The alert enters the pipeline
    // 2. Suppression Phase: Silencer
    if p.Silencer.Mutes(alertLabels) {
        return Outcome{ Status: "silenced" }
    }

    // 3. Suppression Phase: Inhibitor
    if p.Inhibitor.Mutes(alertLabels) {
        return Outcome{ Status: "inhibited" }
    }

    // 4. Routing Phase: Dispatcher
    receivers := p.Router.Match(alertLabels)
    return Outcome{ 
        Status: "active",
        Receivers: receivers,
    }
}
```

## 4. Deterministic State Mapping
To ensure that silences and inhibitions are perfectly reliable in a testing context, `litmus` uses deterministic keys for alert identification:
*   **Label Concatenation:** Instead of Alertmanager's `Fingerprint()` (which has a collision risk), `litmus` uses the sorted concatenation of label key-value pairs as the internal key for its **AlertStore**.
*   **Isolation:** The `StateStore` is re-initialized for every individual test case, ensuring that no state leaks between scenarios.
