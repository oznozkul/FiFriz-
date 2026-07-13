import type { DashboardDocumentTypeStat, DashboardTopFolder } from "../../services/dashboardService";
import { EmptyState } from "../common";

const extensionLabels: Record<string, string> = {
  pdf: "PDF",
  docx: "DOCX",
  xlsx: "XLSX",
  png: "PNG/JPG",
  jpg: "PNG/JPG",
  jpeg: "PNG/JPG",
};

type DocumentInsightsPanelProps = {
  documentTypeStats: DashboardDocumentTypeStat[];
  topFolders: DashboardTopFolder[];
};

export function DocumentInsightsPanel({ documentTypeStats, topFolders }: DocumentInsightsPanelProps) {
  const groupedStats = groupDocumentTypeStats(documentTypeStats);
  const hasStats = groupedStats.length > 0;
  const hasFolders = topFolders.length > 0;

  if (!hasStats && !hasFolders) {
    return (
      <section className="panel p-5">
        <EmptyState title="Henüz istatistik yok." description="Belge yükledikçe dosya türü ve klasör dağılımı burada görünecek." />
      </section>
    );
  }

  return (
    <div className="grid gap-5 xl:grid-cols-2">
      <section className="panel overflow-hidden">
        <div className="px-5 py-4">
          <h2 className="font-bold text-ink">Dosya Türü İstatistikleri</h2>
        </div>
        {hasStats ? (
          <div className="divide-y divide-line">
            {groupedStats.map((stat) => (
              <div key={stat.label} className="flex items-center justify-between px-5 py-4">
                <span className="font-semibold text-ink">{stat.label}</span>
                <span className="rounded-full bg-violet-soft px-3 py-1 text-sm font-bold text-violet-brand">{stat.count}</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="px-5 pb-5">
            <EmptyState title="Dosya türü verisi yok." />
          </div>
        )}
      </section>

      <section className="panel overflow-hidden">
        <div className="px-5 py-4">
          <h2 className="font-bold text-ink">En Çok Belge İçeren Klasörler</h2>
        </div>
        {hasFolders ? (
          <div className="divide-y divide-line">
            {topFolders.map((folder) => (
              <div key={folder.folderId} className="flex items-center justify-between px-5 py-4">
                <span className="truncate font-semibold text-ink">{folder.folderName}</span>
                <span className="rounded-full bg-slate-100 px-3 py-1 text-sm font-bold text-slate-600">{folder.documentCount}</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="px-5 pb-5">
            <EmptyState title="Klasör verisi yok." />
          </div>
        )}
      </section>
    </div>
  );
}

function groupDocumentTypeStats(stats: DashboardDocumentTypeStat[]) {
  const grouped = new Map<string, number>();

  for (const stat of stats) {
    const extension = stat.extension.toLowerCase();
    const label = extensionLabels[extension] ?? extension.toUpperCase();
    grouped.set(label, (grouped.get(label) ?? 0) + stat.count);
  }

  const preferredOrder = ["PDF", "DOCX", "XLSX", "PNG/JPG"];
  const entries = Array.from(grouped.entries()).map(([label, count]) => ({ label, count }));

  return entries.sort((left, right) => {
    const leftIndex = preferredOrder.indexOf(left.label);
    const rightIndex = preferredOrder.indexOf(right.label);
    if (leftIndex !== -1 || rightIndex !== -1) {
      if (leftIndex === -1) return 1;
      if (rightIndex === -1) return -1;
      return leftIndex - rightIndex;
    }
    return right.count - left.count || left.label.localeCompare(right.label, "tr");
  });
}
