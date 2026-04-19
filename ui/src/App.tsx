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
  Clock,
  ChevronRight,
  Zap,
  AlertTriangle,
  Circle,
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

// --- Design Tokens (inline references) ---
// bg-[#181b1f]   = body
// bg-[#1f2128]   = panel
// bg-[#22252b]   = elevated
// bg-[#111217]   = sidebar
// border-[#2c3235] = subtle border
// border-[#34383e] = strong border
// text-[#d9d9d9]  = primary text
// text-[#8e9193]  = secondary text
// text-[#f46800]  = orange accent
// text-[#73bf69]  = green
// text-[#f2495c]  = red
// text-[#f5a623]  = yellow/warning
// text-[#5794f2]  = blue
// text-[#b877d9]  = purple

// --- Shared Components ---

const GfSpinner = ({ size = "md" }: { size?: "sm" | "md" | "lg" }) => {
  const dim = size === "sm" ? "w-3.5 h-3.5" : size === "lg" ? "w-10 h-10" : "w-5 h-5";
  const border = size === "lg" ? "border-[3px]" : "border-2";
  return (
    <div
      className={cn(
        dim,
        border,
        "border-[#f46800]/20 border-t-[#f46800] rounded-full animate-spin",
      )}
    />
  );
};

const StatusDot = ({ status }: { status: "pass" | "fail" | "drift" | "idle" }) => {
  const color = {
    pass: "bg-[#73bf69]",
    fail: "bg-[#f2495c]",
    drift: "bg-[#f5a623]",
    idle: "bg-[#8e9193]",
  }[status];
  return <span className={cn("inline-block w-1.5 h-1.5 rounded-full shrink-0", color)} />;
};

