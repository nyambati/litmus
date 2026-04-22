import { create } from "zustand";
import { persist } from "zustand/middleware";

interface TestResult {
  pass: boolean;
}

interface LabStore {
  results: Record<string, TestResult>;
  lastRunTs: number | null;
  setResults: (r: Record<string, TestResult>) => void;
  setLastRunTs: (ts: number) => void;
}

const RESULTS_KEY = "litmus:lab:results:v2";

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
      partialize: (state) => ({ 
        results: state.results, 
        lastRunTs: state.lastRunTs 
      }),
    },
  ),
);
