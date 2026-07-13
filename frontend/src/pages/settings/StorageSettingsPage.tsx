import { useCallback, useEffect, useState } from "react";
import { PageState } from "../../components/common";
import { formatStorageSize, getDashboardSummary } from "../../services/dashboardService";
import { InfoItem, SettingsSection } from "./shared";

export function StorageSettingsPage() {
  const [totalDocuments, setTotalDocuments] = useState<number | null>(null);
  const [totalStorageBytes, setTotalStorageBytes] = useState<number | null>(null);
  const [trashedDocuments, setTrashedDocuments] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const load = useCallback(() => {
    setIsLoading(true);
    setLoadError(null);
    getDashboardSummary()
      .then((summary) => {
        setTotalDocuments(summary.totalDocuments);
        setTotalStorageBytes(summary.totalStorageBytes);
        setTrashedDocuments(summary.trashedDocuments);
      })
      .catch((error) => {
        setLoadError(error instanceof Error ? error.message : "Depolama bilgisi yüklenemedi.");
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <SettingsSection title="Depolama" description="Workspace depolama kullanımı ve belge istatistikleri.">
      <PageState
        isLoading={isLoading}
        error={loadError}
        variant="inline"
        loadingLabel="Depolama bilgisi yükleniyor..."
        errorTitle="Depolama bilgisi yüklenemedi"
        onRetry={load}
        emptyTitle="Depolama bilgisi yok"
        isEmpty={false}
      >
        <div className="space-y-5">
          <dl className="grid gap-4 sm:grid-cols-2">
            <InfoItem
              label="Toplam kullanılan alan"
              value={totalStorageBytes !== null ? formatStorageSize(totalStorageBytes) : "—"}
            />
            <InfoItem label="Toplam belge" value={totalDocuments !== null ? String(totalDocuments) : "—"} />
            <InfoItem label="Çöp kutusundaki belgeler" value={trashedDocuments !== null ? String(trashedDocuments) : "—"} />
          </dl>
          <div className="rounded-lg border border-dashed border-line bg-slate-50/60 px-4 py-3 text-sm text-slate-500">
            Uygun dosya türleri arka planda sıkıştırılarak depolanır (txt, csv, json, xml, html, eml). PDF ve görseller
            varsayılan olarak sıkıştırılmaz.
          </div>
        </div>
      </PageState>
    </SettingsSection>
  );
}
