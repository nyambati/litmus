import { cn } from "../../utils/ui";

export const FilterTabs = <T extends string>({
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
