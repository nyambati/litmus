import { create } from "zustand";
import { persist } from "zustand/middleware";
import { loadCache, saveCache } from "../utils/persistence";

interface TestResult {
  pass: boolean;
}

interface LabStore {
  results: Record<string, TestResult>;
  lastRunTs: number | null;
  setResults: (r: Record<string, TestResult>) => void;
  setLastRunTs: (ts: number) => void;
}

const RESULTS_KEY = "litmus:lab:results";

export const useLabStore = create<LabStore>()(
  persist(
    (set) => ({
      results: {},
      lastRunTs: null,
      setResults: (r) => set({ results: r }),
      setLastRunTs: (ts) => set({ lastRunTs: ts }),
    }),
    {
      name: RESULTS_KEY,
      partialize: (state) => ({ results: state.results }),
      storage: {
        getItem: (name) => {
          const cache = loadCache<Record<string, TestResult>>(name);
          if (!cache) return null;
          return {
            state: { results: cache.data },
            version: 0,
          };
        },
        setItem: (name, value) => {
          saveCache(name, value.state.results);
        },
        removeItem: (name) => localStorage.removeItem(name),
      },
      onRehydrateStorage: () => (state) => {
        // Recover lastRunTs from cache metadata since it's not in the 'results' data field
        const cache = loadCache<Record<string, TestResult>>(RESULTS_KEY);
        if (cache && state) {
          state.lastRunTs = cache.ts;
        }
      },
    },
  ),
);
