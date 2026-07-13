import { Link } from "react-router-dom";
import { useAuth } from "../../services/AuthContext";
import type { DashboardRecentActivity } from "../../services/dashboardService";
import { formatAuditAction, formatAuditDescription } from "../../services/auditService";
import { DashboardPanelCard } from "./DashboardPanelCard";

type RecentActivitiesPanelProps = {
  activities: DashboardRecentActivity[];
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
};

export function RecentActivitiesPanel({ activities, isLoading = false, error = null, onRetry }: RecentActivitiesPanelProps) {
  const { activeTenant } = useAuth();
  const canViewAllActivities = activeTenant?.role === "owner" || activeTenant?.role === "admin";

  return (
    <DashboardPanelCard
      title="Son aktiviteler"
      subtitle="Çalışma alanındaki son hareketler"
      viewAllHref={canViewAllActivities ? "/activity" : undefined}
      isLoading={isLoading}
      error={error}
      onRetry={onRetry}
      isEmpty={!isLoading && !error && activities.length === 0}
      emptyTitle="Henüz aktivite kaydı yok"
      emptyDescription="Önemli işlemler burada listelenecek."
    >
      <div className="divide-y divide-line">
        {activities.map((activity) => (
          <div key={activity.id} className="flex w-full items-center gap-4 px-5 py-4 text-left">
            <div className="min-w-0 flex-1">
              <p className="truncate font-semibold text-ink">{formatAuditAction(activity.action)}</p>
              <p className="truncate text-sm text-slate-500">{formatAuditDescription(activity)}</p>
            </div>
            <span className="whitespace-nowrap text-xs text-slate-500">{formatDateTime(activity.createdAt)}</span>
          </div>
        ))}
      </div>
      {canViewAllActivities && activities.length > 0 ? (
        <Link className="inline-flex px-5 py-4 text-sm font-semibold text-violet-brand transition hover:text-violet-hover" to="/activity">
          Tümünü Gör →
        </Link>
      ) : null}
    </DashboardPanelCard>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
