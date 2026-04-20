import { cn } from "../../utils/persistence";

export const GfSpinner = ({ size = "md" }: { size?: "sm" | "md" | "lg" }) => {
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
