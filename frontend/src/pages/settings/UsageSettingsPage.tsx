import { useEffect, useState } from "react";
import { LoadingState } from "../../components/common";
import { formatStorageSize, getDashboardSummary } from "../../services/dashboardService";
import { toast } from "../../services/toast";
import { InfoItem, SettingsSection } from "./shared";

export function UsageSettingsPage() {
  const [totalDocuments, setTotalDocuments] = useState<number | null>(null);
  const [totalStorageBytes, setTotalStorageBytes] = useState<number | null>(null);
  const [uploadedThisMonth, setUploadedThisMonth] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    getDashboardSummary()
      .then((summary) => {
        if (cancelled) return;
        setTotalDocuments(summary.totalDocuments);
        setTotalStorageBytes(summary.totalStorageBytes);
        setUploadedThisMonth(summary.documentsUploadedThisMonth);
      })
      .catch((error) => {
        if (!cancelled) toast.error(error instanceof Error ? error.message : "Kullanım bilgisi yüklenemedi.");
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <SettingsSection title="Kullanım" description="Workspace kullanım özetiniz ve limitler.">
      {isLoading ? (
        <LoadingState variant="inline" label="Kullanım bilgisi yükleniyor..." />
      ) : (
        <dl className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <InfoItem label="Toplam belge" value={totalDocuments !== null ? String(totalDocuments) : "—"} />
          <InfoItem label="Kullanılan depolama" value={totalStorageBytes !== null ? formatStorageSize(totalStorageBytes) : "—"} />
          <InfoItem label="Bu ay yüklenen" value={uploadedThisMonth !== null ? String(uploadedThisMonth) : "—"} />
        </dl>
      )}
    </SettingsSection>
  );
}
