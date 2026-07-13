import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { EmptyState, LoadingState } from "../common";
import { formatAuditAction, listAuditLogs, type AuditLog } from "../../services/auditService";
import { formatDateTime } from "./documentWorkspaceUtils";

type DocumentWorkspaceActivitySectionProps = {
  documentId: string;
};

export function DocumentWorkspaceActivitySection({ documentId }: DocumentWorkspaceActivitySectionProps) {
  const [items, setItems] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      const response = await listAuditLogs({ entityType: "document", pageSize: 50 });
      setItems(response.items.filter((item) => item.entityId === documentId).slice(0, 8));
    } catch (error) {
      setItems([]);
      setLoadError(error instanceof Error ? error.message : "Aktivite yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void load();
  }, [load]);

  if (isLoading) {
    return <LoadingState label="Aktivite yükleniyor..." className="py-6" />;
  }

  if (loadError) {
    return (
      <EmptyState
        title="Aktivite yüklenemedi"
        description={loadError}
        action={
          <button type="button" className="text-sm font-semibold text-violet-brand" onClick={() => void load()}>
            Tekrar dene
          </button>
        }
      />
    );
  }

  if (items.length === 0) {
    return <p className="text-sm text-slate-500">Bu belge için kayıtlı aktivite bulunamadı.</p>;
  }

  return (
    <div className="space-y-3">
      <ul className="space-y-2">
        {items.map((item) => (
          <li key={item.id} className="rounded-lg border border-line bg-slate-50 px-3 py-2">
            <p className="text-sm font-semibold text-ink">{formatAuditAction(item.action)}</p>
            <p className="mt-0.5 text-xs text-slate-500">{item.userFullName || item.userEmail}</p>
            <p className="mt-1 text-xs text-slate-400">{formatDateTime(item.createdAt)}</p>
          </li>
        ))}
      </ul>
      <Link className="inline-flex text-xs font-semibold text-violet-brand hover:underline" to="/activity">
        Tüm aktiviteyi gör
      </Link>
    </div>
  );
}
