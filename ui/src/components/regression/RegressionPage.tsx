import React, { useState, useEffect, useCallback } from "react";
import { CheckCircle2, AlertTriangle, ShieldCheck, Tag, ArrowRight, ChevronDown, Save, Play, X } from "lucide-react";
import { cn, API, loadCache, saveCache } from "../../utils/persistence";
import { GfSpinner } from "../ui/Spinner";
import { StatusBadge } from "../ui/Status";
import { LastUpdated } from "../ui/LastUpdated";
import { LabelChip, ReceiverChip } from "../ui/Chips";
import { PrimaryButton } from "../ui/Buttons";
import { Header } from "../layout/Header";
import { EmptyState } from "../ui/EmptyState";
import { FilterTabs } from "../ui/Tabs";

interface RouteStep {
  receiver: string;
  match?: string[];
}

interface MatcherMismatch {
  label: string;
  required: string;
  actual: string;
}

interface RouteDrift {
  receiver: string;
  found: boolean;
  mismatches?: MatcherMismatch[];
}

export interface DiffResult {
  total: number;
  passed: number;
  drifted: number;
  results: Array<{
    name: string;
    pass: boolean;
    kind?: string;
    labels?: Record<string, string>;
    expected?: string[];
    route_path?: RouteStep[];
    why_drifted?: RouteDrift[];
    actual?: string[];
  }>;
}

type ActionType = "analyze" | "update";

