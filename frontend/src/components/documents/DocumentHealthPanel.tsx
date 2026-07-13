import { useEffect, useState } from "react";
import { EmptyState, LoadingState, StatusBadge } from "../common";
import { getDocumentHealth } from "../../services/documentHealthService";
import type { DocumentHealth } from "../../types/document";

type DocumentHealthPanelProps = {
  documentId: string;
};

const HEALTH_ITEMS: Array<{ key: keyof DocumentHealth; label: string }> = [
  { key: "previewGenerated", label: "Önizleme" },
  { key: "ocrGenerated", label: "OCR" },
  { key: "embeddingGenerated", label: "Embedding" },
  { key: "summaryGenerated", label: "Özet" },
  { key: "metadataGenerated", label: "Metadata" },
  { key: "aiCompleted", label: "AI tamamlandı" },
];

export function DocumentHealthPanel({ documentId }: DocumentHealthPanelProps) {
  const [health, setHealth] = useState<DocumentHealth | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    setIsLoading(true);
    setLoadError(null);
    void getDocumentHealth(documentId)
      .then(setHealth)
      .catch((error) => setLoadError(error instanceof Error ? error.message : "Belge sağlık durumu yüklenemedi."))
      .finally(() => setIsLoading(false));
  }, [documentId]);

  if (isLoading) return <LoadingState variant="inline" label="Sağlık durumu yükleniyor..." />;
  if (loadError) {
    return <EmptyState title="Sağlık durumu yüklenemedi" description={loadError} />;
  }
  if (!health) return null;

  return (
    <div className="space-y-4 text-sm">
      <div className="flex flex-wrap items-center gap-2">
        <span className="font-semibold text-slate-600">İşlem durumu</span>
        <StatusBadge status={health.processingStatus} />
      </div>
      <div className="grid gap-2 sm:grid-cols-2">
        {HEALTH_ITEMS.map((item) => (
          <div key={item.key} className="flex items-center justify-between rounded-lg border border-line px-3 py-2">
            <span className="text-slate-600">{item.label}</span>
            <span className={`font-semibold ${health[item.key] ? "text-emerald-600" : "text-slate-400"}`}>
              {typeof health[item.key] === "boolean" && health[item.key] ? "Evet" : typeof health[item.key] === "boolean" ? "Hayır" : String(health[item.key])}
            </span>
          </div>
        ))}
      </div>
      <dl className="space-y-2 text-sm">
        <div className="flex justify-between gap-3">
          <dt className="text-slate-500">Versiyon sayısı</dt>
          <dd className="font-semibold text-ink">{health.versionCount}</dd>
        </div>
        <div className="flex justify-between gap-3">
          <dt className="text-slate-500">Depolama</dt>
          <dd className="font-semibold text-ink">{health.storageStatus}</dd>
        </div>
      </dl>
    </div>
  );
}
