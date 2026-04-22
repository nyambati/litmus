import React from "react";
import {
  Circle,
  CheckCircle2,
  AlertCircle,
  Play,
  Tag,
  Layers,
} from "lucide-react";
import { cn } from "../../utils/ui";
import { GfSpinner } from "../ui/Spinner";
import { StatusBadge } from "../ui/Status";
import { LabelChip, ReceiverChip } from "../ui/Chips";

export interface TestResult {
  pass: boolean;
  error?: string;
  expected?: string[];
  actual?: string[];
  labels?: Record<string, unknown>;
}

export interface Test {
  name: string;
  type: "regression" | "unit";
  tags?: string[];
  expect?: { outcome?: string; receivers?: string[] };
  alert?: { labels?: Record<string, unknown> };
  state?: { silences?: unknown[]; active_alerts?: unknown[] };
  labels?: Record<string, string>[];
  expected?: string[];
}

interface TestCaseShellProps {
  name: string;
  testType: "unit" | "regression";
  tags?: string[];
  result?: TestResult;
  isRunning: boolean;
  globalRunning: boolean;
  onRun: () => void;
  children?: React.ReactNode;
}

const TestCaseShell = ({
  name,
  testType,
  tags,
  result,
  isRunning,
  globalRunning,
  onRun,
  children,
}: TestCaseShellProps) => {
  const leftBorderColor = !result
    ? "border-l-[#34383e]"
    : result.pass
      ? "border-l-[#73bf69]"
      : "border-l-[#f2495c]";

  return (
    <div
      className={cn(
        "flex flex-col bg-[#1f2128] border border-[#2c3235] border-l-[3px] rounded-sm overflow-hidden transition-all hover:border-[#34383e]",
        leftBorderColor,
      )}
    >
      {/* Panel header */}
      <div className="flex items-center gap-3 px-4 py-3 bg-[#22252b] border-b border-[#2c3235]">
        {/* Status icon */}
        <div className="shrink-0">
          {isRunning ? (
            <GfSpinner size="sm" />
          ) : !result ? (
            <Circle size={14} className="text-[#34383e]" />
          ) : result.pass ? (
            <CheckCircle2 size={14} className="text-[#73bf69]" />
          ) : (
            <AlertCircle size={14} className="text-[#f2495c]" />
          )}
        </div>

        {/* Name + tags */}
        <div className="flex-1 min-w-0">
          <h4 className="text-[#d9d9d9] text-sm font-medium truncate">
            {name}
          </h4>
          <div className="flex items-center gap-1.5 mt-0.5 flex-wrap">
            <span
              className={cn(
                "text-[10px] font-semibold uppercase tracking-wider px-1.5 py-px rounded-[2px] border",
                testType === "regression"
                  ? "bg-[#b877d9]/10 border-[#b877d9]/25 text-[#b877d9]"
                  : "bg-[#5794f2]/10 border-[#5794f2]/25 text-[#5794f2]",
              )}
            >
              {testType}
            </span>
            {tags
              ?.filter((t) => t !== "regression")
              .map((t, i) => (
                <span
                  key={i + t}
                  className="text-[10px] text-[#8e9193]/60 px-1 py-px rounded-[2px] bg-[#22252b] border border-[#2c3235]"
                >
                  {t}
                </span>
              ))}
          </div>
        </div>

        {/* Result badge + run button */}
        <div className="flex items-center gap-2 shrink-0">
          {result ? <StatusBadge pass={result.pass} /> : <StatusBadge idle />}
          <button
            onClick={onRun}
            disabled={isRunning || globalRunning}
            title="Run this test"
            className={cn(
              "w-7 h-7 flex items-center justify-center rounded border transition-all",
              "bg-[#22252b] border-[#34383e] text-[#8e9193]",
              "hover:bg-[#f46800]/15 hover:border-[#f46800]/40 hover:text-[#f46800]",
              "disabled:opacity-30 disabled:cursor-not-allowed",
            )}
          >
            {isRunning ? <GfSpinner size="sm" /> : <Play size={11} />}
          </button>
        </div>
      </div>

      {/* Panel body */}
      {children}
    </div>
  );
};

interface TestCaseProps {
  test: Test;
  result?: TestResult;
  isRunning: boolean;
  globalRunning: boolean;
  onRun: () => void;
}

