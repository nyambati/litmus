import { create } from "zustand";
import { persist } from "zustand/middleware";
import { type DiffResult } from "../components/regression/RegressionPage";

interface RegressionStore {
  diff: DiffResult | null;
  lastRunTs: number | null;
  setDiff: (d: DiffResult) => void;
  setLastRunTs: (ts: number) => void;
}

const REGRESSION_KEY = "litmus:regression:diff:v2";

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
      partialize: (state) => ({ 
        diff: state.diff, 
        lastRunTs: state.lastRunTs 
      }),
    },
  ),
);
