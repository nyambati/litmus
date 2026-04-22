import {
  CheckCircle2,
  AlertCircle,
  GitCompare,
  AlertTriangle,
} from "lucide-react";
import { cn, formatAge } from "../../utils/ui";
import { useExplorerStore } from "../../stores/useExplorerStore";
import { useLabStore } from "../../stores/useLabStore";
import { useRegressionStore } from "../../stores/useRegressionStore";
import { StatPanel } from "./StatsSidebar";
import { StatusDot } from "../ui/Status";
import { ReceiverChip } from "../ui/Chips";

export const ExplorerStats = () => {
  const {
    matchedReceivers,
    queryHistory,
    loadHistoryEntry,
    deleteHistoryEntry,
    clearHistory,
  } = useExplorerStore();

  return (
    <div className="space-y-3">
      {/* Matched receivers */}
      <div className="bg-[#1f2128] border border-[#2c3235] rounded-sm p-3">
        <p className="text-[11px] font-semibold text-[#8e9193] uppercase tracking-wider mb-2">
          Matched Receivers
        </p>
        {matchedReceivers.length > 0 ? (
          <div className="flex flex-wrap gap-1.5">
            {matchedReceivers.map((r, idx) => (
              <ReceiverChip key={`${r}-${idx}`} name={r} variant="green" />
            ))}
          </div>
        ) : (
          <p className="text-[#8e9193]/50 text-xs italic">
            Run a query to see matched receivers
          </p>
        )}
      </div>

      {/* Match status */}
      <div className="bg-[#1f2128] border border-[#2c3235] rounded-sm p-3">
        <p className="text-[11px] font-semibold text-[#8e9193] uppercase tracking-wider mb-2">
          Status
        </p>
        <div className="flex items-center gap-2">
          <StatusDot status={matchedReceivers.length > 0 ? "pass" : "idle"} />
          <span
            className={cn(
              "text-sm font-medium",
              matchedReceivers.length > 0 ? "text-[#73bf69]" : "text-[#8e9193]",
            )}
          >
            {matchedReceivers.length > 0 ? "Active Route" : "Idle"}
          </span>
        </div>
      </div>

      {/* Query history */}
      {queryHistory.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-2 pt-1 border-t border-[#2c3235]">
            <p className="text-[11px] font-semibold text-[#8e9193] uppercase tracking-wider">
              Recent
            </p>
            <button
              onClick={clearHistory}
              className="text-[11px] text-[#8e9193]/50 hover:text-[#8e9193] transition-colors"
            >
              Clear
            </button>
          </div>
          <div className="space-y-1.5">
            {queryHistory.map((entry, i) => (
              <div
                key={i}
                className="relative p-2.5 rounded-xs bg-[#1f2128] border border-[#2c3235] hover:border-[#34383e] hover:bg-[#22252b] transition-all group"
              >
                <button
                  onClick={() => loadHistoryEntry(entry)}
                  className="w-full text-left"
                >
                  <p className="font-mono text-[11px] text-[#d9d9d9]/70 truncate pr-5 group-hover:text-[#d9d9d9] transition-colors">
                    {entry.query}
                  </p>
                  <div className="flex items-center justify-between mt-1.5">
                    <div className="flex gap-1 flex-wrap">
                      {entry.receivers.slice(0, 2).map((r, ri) => (
                        <ReceiverChip
                          key={`${r}-${ri}`}
                          name={r}
                          variant="green"
                        />
                      ))}
                      {entry.receivers.length > 2 && (
                        <span className="text-[10px] text-[#8e9193]/50">
                          +{entry.receivers.length - 2}
                        </span>
                      )}
                    </div>
                    <span className="text-[10px] text-[#8e9193]/40 shrink-0 ml-1">
                      {formatAge(entry.ts)}
                    </span>
                  </div>
                </button>
                <button
                  onClick={() => deleteHistoryEntry(entry.query)}
                  title="Remove"
                  className="absolute top-2 right-2 w-4 h-4 flex items-center justify-center text-[#8e9193]/0 group-hover:text-[#8e9193]/50 hover:!text-[#f2495c] transition-colors"
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export const LabStats = () => {
  const results = useLabStore((state) => state.results);

  let passed = 0;
  let failed = 0;
  Object.values(results).forEach((r) => {
    if (r.pass) passed++;
    else failed++;
  });

  const total = passed + failed;

  return (
    <div className="space-y-3">
      <StatPanel
        label="Passed"
        value={passed}
        color="green"
        icon={<CheckCircle2 size={20} />}
      />
      <StatPanel
        label="Failed"
        value={failed}
        color={failed > 0 ? "red" : "default"}
        icon={<AlertCircle size={20} />}
      />
      <StatPanel
        label="Success Rate"
        value={total > 0 ? `${Math.round((passed / total) * 100)}%` : "—"}
        color={total > 0 ? (failed === 0 ? "green" : "default") : "default"}
      />
    </div>
  );
};

export const RegressionStats = () => {
  const diff = useRegressionStore((state) => state.diff);

  return (
    <div className="space-y-3">
      <StatPanel
        label="Baseline Routes"
        value={diff?.total ?? "—"}
        icon={<GitCompare size={20} />}
      />
      <StatPanel
        label="Drifted"
        value={diff?.drifted ?? "—"}
        color={
          diff?.drifted ? (diff.drifted > 0 ? "yellow" : "green") : "default"
        }
        icon={<AlertTriangle size={20} />}
      />
      {diff && (
        <div className="bg-[#1f2128] border border-[#2c3235] rounded-sm p-3">
          <p className="text-[11px] font-semibold text-[#8e9193] uppercase tracking-wider mb-2">
            Status
          </p>
          <div className="flex items-center gap-2">
            <StatusDot status={diff.drifted > 0 ? "drift" : "pass"} />
            <span
              className={cn(
                "text-sm font-medium",
                diff.drifted > 0 ? "text-[#f5a623]" : "text-[#73bf69]",
              )}
            >
              {diff.drifted > 0 ? "Drift Detected" : "Clean"}
            </span>
          </div>
        </div>
      )}
    </div>
  );
};
