import React, { useState, useEffect, useRef, useCallback } from "react";
import {
  Search,
  Activity,
  Zap,
  CheckCircle2,
  AlertTriangle,
  X,
} from "lucide-react";
import { cn, API, minDelay } from "../../utils/persistence";
import { useExplorerStore } from "../../stores/useExplorerStore";
import { ReceiverChip } from "../ui/Chips";
import { PrimaryButton } from "../ui/Buttons";
import { Header } from "../layout/Header";
import { EmptyState } from "../ui/EmptyState";
import { Autocomplete } from "./Autocomplete";

interface RouteNode {
  receiver?: string;
  matched: boolean;
  match?: string[];
  continue?: boolean;
  group_by?: string[];
  groupBy?: string[];
  group_wait?: string;
  groupWait?: string;
  group_interval?: string;
  groupInterval?: string;
  repeat_interval?: string;
  repeatInterval?: string;
  children?: RouteNode[];
}

interface EvaluationResult {
  receivers?: string[];
  path?: RouteNode;
  [key: string]: unknown;
}

export const ExplorerPage = () => {
  const { labels, setLabels, runTrigger, setMatchedReceivers, saveQuery } =
    useExplorerStore();

  const [result, setResult] = useState<EvaluationResult | null>(null);
  const [loading, setLoading] = useState(false);
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

  const [suggestions, setSuggestions] = useState<{
    labels: string[];
    values: Record<string, string[]>;
  }>({ labels: [], values: {} });
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [cursorPos, setCursorPos] = useState(0);
  const [textareaHeight, setTextareaHeight] = useState("auto");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const adjustTextareaHeight = useCallback(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      const scrollHeight = textareaRef.current.scrollHeight;
      setTextareaHeight(`${Math.max(scrollHeight, 40)}px`);
    }
  }, []);

  useEffect(() => {
    fetch(`${API}/api/v1/label_values`)
      .then((r) => r.json())
      .then(setSuggestions)
      .catch(console.error);
  }, []);

  useEffect(() => {
    adjustTextareaHeight();
  }, [adjustTextareaHeight, labels]);

  const runEvaluation = useCallback(
    async (overrideLabels?: string, silent = false) => {
      setLoading(true);
      if (!silent) setNotification(null);
      try {
        const labelMap: Record<string, string> = {};
        const src = (overrideLabels ?? labels).trim();

        if (!src) throw new Error("Labels cannot be empty");

        const pairs = src.includes(",") ? src.split(",") : src.split("\n");
        pairs.forEach((pair) => {
          const [k, v] = pair
            .split("=")
            .map((s) => s.trim().replace(/^["']|["']$/g, ""));
          if (k && v) labelMap[k] = v;
        });

        const fetchPromise = fetch(`${API}/api/v1/evaluate`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ labels: labelMap }),
        });

        const resp = await minDelay(fetchPromise);

        if (!resp.ok) {
          const text = await resp.text();
          throw new Error(text || `Server returned ${resp.status}`);
        }

        const data = await resp.json();
        setResult(data);
        const receivers = data.receivers || [];
        setMatchedReceivers(receivers);
        saveQuery(overrideLabels ?? labels, receivers);

        if (!silent) {
          setNotification({
            type: "success",
            message: `Evaluation completed: ${receivers.length} matched receivers`,
          });
        }
      } catch (err: any) {
        console.error("Evaluation failed:", err);
        setNotification({
          type: "error",
          message: err.message || "Failed to parse labels. Use k=v format.",
        });
      } finally {
        setLoading(false);
      }
    },
    [labels, setMatchedReceivers, saveQuery],
  );

  useEffect(() => {
    if (runTrigger > 0) {
      runEvaluation(labels, true);
    }
    // Only re-run when runTrigger changes (history entry load), not on every keystroke.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [runTrigger]);

  const handleSelect = useCallback(
    (selected: string) => {
      const beforeCursor = labels.slice(0, cursorPos);
      const afterCursor = labels.slice(cursorPos);

      const lastComma = beforeCursor.lastIndexOf(",");
      const lastNewline = beforeCursor.lastIndexOf("\n");
      const lastStart = Math.max(lastComma, lastNewline);

      const partBeforeToken = beforeCursor.slice(0, lastStart + 1);
      const currentToken = beforeCursor.slice(lastStart + 1);

      let newToken = "";
      if (currentToken.includes("=")) {
        const [labelPart] = currentToken.split("=");
        newToken = `${labelPart}="${selected}"`;
      } else {
        const match = currentToken.match(/^(\s*)/);
        const indent = match ? match[1] : "";
        newToken = indent + selected;
      }

      const nextValue = partBeforeToken + newToken + afterCursor;
      setLabels(nextValue);
      setShowSuggestions(false);

      setTimeout(() => {
        if (textareaRef.current) {
          textareaRef.current.focus();
          const nextPos = partBeforeToken.length + newToken.length;
          textareaRef.current.setSelectionRange(nextPos, nextPos);
        }
      }, 0);
    },
    [labels, cursorPos, setLabels],
  );

  const handleCloseSuggestions = useCallback(
    () => setShowSuggestions(false),
    [],
  );

  const renderPath = (
    node: RouteNode | null,
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
            {((node.match?.length ?? 0) > 0 || node.matched) && (
              <div className="px-4 pb-3 border-t border-[#2c3235] pt-2.5 grid grid-cols-2 gap-4">
                {(node.match?.length ?? 0) > 0 && (
                  <div className="space-y-1">
                    {node.match!.map((m: string, idx: number) => (
                      <div
                        key={`${currentNodeId}-match-${idx}`}
                        className="flex gap-2 font-mono text-[11px]"
                      >
                        <span className="text-[#5794f2]/50 shrink-0">
                          match:
                        </span>
                        <span className="text-[#8e9193] truncate">{m}</span>
                      </div>
                    ))}
                  </div>
                )}

                {node.matched && (
                  <div className="space-y-1 text-[11px] font-mono">
                    {((node.group_by || node.groupBy)?.length ?? 0) > 0 && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">
                          group
                        </span>
                        <span className="text-[#8e9193]">
                          [{(node.group_by ?? node.groupBy ?? []).join(", ")}]
                        </span>
                      </div>
                    )}
                    {(node.group_wait || node.groupWait) && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">
                          wait
                        </span>
                        <span className="text-[#8e9193]">
                          {node.group_wait || node.groupWait}
                        </span>
                      </div>
                    )}
                    {(node.group_interval || node.groupInterval) && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">
                          interval
                        </span>
                        <span className="text-[#8e9193]">
                          {node.group_interval || node.groupInterval}
                        </span>
                      </div>
                    )}
                    {(node.repeat_interval || node.repeatInterval) && (
                      <div className="flex gap-2">
                        <span className="text-[#34383e] w-14 shrink-0">
                          repeat
                        </span>
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
            {node.children.map((child, idx: number) =>
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

        {!result && !loading && (
          <EmptyState
            icon={Search}
            title="No evaluation run yet"
            description="Enter alert labels below and click Run Query to trace the routing path"
          />
        )}

        {result && (
          <div className="pb-8 space-y-1 animate-fade-in-up">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <Activity size={14} className="text-[#f46800]" />
                <span className="text-sm font-medium text-[#d9d9d9]">
                  Route Path
                </span>
                {(result.receivers?.length ?? 0) > 0 && (
                  <div className="flex items-center gap-1.5 ml-2">
                    {result.receivers!.map((r: string, i: number) => (
                      <ReceiverChip
                        key={`${r}-${i}`}
                        name={r}
                        variant="green"
                      />
                    ))}
                  </div>
                )}
              </div>
            </div>
            {renderPath(result.path ?? null)}
          </div>
        )}
      </main>

      {/* Query input bar — Grafana Explore style */}
      <div className="border-t border-[#2c3235] bg-[#1f2128] p-4 shrink-0">
        <div className="flex items-center gap-2 mb-3">
          <div className="flex-1 relative flex items-center gap-2 bg-[#111217] border border-[#2c3235] rounded px-3 py-0 transition-colors">
            <textarea
              ref={textareaRef}
              value={labels}
              onChange={(e) => {
                setLabels(e.target.value);
                setCursorPos(e.target.selectionStart);
                setShowSuggestions(true);
                adjustTextareaHeight();
              }}
              onKeyUp={(e) => {
                setCursorPos((e.target as HTMLTextAreaElement).selectionStart);
              }}
              onFocus={(e) => {
                setCursorPos(e.target.selectionStart);
              }}
              onBlur={() => {
                setTimeout(() => setShowSuggestions(false), 200);
              }}
              style={{ height: textareaHeight }}
              className="flex-1 bg-transparent py-2.5 font-mono text-[13px] text-[#5794f2] focus:outline-none resize-none placeholder:text-[#34383e] overflow-hidden"
              placeholder="severity=critical, team=database"
              onKeyDown={(e) => {
                if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
                  e.preventDefault();
                  runEvaluation();
                }
              }}
            />
            {showSuggestions && (
              <Autocomplete
                suggestions={suggestions}
                text={labels}
                cursorPos={cursorPos}
                onSelect={handleSelect}
                onClose={handleCloseSuggestions}
              />
            )}
          </div>
          <PrimaryButton
            onClick={() => runEvaluation()}
            loading={loading}
            icon={<Zap size={14} />}
          >
            Query
          </PrimaryButton>
        </div>
        <p className="text-[10px] text-[#8e9193]/50 font-mono">
          ⌘ + Enter to run · k=v comma-separated or newline-separated
        </p>
      </div>
    </div>
  );
};
