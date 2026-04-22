import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface QueryHistoryEntry {
  query: string;
  receivers: string[];
  ts: number;
}

interface ExplorerStore {
  labels: string;
  runTrigger: number;
  matchedReceivers: string[];
  queryHistory: QueryHistoryEntry[];
  setLabels: (v: string) => void;
  triggerRun: () => void;
  setMatchedReceivers: (r: string[]) => void;
  saveQuery: (query: string, receivers: string[]) => void;
  loadHistoryEntry: (entry: QueryHistoryEntry) => void;
  deleteHistoryEntry: (query: string) => void;
  clearHistory: () => void;
}

const HISTORY_KEY = "litmus:explorer:history:v2";
const HISTORY_MAX = 20;

export const useExplorerStore = create<ExplorerStore>()(
  persist(
    (set) => ({
      labels: "",
      runTrigger: 0,
      matchedReceivers: [],
      queryHistory: [],
      setLabels: (v) => set({ labels: v }),
      triggerRun: () => set((state) => ({ runTrigger: state.runTrigger + 1 })),
      setMatchedReceivers: (r) => set({ matchedReceivers: r }),
      saveQuery: (query, receivers) =>
        set((state) => {
          const deduped = state.queryHistory.filter((e) => e.query !== query);
          const next = [{ query, receivers, ts: Date.now() }, ...deduped].slice(
            0,
            HISTORY_MAX,
          );
          return { queryHistory: next };
        }),
      loadHistoryEntry: (entry) =>
        set((state) => ({
          labels: entry.query,
          runTrigger: state.runTrigger + 1,
        })),
      deleteHistoryEntry: (query) =>
        set((state) => ({
          queryHistory: state.queryHistory.filter((e) => e.query !== query),
        })),
      clearHistory: () => set({ queryHistory: [] }),
    }),
    {
      name: HISTORY_KEY,
      partialize: (state) => ({ queryHistory: state.queryHistory }),
    },
  ),
);
