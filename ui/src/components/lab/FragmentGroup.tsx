import { useState } from "react";
import { ChevronDown, ChevronRight, Play, Layers } from "lucide-react";
import { cn } from "../../utils/ui";
import { GfSpinner } from "../ui/Spinner";
import { LabelChip } from "../ui/Chips";
import { TestCaseCard, type Test, type TestResult } from "./TestCase";

export interface FragmentGroupInfo {
  match: Record<string, string>;
  receiver?: string;
}

export interface FragmentTestGroup {
  name: string;
  namespace?: string;
  group?: FragmentGroupInfo;
  tests: Test[];
}

interface Props {
  group: FragmentTestGroup;
  results: Record<string, TestResult>;
  runningFragment: string | null;
  runningTest: string | null;
  globalRunning: boolean;
  onRunFragment: (name: string) => void;
  onRunTest: (testName: string, testType: string) => void;
}

export const FragmentGroup = ({
  group,
  results,
  runningFragment,
  runningTest,
  globalRunning,
  onRunFragment,
  onRunTest,
}: Props) => {
  const [open, setOpen] = useState(false);
  const isRoot = group.name === "root";
  const isRunning = runningFragment === group.name;

  const passed = group.tests.filter((t) => results[t.name]?.pass === true).length;
  const failed = group.tests.filter((t) => results[t.name]?.pass === false).length;
  const ran = passed + failed;

  return (
    <div className="border border-[#2c3235] rounded-sm overflow-hidden">
      {/* Header */}
      <div
        className={cn(
          "flex items-center gap-3 px-4 py-3 bg-[#22252b] border-b border-[#2c3235]",
          open ? "border-b border-[#2c3235]" : "border-b-0",
        )}
      >
        {/* Collapse toggle */}
        <button
          onClick={() => setOpen((v) => !v)}
          className="text-[#8e9193] hover:text-[#d9d9d9] transition-colors shrink-0"
        >
          {open ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        </button>

        {/* Icon */}
        <div
          className={cn(
            "w-6 h-6 rounded flex items-center justify-center shrink-0",
            isRoot ? "bg-[#f46800]/15" : "bg-[#5794f2]/15",
          )}
        >
          <Layers size={12} className={isRoot ? "text-[#f46800]" : "text-[#5794f2]"} />
        </div>

        {/* Name + namespace */}
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <span className="text-[#d9d9d9] font-semibold text-sm">{group.name}</span>
          {group.namespace && (
            <span className="text-[11px] text-[#8e9193] font-mono bg-[#181b1f] border border-[#2c3235] px-1.5 py-0.5 rounded-[2px] shrink-0">
              ns:{group.namespace}
            </span>
          )}
          {isRoot && (
            <span className="text-[10px] font-semibold text-[#f46800] bg-[#f46800]/10 border border-[#f46800]/20 px-1.5 py-0.5 rounded-sm uppercase tracking-wide shrink-0">
              Root
            </span>
          )}
          {/* Group labels */}
          {group.group && (
            <div className="flex items-center gap-1 flex-wrap">
              {Object.entries(group.group.match).map(([k, v]) => (
                <LabelChip key={k} labelKey={k} value={v} />
              ))}
            </div>
          )}
        </div>

        {/* Result summary */}
        <div className="flex items-center gap-3 shrink-0">
          <span className="text-[11px] text-[#8e9193]">
            {group.tests.length} test{group.tests.length !== 1 ? "s" : ""}
          </span>
          {ran > 0 && (
            <div className="flex items-center gap-2 font-mono text-[11px]">
              {passed > 0 && (
                <span className="text-[#73bf69]">{passed} pass</span>
              )}
              {failed > 0 && (
                <span className="text-[#f2495c]">{failed} fail</span>
              )}
            </div>
          )}

          {/* Run fragment button */}
          <button
            onClick={() => onRunFragment(group.name)}
            disabled={isRunning || globalRunning}
            title={`Run all ${group.name} tests`}
            className={cn(
              "flex items-center gap-1.5 px-3 py-1.5 rounded border text-[11px] font-semibold transition-all",
              "bg-[#181b1f] border-[#34383e] text-[#8e9193]",
              "hover:bg-[#f46800]/10 hover:border-[#f46800]/40 hover:text-[#f46800]",
              "disabled:opacity-30 disabled:cursor-not-allowed",
            )}
          >
            {isRunning ? <GfSpinner size="sm" /> : <Play size={10} fill="currentColor" />}
            {isRunning ? "Running…" : "Run"}
          </button>
        </div>
      </div>

      {/* Tests */}
      {open && (
        <div className="divide-y divide-[#1e2228]">
          {group.tests.map((test) => (
            <div key={test.name} className="px-4 py-3 bg-[#181b1f]">
              <TestCaseCard
                test={test}
                result={results[test.name]}
                isRunning={runningTest === test.name}
                globalRunning={globalRunning || isRunning}
                onRun={() => onRunTest(test.name, test.type)}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
