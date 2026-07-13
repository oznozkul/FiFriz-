import { Link } from "react-router-dom";
import type { DashboardRecentActivity } from "../../services/dashboardService";
import { formatAuditAction, formatAuditDescription } from "../../services/auditService";
import { DashboardPanelCard } from "./DashboardPanelCard";

type RecentSharesPanelProps = {
  activities: DashboardRecentActivity[];
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
};

const SHARE_ACTION_PREFIXES = ["share_", "document_internal_shared", "shared_document"];

function isShareActivity(action: string) {
  const normalized = action.toLowerCase();
  return SHARE_ACTION_PREFIXES.some((prefix) => normalized.startsWith(prefix) || normalized.includes("share"));
}

export function RecentSharesPanel({ activities, isLoading = false, error = null, onRetry }: RecentSharesPanelProps) {
  const shareActivities = activities.filter((item) => isShareActivity(item.action));

  return (
    <DashboardPanelCard
      title="Son paylaşımlar"
      subtitle="Paylaşım ve erişim hareketleri"
      viewAllHref="/activity"
      isLoading={isLoading}
      error={error}
      onRetry={onRetry}
      isEmpty={!isLoading && !error && shareActivities.length === 0}
      emptyTitle="Henüz paylaşım kaydı yok"
      emptyDescription="Belge paylaşımları ve link erişimleri burada listelenecek."
    >
      <div className="divide-y divide-line">
        {shareActivities.map((activity) => (
          <div key={activity.id} className="flex items-center gap-4 px-5 py-4">
            <div className="min-w-0 flex-1">
              <p className="truncate font-semibold text-ink">{formatAuditAction(activity.action)}</p>
              <p className="truncate text-sm text-slate-500">{formatAuditDescription(activity)}</p>
            </div>
            <span className="whitespace-nowrap text-xs text-slate-500">{formatDateTime(activity.createdAt)}</span>
          </div>
        ))}
      </div>
      {shareActivities.length > 0 ? (
        <Link className="inline-flex px-5 py-4 text-sm font-semibold text-violet-brand transition hover:text-violet-hover" to="/shared-with-me">
          Paylaşılan belgeler →
        </Link>
      ) : null}
    </DashboardPanelCard>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
