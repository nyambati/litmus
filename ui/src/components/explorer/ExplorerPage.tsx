import React, { useState, useEffect, useRef, useCallback } from "react";
import { Search, Activity, Zap } from "lucide-react";
import { cn, API } from "../../utils/persistence";
import { GfSpinner } from "../ui/Spinner";
import { ReceiverChip } from "../ui/Chips";
import { PrimaryButton, GhostButton } from "../ui/Buttons";
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
  route?: RouteNode;
  [key: string]: unknown;
}

export const ExplorerPage = ({
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
  const [result, setResult] = useState<EvaluationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [suggestions, setSuggestions] = useState<{ labels: string[], values: Record<string, string[]> }>({ labels: [], values: {} });
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [cursorPos, setCursorPos] = useState(0);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    fetch(`${API}/api/v1/suggest`)
      .then(r => r.json())
      .then(setSuggestions)
      .catch(console.error);
  }, []);

  const runEvaluation = useCallback(async (overrideLabels?: string) => {
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
  }, [labels, onEvaluate, onQuerySaved]);

  useEffect(() => {
    if (runTrigger > 0) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      runEvaluation(labels);
    }
  }, [runTrigger, runEvaluation, labels]);

  const handleSelect = (selected: string) => {
    const beforeCursor = labels.slice(0, cursorPos);
    const afterCursor = labels.slice(cursorPos);
    
    const lastComma = beforeCursor.lastIndexOf(",");
    const lastNewline = beforeCursor.lastIndexOf("\n");
    const lastStart = Math.max(lastComma, lastNewline);
    
    const partBeforeToken = beforeCursor.slice(0, lastStart + 1);
    const currentToken = beforeCursor.slice(lastStart + 1);
    
    let newToken = "";
    if (currentToken.includes("=") || currentToken.includes(":")) {
      const delimiter = currentToken.includes("=") ? "=" : ":";
      const [labelPart] = currentToken.split(delimiter);
      newToken = `${labelPart}${delimiter}"${selected}"`;
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
  };

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
          <div className="flex-1 relative flex items-center gap-2 bg-[#111217] border border-[#2c3235] rounded px-3 py-0 transition-colors">
            <Search size={13} className="text-[#8e9193] shrink-0" />
            <textarea
              ref={textareaRef}
              value={labels}
              onChange={(e) => {
                setLabels(e.target.value);
                setCursorPos(e.target.selectionStart);
                setShowSuggestions(true);
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
            {showSuggestions && (
              <Autocomplete 
                suggestions={suggestions}
                text={labels}
                cursorPos={cursorPos}
                onSelect={handleSelect}
                onClose={() => setShowSuggestions(false)}
              />
            )}
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
