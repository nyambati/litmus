# LIT-0: Workflow: Graphify Knowledge Mapping
**Summary:** Transform documentation and code into a queryable knowledge graph for AI assistance.
**Component:** /docs, /internal, /graphify-out
**Priority:** Low

### Description:
Install and configure **Graphify** to index the project's documentation, architecture diagrams, and internal Go code. This provides AI agents with a high-fidelity, graph-based context for implementation and research.

### Tasks:
*   [ ] Install Graphify (`pip install graphifyy`).
*   [ ] Run initial mapping: `graphify .`.
*   [ ] Verify `graphify-out/GRAPH_REPORT.md` accurately captures the "Three-Tier Engine" architecture.
*   [ ] Configure the agent-specific hook (e.g., `CLAUDE.md` or `.cursor/rules/graphify.mdc`) to enable "always-on" graph context.

### Acceptance Criteria:
*   [ ] The project knowledge graph is generated successfully.
*   [ ] `graphify query` returns relevant subgraphs for core architectural questions.
*   [ ] All markdown documentation in `docs/` is indexed and linked to the corresponding `internal/` packages.