const StatusBadge = ({
  pass,
  idle,
  drifted,
}: {
  pass?: boolean;
  idle?: boolean;
  drifted?: boolean;
}) => {
  if (idle)
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[2px] text-[11px] font-medium bg-[#8e9193]/10 text-[#8e9193] border border-[#34383e]">
        <StatusDot status="idle" />
        Pending
      </span>
    );
  if (drifted)
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[2px] text-[11px] font-medium bg-[#f5a623]/10 text-[#f5a623] border border-[#f5a623]/20">
        <StatusDot status="drift" />
        Drifted
      </span>
    );
  if (pass)
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[2px] text-[11px] font-medium bg-[#73bf69]/10 text-[#73bf69] border border-[#73bf69]/20">
        <StatusDot status="pass" />
        Passing
      </span>
    );
  return (
    <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[2px] text-[11px] font-medium bg-[#f2495c]/10 text-[#f2495c] border border-[#f2495c]/20">
      <StatusDot status="fail" />
      Failed
    </span>
  );
};

const LastUpdated = ({ ts }: { ts: number | null }) => {
  if (!ts) return null;
  return (
    <span className="inline-flex items-center gap-1 text-[11px] text-[#8e9193]">
      <Clock size={11} />
      {formatAge(ts)}
    </span>
  );
};

const LabelChip = ({ labelKey, value }: { labelKey: string; value: string }) => (
  <span className="inline-flex items-center font-mono text-[11px] px-1.5 py-0.5 bg-[#22252b] border border-[#34383e] rounded-[2px] text-[#d9d9d9]/80">
    <span className="text-[#8e9193]">{labelKey}=</span>
    {value}
  </span>
);

const ReceiverChip = ({
  name,
  variant = "blue",
}: {
  name: string;
  variant?: "blue" | "purple" | "green" | "amber";
}) => {
  const colors = {
    blue: "bg-[#5794f2]/10 border-[#5794f2]/25 text-[#5794f2]",
    purple: "bg-[#b877d9]/10 border-[#b877d9]/25 text-[#b877d9]",
    green: "bg-[#73bf69]/10 border-[#73bf69]/25 text-[#73bf69]",
    amber: "bg-[#f5a623]/10 border-[#f5a623]/25 text-[#f5a623]",
  }[variant];
  return (
    <span
      className={cn(
        "inline-flex items-center font-mono text-[11px] px-2 py-0.5 rounded-[2px] border",
        colors,
      )}
    >
      {name}
    </span>
  );
};

// Primary / secondary buttons
const PrimaryButton = ({
  onClick,
  disabled,
  loading,
  icon,
  children,
}: {
  onClick: () => void;
  disabled?: boolean;
  loading?: boolean;
  icon?: React.ReactNode;
  children: React.ReactNode;
}) => (
  <button
    onClick={onClick}
    disabled={disabled || loading}
    className="inline-flex items-center gap-2 px-4 py-[7px] rounded bg-[#f46800] hover:bg-[#ff7f2a] disabled:opacity-40 disabled:cursor-not-allowed text-white text-sm font-semibold transition-colors"
  >
    {loading ? <GfSpinner size="sm" /> : icon}
    {children}
  </button>
);

const GhostButton = ({
  onClick,
  children,
  className,
}: {
  onClick?: () => void;
  children: React.ReactNode;
  className?: string;
}) => (
  <button
    onClick={onClick}
    className={cn(
      "inline-flex items-center gap-1.5 px-3 py-[5px] rounded border border-[#34383e] text-[#d9d9d9] text-sm hover:bg-[#ffffff08] transition-colors",
      className,
    )}
  >
    {children}
  </button>
);

// --- Sidebar ---

const Sidebar = () => {
  const location = useLocation();

  const navItems = [
    { name: "Explorer", path: "/", icon: Search, description: "Route evaluator" },
    { name: "Lab", path: "/lab", icon: FlaskConical, description: "Test runner" },
    { name: "Regression", path: "/regression", icon: History, description: "Drift detection" },
  ];

  return (
    <aside
      className="w-[220px] flex flex-col shrink-0 border-r border-[#1e2228]"
      style={{ background: "#111217" }}
    >
      {/* Logo */}
      <div className="h-12 flex items-center px-4 border-b border-[#1e2228] shrink-0">
        <div className="flex items-center gap-2.5">
          <div className="w-6 h-6 rounded flex items-center justify-center bg-[#f46800]/15">
            <Activity size={14} className="text-[#f46800]" />
          </div>
          <span className="text-[#d9d9d9] font-semibold text-[15px] tracking-tight">
            Litmus
          </span>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 py-2">
        {navItems.map((item) => {
          const active = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                "relative flex items-center gap-3 px-4 py-2.5 text-sm transition-colors group",
                active
                  ? "text-[#d9d9d9] bg-[#f46800]/8"
                  : "text-[#8e9193] hover:text-[#d9d9d9] hover:bg-[#ffffff06]",
              )}
            >
              {/* Active indicator */}
              {active && (
                <span className="absolute left-0 top-0 bottom-0 w-[3px] bg-[#f46800] rounded-r" />
              )}
              <item.icon size={16} className={active ? "text-[#f46800]" : ""} />
              <span className="font-medium">{item.name}</span>
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-[#1e2228]">
        <span className="text-[11px] text-[#8e9193]/60 font-mono">v0.1.0-alpha</span>
      </div>
    </aside>
  );
};

// --- Header ---

const Header = ({ title, subtitle }: { title: string; subtitle?: string }) => (
  <header className="h-12 border-b border-[#2c3235] flex items-center px-6 bg-[#1f2128] shrink-0">
    <div className="flex items-center gap-2 text-[#8e9193] text-sm">
      <span className="text-[#d9d9d9] font-medium">{title}</span>
      {subtitle && (
        <>
          <ChevronRight size={14} className="text-[#34383e]" />
          <span>{subtitle}</span>
        </>
      )}
    </div>
  </header>
);

// --- Stats Sidebar ---

const StatPanel = ({
  label,
  value,
  color = "default",
  icon,
}: {
  label: string;
  value: React.ReactNode;
  color?: "default" | "green" | "red" | "yellow" | "orange";
  icon?: React.ReactNode;
}) => {
  const valueColor = {
    default: "text-[#d9d9d9]",
    green: "text-[#73bf69]",
    red: "text-[#f2495c]",
    yellow: "text-[#f5a623]",
    orange: "text-[#f46800]",
  }[color];

  return (
    <div className="bg-[#1f2128] border border-[#2c3235] rounded-sm p-3">
      <div className="flex items-center justify-between mb-1">
        <span className="text-[11px] font-medium text-[#8e9193] uppercase tracking-wider">
          {label}
        </span>
        {icon && <span className="text-[#34383e]">{icon}</span>}
      </div>
      <p className={cn("text-2xl font-bold tabular-nums", valueColor)}>{value}</p>
    </div>
  );
};

const StatsSidebar = ({ children }: { children?: React.ReactNode }) => (
  <aside className="w-72 border-l border-[#2c3235] flex flex-col bg-[#181b1f] overflow-y-auto shrink-0">
    <div className="h-12 border-b border-[#2c3235] flex items-center px-4 bg-[#1f2128] shrink-0">
      <span className="text-[11px] font-semibold text-[#8e9193] uppercase tracking-widest">
        Overview
      </span>
    </div>
    <div className="flex-1 p-4 space-y-3">
      {children || (
        <div className="p-4 rounded-sm bg-[#1f2128] border border-[#2c3235]">
          <p className="text-[#8e9193] text-sm">Select a page to see stats</p>
        </div>
      )}
    </div>
  </aside>
);

// --- Empty States ---

const EmptyState = ({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon: React.ElementType;
  title: string;
  description?: string;
  action?: React.ReactNode;
}) => (
  <div className="flex flex-col items-center justify-center py-20 gap-4">
    <div className="w-14 h-14 rounded-full bg-[#1f2128] border border-[#2c3235] flex items-center justify-center">
      <Icon size={24} className="text-[#34383e]" />
    </div>
    <div className="text-center space-y-1">
      <p className="text-[#d9d9d9] font-medium text-sm">{title}</p>
      {description && (
        <p className="text-[#8e9193] text-xs max-w-xs">{description}</p>
      )}
    </div>
    {action && <div className="mt-2">{action}</div>}
  </div>
);

// --- Explorer Page ---

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
      <div key={currentNodeId} className="space-y-2 relative">
        {/* Vertical connector line */}
        {depth > 0 && (
          <div className="absolute left-[15px] -top-2 w-px h-2 bg-[#2c3235]" />
        )}

        <div className="flex items-start gap-3 group">
          {/* Depth dot */}
          <div
            className={cn(
              "w-8 h-8 shrink-0 rounded-sm flex items-center justify-center text-xs font-bold z-10 border transition-all",
              node.matched
                ? "bg-[#f46800]/15 border-[#f46800]/40 text-[#f46800]"
                : "bg-[#1f2128] border-[#2c3235] text-[#8e9193]",
            )}
          >
            {depth + 1}
          </div>

          {/* Route node panel */}
          <div
            className={cn(
              "flex-1 border rounded-sm transition-all overflow-hidden",
              node.matched
                ? "bg-[#1f2128] border-[#f46800]/30 border-l-[3px] border-l-[#f46800]"
                : "bg-[#1f2128] border-[#2c3235] hover:border-[#34383e]",
            )}
          >
            <div className="flex items-center justify-between px-4 py-2.5">
              <span
                className={cn(
                  "font-medium text-sm",
                  node.matched ? "text-[#d9d9d9]" : "text-[#8e9193]",
                )}
              >
                {node.receiver || "root"}
              </span>
              <div className="flex items-center gap-2">
                {node.continue && (
                  <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 bg-[#5794f2]/10 text-[#5794f2] border border-[#5794f2]/20 rounded-[2px]">
                    continue
                  </span>
                )}
                <span
                  className={cn(
                    "text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded-[2px] border",
                    node.matched
                      ? "bg-[#f46800]/10 text-[#f46800] border-[#f46800]/25"
                      : "bg-[#22252b] text-[#8e9193] border-[#34383e]",
                  )}
                >
                  {node.matched ? "matched" : "no match"}
                </span>
              </div>
            </div>

            {/* Matchers + timing */}
            {(node.match?.length > 0 || node.matched) && (
              <div className="px-4 pb-3 border-t border-[#2c3235] pt-2.5 grid grid-cols-2 gap-4">
                {node.match?.length > 0 && (
                  <div className="space-y-1">
                    {node.match.map((m: string, idx: number) => (
                      <div
                        key={`${currentNodeId}-match-${idx}`}
                        className="flex gap-2 font-mono text-[11px]"
                      >
                        <span className="text-[#5794f2]/50 shrink-0">match:</span>
                        <span className="text-[#8e9193] truncate">{m}</span>
                      </div>
                    ))}
                  </div>
                )}

                {node.matched && (
                  <div className="space-y-1 text-[11px] font-mono">
                    {(node.group_by || node.groupBy)?.length > 0 && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">group</span>
                        <span className="text-[#8e9193]">
                          [{(node.group_by || node.groupBy).join(", ")}]
                        </span>
                      </div>
                    )}
                    {(node.group_wait || node.groupWait) && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">wait</span>
                        <span className="text-[#8e9193]">
                          {node.group_wait || node.groupWait}
                        </span>
                      </div>
                    )}
                    {(node.group_interval || node.groupInterval) && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">interval</span>
                        <span className="text-[#8e9193]">
                          {node.group_interval || node.groupInterval}
                        </span>
                      </div>
                    )}
                    {(node.repeat_interval || node.repeatInterval) && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">repeat</span>
                        <span className="text-[#8e9193]">
                          {node.repeat_interval || node.repeatInterval}
                        </span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Children */}
        {node.children && node.children.length > 0 && (
          <div className="ml-11 space-y-2 border-l border-[#2c3235] pl-4">
            {node.children.map((child: any, idx: number) =>
              renderPath(child, depth + 1, `${currentNodeId}-${idx}`),
            )}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="flex-1 flex flex-col h-screen min-h-0 bg-[#181b1f]">
      <Header title="Route Explorer" />

      {/* Route path result */}
      <main className="flex-1 overflow-y-auto p-6">
        {!result && !loading && (
          <EmptyState
            icon={Search}
            title="No evaluation run yet"
            description="Enter alert labels below and click Run Query to trace the routing path"
          />
        )}

        {loading && (
          <div className="flex items-center justify-center py-20">
            <GfSpinner size="lg" />
          </div>
        )}

        {result && (
          <div className="pb-8 space-y-1">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <Activity size={14} className="text-[#f46800]" />
                <span className="text-sm font-medium text-[#d9d9d9]">
                  Route Path
                </span>
                {result.receivers?.length > 0 && (
                  <div className="flex items-center gap-1.5 ml-2">
                    {result.receivers.map((r: string) => (
                      <ReceiverChip key={r} name={r} variant="green" />
                    ))}
                  </div>
                )}
              </div>
            </div>
            {renderPath(result.path)}
          </div>
        )}
      </main>

      {/* Query input bar — Grafana Explore style */}
      <div className="border-t border-[#2c3235] bg-[#1f2128] p-4 shrink-0">
        <div className="flex items-center gap-2 mb-3">
          <div className="flex-1 flex items-center gap-2 bg-[#111217] border border-[#2c3235] rounded px-3 py-0 focus-within:border-[#f46800]/50 transition-colors">
            <Search size={13} className="text-[#8e9193] shrink-0" />
            <textarea
              value={labels}
              onChange={(e) => setLabels(e.target.value)}
              rows={2}
              className="flex-1 bg-transparent py-2.5 font-mono text-[13px] text-[#5794f2] focus:outline-none resize-none placeholder:text-[#34383e]"
              placeholder='severity="critical", team="database"'
              onKeyDown={(e) => {
                if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
                  e.preventDefault();
                  runEvaluation();
                }
              }}
            />
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <GhostButton onClick={() => setLabels("")}>Clear</GhostButton>
            <PrimaryButton
              onClick={() => runEvaluation()}
              loading={loading}
              icon={<Zap size={14} />}
            >
              Run Query
            </PrimaryButton>
          </div>
        </div>
        <p className="text-[10px] text-[#8e9193]/50 font-mono">
          ⌘ + Enter to run · k=v,k=v · JSON · newline-separated
        </p>
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
          <h4 className="text-[#d9d9d9] text-sm font-medium truncate">{name}</h4>
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
                  key={i}
                  className="text-[10px] text-[#8e9193]/60 px-1 py-px rounded-[2px] bg-[#22252b] border border-[#2c3235]"
                >
                  {t}
                </span>
              ))}
          </div>
        </div>

        {/* Result badge + run button */}
        <div className="flex items-center gap-2 shrink-0">
          {result ? (
            <StatusBadge pass={result.pass} />
          ) : (
            <StatusBadge idle />
          )}
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
            {isRunning ? (
              <GfSpinner size="sm" />
            ) : (
              <Play size={11} />
            )}
          </button>
        </div>
      </div>

      {/* Panel body */}
      {children}
    </div>
  );
};

interface TestCaseProps {
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
}: TestCaseProps) => {
  const outcome = test.expect?.outcome;
  const receivers = test.expect?.receivers || [];
  const alertLabels = test.alert?.labels || {};
  const hasState =
    test.state &&
    (test.state.silences?.length > 0 || test.state.active_alerts?.length > 0);

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
          {receivers.map((r: string) => (
            <ReceiverChip key={r} name={r} variant="blue" />
          ))}
        </div>

        {/* State hint */}
        {hasState && (
          <div className="flex items-center gap-3 text-[11px] text-[#8e9193]/50 font-mono">
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
          {expected.map((r) => (
            <ReceiverChip key={r} name={r} variant="purple" />
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
                  <span className="text-[#8e9193]/40 w-16 shrink-0">expected</span>
                  <span className="text-[#d9d9d9]/70">
                    [{result.expected.join(", ")}]
                  </span>
                </div>
                <div className="flex gap-2 items-baseline">
                  <span className="text-[#8e9193]/40 w-16 shrink-0">actual</span>
                  <span className="text-[#f2495c]">
                    [{(result.actual || []).join(", ")}]
                  </span>
                </div>
                {result.labels && (
                  <div className="flex gap-2 pt-1.5 border-t border-[#2c3235]">
                    <span className="text-[#8e9193]/40 w-16 shrink-0">labels</span>
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

const TestCaseCard = (props: TestCaseProps) =>
  props.test.type === "regression" ? (
    <RegressionTestCase {...props} />
  ) : (
    <UnitTestCase {...props} />
  );

// --- Filter Tabs (Grafana underline style) ---

type FilterType = "all" | "unit" | "regression";

const FilterTabs = <T extends string>({
  tabs,
  active,
  onChange,
}: {
  tabs: { label: string; value: T; count?: number }[];
  active: T;
  onChange: (v: T) => void;
}) => (
  <div className="gf-tabs mb-5">
    {tabs.map((tab) => (
      <button
        key={tab.value}
        onClick={() => onChange(tab.value)}
        className={cn("gf-tab", active === tab.value && "active")}
      >
        {tab.label}
        {tab.count !== undefined && (
          <span className="ml-1.5 text-[10px] opacity-60">({tab.count})</span>
        )}
      </button>
    ))}
  </div>
);

// --- Lab Page ---

const LabPage = ({
  onTestsRun,
}: {
  onTestsRun: (passed: number, failed: number) => void;
}) => {
  const [tests, setTests] = useState<any[]>([]);
  const _labCache = loadCache<Record<string, any>>("litmus:lab:results");
  const [results, setResults] = useState<Record<string, any>>(
    _labCache?.data ?? {},
  );
  const [lastRunTs, setLastRunTs] = useState<number | null>(
    _labCache?.ts ?? null,
  );
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
        console.error("Failed to fetch regression tests:", regressionSettled.reason);

      const unitTests = unitData.map((t) => ({ ...t, type: "unit" }));
      const regressionTests = regressionData.map((t) => ({
        ...t,
        type: "regression",
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

  const filterTabs = [
    { label: "All", value: "all" as FilterType, count: tests.length },
    {
      label: "Unit",
      value: "unit" as FilterType,
      count: tests.filter((t) => getTestType(t) === "unit").length,
    },
    {
      label: "Regression",
      value: "regression" as FilterType,
      count: tests.filter((t) => getTestType(t) === "regression").length,
    },
  ];

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#181b1f]">
      <Header title="Test Lab" />
      <main className="flex-1 p-6 overflow-y-auto">
        {/* Toolbar */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <h3 className="text-[#d9d9d9] font-semibold text-sm">
              {filter === "unit"
                ? "Unit Tests"
                : filter === "regression"
                  ? "Regression Tests"
                  : "All Tests"}
            </h3>
            <span className="text-[11px] text-[#8e9193]">
              {filteredTests.length} of {tests.length}
            </span>
            <LastUpdated ts={lastRunTs} />
          </div>
          <PrimaryButton
            onClick={runAllTests}
            disabled={running || filteredTests.length === 0}
            loading={running}
            icon={<FlaskConical size={14} />}
          >
            {running ? "Running…" : "Run All"}
          </PrimaryButton>
        </div>

        {/* Underline tabs */}
        <FilterTabs tabs={filterTabs} active={filter} onChange={setFilter} />

        {/* Content */}
        {loading ? (
          <div className="flex items-center justify-center h-48">
            <GfSpinner size="lg" />
          </div>
        ) : filteredTests.length === 0 ? (
          <EmptyState
            icon={FlaskConical}
            title="No tests found"
            description={
              filter === "regression"
                ? "Run 'litmus snapshot' to generate a regression baseline"
                : "Add YAML test files to the tests/ directory"
            }
          />
        ) : (
          <div className="space-y-2">
            {filteredTests.map((test, testIdx) => (
              <TestCaseCard
                key={`${test.name}-${testIdx}`}
                test={test}
                result={results[test.name]}
                isRunning={runningTest === test.name}
                globalRunning={running}
                onRun={() => runSingleTest(test.name, getTestType(test))}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
};

// --- Regression / Diff Page ---

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
  const [diff, setDiff] = useState<DiffResult | null>(
    _diffCache?.data ?? null,
  );
  const [lastRunTs, setLastRunTs] = useState<number | null>(
    _diffCache?.ts ?? null,
  );
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
    diff?.results.filter((r) => (filter === "drifted" ? !r.pass : r.pass)) ??
    [];

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#181b1f]">
      <Header title="Regression" subtitle="Baseline Comparison" />
      <main className="flex-1 p-6 overflow-y-auto">
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
          <PrimaryButton
            onClick={runDiff}
            loading={loading}
            icon={<RefreshCw size={14} />}
          >
            {loading ? "Analyzing…" : "Run Diff"}
          </PrimaryButton>
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
            icon={GitCompare}
            title="No baseline comparison run"
            description="Click Run Diff to compare current routing against the snapshot baseline"
            action={
              <PrimaryButton
                onClick={runDiff}
                loading={loading}
                icon={<RefreshCw size={14} />}
              >
                Run Diff
              </PrimaryButton>
            }
          />
        )}

        {diff && (
          <>
            {/* Summary row */}
            <div className="grid grid-cols-3 gap-3 mb-5">
              <StatPanel
                label="Total"
                value={diff.total}
                icon={<GitCompare size={20} />}
              />
              <StatPanel
                label="Passing"
                value={diff.passed}
                color="green"
                icon={<CheckCircle2 size={20} />}
              />
              <StatPanel
                label="Drifted"
                value={diff.drifted}
                color={diff.drifted > 0 ? "yellow" : "default"}
                icon={<AlertTriangle size={20} />}
              />
            </div>

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
                { label: "Drifted", value: "drifted" as const, count: diff.drifted },
                { label: "Passing", value: "passing" as const, count: diff.passed },
              ]}
              active={filter}
              onChange={setFilter}
            />

            {/* Result cards */}
            <div className="space-y-2">
              {visibleResults.length === 0 && (
                <div className="py-8 text-center text-[#8e9193] text-sm">
                  No {filter} results
                </div>
              )}
              {visibleResults.map((result, i) => (
                <div
                  key={`${result.name}-${i}`}
                  className={cn(
                    "bg-[#1f2128] border border-[#2c3235] border-l-[3px] rounded-sm overflow-hidden",
                    result.pass
                      ? "border-l-[#73bf69]"
                      : "border-l-[#f5a623]",
                  )}
                >
                  {/* Row header */}
                  <div className="flex items-center gap-3 px-4 py-3 bg-[#22252b] border-b border-[#2c3235]">
                    <div className="shrink-0">
                      {result.pass ? (
                        <CheckCircle2 size={14} className="text-[#73bf69]" />
                      ) : (
                        <AlertTriangle size={14} className="text-[#f5a623]" />
                      )}
                    </div>
                    <p className="flex-1 text-[#d9d9d9] text-sm font-medium truncate">
                      {result.name}
                    </p>
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
                      {result.expected && (
                        <div className="flex items-center gap-2 flex-wrap font-mono text-xs">
                          <div className="flex items-center gap-1.5 flex-wrap">
                            <span className="text-[11px] text-[#8e9193]/50 uppercase tracking-wider font-sans font-bold">
                              baseline
                            </span>
                            {result.expected.map((r) => (
                              <ReceiverChip key={r} name={r} variant="blue" />
                            ))}
                          </div>
                          <ArrowRight size={13} className="text-[#34383e] shrink-0" />
                          <div className="flex items-center gap-1.5 flex-wrap">
                            <span className="text-[11px] text-[#8e9193]/50 uppercase tracking-wider font-sans font-bold">
                              current
                            </span>
                            {(result.actual || []).length > 0 ? (
                              (result.actual || []).map((r) => (
                                <ReceiverChip key={r} name={r} variant="amber" />
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
                                      <ArrowRight size={10} className="text-[#34383e] shrink-0" />
                                      <span className="px-2 py-0.5 rounded-[2px] bg-[#f5a623]/10 border border-[#f5a623]/25 text-[#f5a623] font-semibold">
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
                        <p className="font-mono text-xs text-[#f2495c]/80">
                          {result.error}
                        </p>
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

// --- Layout & Root ---

const AppLayout = ({
  children,
  stats,
}: {
  children: React.ReactNode;
  stats?: React.ReactNode;
}) => (
  <div className="flex h-screen bg-[#181b1f] text-[#d9d9d9] w-full overflow-hidden">
    <Sidebar />
    <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
      {children}
    </main>
    <StatsSidebar>{stats}</StatsSidebar>
  </div>
);

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
  const [testResults, setTestResults] = useState<{
    passed: number;
    failed: number;
  }>(() => {
    if (!_labCache?.data) return { passed: 0, failed: 0 };
    let passed = 0;
    let failed = 0;
    Object.values(_labCache.data).forEach((r: any) => {
      if (r.pass) passed++;
      else failed++;
    });
    return { passed, failed };
  });

  const _diffCache = loadCache<DiffResult>("litmus:regression:diff");
  const [diffStats, setDiffStats] = useState<{
    total: number;
    drifted: number;
  } | null>(
    _diffCache?.data
      ? { total: _diffCache.data.total, drifted: _diffCache.data.drifted }
      : null,
  );

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
                          <button
                            key={i}
                            onClick={() => loadHistoryEntry(entry)}
                            className="w-full text-left p-2.5 rounded-[2px] bg-[#1f2128] border border-[#2c3235] hover:border-[#34383e] hover:bg-[#22252b] transition-all group"
                          >
                            <p className="font-mono text-[11px] text-[#d9d9d9]/70 truncate group-hover:text-[#d9d9d9] transition-colors">
                              {entry.query}
                            </p>
                            <div className="flex items-center justify-between mt-1.5">
                              <div className="flex gap-1 flex-wrap">
                                {entry.receivers.slice(0, 2).map((r) => (
                                  <ReceiverChip
                                    key={r}
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
                onDiffRun={(total, drifted) =>
                  setDiffStats({ total, drifted })
                }
              />
            </AppLayout>
          }
        />
      </Routes>
    </Router>
  );
}

export default App;
