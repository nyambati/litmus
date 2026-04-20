import React, { useState, useEffect, useRef } from "react";
import { cn } from "../../utils/persistence";

interface Suggestions {
  labels: string[];
  values: Record<string, string[]>;
}

export const Autocomplete = ({
  suggestions,
  text,
  cursorPos,
  onSelect,
  onClose,
}: {
  suggestions: Suggestions;
  text: string;
  cursorPos: number;
  onSelect: (val: string) => void;
  onClose: () => void;
}) => {
  const [activeIndex, setActiveIndex] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);

  // Determine what we are suggesting
  const beforeCursor = text.slice(0, cursorPos);
  const lines = beforeCursor.split("\n");
  const lastLine = lines[lines.length - 1];
  const parts = lastLine.split(",");
  const currentPart = parts[parts.length - 1].trimStart();

  let type: "label" | "value" = "label";
  let filter = currentPart;
  let activeLabel = "";

  if (currentPart.includes("=") || currentPart.includes(":")) {
    type = "value";
    const delimiter = currentPart.includes("=") ? "=" : ":";
    const [label, ...rest] = currentPart.split(delimiter);
    activeLabel = label.trim();
    filter = rest.join(delimiter).trim().replace(/^["']/, "");
  }

  const list = type === "label" 
    ? suggestions.labels.filter(l => l.startsWith(filter))
    : (suggestions.values[activeLabel] || []).filter(v => v.startsWith(filter));

  useEffect(() => {
    setActiveIndex(0);
  }, [filter, type]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setActiveIndex(prev => (prev + 1) % Math.max(1, list.length));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setActiveIndex(prev => (prev - 1 + list.length) % Math.max(1, list.length));
      } else if (e.key === "Enter" || e.key === "Tab") {
        if (list.length > 0) {
          e.preventDefault();
          onSelect(list[activeIndex]);
        }
      } else if (e.key === "Escape") {
        onClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [list, activeIndex, onSelect, onClose]);

  if (list.length === 0) return null;

  return (
    <div 
      ref={containerRef}
      className="absolute bottom-full mb-1 left-0 z-50 w-64 max-h-48 overflow-y-auto bg-[#1f2128] border border-[#34383e] rounded shadow-xl py-1"
    >
      <div className="px-2 py-1 border-b border-[#34383e] bg-[#111217]">
        <span className="text-[10px] font-bold uppercase tracking-widest text-[#8e9193]">
          {type === "label" ? "Labels" : `Values for ${activeLabel}`}
        </span>
      </div>
      {list.map((item, i) => (
        <button
          key={item}
          className={cn(
            "w-full text-left px-3 py-1.5 text-xs font-mono transition-colors",
            i === activeIndex 
              ? "bg-[#f46800] text-white" 
              : "text-[#d9d9d9] hover:bg-[#ffffff08]"
          )}
          onClick={() => onSelect(item)}
        >
          {item}
        </button>
      ))}
    </div>
  );
};
