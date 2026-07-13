import { Link } from "react-router-dom";
import type { DashboardImportantDateItem } from "../../services/dashboardService";
import { useImportantDateLabels } from "../../hooks/useImportantDateLabels";
import { importantDateTypeLabel } from "../../types/importantDate";
import { DashboardPanelCard } from "./DashboardPanelCard";
import { ExternalLink } from "lucide-react";

type UpcomingImportantDatesPanelProps = {
  items: DashboardImportantDateItem[];
  upcomingCount: number;
  overdueCount: number;
  assignedCount: number;
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
};

export function UpcomingImportantDatesPanel({
  items,
  upcomingCount,
  overdueCount,
  assignedCount,
  isLoading = false,
  error = null,
  onRetry,
}: UpcomingImportantDatesPanelProps) {
  const labels = useImportantDateLabels();

  return (
    <DashboardPanelCard
      title={labels.dashboardUpcoming}
      subtitle={`${upcomingCount} yaklaşan · ${overdueCount} ${labels.dashboardOverdue.toLowerCase()} · ${assignedCount} ${labels.dashboardAssigned.toLowerCase()}`}
      viewAllHref="/important-dates"
      isLoading={isLoading}
      error={error}
      onRetry={onRetry}
      isEmpty={!isLoading && !error && items.length === 0}
      emptyTitle="Yaklaşan tarih yok"
      emptyDescription="Önemli tarih eklediğinizde burada görünecek."
    >
      <div className="divide-y divide-line">
        {items.map((item) => (
          <div key={item.id} className="flex items-center justify-between gap-3 px-5 py-3">
            <div className="min-w-0">
              <p className="truncate font-semibold text-ink">{item.title}</p>
              <p className="mt-0.5 truncate text-xs text-slate-500">
                {formatDisplayDate(item.date)} · {importantDateTypeLabel(item.type as never)} · {item.documentName}
              </p>
            </div>
            <Link
              className="inline-flex h-8 shrink-0 items-center gap-1 rounded-lg px-2 text-xs font-semibold text-violet-brand transition hover:bg-violet-soft/40"
              to={`/documents/${item.documentId}`}
            >
              <ExternalLink className="h-3.5 w-3.5" />
              {labels.goToDocument}
            </Link>
          </div>
        ))}
      </div>
    </DashboardPanelCard>
  );
}

function formatDisplayDate(value: string) {
  const parsed = new Date(`${value}T00:00:00`);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleDateString("tr-TR", { day: "numeric", month: "short" });
}
