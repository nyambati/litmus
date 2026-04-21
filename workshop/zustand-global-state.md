# Global State with Zustand

## Context

`App.tsx` has become a god-component: it owns state that belongs to individual pages (test results, diff stats, explorer history), props-drills callbacks into every page (`onTestsRun`, `onDiffRun`, `onEvaluate`, `onQuerySaved`, `labels`, `setLabels`, `runTrigger`), and manually syncs state to/from localStorage. This tight coupling makes pages hard to reason about and the sidebar hard to extend.

Introducing Zustand stores per feature decouples the sidebar from pages — each page writes its own store, the sidebar reads directly, and App.tsx becomes a pure layout/routing shell.

---

## Install

```
npm install zustand
```

---

## Stores to create (`src/stores/`)

### `useExplorerStore.ts`
Owns: `labels`, `runTrigger`, `matchedReceivers`, `queryHistory`  
Replaces: `explorerLabels`, `explorerRunTrigger`, `matchedReceivers`, `queryHistory` + all history helpers in `App.tsx`  
Persist: `queryHistory` → key `litmus:explorer:history`

### `useLabStore.ts`
Owns: `results: Record<string, TestResult>`, `lastRunTs: number | null`  
Replaces: `testResults` in `App.tsx` + the `loadCache` init block  
Persist: `results` → key `litmus:lab:results`

### `useRegressionStore.ts`
Owns: `diff: DiffResult | null`, `lastRunTs: number | null`  
Replaces: `diffStats` in `App.tsx` + `loadCache` init block  
Persist: `diff` → key `litmus:regression:diff`

---

## Files to change

| File | Change |
|---|---|
| `ui/package.json` | Add `zustand` dependency |
| `ui/src/stores/useExplorerStore.ts` | **Create** — explorer state + history |
| `ui/src/stores/useLabStore.ts` | **Create** — test results |
| `ui/src/stores/useRegressionStore.ts` | **Create** — diff/regression state |
| `ui/src/App.tsx` | Remove all page state + callbacks; sidebar reads from stores directly |
| `ui/src/components/explorer/ExplorerPage.tsx` | Remove all 5 props; read/write `useExplorerStore` directly |
| `ui/src/components/lab/LabPage.tsx` | Remove `onTestsRun` prop; write `useLabStore` directly |
| `ui/src/components/regression/RegressionPage.tsx` | Remove `onDiffRun` prop; write `useRegressionStore` directly |

---

## Store shapes

```ts
// useExplorerStore.ts
interface ExplorerStore {
  labels: string
  runTrigger: number
  matchedReceivers: string[]
  queryHistory: QueryHistoryEntry[]
  setLabels: (v: string) => void
  triggerRun: () => void
  setMatchedReceivers: (r: string[]) => void
  saveQuery: (query: string, receivers: string[]) => void
  loadHistoryEntry: (entry: QueryHistoryEntry) => void
  deleteHistoryEntry: (query: string) => void
  clearHistory: () => void
}

// useLabStore.ts
interface LabStore {
  results: Record<string, { pass: boolean }>
  lastRunTs: number | null
  setResults: (r: Record<string, { pass: boolean }>) => void
  setLastRunTs: (ts: number) => void
}

// useRegressionStore.ts
interface RegressionStore {
  diff: DiffResult | null
  lastRunTs: number | null
  setDiff: (d: DiffResult) => void
  setLastRunTs: (ts: number) => void
}
```

---

## What App.tsx becomes

```tsx
// After: pure layout + routing, no state
function App() {
  return (
    <Router>
      <Routes>
        <Route path="/"           element={<AppLayout stats={<ExplorerStats />}><ExplorerPage /></AppLayout>} />
        <Route path="/lab"        element={<AppLayout stats={<LabStats />}><LabPage /></AppLayout>} />
        <Route path="/regression" element={<AppLayout stats={<RegressionStats />}><RegressionPage /></AppLayout>} />
      </Routes>
    </Router>
  )
}
```

Stats sidebar sections become small components (`ExplorerStats`, `LabStats`, `RegressionStats`) that read directly from stores — co-located with each page or placed in `src/components/layout/`.

---

## Persistence

Use Zustand `persist` middleware to replace manual `loadCache`/`saveCache`. Map to existing localStorage keys to avoid breaking cached data:

| Store | Key | Persisted fields |
|---|---|---|
| `useExplorerStore` | `litmus:explorer:history` | `queryHistory` only |
| `useLabStore` | `litmus:lab:results` | `results` only |
| `useRegressionStore` | `litmus:regression:diff` | `diff` only |

---

## What does NOT change

- `utils/persistence.ts` — keep `loadCache`/`saveCache`/`cn`/`API`/`formatAge`
- API fetch logic inside each page — stays local, store updated after fetch completes
- Component layout, styling, UI logic

---

## Verification

1. `npm install` — zustand installs cleanly
2. `npm run build` — TypeScript compiles, no type errors
3. Dev server: navigate between pages — sidebar stats update without prop callbacks
4. Refresh page — persisted state (history, results, diff) restores from localStorage
5. Run tests in Lab — sidebar passed/failed counts update immediately
6. Run diff in Regression — sidebar baseline/drifted counts update immediately
