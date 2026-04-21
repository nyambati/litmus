import { create } from "zustand";
import { persist } from "zustand/middleware";
import { loadCache, saveCache } from "../utils/persistence";
import { type DiffResult } from "../components/regression/RegressionPage";

interface RegressionStore {
  diff: DiffResult | null;
  lastRunTs: number | null;
  setDiff: (d: DiffResult) => void;
  setLastRunTs: (ts: number) => void;
}

const REGRESSION_KEY = "litmus:regression:diff";

export const useRegressionStore = create<RegressionStore>()(
  persist(
    (set) => ({
      diff: null,
      lastRunTs: null,
      setDiff: (d) => set({ diff: d }),
      setLastRunTs: (ts) => set({ lastRunTs: ts }),
    }),
    {
      name: REGRESSION_KEY,
      partialize: (state) => ({ diff: state.diff }),
      storage: {
        getItem: (name) => {
          const cache = loadCache<DiffResult>(name);
          if (!cache) return null;
          return {
            state: { diff: cache.data },
            version: 0,
          };
        },
        setItem: (name, value) => {
          if (value.state.diff) {
            saveCache(name, value.state.diff);
          }
        },
        removeItem: (name) => localStorage.removeItem(name),
      },
      onRehydrateStorage: () => (state) => {
        // Recover lastRunTs from cache metadata
        const cache = loadCache<DiffResult>(REGRESSION_KEY);
        if (cache && state) {
          state.lastRunTs = cache.ts;
        }
      },
    },
  ),
);
