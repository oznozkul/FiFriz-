import { HYBRID_MATCH_TYPE_LABELS, type HybridSearchMatchType } from "../../types/hybridSearch";

type SearchMatchTypeBadgeProps = {
  matchType: HybridSearchMatchType;
};

export function SearchMatchTypeBadge({ matchType }: SearchMatchTypeBadgeProps) {
  const label = HYBRID_MATCH_TYPE_LABELS[matchType] ?? matchType;
  return (
    <span className="inline-flex items-center rounded-full border border-violet-200 bg-violet-50 px-2 py-0.5 text-[11px] font-semibold text-violet-700">
      {label}
    </span>
  );
}