const UnitTestCase = ({
  test,
  result,
  isRunning,
  globalRunning,
  onRun,
}: TestCaseProps) => {
  const outcome = test.expect?.outcome;
  const receivers = test.expect?.receivers || [];
  const alertLabels = test.alert?.labels || {};
  const hasState =
    test.state &&
    ((test.state.silences?.length ?? 0) > 0 || (test.state.active_alerts?.length ?? 0) > 0);

  const outcomeColors: Record<string, string> = {
    active: "bg-[#73bf69]/10 border-[#73bf69]/25 text-[#73bf69]",
    silenced: "bg-[#f5a623]/10 border-[#f5a623]/25 text-[#f5a623]",
    inhibited: "bg-[#f46800]/10 border-[#f46800]/25 text-[#f46800]",
  };
  const outcomeColor =
    outcomeColors[outcome as string] ??
    "bg-[#22252b] border-[#34383e] text-[#8e9193]";

  return (
    <TestCaseShell
      name={test.name}
      testType="unit"
      tags={test.tags}
      result={result}
      isRunning={isRunning}
      globalRunning={globalRunning}
      onRun={onRun}
    >
      <div className="px-4 py-3 space-y-2.5">
        {/* Alert labels */}
        <div className="flex items-start gap-2">
          <Tag size={12} className="text-[#8e9193]/50 mt-0.5 shrink-0" />
          <div className="flex flex-wrap gap-1.5">
            {Object.entries(alertLabels).map(([k, v]) => (
              <LabelChip key={k} labelKey={k} value={String(v)} />
            ))}
          </div>
        </div>

        {/* Expect row */}
        <div className="flex items-center gap-2 flex-wrap">
          <span
            className={cn(
              "text-[10px] font-bold uppercase px-2 py-0.5 rounded-[2px] border tracking-wider",
              outcomeColor,
            )}
          >
            {outcome || "active"}
          </span>
          {receivers.map((r: string, i: number) => (
            <ReceiverChip key={`${r}-${i}`} name={r} variant="blue" />
          ))}
        </div>

        {/* State hint */}
        {hasState && (
          <div className="flex items-center gap-3 text-[11px] text-[#8e9193]/50 font-mono">
            {(test.state?.silences?.length ?? 0) > 0 && (
              <span>{test.state?.silences?.length} silence(s)</span>
            )}
            {(test.state?.active_alerts?.length ?? 0) > 0 && (
              <span>{test.state?.active_alerts?.length} active alert(s)</span>
            )}
          </div>
        )}

        {/* Failure */}
        {result && !result.pass && result.error && (
          <div className="p-3 bg-[#111217] rounded-[2px] border border-[#f2495c]/15 font-mono text-xs text-[#f2495c]/80 whitespace-pre-wrap">
            {result.error}
          </div>
        )}
      </div>
    </TestCaseShell>
  );
};

const RegressionTestCase = ({
  test,
  result,
  isRunning,
  globalRunning,
  onRun,
}: TestCaseProps) => {
  const labelSets: Record<string, string>[] = test.labels || [];
  const expected: string[] = test.expected || [];

  return (
    <TestCaseShell
      name={test.name}
      testType="regression"
      tags={test.tags}
      result={result}
      isRunning={isRunning}
      globalRunning={globalRunning}
      onRun={onRun}
    >
      <div className="px-4 py-3 space-y-2.5">
        {/* Label sets */}
        {labelSets.map((labelSet, i) => (
          <div key={i} className="flex items-start gap-2">
            <Layers size={12} className="text-[#8e9193]/50 mt-0.5 shrink-0" />
            <div className="flex flex-wrap gap-1.5">
              {Object.entries(labelSet).map(([k, v]) => (
                <LabelChip key={k} labelKey={k} value={v} />
              ))}
            </div>
          </div>
        ))}

        {/* Expected receivers */}
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-[10px] text-[#8e9193]/50 uppercase font-bold tracking-wider">
            expect
          </span>
          {expected.map((r, i) => (
            <ReceiverChip key={`${r}-${i}`} name={r} variant="purple" />
          ))}
        </div>

        {/* Failure diff */}
        {result && !result.pass && (
          <div className="p-3 bg-[#111217] rounded-[2px] border border-[#f2495c]/15 space-y-2 font-mono text-xs">
            {result.error && (
              <div className="text-[#f2495c]/80">{result.error}</div>
            )}
            {result.expected && (
              <>
                <div className="flex gap-2 items-baseline">
                  <span className="text-[#8e9193]/40 w-16 shrink-0">
                    expected
                  </span>
                  <span className="text-[#d9d9d9]/70">
                    [{result.expected.join(", ")}]
                  </span>
                </div>
                <div className="flex gap-2 items-baseline">
                  <span className="text-[#8e9193]/40 w-16 shrink-0">
                    actual
                  </span>
                  <span className="text-[#f2495c]">
                    [{(result.actual || []).join(", ")}]
                  </span>
                </div>
                {result.labels && (
                  <div className="flex gap-2 pt-1.5 border-t border-[#2c3235]">
                    <span className="text-[#8e9193]/40 w-16 shrink-0">
                      labels
                    </span>
                    <span className="flex flex-wrap gap-1 text-[#8e9193]/60">
                      {Object.entries(result.labels).map(([k, v]) => (
                        <span key={k}>
                          {k}={String(v)}
                        </span>
                      ))}
                    </span>
                  </div>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </TestCaseShell>
  );
};

export const TestCaseCard = (props: TestCaseProps) =>
  props.test.type === "regression" ? (
    <RegressionTestCase {...props} />
  ) : (
    <UnitTestCase {...props} />
  );
