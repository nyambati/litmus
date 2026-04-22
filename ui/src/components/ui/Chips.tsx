import { cn } from "../../utils/ui";

export const LabelChip = ({ labelKey, value }: { labelKey: string; value: string }) => (
  <span className="inline-flex items-center font-mono text-[11px] px-1.5 py-0.5 bg-[#22252b] border border-[#34383e] rounded-[2px] text-[#d9d9d9]/80">
    <span className="text-[#8e9193]">{labelKey}=</span>
    {value}
  </span>
);

export const ReceiverChip = ({
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
