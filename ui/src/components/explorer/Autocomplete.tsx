import { useState, useEffect, useLayoutEffect, useRef } from "react";
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

  const beforeCursor = text.slice(0, cursorPos);
  const lines = beforeCursor.split("\n");
  const lastLine = lines[lines.length - 1];
  const parts = lastLine.split(",");
  const currentPart = parts[parts.length - 1].trimStart();

  let type: "label" | "value" = "label";
  let filter = currentPart;
  let activeLabel = "";

  if (currentPart.includes("=")) {
    type = "value";
    const [label, ...rest] = currentPart.split("=");
    activeLabel = label.trim();
    filter = rest.join("=").trim().replace(/^["']/, "");
  }

  const list =
    type === "label"
      ? suggestions.labels.filter((l) => l.startsWith(filter))
      : (suggestions.values[activeLabel] || []).filter((v) =>
          v.startsWith(filter),
        );

  // Reset active index when filter or type changes
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setActiveIndex(0);
  }, [filter, type]);

  // Use refs so the keydown handler is registered once, not re-registered on every render.
  // useLayoutEffect keeps refs in sync before browser paint, so handlers always see current values.
  const listRef = useRef(list);
  const activeIndexRef = useRef(activeIndex);
  const onSelectRef = useRef(onSelect);
  const onCloseRef = useRef(onClose);

  useLayoutEffect(() => {
    listRef.current = list;
    activeIndexRef.current = activeIndex;
    onSelectRef.current = onSelect;
    onCloseRef.current = onClose;
  });

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const currentList = listRef.current;
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setActiveIndex((prev) => (prev + 1) % Math.max(1, currentList.length));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setActiveIndex(
          (prev) =>
            (prev - 1 + currentList.length) % Math.max(1, currentList.length),
        );
      } else if (e.key === "Enter" || e.key === "Tab") {
        if (currentList.length > 0) {
          e.preventDefault();
          onSelectRef.current(currentList[activeIndexRef.current]);
        }
      } else if (e.key === "Escape") {
        onCloseRef.current();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []); // Registered once on mount, refs keep values current

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
              : "text-[#d9d9d9] hover:bg-[#ffffff08]",
          )}
          onClick={() => onSelect(item)}
        >
          {item}
        </button>
      ))}
    </div>
  );
};
