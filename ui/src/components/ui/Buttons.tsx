import React from "react";
import { cn } from "../../utils/ui";
import { GfSpinner } from "./Spinner";

export const PrimaryButton = ({
  onClick,
  disabled,
  loading,
  icon,
  children,
  className,
}: {
  onClick: () => void;
  disabled?: boolean;
  loading?: boolean;
  icon?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) => (
  <button
    onClick={onClick}
    disabled={disabled || loading}
    className={cn(
      "inline-flex items-center gap-2 px-4 py-[7px] rounded bg-[#f46800] hover:bg-[#ff7f2a] disabled:opacity-40 disabled:cursor-not-allowed text-white text-sm font-semibold transition-colors",
      className,
    )}
  >
    {loading ? <GfSpinner size="sm" /> : icon}
    {children}
  </button>
);

export const GhostButton = ({
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
