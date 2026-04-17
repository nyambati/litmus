# Backlog: Future Enhancements

This document tracks high-value features and research areas for `litmus` that are currently excluded from the MVP to maintain focus on core routing and suppression logic.

## 1. Negative Matcher Synthesis Paradox
*   **The Problem:** Routes defined *only* with `!=` or `!~` matchers have no positive labels to synthesize, making automated regression testing difficult for these branches.
*   **The Goal:** Implement a **"Non-Matching Seed"** generator. This would require an engine that analyzes a regex and produces a string guaranteed to *fail* that regex (thereby passing the negative route). 

## 2. The Time Dimension (BUT)
*   **The Problem:** `active_time_intervals` can cause tests to pass on weekdays but fail on weekends, introducing "flaky" CI based on the system clock.
*   **The Goal:** Implement full **Time Simulation**. This requires mocking the internal clock used by the Alertmanager pipeline so that unit tests can assert behavior at specific timestamps (e.g., "Saturday at 2 AM") in a deterministic way.

## 3. Template Fragility
*   **The Problem:** A route may be syntactically correct, but the destination receiver's template (e.g., Slack or Email) might require labels (like `owner` or `service_id`) that are not guaranteed to exist on alerts reaching that path.
*   **The Goal:** Implement **Template-Aware Sanity Checks** that parse the `templates/` and verify that the labels matched in a route path satisfy the requirements of the associated receiver.

## 4. Modular Configuration Support
*   **The Problem:** Large organizations struggle with massive, monolithic `alertmanager.yml` files that cause Git merge conflicts and make team-based ownership difficult.
*   **The Goal:** Support for **Configuration Fragments**. Teams define their own routes and receivers in isolated files (e.g., `teams/database.yml`). 
*   **The Mechanism:** Litmus would "assemble" these fragments into a virtual routing tree for validation. This enables **Cross-Impact Detection**, where Litmus can warn a team if their local change has accidentally shadowed or inhibited another team's alerts in the global tree.
