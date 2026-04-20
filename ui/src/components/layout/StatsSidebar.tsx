import React from "react";
import { cn } from "../../utils/persistence";

export const StatPanel = ({
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

export const StatsSidebar = ({ children }: { children?: React.ReactNode }) => (
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
