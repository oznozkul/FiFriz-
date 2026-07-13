import { FileArchive, FileText, Folder, Star, Trash2 } from "lucide-react";
import type { DashboardSummary } from "../../services/dashboardService";
import { formatStorageSize } from "../../services/dashboardService";
import { StatCardSkeleton } from "../common";
import { StatCard } from "./StatCard";

type StatsGridProps = {
  summary: DashboardSummary | null;
  isLoading?: boolean;
};

export function StatsGrid({ summary, isLoading = false }: StatsGridProps) {
  if (isLoading || !summary) {
    return (
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        {[1, 2, 3, 4, 5].map((item) => (
          <StatCardSkeleton key={item} />
        ))}
      </div>
    );
  }

  const stats = [
    {
      label: "Toplam Belge",
      value: String(summary.totalDocuments),
      change: summary.documentsUploadedThisMonth > 0 ? `+${summary.documentsUploadedThisMonth} bu ay` : undefined,
      icon: FileText,
    },
    {
      label: "Klasör",
      value: String(summary.totalFolders),
      icon: Folder,
    },
    {
      label: "Favoriler",
      value: String(summary.favoriteDocuments),
      icon: Star,
    },
    {
      label: "Çöp Kutusu",
      value: String(summary.trashedDocuments),
      icon: Trash2,
    },
    {
      label: "Toplam Storage",
      value: formatStorageSize(summary.totalStorageBytes),
      icon: FileArchive,
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
      {stats.map((stat) => (
        <StatCard key={stat.label} {...stat} />
      ))}
    </div>
  );
}
