import { ChevronRight } from "lucide-react";

export const Header = ({ title, subtitle }: { title: string; subtitle?: string }) => (
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
