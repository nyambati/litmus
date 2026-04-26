import { useEffect, useState } from "react";
import { Layers, GitBranch, Radio, FlaskConical } from "lucide-react";
import { API } from "../../utils/ui";
import { Header } from "../layout/Header";
import { GfSpinner } from "../ui/Spinner";
import { EmptyState } from "../ui/EmptyState";
import { LabelChip } from "../ui/Chips";
import { ReceiverChip } from "../ui/Chips";
import {
  useFragmentsStore,
  type FragmentInfo,
} from "../../stores/useFragmentsStore";

export const FragmentsPage = () => {
  const { fragments, setFragments } = useFragmentsStore();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetch(`${API}/api/v1/fragments`)
      .then((r) => r.json())
      .then((data: FragmentInfo[]) => setFragments(data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [setFragments]);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#181b1f]">
      <Header title="Fragments" />
      <main className="flex-1 p-6 overflow-y-auto">
        {loading ? (
          <div className="flex items-center justify-center h-48">
            <GfSpinner size="lg" />
          </div>
        ) : fragments.length === 0 ? (
          <EmptyState
            icon={Layers}
            title="No fragments found"
            description="Add YAML fragment files to the configured fragments directory"
          />
        ) : (
          <div className="space-y-2 animate-fade-in-up">
            {fragments.map((frag) => (
              <FragmentCard key={frag.name} fragment={frag} />
            ))}
          </div>
        )}
      </main>
    </div>
  );
};

const FragmentCard = ({ fragment }: { fragment: FragmentInfo }) => {
  const isRoot = fragment.name === "root";

  return (
    <div className="bg-[#1f2128] border border-[#2c3235] rounded-sm p-4 hover:border-[#34383e] transition-colors">
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2.5">
          <div
            className={`w-7 h-7 rounded flex items-center justify-center shrink-0 ${
              isRoot ? "bg-[#f46800]/15" : "bg-[#5794f2]/15"
            }`}
          >
            <Layers
              size={13}
              className={isRoot ? "text-[#f46800]" : "text-[#5794f2]"}
            />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-[#d9d9d9] font-semibold text-sm">
              {fragment.name}
            </span>
            {fragment.namespace && (
              <span className="text-[11px] text-[#8e9193] font-mono bg-[#22252b] border border-[#2c3235] px-1.5 py-0.5 rounded-[2px]">
                ns:{fragment.namespace}
              </span>
            )}
          </div>
        </div>
        {isRoot && (
          <span className="text-[10px] font-semibold text-[#f46800] bg-[#f46800]/10 border border-[#f46800]/20 px-2 py-0.5 rounded-sm uppercase tracking-wide">
            Root
          </span>
        )}
      </div>

      {/* Group match labels */}
      {fragment.group && (
        <div className="mb-3 flex items-center gap-2 flex-wrap">
          <span className="text-[11px] text-[#8e9193] font-medium shrink-0">
            group
          </span>
          <div className="flex items-center gap-1.5 flex-wrap">
            {Object.entries(fragment.group.match).map(([k, v]) => (
              <LabelChip key={k} labelKey={k} value={v} />
            ))}
            {fragment.group.receiver && (
              <ReceiverChip name={fragment.group.receiver} variant="purple" />
            )}
          </div>
        </div>
      )}

      {/* Counts */}
      <div className="flex items-center gap-5 pt-2.5 border-t border-[#2c3235]">
        <CountItem
          icon={<GitBranch size={12} />}
          label="Routes"
          value={fragment.routes}
        />
        {!isRoot && (
          <CountItem
            icon={<Radio size={12} />}
            label="Receivers"
            value={fragment.receivers}
          />
        )}
        <CountItem
          icon={<FlaskConical size={12} />}
          label="Tests"
          value={fragment.tests}
          dim={fragment.tests === 0}
        />
      </div>
    </div>
  );
};

const CountItem = ({
  icon,
  label,
  value,
  dim = false,
}: {
  icon: React.ReactNode;
  label: string;
  value: number;
  dim?: boolean;
}) => (
  <div className="flex items-center gap-1.5">
    <span className="text-[#8e9193]">{icon}</span>
    <span
      className={`text-sm font-semibold tabular-nums ${
        dim ? "text-[#8e9193]" : "text-[#d9d9d9]"
      }`}
    >
      {value}
    </span>
    <span className="text-[11px] text-[#8e9193]">{label}</span>
  </div>
);
