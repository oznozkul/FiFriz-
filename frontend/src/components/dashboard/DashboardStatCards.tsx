import { CalendarDays, ClipboardCheck, FileText, Star } from "lucide-react";
import type { DashboardSummary } from "../../services/dashboardService";
import { StatCardSkeleton } from "../common";
import { StatCard } from "./StatCard";

type DashboardStatCardsProps = {
  isPersonal: boolean;
  summary: DashboardSummary | null;
  pendingApprovalsCount?: number;
  isLoading?: boolean;
  upcomingLabel: string;
};

export function DashboardStatCards({
  isPersonal,
  summary,
  pendingApprovalsCount = 0,
  isLoading = false,
  upcomingLabel,
}: DashboardStatCardsProps) {
  if (isLoading || !summary) {
    return (
      <div className={`grid gap-4 ${isPersonal ? "md:grid-cols-3" : "md:grid-cols-3"}`}>
        {[1, 2, 3].map((item) => (
          <StatCardSkeleton key={item} />
        ))}
      </div>
    );
  }

  if (isPersonal) {
    const stats = [
      {
        label: "Toplam belge",
        value: String(summary.totalDocuments),
        change: summary.documentsUploadedThisMonth > 0 ? `+${summary.documentsUploadedThisMonth} bu ay` : undefined,
        icon: FileText,
      },
      {
        label: upcomingLabel,
        value: String(summary.upcomingImportantDates),
        change: summary.overdueImportantDates > 0 ? `${summary.overdueImportantDates} geciken` : undefined,
        icon: CalendarDays,
      },
      {
        label: "Favoriler",
        value: String(summary.favoriteDocuments),
        icon: Star,
      },
    ];
    return (
      <div className="grid gap-4 md:grid-cols-3">
        {stats.map((stat) => (
          <StatCard key={stat.label} {...stat} />
        ))}
      </div>
    );
  }

  const stats = [
    {
      label: "Toplam belge",
      value: String(summary.totalDocuments),
      change: summary.documentsUploadedThisMonth > 0 ? `+${summary.documentsUploadedThisMonth} bu ay` : undefined,
      icon: FileText,
    },
    {
      label: "Bekleyen onaylar",
      value: String(pendingApprovalsCount),
      icon: ClipboardCheck,
    },
    {
      label: upcomingLabel,
      value: String(summary.upcomingImportantDates),
      change: summary.overdueImportantDates > 0 ? `${summary.overdueImportantDates} geciken` : undefined,
      icon: CalendarDays,
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-3">
      {stats.map((stat) => (
        <StatCard key={stat.label} {...stat} />
      ))}
    </div>
  );
}
