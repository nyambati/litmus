import { useState } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import {
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  GitCompare,
} from "lucide-react";
import { cn, loadCache, saveCache, formatAge } from "./utils/persistence";

interface TestResult {
  pass: boolean;
}

// Layout components
import { AppLayout } from "./components/layout/AppLayout";
import { StatPanel } from "./components/layout/StatsSidebar";

// Page components
import { ExplorerPage } from "./components/explorer/ExplorerPage";
import { LabPage } from "./components/lab/LabPage";
import {
  RegressionPage,
  type DiffResult,
} from "./components/regression/RegressionPage";

// UI components
import { StatusDot } from "./components/ui/Status";
import { ReceiverChip } from "./components/ui/Chips";

// --- Query history + App root ---

interface QueryHistoryEntry {
  query: string;
  receivers: string[];
  ts: number;
}

const HISTORY_KEY = "litmus:explorer:history";
const HISTORY_MAX = 20;

function App() {
  const [matchedReceivers, setMatchedReceivers] = useState<string[]>([]);

  const [queryHistory, setQueryHistory] = useState<QueryHistoryEntry[]>(
    () => loadCache<QueryHistoryEntry[]>(HISTORY_KEY)?.data ?? [],
  );
  const [explorerLabels, setExplorerLabels] = useState("");
  const [explorerRunTrigger, setExplorerRunTrigger] = useState(0);

  const saveQuery = (query: string, receivers: string[]) => {
    setQueryHistory((prev) => {
      const deduped = prev.filter((e) => e.query !== query);
      const next = [{ query, receivers, ts: Date.now() }, ...deduped].slice(
        0,
        HISTORY_MAX,
      );
      saveCache(HISTORY_KEY, next);
      return next;
    });
  };

  const loadHistoryEntry = (entry: QueryHistoryEntry) => {
    setExplorerLabels(entry.query);
    setExplorerRunTrigger((n) => n + 1);
  };

  const deleteHistoryEntry = (query: string) => {
    setQueryHistory((prev) => {
      const next = prev.filter((e) => e.query !== query);
      saveCache(HISTORY_KEY, next);
      return next;
    });
  };

  const [testResults, setTestResults] = useState<{
    passed: number;
    failed: number;
  }>(() => {
    const labCache = loadCache<Record<string, TestResult>>("litmus:lab:results");
    if (!labCache?.data) return { passed: 0, failed: 0 };
    let passed = 0;
    let failed = 0;
    Object.values(labCache.data).forEach((r) => {
      if (r.pass) passed++;
      else failed++;
    });
    return { passed, failed };
  });

  const [diffStats, setDiffStats] = useState<{
    total: number;
    drifted: number;
  } | null>(() => {
    const diffCache = loadCache<DiffResult>("litmus:regression:diff");
    return diffCache?.data
      ? { total: diffCache.data.total, drifted: diffCache.data.drifted }
      : null;
  });

  return (
    <Router>
      <Routes>
        {/* Explorer */}
        <Route
          path="/"
          element={
            <AppLayout
              stats={
                <div className="space-y-3">
                  {/* Matched receivers */}
                  <div className="bg-[#1f2128] border border-[#2c3235] rounded-sm p-3">
                    <p className="text-[11px] font-semibold text-[#8e9193] uppercase tracking-wider mb-2">
                      Matched Receivers
                    </p>
                    {matchedReceivers.length > 0 ? (
                      <div className="flex flex-wrap gap-1.5">
                        {matchedReceivers.map((r, idx) => (
                          <ReceiverChip
                            key={`${r}-${idx}`}
                            name={r}
                            variant="green"
                          />
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
                      <StatusDot
                        status={matchedReceivers.length > 0 ? "pass" : "idle"}
                      />
                      <span
                        className={cn(
                          "text-sm font-medium",
                          matchedReceivers.length > 0
                            ? "text-[#73bf69]"
                            : "text-[#8e9193]",
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
                          onClick={() => {
                            setQueryHistory([]);
                            saveCache(HISTORY_KEY, []);
                          }}
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
              }
            >
              <ExplorerPage
                onEvaluate={setMatchedReceivers}
                onQuerySaved={saveQuery}
                labels={explorerLabels}
                setLabels={setExplorerLabels}
                runTrigger={explorerRunTrigger}
              />
            </AppLayout>
          }
        />

        {/* Lab */}
        <Route
          path="/lab"
          element={
            <AppLayout
              stats={
                <div className="space-y-3">
                  <StatPanel
                    label="Passed"
                    value={testResults.passed}
                    color="green"
                    icon={<CheckCircle2 size={20} />}
                  />
                  <StatPanel
                    label="Failed"
                    value={testResults.failed}
                    color={testResults.failed > 0 ? "red" : "default"}
                    icon={<AlertCircle size={20} />}
                  />
                  <StatPanel
                    label="Success Rate"
                    value={
                      testResults.passed + testResults.failed > 0
                        ? `${Math.round((testResults.passed / (testResults.passed + testResults.failed)) * 100)}%`
                        : "—"
                    }
                    color={
                      testResults.passed + testResults.failed > 0
                        ? testResults.failed === 0
                          ? "green"
                          : "default"
                        : "default"
                    }
                  />
                </div>
              }
            >
              <LabPage
                onTestsRun={(p, f) => setTestResults({ passed: p, failed: f })}
              />
            </AppLayout>
          }
        />

        {/* Regression */}
        <Route
          path="/regression"
          element={
            <AppLayout
              stats={
                <div className="space-y-3">
                  <StatPanel
                    label="Baseline Routes"
                    value={diffStats?.total ?? "—"}
                    icon={<GitCompare size={20} />}
                  />
                  <StatPanel
                    label="Drifted"
                    value={diffStats?.drifted ?? "—"}
                    color={
                      diffStats?.drifted
                        ? diffStats.drifted > 0
                          ? "yellow"
                          : "green"
                        : "default"
                    }
                    icon={<AlertTriangle size={20} />}
                  />
                  {diffStats && (
                    <div className="bg-[#1f2128] border border-[#2c3235] rounded-sm p-3">
                      <p className="text-[11px] font-semibold text-[#8e9193] uppercase tracking-wider mb-2">
                        Status
                      </p>
                      <div className="flex items-center gap-2">
                        <StatusDot
                          status={diffStats.drifted > 0 ? "drift" : "pass"}
                        />
                        <span
                          className={cn(
                            "text-sm font-medium",
                            diffStats.drifted > 0
                              ? "text-[#f5a623]"
                              : "text-[#73bf69]",
                          )}
                        >
                          {diffStats.drifted > 0 ? "Drift Detected" : "Clean"}
                        </span>
                      </div>
                    </div>
                  )}
                </div>
              }
            >
              <RegressionPage
                onDiffRun={(total, drifted) => setDiffStats({ total, drifted })}
              />
            </AppLayout>
          }
        />
      </Routes>
    </Router>
  );
}

export default App;
