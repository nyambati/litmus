import { create } from "zustand";

export interface FragmentGroupInfo {
  match: Record<string, string>;
  receiver?: string;
}

export interface FragmentInfo {
  name: string;
  namespace?: string;
  group?: FragmentGroupInfo;
  routes: number;
  receivers: number;
  tests: number;
}

interface FragmentsStore {
  fragments: FragmentInfo[];
  setFragments: (fragments: FragmentInfo[]) => void;
}

export const useFragmentsStore = create<FragmentsStore>()((set) => ({
  fragments: [],
  setFragments: (fragments) => set({ fragments }),
}));