export const RegressionPage = ({
  onDiffRun,
}: {
  onDiffRun?: (total: number, drifted: number) => void;
}) => {
  const _diffCache = loadCache<DiffResult>("litmus:regression:diff");
  const [diff, setDiff] = useState<DiffResult | null>(_diffCache?.data ?? null);
  const [lastRunTs, setLastRunTs] = useState<number | null>(
    _diffCache?.ts ?? null,
  );
  const [loading, setLoading] = useState(false);
  const [filter, setFilter] = useState<"all" | "passing">("all");
  const [activeAction, setActiveAction] = useState<ActionType>("analyze");
  const [showDropdown, setShowDropdown] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [notification, setNotification] = useState<{
    type: "success" | "error";
    message: string;
  } | null>(null);

  useEffect(() => {
    if (notification) {
      const timer = setTimeout(() => setNotification(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [notification]);

  const runAnalyse = useCallback(async (silent = false) => {
    setLoading(true);
    if (!silent) setNotification(null);
    try {
      const resp = await fetch(`${API}/api/v1/diff`);
      if (!resp.ok) throw new Error(await resp.text());
      const data: DiffResult = await resp.json();
      const now = Date.now();
      setDiff(data);
      setLastRunTs(now);
      saveCache("litmus:regression:diff", data);
      onDiffRun?.(data.total, data.drifted);
      if (!silent) {
        setNotification({
          type: "success",
          message: "Routing analysis completed",
        });
      }
    } catch (err) {
      console.error("Failed to run analysis:", err);
      setNotification({
        type: "error",
        message: `Analysis failed: ${err}`,
      });
    } finally {
      setLoading(false);
    }
  }, [onDiffRun]);

  const handleAction = async () => {
    if (activeAction === "analyze") {
      runAnalyse();
      return;
    }
    setShowConfirm(true);
  };

  const performUpdate = async () => {
    setShowConfirm(false);
    setLoading(true);
    setNotification(null);
    try {
      const resp = await fetch(`${API}/api/v1/snapshot?update=true`, {
        method: "POST",
      });
      if (!resp.ok) throw new Error(await resp.text());

      setNotification({
        type: "success",
        message: "Baseline updated successfully",
      });

      // After update, refresh analysis to show clean state
      await runAnalyse(true);
    } catch (err) {
      console.error(`Failed to update baseline:`, err);
      setNotification({
        type: "error",
        message: String(err),
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!_diffCache) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      runAnalyse(true);
    }
  }, [_diffCache, runAnalyse]);

  const visibleResults =
    diff?.results.filter((r) => {
      if (filter === "all") return r.pass === false || r.kind !== "passing";
      return r.pass === true;
    }) ?? [];

  const actions = [
    {
      id: "analyze" as const,
      label: "Analyse",
      icon: Play,
      desc: "Compare current routing with baseline",
    },
    {
      id: "update" as const,
      label: "Update Baseline",
      icon: Save,
      desc: "Accept changes and update baseline",
    },
  ];

  const currentAction =
    actions.find((a) => a.id === activeAction) || actions[0];

  return (
    <div className="flex-1 flex flex-col h-screen min-h-0 bg-[#181b1f]">
      <Header title="Regression" subtitle="Baseline Comparison" />
      <main className="flex-1 p-6 overflow-y-auto">
        {/* Confirm Modal */}
        {showConfirm && (
          <>
            <div className="fixed inset-0 bg-black/60 backdrop-blur-[2px] z-[100]" onClick={() => setShowConfirm(false)} />
            <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-md bg-[#1f2128] border border-[#34383e] rounded shadow-2xl z-[101] overflow-hidden animate-in zoom-in-95 duration-200">
              <div className="px-6 py-4 border-b border-[#34383e] flex items-center justify-between bg-[#111217]">
                <h3 className="text-[#d9d9d9] font-semibold flex items-center gap-2">
                  <Save size={16} className="text-[#f46800]" />
                  Update Baseline
                </h3>
                <button onClick={() => setShowConfirm(false)} className="text-[#8e9193] hover:text-[#d9d9d9] transition-colors">
                  <X size={18} />
                </button>
              </div>
              <div className="px-6 py-5">
                <p className="text-[#d9d9d9] text-sm leading-relaxed">
                  Are you sure you want to update the regression baseline? This will overwrite the current baseline with the detected routing behavior.
                </p>
                <div className="mt-6 flex items-center justify-end gap-3">
                  <button 
                    onClick={() => setShowConfirm(false)}
                    className="px-4 py-[7px] rounded border border-[#34383e] text-[#d9d9d9] text-sm font-medium hover:bg-[#ffffff08] transition-colors"
                  >
                    Cancel
                  </button>
                  <button 
                    onClick={performUpdate}
                    className="px-4 py-[7px] rounded bg-[#f46800] hover:bg-[#ff7f2a] text-white text-sm font-semibold transition-colors"
                  >
                    Confirm Update
                  </button>
                </div>
              </div>
            </div>
          </>
        )}

        {/* Notifications */}
        {notification && (
          <div
            className={cn(
              "mb-5 px-4 py-3 rounded border flex items-center justify-between animate-in fade-in slide-in-from-top-2",
              notification.type === "success"
                ? "bg-[#73bf69]/10 border-[#73bf69]/20 text-[#73bf69]"
                : "bg-[#f2495c]/10 border-[#f2495c]/20 text-[#f2495c]",
            )}
          >
            <div className="flex items-center gap-2 text-sm font-medium">
              {notification.type === "success" ? (
                <CheckCircle2 size={16} />
              ) : (
                <AlertTriangle size={16} />
              )}
              {notification.message}
            </div>
            <button
              onClick={() => setNotification(null)}
              className="opacity-50 hover:opacity-100 transition-opacity"
            >
              <X size={14} />
            </button>
          </div>
        )}

        {/* Toolbar */}
        <div className="flex items-center justify-between mb-5">
          <div className="flex items-center gap-3">
            <h3 className="text-[#d9d9d9] font-semibold text-sm">
              Routing Drift Analysis
            </h3>
            {diff && (
              <span className="text-[11px] text-[#8e9193]">
                {diff.total} baseline routes
              </span>
            )}
            <LastUpdated ts={lastRunTs} />
          </div>

          <div className="flex items-center">
            <div className="relative flex items-stretch">
              <button
                onClick={handleAction}
                disabled={loading}
                className="flex items-center gap-2 px-4 py-[7px] rounded-l bg-[#f46800] hover:bg-[#ff7f2a] disabled:opacity-40 text-white text-sm font-semibold transition-colors border-r border-black/20"
              >
                {loading ? (
                  <GfSpinner size="sm" />
                ) : (
                  <currentAction.icon size={12} fill={currentAction.id === "analyze" ? "currentColor" : "none"} />
                )}
                {loading ? "Processing…" : currentAction.label}
              </button>
              <button
                onClick={() => setShowDropdown(!showDropdown)}
                disabled={loading}
                className="flex items-center px-2 py-[7px] rounded-r bg-[#f46800] hover:bg-[#ff7f2a] disabled:opacity-40 text-white transition-colors"
              >
                <ChevronDown size={14} />
              </button>

              {showDropdown && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setShowDropdown(false)}
                  />
                  <div className="absolute right-0 mt-1 w-64 bg-[#1f2128] border border-[#34383e] rounded shadow-xl z-20 py-1">
                    {actions.map((action) => (
                      <button
                        key={action.id}
                        onClick={() => {
                          setActiveAction(action.id);
                          setShowDropdown(false);
                        }}
                        className={cn(
                          "w-full text-left px-4 py-2.5 hover:bg-[#ffffff08] transition-colors group",
                          activeAction === action.id && "bg-[#f46800]/10",
                        )}
                      >
                        <div className="flex items-center gap-2">
                          <action.icon
                            size={14}
                            className={
                              activeAction === action.id
                                ? "text-[#f46800]"
                                : "text-[#8e9193] group-hover:text-[#d9d9d9]"
                            }
                          />
                          <span
                            className={cn(
                              "text-sm font-medium",
                              activeAction === action.id
                                ? "text-[#f46800]"
                                : "text-[#d9d9d9]",
                            )}
                          >
                            {action.label}
                          </span>
                        </div>
                        <p className="text-[11px] text-[#8e9193] mt-0.5 ml-6">
                          {action.desc}
                        </p>
                      </button>
                    ))}
                  </div>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Loading */}
        {loading && !diff && (
          <div className="flex items-center justify-center h-48">
            <GfSpinner size="lg" />
          </div>
        )}

        {/* Idle */}
        {!loading && !diff && (
          <EmptyState
            icon={Play}
            title="No baseline comparison run"
            description="Run an analysis to compare current routing against the snapshot baseline"
            action={
              <PrimaryButton
                onClick={handleAction}
                loading={loading}
                icon={<Play size={12} fill="currentColor" />}
              >
                Analyse
              </PrimaryButton>
            }
          />
        )}

        {diff && (
          <>
            {/* All-green state */}
            {diff.drifted === 0 && (
              <div className="flex flex-col items-center gap-3 py-12">
                <div className="w-12 h-12 rounded-full bg-[#73bf69]/10 border border-[#73bf69]/25 flex items-center justify-center">
                  <ShieldCheck size={22} className="text-[#73bf69]" />
                </div>
                <p className="text-[#73bf69] font-medium text-sm">
                  All {diff.total} baseline routes match — no drift detected
                </p>
              </div>
            )}

            {/* Filter tabs */}
            <FilterTabs
              tabs={[
                {
                  label: "All Changes",
                  value: "all" as const,
                  count: diff.drifted,
                },
                {
                  label: "Passing",
                  value: "passing" as const,
                  count: diff.passed,
                },
              ]}
              active={filter}
              onChange={setFilter}
            />

            {/* Result cards */}
            <div className="space-y-2">
              {visibleResults.length === 0 && (
                <div className="py-8 text-center text-[#8e9193] text-sm">
                  {filter === "all" ? "No changes detected" : "No passing results"}
                </div>
              )}
              {visibleResults.map((result, i) => (
                <div
                  key={`${result.name}-${i}`}
                  className={cn(
                    "bg-[#1f2128] border border-[#2c3235] border-l-[3px] rounded-sm overflow-hidden",
                    result.pass 
                      ? "border-l-[#73bf69]" 
                      : result.kind === "added" 
                        ? "border-l-[#5794f2]" 
                        : "border-l-[#f5a623]",
                  )}
                >
                  {/* Row header */}
                  <div className="flex items-center gap-3 px-4 py-3 bg-[#22252b] border-b border-[#2c3235]">
                    <div className="shrink-0">
                      {result.pass ? (
                        <CheckCircle2 size={14} className="text-[#73bf69]" />
                      ) : result.kind === "added" ? (
                        <div className="w-3.5 h-3.5 rounded-full bg-[#5794f2]" />
                      ) : (
                        <AlertTriangle size={14} className="text-[#f5a623]" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <p className="text-[#d9d9d9] text-sm font-medium truncate">
                          {result.name}
                        </p>
                        {result.kind && !result.pass && (
                          <span className={cn(
                            "text-[10px] font-bold uppercase tracking-tighter px-1 rounded-[2px]",
                            result.kind === "added" ? "bg-[#5794f2]/20 text-[#5794f2]" : "bg-[#f5a623]/20 text-[#f5a623]"
                          )}>
                            {result.kind}
                          </span>
                        )}
                      </div>
                    </div>
                    <StatusBadge pass={result.pass} drifted={!result.pass} />
                  </div>

                  {/* Drift detail */}
                  {!result.pass && (
                    <div className="px-4 py-3 space-y-3">
                      {/* Failing labels */}
                      {result.labels &&
                        Object.keys(result.labels).length > 0 && (
                          <div className="flex items-start gap-2">
                            <Tag
                              size={12}
                              className="text-[#8e9193]/50 mt-0.5 shrink-0"
                            />
                            <div className="flex flex-wrap gap-1.5">
                              {Object.entries(result.labels).map(([k, v]) => (
                                <LabelChip key={k} labelKey={k} value={v} />
                              ))}
                            </div>
                          </div>
                        )}

                      {/* Expected → Actual */}
                      {(result.expected || result.actual) && (
                        <div className="flex items-center gap-2 flex-wrap font-mono text-xs">
                          <div className="flex items-center gap-1.5 flex-wrap">
                            <span className="text-[11px] text-[#8e9193]/50 uppercase tracking-wider font-sans font-bold">
                              baseline
                            </span>
                            {result.expected && result.expected.length > 0 ? (
                              result.expected.map((r) => (
                                <ReceiverChip key={r} name={r} variant="blue" />
                              ))
                            ) : (
                                <span className="text-[#8e9193]/40 italic text-[11px]">none</span>
                            )}
                          </div>
                          <ArrowRight
                            size={13}
                            className="text-[#34383e] shrink-0"
                          />
                          <div className="flex items-center gap-1.5 flex-wrap">
                            <span className="text-[11px] text-[#8e9193]/50 uppercase tracking-wider font-sans font-bold">
                              current
                            </span>
                            {result.actual && result.actual.length > 0 ? (
                              result.actual.map((r) => (
                                <ReceiverChip
                                  key={r}
                                  name={r}
                                  variant="amber"
                                />
                              ))
                            ) : (
                              <span className="text-[#8e9193]/40 italic text-[11px]">
                                none
                              </span>
                            )}
                          </div>
                        </div>
                      )}

                      {/* Current route path */}
                      {result.route_path && result.route_path.length > 0 && (
                        <div className="rounded-[2px] bg-[#111217] border border-[#2c3235] p-3 space-y-2">
                          <p className="text-[10px] font-bold uppercase tracking-widest text-[#8e9193]/50">
                            Current Route Path
                          </p>
                          <div className="flex flex-wrap items-center gap-1 font-mono text-xs">
                            {result.route_path.map((step, si) => (
                              <React.Fragment key={si}>
                                {si > 0 && (
                                  <ArrowRight
                                    size={10}
                                    className="text-[#34383e] shrink-0"
                                  />
                                )}
                                <div className="flex flex-col gap-0.5">
                                  <span
                                    className={cn(
                                      "px-2 py-0.5 rounded-[2px] border text-[11px]",
                                      si === result.route_path!.length - 1
                                        ? "bg-[#f5a623]/10 border-[#f5a623]/25 text-[#f5a623]"
                                        : "bg-[#22252b] border-[#34383e] text-[#8e9193]",
                                    )}
                                  >
                                    {step.receiver || "root"}
                                  </span>
                                  {step.match && step.match.length > 0 && (
                                    <span
                                      className="text-[#8e9193]/40 text-[10px] px-1 truncate max-w-[180px]"
                                      title={step.match.join(", ")}
                                    >
                                      {step.match.join(", ")}
                                    </span>
                                  )}
                                </div>
                              </React.Fragment>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Root cause */}
                      {result.why_drifted && result.why_drifted.length > 0 && (
                        <div className="rounded-[2px] border border-[#f2495c]/15 bg-[#f2495c]/5 p-3 space-y-2">
                          <p className="text-[10px] font-bold uppercase tracking-widest text-[#f2495c]/50">
                            Root Cause
                          </p>
                          {result.why_drifted.map((drift, di) => (
                            <div key={di} className="space-y-1.5">
                              {!drift.found ? (
                                <p className="font-mono text-xs text-[#f2495c]/70">
                                  Route{" "}
                                  <span className="text-[#f2495c] font-semibold">
                                    {drift.receiver}
                                  </span>{" "}
                                  was removed from the config
                                </p>
                              ) : (
                                <>
                                  <p className="font-mono text-[11px] text-[#8e9193]/60">
                                    <span className="text-[#f2495c]/80 font-semibold">
                                      {drift.receiver}
                                    </span>{" "}
                                    matcher changed:
                                  </p>
                                  {drift.mismatches?.map((m, mi) => (
                                    <div
                                      key={mi}
                                      className="flex items-center gap-2 font-mono text-xs pl-2"
                                    >
                                      <span
                                        className="text-[#8e9193]/60 w-24 shrink-0 truncate"
                                        title={m.label}
                                      >
                                        {m.label}
                                      </span>
                                      <span className="px-2 py-0.5 rounded-[2px] bg-[#22252b] border border-[#34383e] text-[#8e9193]">
                                        {m.actual || (
                                          <span className="italic text-[#8e9193]/40">
                                            not set
                                          </span>
                                        )}
                                      </span>
                                      <ArrowRight
                                        size={10}
                                        className="text-[#34383e] shrink-0"
                                      />
                                      <span className="px-2 py-0.5 rounded-[2px] bg-[#f5a623]/10 border-[#f5a623]/25 text-[#f5a623] font-semibold">
                                        {m.required}
                                      </span>
                                    </div>
                                  ))}
                                </>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </>
        )}
      </main>
    </div>
  );
};
