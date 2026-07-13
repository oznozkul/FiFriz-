import { aiDocumentTypeLabel } from "../../types/documentClassification";
import type { ApiDocument } from "../../types/document";

type DocumentClassificationBadgeProps = {
  document: Pick<ApiDocument, "aiDocumentType">;
  compact?: boolean;
};

export function DocumentClassificationBadge({ document, compact = false }: DocumentClassificationBadgeProps) {
  const label = aiDocumentTypeLabel(document.aiDocumentType);
  if (!label) {
    return null;
  }

  return (
    <span
      className={`inline-flex items-center rounded-full border border-violet-200 bg-violet-50 font-semibold text-violet-700 ${
        compact ? "px-2 py-0.5 text-[11px]" : "px-2.5 py-1 text-xs"
      }`}
    >
      {label}
    </span>
  );
}
