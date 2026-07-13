import { Link } from "react-router-dom";
import type { PendingApproval } from "../../services/workflowService";
import { FileTypeIcon } from "../common";
import { formatDateTime } from "../documents/documentPresentation";
import { DashboardPanelCard } from "./DashboardPanelCard";

type PendingApprovalsPanelProps = {
  items: PendingApproval[];
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
};

export function PendingApprovalsPanel({ items, isLoading = false, error = null, onRetry }: PendingApprovalsPanelProps) {
  return (
    <DashboardPanelCard
      title="Bekleyen onaylar"
      subtitle="Onayınızı bekleyen belgeler"
      viewAllHref="/approvals"
      isLoading={isLoading}
      error={error}
      onRetry={onRetry}
      isEmpty={!isLoading && !error && items.length === 0}
      emptyTitle="Bekleyen onay yok"
      emptyDescription="Size atanmış onay adımı bulunmuyor."
    >
      <div className="divide-y divide-line">
        {items.slice(0, 5).map((item) => (
          <Link
            key={`${item.workflowId}-${item.stepId}`}
            to={`/documents/${item.documentId}?panel=workflow`}
            className="flex items-center gap-3 px-5 py-4 transition hover:bg-violet-soft/35"
          >
            <FileTypeIcon extension="pdf" />
            <div className="min-w-0 flex-1">
              <p className="truncate font-semibold text-ink">{item.documentName}</p>
              <p className="mt-0.5 truncate text-xs text-slate-500">
                {item.workflowTitle} · {item.createdByFullName}
              </p>
            </div>
            <span className="hidden shrink-0 text-xs text-slate-500 sm:block">{formatDateTime(item.createdAt)}</span>
          </Link>
        ))}
      </div>
    </DashboardPanelCard>
  );
}
