import React, { useState, useEffect } from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  useLocation,
} from "react-router-dom";
import {
  FlaskConical,
  History,
  Search,
  Activity,
  CheckCircle2,
  AlertCircle,
  Play,
  Tag,
  Layers,
  RefreshCw,
  GitCompare,
  ArrowRight,
  ShieldCheck,
} from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

const API = import.meta.env.DEV ? "http://localhost:8080" : "";

// --- Persistence ---

function loadCache<T>(key: string): { data: T; ts: number } | null {
  try {
    const raw = localStorage.getItem(key);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function saveCache<T>(key: string, data: T) {
  try {
    localStorage.setItem(key, JSON.stringify({ data, ts: Date.now() }));
  } catch {}
}

function formatAge(ts: number): string {
  const secs = Math.floor((Date.now() - ts) / 1000);
  if (secs < 60) return "just now";
  if (secs < 3600) return `${Math.floor(secs / 60)}m ago`;
  if (secs < 86400) return `${Math.floor(secs / 3600)}h ago`;
  return `${Math.floor(secs / 86400)}d ago`;
}

// --- Components ---

const Sidebar = () => {
  const location = useLocation();

  const navItems = [
    { name: "Explorer", path: "/", icon: Search },
    { name: "Lab", path: "/lab", icon: FlaskConical },
    { name: "Regression", path: "/regression", icon: History },
  ];

  return (
    <aside className="w-64 border-r border-slate-800 flex flex-col bg-slate-900 text-slate-300">
      <div className="p-6 border-b border-slate-800">
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <Activity className="text-blue-500" />
          Litmus
        </h1>
      </div>

      <nav className="flex-1 p-4 space-y-2">
        {navItems.map((item) => (
          <Link
            key={item.path}
            to={item.path}
            className={cn(
              "flex items-center gap-3 px-3 py-2 rounded-lg transition-colors",
              location.pathname === item.path
                ? "bg-blue-600/10 text-blue-400 border border-blue-600/20"
                : "hover:bg-slate-800 hover:text-white",
            )}
          >
            <item.icon size={20} />
            <span className="font-medium">{item.name}</span>
          </Link>
        ))}
      </nav>

      <div className="p-4 border-t border-slate-800 text-xs text-slate-500">
        v0.1.0-alpha
      </div>
    </aside>
  );
};

const Header = ({ title }: { title: string }) => (
  <header className="h-16 border-b border-slate-800 flex items-center px-8 bg-slate-900/50 backdrop-blur-sm sticky top-0 z-10">
    <h2 className="text-lg font-semibold text-white uppercase tracking-wider">
      {title}
    </h2>
  </header>
);

const StatsSidebar = ({ children }: { children?: React.ReactNode }) => (
  <aside className="w-80 border-l border-slate-800 p-6 flex flex-col bg-slate-900 overflow-y-auto">
    <h3 className="text-sm font-bold text-slate-500 uppercase mb-6 tracking-widest">
      Statistics
    </h3>
    <div className="space-y-4">
      {children || (
        <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
          <p className="text-slate-400 text-sm">Select a page to see stats</p>
        </div>
      )}
    </div>
  </aside>
);

// --- Pages ---

const ExplorerPage = ({
  onEvaluate,
  onQuerySaved,
  labels,
  setLabels,
  runTrigger,
}: {
  onEvaluate: (receivers: string[]) => void;
  onQuerySaved: (query: string, receivers: string[]) => void;
  labels: string;
  setLabels: (v: string) => void;
  runTrigger: number;
}) => {
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const runEvaluation = async (overrideLabels?: string) => {
    setLoading(true);
    try {
      let labelMap: Record<string, string> = {};
      const src = (overrideLabels ?? labels).trim();

      if (src.startsWith("{")) {
        labelMap = JSON.parse(src);
      } else {
        // Support both comma-separated and newline-separated
        // and both "=" and ":" delimiters
        const pairs = src.includes(",") ? src.split(",") : src.split("\n");
        pairs.forEach((pair) => {
          const delimiter = pair.includes("=") ? "=" : ":";
          const [k, v] = pair
            .split(delimiter)
            .map((s) => s.trim().replace(/^["']|["']$/g, ""));
          if (k && v) labelMap[k] = v;
        });
      }

      const resp = await fetch(`${API}/api/v1/evaluate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ labels: labelMap }),
      });
      const data = await resp.json();
      setResult(data);
      onEvaluate(data.receivers || []);
      onQuerySaved(overrideLabels ?? labels, data.receivers || []);
    } catch (err) {
      console.error("Evaluation failed:", err);
      alert("Failed to parse labels. Use k=v,k=v or JSON.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (runTrigger > 0) runEvaluation(labels);
  }, [runTrigger]);

  const renderPath = (
    node: any,
    depth = 0,
    pathId = "root",
  ): React.ReactNode => {
    if (!node) return null;
    const currentNodeId = `${pathId}-${node.receiver || "root"}-${depth}`;

    return (
      <div key={currentNodeId} className="space-y-6 relative">
        <div className="flex items-start gap-4 group relative">
          <div
            className={cn(
              "w-8 h-8 shrink-0 rounded-full flex items-center justify-center text-sm font-bold z-10 transition-transform group-hover:scale-110",
              node.matched
                ? "bg-blue-600 text-white ring-4 ring-blue-600/20"
                : "bg-slate-800 text-slate-500 border border-slate-700",
            )}
          >
            {depth + 1}
          </div>

          <div
            className={cn(
              "flex-1 p-4 rounded-xl bg-slate-900/50 border transition-all shadow-sm",
              node.matched
                ? "border-blue-500/50 group-hover:border-blue-400"
                : "border-slate-800 group-hover:border-slate-600",
            )}
          >
            <div className="flex items-center justify-between mb-2">
              <span
                className={cn(
                  "font-semibold tracking-tight",
                  node.matched ? "text-blue-400" : "text-slate-400",
                )}
              >
                {node.receiver || "root"}
              </span>
              <div className="flex gap-2">
                {node.continue && (
                  <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 text-[10px] font-bold uppercase tracking-tighter border border-blue-500/20">
                    continue
                  </span>
                )}
                <span
                  className={cn(
                    "px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-tighter border",
                    node.matched
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                      : "bg-slate-800 text-slate-500 border-slate-700",
                  )}
                >
                  {node.matched ? "matched" : "no match"}
                </span>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              {node.match && node.match.length > 0 && (
                <div className="text-sm text-slate-500 font-mono space-y-1">
                  {node.match.map((m: string, idx: number) => (
                    <div
                      key={`${currentNodeId}-match-${idx}`}
                      className="flex gap-2"
                    >
                      <span className="text-blue-400/40">match:</span>
                      <span className="truncate">{m}</span>
                    </div>
                  ))}
                </div>
              )}

              {node.matched && (
                <div className="text-[11px] text-slate-500 space-y-1 border-l border-slate-800 pl-4">
                  {(node.group_by || node.groupBy) &&
                    (node.group_by || node.groupBy).length > 0 && (
                      <div className="flex gap-2">
                        <span className="text-slate-600 font-bold uppercase tracking-tighter w-16 text-right">
                          Group:
                        </span>
                        <span className="text-slate-400 truncate">
                          [{(node.group_by || node.groupBy).join(", ")}]
                        </span>
                      </div>
                    )}
                  {(node.group_wait || node.groupWait) && (
                    <div className="flex gap-2">
                      <span className="text-slate-600 font-bold uppercase tracking-tighter w-16 text-right">
                        Wait:
                      </span>
                      <span className="text-slate-400">
                        {node.group_wait || node.groupWait}
                      </span>
                    </div>
                  )}
                  {(node.group_interval || node.groupInterval) && (
                    <div className="flex gap-2">
                      <span className="text-slate-600 font-bold uppercase tracking-tighter w-16 text-right">
                        Interval:
                      </span>
                      <span className="text-slate-400">
                        {node.group_interval || node.groupInterval}
                      </span>
                    </div>
                  )}
                  {(node.repeat_interval || node.repeatInterval) && (
                    <div className="flex gap-2">
                      <span className="text-slate-600 font-bold uppercase tracking-tighter w-16 text-right">
                        Repeat:
                      </span>
                      <span className="text-slate-400">
                        {node.repeat_interval || node.repeatInterval}
                      </span>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>

        {node.children &&
          node.children.map((child: any, idx: number) =>
            renderPath(child, depth + 1, `${currentNodeId}-${idx}`),
          )}
      </div>
    );
  };

  return (
    <div className="flex-1 flex flex-col h-screen min-h-0 bg-slate-950/20">
      <Header title="Route Explorer" />

      {/* Scrollable Center Content */}
      <main className="flex-1 overflow-y-auto p-6 scroll-smooth">
        <div className="mb-6 flex items-center justify-between">
          <h3 className="text-white font-medium flex items-center gap-2">
            <Activity size={18} className="text-blue-500" />
            Evaluation Path
          </h3>
          {result && (
            <span className="text-[10px] text-slate-500 font-mono uppercase tracking-widest bg-slate-900 px-3 py-1 rounded-full border border-slate-800">
              Live Path Generated
            </span>
          )}
        </div>

        {!result && !loading && (
          <div className="h-full flex flex-col items-center justify-center text-slate-500 space-y-4 py-20">
            <Search size={48} className="opacity-20" />
            <p className="text-lg">
              Enter labels below and run query to explore routes
            </p>
          </div>
        )}

        {loading && (
          <div className="h-full flex items-center justify-center py-20">
            <div className="w-10 h-10 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin" />
          </div>
        )}

        {result && (
          <div className="pb-12 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {renderPath(result.path)}
          </div>
        )}
      </main>

      {/* Fixed Bottom Input Section - Snug to sides */}
      <div className="border-t border-slate-800 bg-slate-900/90 backdrop-blur-xl p-6 shadow-2xl z-20">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <Search size={18} className="text-slate-500" />
            <h3 className="text-white font-medium uppercase text-xs tracking-widest">
              Alert Labels
            </h3>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setLabels("")}
              className="px-3 py-1.5 text-xs font-semibold bg-slate-800 text-slate-300 rounded-md hover:bg-slate-700 transition-colors border border-slate-700"
            >
              Clear
            </button>
            <button
              onClick={() => runEvaluation()}
              disabled={loading}
              className="px-6 py-1.5 text-xs font-bold bg-blue-600 text-white rounded-md hover:bg-blue-500 disabled:opacity-50 transition-all shadow-lg shadow-blue-900/40 active:scale-95 flex items-center gap-2"
            >
              {loading ? (
                <>
                  <div className="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin" />
                  Evaluating...
                </>
              ) : (
                "Run Query"
              )}
            </button>
          </div>
        </div>
        <textarea
          value={labels}
          onChange={(e) => setLabels(e.target.value)}
          className="w-full h-20 bg-slate-950 border border-slate-800 rounded-lg p-4 font-mono text-sm text-blue-400 focus:ring-1 focus:ring-blue-500/50 focus:outline-none resize-none shadow-inner transition-all hover:border-slate-700"
          placeholder='severity="critical", team="database"'
        />
      </div>
    </div>
  );
};

// --- Test Case Components ---

interface TestCaseShellProps {
  name: string;
  testType: "unit" | "regression";
  tags?: string[];
  result?: any;
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
}: TestCaseShellProps) => (
  <div
    className={cn(
      "flex flex-col rounded-2xl border transition-all shadow-sm overflow-hidden",
      !result
        ? "bg-slate-800/30 border-slate-800"
        : result.pass
          ? "bg-emerald-500/5 border-emerald-500/20"
          : "bg-rose-500/5 border-rose-500/20",
    )}
  >
    {/* Header row */}
    <div className="flex items-center gap-4 p-5">
      <div
        className={cn(
          "w-9 h-9 rounded-full flex items-center justify-center border shrink-0",
          !result
            ? "bg-slate-800/50 border-slate-700 text-slate-500"
            : result.pass
              ? "bg-emerald-500/10 border-emerald-500/20 text-emerald-500"
              : "bg-rose-500/10 border-rose-500/20 text-rose-500",
        )}
      >
        {isRunning ? (
          <div className="w-4 h-4 border-2 border-current/30 border-t-current rounded-full animate-spin" />
        ) : !result ? (
          <div className="w-2 h-2 rounded-full bg-slate-600" />
        ) : result.pass ? (
          <CheckCircle2 size={18} />
        ) : (
          <AlertCircle size={18} />
        )}
      </div>

      <div className="flex-1 min-w-0">
        <h4 className="text-white font-semibold tracking-tight truncate">
          {name}
        </h4>
        <div className="flex flex-wrap items-center gap-1.5 mt-1">
          <span
            className={cn(
              "px-1.5 py-0.5 rounded text-[10px] uppercase border tracking-tighter font-bold",
              testType === "regression"
                ? "bg-violet-500/10 border-violet-500/20 text-violet-400"
                : "bg-blue-500/10 border-blue-500/20 text-blue-400",
            )}
          >
            {testType}
          </span>
          {tags
            ?.filter((t) => t !== "regression")
            .map((t, i) => (
              <span
                key={i}
                className="px-1.5 py-0.5 bg-slate-900 rounded text-[10px] uppercase border border-slate-800 tracking-tighter text-slate-500"
              >
                #{t}
              </span>
            ))}
        </div>
      </div>

      <div className="flex items-center gap-2 shrink-0">
        {result && (
          <span
            className={cn(
              "text-[10px] font-bold uppercase tracking-widest px-2.5 py-1 rounded-full border",
              result.pass
                ? "text-emerald-500 border-emerald-500/20 bg-emerald-500/5"
                : "text-rose-500 border-rose-500/20 bg-rose-500/5",
            )}
          >
            {result.pass ? "Passed" : "Failed"}
          </span>
        )}
        <button
          onClick={onRun}
          disabled={isRunning || globalRunning}
          title="Run this test"
          className="p-1.5 rounded-lg bg-slate-800 hover:bg-blue-600 border border-slate-700 hover:border-blue-500 text-slate-400 hover:text-white transition-all disabled:opacity-40 disabled:cursor-not-allowed"
        >
          {isRunning ? (
            <div className="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          ) : (
            <Play size={13} />
          )}
        </button>
      </div>
    </div>

    {/* Type-specific body */}
    {children}
  </div>
);

interface UnitTestCaseProps {
  test: any;
  result?: any;
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
}: UnitTestCaseProps) => {
  const outcome = test.expect?.outcome;
  const receivers = test.expect?.receivers || [];
  const alertLabels = test.alert?.labels || {};
  const hasState =
    test.state &&
    (test.state.silences?.length > 0 || test.state.active_alerts?.length > 0);

  const outcomeColor =
    {
      active: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20",
      silenced: "text-amber-400 bg-amber-500/10 border-amber-500/20",
      inhibited: "text-orange-400 bg-orange-500/10 border-orange-500/20",
    }[outcome as string] || "text-slate-400 bg-slate-800 border-slate-700";

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
      <div className="px-5 pb-5 border-t border-slate-800/60 pt-4 space-y-3">
        {/* Alert labels */}
        <div className="flex items-start gap-3">
          <Tag size={13} className="text-slate-600 mt-0.5 shrink-0" />
          <div className="flex flex-wrap gap-1.5">
            {Object.entries(alertLabels).map(([k, v]) => (
              <span
                key={k}
                className="font-mono text-[11px] px-2 py-0.5 bg-slate-900 border border-slate-800 rounded text-slate-400"
              >
                <span className="text-slate-500">{k}=</span>
                {String(v)}
              </span>
            ))}
          </div>
        </div>

        {/* Expect */}
        <div className="flex items-center gap-3 flex-wrap">
          <span
            className={cn(
              "text-[10px] font-bold uppercase px-2 py-0.5 rounded border tracking-wider",
              outcomeColor,
            )}
          >
            {outcome || "active"}
          </span>
          {receivers.map((r: string) => (
            <span
              key={r}
              className="text-[11px] font-mono px-2 py-0.5 rounded bg-blue-500/10 border border-blue-500/20 text-blue-400"
            >
              {r}
            </span>
          ))}
        </div>

        {/* State hint */}
        {hasState && (
          <div className="text-[11px] text-slate-600 flex gap-3">
            {test.state.silences?.length > 0 && (
              <span>{test.state.silences.length} silence(s)</span>
            )}
            {test.state.active_alerts?.length > 0 && (
              <span>{test.state.active_alerts.length} active alert(s)</span>
            )}
          </div>
        )}

        {/* Failure */}
        {result && !result.pass && result.error && (
          <div className="p-3 bg-slate-950/60 rounded-xl border border-rose-500/10 font-mono text-xs text-rose-400/80 whitespace-pre-wrap">
            {result.error}
          </div>
        )}
      </div>
    </TestCaseShell>
  );
};

interface RegressionTestCaseProps {
  test: any;
  result?: any;
  isRunning: boolean;
  globalRunning: boolean;
  onRun: () => void;
}

const RegressionTestCase = ({
  test,
  result,
  isRunning,
  globalRunning,
  onRun,
}: RegressionTestCaseProps) => {
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
      <div className="px-5 pb-5 border-t border-slate-800/60 pt-4 space-y-3">
        {/* Label sets */}
        {labelSets.map((labelSet, i) => (
          <div key={i} className="flex items-start gap-3">
            <Layers size={13} className="text-slate-600 mt-0.5 shrink-0" />
            <div className="flex flex-wrap gap-1.5">
              {Object.entries(labelSet).map(([k, v]) => (
                <span
                  key={k}
                  className="font-mono text-[11px] px-2 py-0.5 bg-slate-900 border border-slate-800 rounded text-slate-400"
                >
                  <span className="text-slate-500">{k}=</span>
                  {v}
                </span>
              ))}
            </div>
          </div>
        ))}

        {/* Expected receivers */}
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-[10px] text-slate-600 uppercase font-bold tracking-wider">
            expect:
          </span>
          {expected.map((r) => (
            <span
              key={r}
              className="text-[11px] font-mono px-2 py-0.5 rounded bg-violet-500/10 border border-violet-500/20 text-violet-400"
            >
              {r}
            </span>
          ))}
        </div>

        {/* Failure diff */}
        {result && !result.pass && (
          <div className="p-3 bg-slate-950/60 rounded-xl border border-rose-500/10 space-y-2 font-mono text-xs">
            {result.error && (
              <div className="text-rose-400/80">{result.error}</div>
            )}
            {result.expected && (
              <>
                <div className="flex gap-2">
                  <span className="text-slate-600 w-16 shrink-0">
                    expected:
                  </span>
                  <span className="text-slate-300">
                    [{result.expected.join(", ")}]
                  </span>
                </div>
                <div className="flex gap-2">
                  <span className="text-slate-600 w-16 shrink-0">actual:</span>
                  <span className="text-rose-400">
                    [{(result.actual || []).join(", ")}]
                  </span>
                </div>
                {result.labels && (
                  <div className="flex gap-2 pt-1 border-t border-slate-800">
                    <span className="text-slate-600 w-16 shrink-0">
                      labels:
                    </span>
                    <span className="text-slate-500 flex flex-wrap gap-1">
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

type FilterType = "all" | "unit" | "regression";

const LabPage = ({
  onTestsRun,
}: {
  onTestsRun: (passed: number, failed: number) => void;
}) => {
  const [tests, setTests] = useState<any[]>([]);
  const _labCache = loadCache<Record<string, any>>("litmus:lab:results");
  const [results, setResults] = useState<Record<string, any>>(_labCache?.data ?? {});
  const [lastRunTs, setLastRunTs] = useState<number | null>(_labCache?.ts ?? null);
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [runningTest, setRunningTest] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterType>("all");

  const fetchTests = async () => {
    setLoading(true);
    try {
      const [unitSettled, regressionSettled] = await Promise.allSettled([
        fetch(`${API}/api/v1/tests`).then((r) => r.json()),
        fetch(`${API}/api/v1/regressions`).then((r) => r.json()),
      ]);

      const unitData: any[] =
        unitSettled.status === "fulfilled" ? unitSettled.value || [] : [];
      const regressionData: any[] =
        regressionSettled.status === "fulfilled"
          ? regressionSettled.value || []
          : [];

      if (unitSettled.status === "rejected")
        console.error("Failed to fetch unit tests:", unitSettled.reason);
      if (regressionSettled.status === "rejected")
        console.error(
          "Failed to fetch regression tests:",
          regressionSettled.reason,
        );

      const unitTests = unitData.map((t) => ({ ...t, type: "unit" }));
      // Go serializes RegressionTest with capitalized field names (no json tags)
      const regressionTests = regressionData.map((t) => ({
        type: "regression",
        name: t.Name,
        tags: t.Tags || [],
        labels: t.Labels || [],
        expected: t.Expected || [],
      }));

      setTests([...unitTests, ...regressionTests]);
    } catch (err) {
      console.error("Failed to fetch tests:", err);
    } finally {
      setLoading(false);
    }
  };

  const applyResults = (data: any[]) => {
    setResults((prev) => {
      const next = { ...prev };
      data.forEach((res: any) => {
        next[res.name] = res;
      });
      saveCache("litmus:lab:results", next);
      return next;
    });
    setLastRunTs(Date.now());
  };

  const runAllTests = async () => {
    setRunning(true);
    try {
      const toRun: Promise<any[]>[] = [];
      if (filter === "all" || filter === "unit") {
        toRun.push(
          fetch(`${API}/api/v1/tests/run`, { method: "POST" }).then((r) =>
            r.json(),
          ),
        );
      }
      if (filter === "all" || filter === "regression") {
        toRun.push(
          fetch(`${API}/api/v1/regressions/run`, { method: "POST" }).then((r) =>
            r.json(),
          ),
        );
      }

      const allResults = (await Promise.all(toRun)).flat();
      const incoming: Record<string, any> = {};
      allResults.forEach((res: any) => {
        incoming[res.name] = res;
      });

      // Merge with current state and derive stats from the full merged map
      setResults((prev) => {
        const merged = { ...prev, ...incoming };
        let passed = 0;
        let failed = 0;
        Object.values(merged).forEach((r: any) => {
          if (r.pass) passed++;
          else failed++;
        });
        onTestsRun(passed, failed);
        saveCache("litmus:lab:results", merged);
        return merged;
      });
      setLastRunTs(Date.now());
    } catch (err) {
      console.error("Failed to run tests:", err);
    } finally {
      setRunning(false);
    }
  };

  const runSingleTest = async (testName: string, testType: string) => {
    setRunningTest(testName);
    try {
      const endpoint =
        testType === "regression"
          ? `${API}/api/v1/regressions/run?name=${encodeURIComponent(testName)}`
          : `${API}/api/v1/tests/run?name=${encodeURIComponent(testName)}`;
      const resp = await fetch(endpoint, { method: "POST" });
      const data = await resp.json();
      applyResults(data);
    } catch (err) {
      console.error("Failed to run test:", err);
    } finally {
      setRunningTest(null);
    }
  };

  const getTestType = (test: any): string => test.type || "unit";

  const filteredTests = tests.filter((test) => {
    if (filter === "all") return true;
    return getTestType(test) === filter;
  });

  useEffect(() => {
    fetchTests();
  }, []);

  const filterTabs: { label: string; value: FilterType }[] = [
    { label: "All", value: "all" },
    { label: "Unit", value: "unit" },
    { label: "Regression", value: "regression" },
  ];

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <Header title="Test Lab" />
      <main className="flex-1 p-8 overflow-y-auto">
        <div className="flex justify-between items-center mb-6">
          <div>
            <h3 className="text-white text-xl font-bold">
              {filter === "unit" ? "Unit Tests" : filter === "regression" ? "Regression Tests" : "All Tests"}
            </h3>
            <p className="text-slate-500 text-sm mt-1">
              {filteredTests.length} of {tests.length} tests
              {filter === "unit" ? " · from tests/" : filter === "regression" ? " · from regressions/" : ""}
              {lastRunTs && <span className="text-slate-600"> · last run {formatAge(lastRunTs)}</span>}
            </p>
          </div>
          <button
            onClick={runAllTests}
            disabled={running || filteredTests.length === 0}
            className="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-900/20 flex items-center gap-2"
          >
            {running ? (
              <div className="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin" />
            ) : (
              <FlaskConical size={18} />
            )}
            {running ? "Running..." : "Run"}
          </button>
        </div>

        {/* Filter Tabs */}
        <div className="flex gap-1 mb-6 p-1 bg-slate-900 rounded-xl border border-slate-800 w-fit">
          {filterTabs.map((tab) => (
            <button
              key={tab.value}
              onClick={() => setFilter(tab.value)}
              className={cn(
                "px-4 py-1.5 text-xs font-bold uppercase tracking-wider rounded-lg transition-all",
                filter === tab.value
                  ? "bg-blue-600 text-white shadow"
                  : "text-slate-500 hover:text-white hover:bg-slate-800",
              )}
            >
              {tab.label}
              {tab.value !== "all" && (
                <span className="ml-1.5 text-[10px] opacity-60">
                  ({tests.filter((t) => getTestType(t) === tab.value).length})
                </span>
              )}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="h-64 flex items-center justify-center">
            <div className="w-10 h-10 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin" />
          </div>
        ) : (
          <div className="space-y-3">
            {filteredTests.map((test, testIdx) => {
              const res = results[test.name];
              const isRunning = runningTest === test.name;
              const testType = getTestType(test) as "unit" | "regression";
              const key = `${test.name}-${testIdx}`;

              if (testType === "regression") {
                return (
                  <RegressionTestCase
                    key={key}
                    test={test}
                    result={res}
                    isRunning={isRunning}
                    globalRunning={running}
                    onRun={() => runSingleTest(test.name, "regression")}
                  />
                );
              }

              return (
                <UnitTestCase
                  key={key}
                  test={test}
                  result={res}
                  isRunning={isRunning}
                  globalRunning={running}
                  onRun={() => runSingleTest(test.name, "unit")}
                />
              );
            })}
          </div>
        )}
      </main>
    </div>
  );
};

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

interface DiffResult {
  total: number;
  passed: number;
  drifted: number;
  results: Array<{
    name: string;
    pass: boolean;
    error?: string;
    labels?: Record<string, string>;
    expected?: string[];
    route_path?: RouteStep[];
    why_drifted?: RouteDrift[];
    actual?: string[];
  }>;
}

const RegressionPage = ({
  onDiffRun,
}: {
  onDiffRun?: (total: number, drifted: number) => void;
}) => {
  const _diffCache = loadCache<DiffResult>("litmus:regression:diff");
  const [diff, setDiff] = useState<DiffResult | null>(_diffCache?.data ?? null);
  const [lastRunTs, setLastRunTs] = useState<number | null>(_diffCache?.ts ?? null);
  const [loading, setLoading] = useState(false);
  const [filter, setFilter] = useState<"drifted" | "passing">("drifted");

  const runDiff = async () => {
    setLoading(true);
    try {
      const resp = await fetch(`${API}/api/v1/diff`);
      const data: DiffResult = await resp.json();
      setDiff(data);
      setLastRunTs(Date.now());
      saveCache("litmus:regression:diff", data);
      onDiffRun?.(data.total, data.drifted);
    } catch (err) {
      console.error("Failed to run diff:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!_diffCache) runDiff();
  }, []);

  const visibleResults =
    diff?.results.filter((r) =>
      filter === "drifted" ? !r.pass : r.pass,
    ) ?? [];

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <Header title="Regression Diff" />
      <main className="flex-1 p-8 overflow-y-auto">
        {/* Toolbar */}
        <div className="flex justify-between items-center mb-6">
          <div>
            <h3 className="text-white text-xl font-bold">
              Baseline Comparison
            </h3>
            <p className="text-slate-500 text-sm mt-1">
              {diff
                ? `${diff.total} baseline tests — ${diff.drifted} drifted`
                : "Run diff to compare current routing against baseline"}
              {lastRunTs && <span className="text-slate-600"> · last run {formatAge(lastRunTs)}</span>}
            </p>
          </div>
          <button
            onClick={runDiff}
            disabled={loading}
            className="px-5 py-2.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-900/20 flex items-center gap-2"
          >
            {loading ? (
              <div className="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin" />
            ) : (
              <RefreshCw size={16} />
            )}
            {loading ? "Running..." : "Run Diff"}
          </button>
        </div>

        {/* Loading state */}
        {loading && !diff && (
          <div className="h-64 flex items-center justify-center">
            <div className="w-10 h-10 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin" />
          </div>
        )}

        {/* Idle state */}
        {!loading && !diff && (
          <div className="h-64 flex flex-col items-center justify-center gap-4 text-slate-600">
            <GitCompare size={48} className="opacity-30" />
            <p>
              Click "Run Diff" to compare current routing against the baseline
            </p>
          </div>
        )}

        {diff && (
          <>
            {/* Summary cards */}
            <div className="grid grid-cols-3 gap-4 mb-6">
              <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                  Total
                </p>
                <p className="text-white text-2xl font-bold">{diff.total}</p>
              </div>
              <div className="p-4 rounded-xl bg-emerald-500/5 border border-emerald-500/20">
                <p className="text-emerald-500 text-xs uppercase font-bold mb-1">
                  Passing
                </p>
                <p className="text-emerald-400 text-2xl font-bold">
                  {diff.passed}
                </p>
              </div>
              <div
                className={cn(
                  "p-4 rounded-xl border",
                  diff.drifted > 0
                    ? "bg-amber-500/5 border-amber-500/20"
                    : "bg-slate-800/50 border-slate-700",
                )}
              >
                <p
                  className={cn(
                    "text-xs uppercase font-bold mb-1",
                    diff.drifted > 0 ? "text-amber-400" : "text-slate-500",
                  )}
                >
                  Drifted
                </p>
                <p
                  className={cn(
                    "text-2xl font-bold",
                    diff.drifted > 0 ? "text-amber-400" : "text-slate-400",
                  )}
                >
                  {diff.drifted}
                </p>
              </div>
            </div>

            {/* Clean state */}
            {diff.drifted === 0 && (
              <div className="flex flex-col items-center justify-center py-16 gap-4 text-slate-500">
                <ShieldCheck size={48} className="text-emerald-500/60" />
                <p className="text-emerald-400 font-semibold">
                  All {diff.total} baseline tests pass — no drift detected
                </p>
              </div>
            )}

            {/* Filter tabs — always visible once diff is loaded */}
            <>
              <div className="flex gap-1 mb-4 p-1 bg-slate-900 rounded-xl border border-slate-800 w-fit">
                  {(
                    [
                      ["drifted", "Drifted"],
                      ["passing", "Passed"],
                    ] as const
                  ).map(([val, label]) => (
                    <button
                      key={val}
                      onClick={() => setFilter(val)}
                      className={cn(
                        "px-4 py-1.5 text-xs font-bold uppercase tracking-wider rounded-lg transition-all",
                        filter === val
                          ? "bg-blue-600 text-white shadow"
                          : "text-slate-500 hover:text-white hover:bg-slate-800",
                      )}
                    >
                      {label}
                    </button>
                  ))}
                </div>

                <div className="space-y-3">
                  {visibleResults.map((result, i) => (
                    <div
                      key={`${result.name}-${i}`}
                      className={cn(
                        "rounded-2xl border overflow-hidden",
                        result.pass
                          ? "bg-slate-800/20 border-slate-800"
                          : "bg-amber-500/5 border-amber-500/20",
                      )}
                    >
                      {/* Row header */}
                      <div className="flex items-center gap-4 px-5 py-4">
                        <div
                          className={cn(
                            "w-8 h-8 rounded-full flex items-center justify-center border shrink-0",
                            result.pass
                              ? "bg-emerald-500/10 border-emerald-500/20 text-emerald-500"
                              : "bg-amber-500/10 border-amber-500/20 text-amber-500",
                          )}
                        >
                          {result.pass ? (
                            <CheckCircle2 size={16} />
                          ) : (
                            <AlertCircle size={16} />
                          )}
                        </div>
                        <p className="flex-1 text-white font-medium text-sm truncate">
                          {result.name}
                        </p>
                        <span
                          className={cn(
                            "text-[10px] font-bold uppercase tracking-widest px-2.5 py-1 rounded-full border shrink-0",
                            result.pass
                              ? "text-emerald-500 border-emerald-500/20 bg-emerald-500/5"
                              : "text-amber-400 border-amber-500/20 bg-amber-500/5",
                          )}
                        >
                          {result.pass ? "Passing" : "Drifted"}
                        </span>
                      </div>

                      {/* Drift detail */}
                      {!result.pass && (
                        <div className="px-5 pb-5 space-y-4 border-t border-slate-800/60 pt-4">
                          {/* Failing labels */}
                          {result.labels &&
                            Object.keys(result.labels).length > 0 && (
                              <div className="flex items-start gap-3">
                                <Tag
                                  size={13}
                                  className="text-slate-600 mt-0.5 shrink-0"
                                />
                                <div className="flex flex-wrap gap-1.5">
                                  {Object.entries(result.labels).map(
                                    ([k, v]) => (
                                      <span
                                        key={k}
                                        className="font-mono text-[11px] px-2 py-0.5 bg-slate-900 border border-slate-800 rounded text-slate-400"
                                      >
                                        <span className="text-slate-600">
                                          {k}=
                                        </span>
                                        {v}
                                      </span>
                                    ),
                                  )}
                                </div>
                              </div>
                            )}

                          {/* Expected → Actual */}
                          {result.expected && (
                            <div className="flex items-center gap-3 flex-wrap font-mono text-xs">
                              <div className="flex items-center gap-1.5 flex-wrap">
                                <span className="text-slate-600 uppercase tracking-wider text-[10px] font-bold">
                                  baseline:
                                </span>
                                {result.expected.map((r) => (
                                  <span
                                    key={r}
                                    className="px-2 py-0.5 rounded bg-slate-800 border border-slate-700 text-slate-300"
                                  >
                                    {r}
                                  </span>
                                ))}
                              </div>
                              <ArrowRight
                                size={14}
                                className="text-slate-600 shrink-0"
                              />
                              <div className="flex items-center gap-1.5 flex-wrap">
                                <span className="text-slate-600 uppercase tracking-wider text-[10px] font-bold">
                                  current:
                                </span>
                                {(result.actual || []).length > 0 ? (
                                  (result.actual || []).map((r) => (
                                    <span
                                      key={r}
                                      className="px-2 py-0.5 rounded bg-amber-500/10 border border-amber-500/20 text-amber-400"
                                    >
                                      {r}
                                    </span>
                                  ))
                                ) : (
                                  <span className="text-slate-600 italic">
                                    none
                                  </span>
                                )}
                              </div>
                            </div>
                          )}

                          {/* Route path trace — shows which routes now match and why */}
                          {result.route_path &&
                            result.route_path.length > 0 && (
                              <div className="rounded-xl bg-slate-950/60 border border-slate-800 p-3 space-y-2">
                                <p className="text-[10px] font-bold uppercase tracking-widest text-slate-600">
                                  Current Route Path
                                </p>
                                <div className="flex flex-wrap items-center gap-1 font-mono text-xs">
                                  {result.route_path.map((step, si) => (
                                    <React.Fragment key={si}>
                                      {si > 0 && (
                                        <ArrowRight
                                          size={11}
                                          className="text-slate-700 shrink-0"
                                        />
                                      )}
                                      <div className="flex flex-col gap-0.5">
                                        <span
                                          className={cn(
                                            "px-2 py-0.5 rounded border",
                                            si === result.route_path!.length - 1
                                              ? "bg-amber-500/10 border-amber-500/20 text-amber-400"
                                              : "bg-slate-800/80 border-slate-700 text-slate-400",
                                          )}
                                        >
                                          {step.receiver || "root"}
                                        </span>
                                        {step.match &&
                                          step.match.length > 0 && (
                                            <span
                                              className="text-slate-600 text-[10px] px-1 truncate max-w-[200px]"
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

                          {/* Root cause — which matchers on the expected route changed */}
                          {result.why_drifted &&
                            result.why_drifted.length > 0 && (
                              <div className="rounded-xl border border-rose-500/15 bg-rose-500/5 p-3 space-y-2">
                                <p className="text-[10px] font-bold uppercase tracking-widest text-rose-400/60">
                                  Root Cause
                                </p>
                                {result.why_drifted.map((drift, di) => (
                                  <div key={di} className="space-y-1.5">
                                    {!drift.found ? (
                                      <p className="font-mono text-xs text-rose-400/70">
                                        Route{" "}
                                        <span className="text-rose-300 font-semibold">
                                          {drift.receiver}
                                        </span>{" "}
                                        was removed from the config
                                      </p>
                                    ) : (
                                      <>
                                        <p className="font-mono text-[11px] text-slate-500">
                                          <span className="text-rose-300/80 font-semibold">
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
                                              className="text-slate-500 w-24 shrink-0 truncate"
                                              title={m.label}
                                            >
                                              {m.label}
                                            </span>
                                            <span className="px-2 py-0.5 rounded bg-slate-800 border border-slate-700 text-slate-400">
                                              {m.actual || (
                                                <span className="italic text-slate-600">
                                                  not set
                                                </span>
                                              )}
                                            </span>
                                            <ArrowRight
                                              size={11}
                                              className="text-slate-600 shrink-0"
                                            />
                                            <span className="px-2 py-0.5 rounded bg-amber-500/10 border border-amber-500/25 text-amber-300 font-semibold">
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

                          {result.error && !result.expected && (
                            <p className="font-mono text-xs text-rose-400/80">
                              {result.error}
                            </p>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
            </>
          </>
        )}
      </main>
    </div>
  );
};

// --- Layout & Root ---

const AppLayout = ({
  children,
  stats,
}: {
  children: React.ReactNode;
  stats?: React.ReactNode;
}) => (
  <div className="flex h-screen bg-slate-950 text-slate-300 w-full overflow-hidden">
    <Sidebar />
    <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
      {children}
    </main>
    <StatsSidebar>{stats}</StatsSidebar>
  </div>
);

interface QueryHistoryEntry {
  query: string;
  receivers: string[];
  ts: number;
}

const HISTORY_KEY = "litmus:explorer:history";
const HISTORY_MAX = 20;

function App() {
  const [matchedReceivers, setMatchedReceivers] = useState<string[]>([]);

  // Explorer query history
  const [queryHistory, setQueryHistory] = useState<QueryHistoryEntry[]>(
    () => loadCache<QueryHistoryEntry[]>(HISTORY_KEY)?.data ?? [],
  );
  const [explorerLabels, setExplorerLabels] = useState(
    'severity="critical", team="database"',
  );
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
  const _labCache = loadCache<Record<string, any>>("litmus:lab:results");
  const [testResults, setTestResults] = useState<{ passed: number; failed: number }>(() => {
    if (!_labCache?.data) return { passed: 0, failed: 0 };
    let passed = 0; let failed = 0;
    Object.values(_labCache.data).forEach((r: any) => { if (r.pass) passed++; else failed++; });
    return { passed, failed };
  });

  const _diffCache = loadCache<DiffResult>("litmus:regression:diff");
  const [diffStats, setDiffStats] = useState<{ total: number; drifted: number } | null>(
    _diffCache?.data ? { total: _diffCache.data.total, drifted: _diffCache.data.drifted } : null,
  );

  return (
    <Router>
      <Routes>
        <Route
          path="/"
          element={
            <AppLayout
              stats={
                <div className="space-y-4">
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                    <p className="text-slate-500 text-xs uppercase font-bold mb-2">
                      Matched Receivers
                    </p>
                    {matchedReceivers.length > 0 ? (
                      <div className="flex flex-wrap gap-2">
                        {matchedReceivers.map((r, idx) => (
                          <span
                            key={`${r}-${idx}`}
                            className="px-3 py-1 rounded bg-emerald-500/10 text-emerald-400 text-sm font-mono border border-emerald-500/20"
                          >
                            {r}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <p className="text-slate-500 text-sm italic">
                        No evaluation run yet
                      </p>
                    )}
                  </div>
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                    <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                      Match Status
                    </p>
                    <div className="flex items-center gap-2">
                      <div
                        className={cn(
                          "w-2 h-2 rounded-full",
                          matchedReceivers.length > 0
                            ? "bg-emerald-500 animate-pulse"
                            : "bg-slate-600",
                        )}
                      />
                      <p
                        className={cn(
                          "font-bold text-lg",
                          matchedReceivers.length > 0
                            ? "text-white"
                            : "text-slate-500",
                        )}
                      >
                        {matchedReceivers.length > 0 ? "Active Route" : "Idle"}
                      </p>
                    </div>
                  </div>

                  {/* Query History */}
                  {queryHistory.length > 0 && (
                    <div>
                      <div className="flex items-center justify-between mb-3 pt-2 border-t border-slate-800">
                        <p className="text-slate-500 text-xs uppercase font-bold tracking-widest">
                          Recent Queries
                        </p>
                        <button
                          onClick={() => {
                            setQueryHistory([]);
                            saveCache(HISTORY_KEY, []);
                          }}
                          className="text-slate-600 hover:text-slate-400 text-xs transition-colors"
                        >
                          Clear
                        </button>
                      </div>
                      <div className="space-y-2">
                        {queryHistory.map((entry, i) => (
                          <button
                            key={i}
                            onClick={() => loadHistoryEntry(entry)}
                            className="w-full text-left p-3 rounded-xl bg-slate-800/40 border border-slate-800 hover:border-slate-600 hover:bg-slate-800/80 transition-all group"
                          >
                            <p className="font-mono text-xs text-slate-300 truncate group-hover:text-white transition-colors">
                              {entry.query}
                            </p>
                            <div className="flex items-center justify-between mt-1.5">
                              <div className="flex gap-1 flex-wrap">
                                {entry.receivers.slice(0, 2).map((r) => (
                                  <span
                                    key={r}
                                    className="px-1.5 py-0.5 rounded text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-mono"
                                  >
                                    {r}
                                  </span>
                                ))}
                                {entry.receivers.length > 2 && (
                                  <span className="text-[10px] text-slate-600">
                                    +{entry.receivers.length - 2}
                                  </span>
                                )}
                              </div>
                              <span className="text-[10px] text-slate-600 shrink-0 ml-1">
                                {formatAge(entry.ts)}
                              </span>
                            </div>
                          </button>
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
        <Route
          path="/lab"
          element={
            <AppLayout
              stats={
                <div className="space-y-4">
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700 flex items-center justify-between">
                    <div>
                      <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                        Passed
                      </p>
                      <p className="text-emerald-400 text-2xl font-bold">
                        {testResults.passed}
                      </p>
                    </div>
                    <CheckCircle2 className="text-emerald-500/20" size={40} />
                  </div>
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700 flex items-center justify-between">
                    <div>
                      <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                        Failed
                      </p>
                      <p className="text-rose-400 text-2xl font-bold">
                        {testResults.failed}
                      </p>
                    </div>
                    <AlertCircle className="text-rose-500/20" size={40} />
                  </div>
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                    <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                      Success Rate
                    </p>
                    <div className="flex items-end gap-2">
                      <p className="text-white text-2xl font-bold">
                        {testResults.passed + testResults.failed > 0
                          ? Math.round(
                              (testResults.passed /
                                (testResults.passed + testResults.failed)) *
                                100,
                            )
                          : 0}
                        %
                      </p>
                    </div>
                  </div>
                </div>
              }
            >
              <LabPage
                onTestsRun={(p, f) => setTestResults({ passed: p, failed: f })}
              />
            </AppLayout>
          }
        />
        <Route
          path="/regression"
          element={
            <AppLayout
              stats={
                <div className="space-y-4">
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700 flex items-center justify-between">
                    <div>
                      <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                        Baseline Tests
                      </p>
                      <p className="text-white text-2xl font-bold">
                        {diffStats?.total ?? "—"}
                      </p>
                    </div>
                    <GitCompare className="text-slate-700" size={36} />
                  </div>
                  <div
                    className={cn(
                      "p-4 rounded-xl border flex items-center justify-between",
                      diffStats && diffStats.drifted > 0
                        ? "bg-amber-500/5 border-amber-500/20"
                        : "bg-slate-800/50 border-slate-700",
                    )}
                  >
                    <div>
                      <p
                        className={cn(
                          "text-xs uppercase font-bold mb-1",
                          diffStats?.drifted
                            ? "text-amber-400"
                            : "text-slate-500",
                        )}
                      >
                        Drifted
                      </p>
                      <p
                        className={cn(
                          "text-2xl font-bold",
                          diffStats?.drifted
                            ? "text-amber-400"
                            : "text-slate-400",
                        )}
                      >
                        {diffStats?.drifted ?? "—"}
                      </p>
                    </div>
                    {diffStats?.drifted ? (
                      <AlertCircle className="text-amber-500/30" size={36} />
                    ) : (
                      <ShieldCheck className="text-slate-700" size={36} />
                    )}
                  </div>
                  {diffStats && (
                    <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                      <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                        Status
                      </p>
                      <div className="flex items-center gap-2 mt-1">
                        <div
                          className={cn(
                            "w-2 h-2 rounded-full",
                            diffStats.drifted > 0
                              ? "bg-amber-500 animate-pulse"
                              : "bg-emerald-500",
                          )}
                        />
                        <span
                          className={cn(
                            "font-medium",
                            diffStats.drifted > 0
                              ? "text-amber-400"
                              : "text-emerald-400",
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
