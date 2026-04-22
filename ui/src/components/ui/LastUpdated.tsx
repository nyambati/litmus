import { Clock } from "lucide-react";
import { formatAge } from "../../utils/ui";

export const LastUpdated = ({ ts }: { ts: number | null }) => {
  if (!ts) return null;
  return (
    <span className="inline-flex items-center gap-1 text-[11px] text-[#8e9193]">
      <Clock size={11} />
      {formatAge(ts)}
    </span>
  );
};
