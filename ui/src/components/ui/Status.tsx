import { cn } from "../../utils/persistence";

export const StatusDot = ({
  status,
}: {
  status: "pass" | "fail" | "drift" | "idle";
}) => {
  const color = {
    pass: "bg-[#73bf69]",
    fail: "bg-[#f2495c]",
    drift: "bg-[#f5a623]",
    idle: "bg-[#8e9193]",
  }[status];
  return (
    <span
      className={cn("inline-block w-1.5 h-1.5 rounded-full shrink-0", color)}
    />
  );
};

export const StatusBadge = ({
  pass,
  idle,
  drifted,
}: {
  pass?: boolean;
  idle?: boolean;
  drifted?: boolean;
}) => {
  if (idle)
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-xs text-[11px] font-medium bg-[#8e9193]/10 text-[#8e9193] border border-[#34383e]">
        <StatusDot status="idle" />
        Pending
      </span>
    );
  if (drifted)
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-xs text-[11px] font-medium bg-[#f5a623]/10 text-[#f5a623] border border-[#f5a623]/20">
        <StatusDot status="drift" />
        Drifted
      </span>
    );
  if (pass)
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-xs text-[11px] font-medium bg-[#73bf69]/10 text-[#73bf69] border border-[#73bf69]/20">
        <StatusDot status="pass" />
        Passing
      </span>
    );
  return (
    <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-xs text-[11px] font-medium bg-[#f2495c]/10 text-[#f2495c] border border-[#f2495c]/20">
      <StatusDot status="fail" />
      Failed
    </span>
  );
};
