import { Share2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { EmptyState, LoadingState } from "../common";
import { listShareLinks, type DocumentShareLink } from "../../services/shareService";
import { formatDateTime } from "./documentWorkspaceUtils";

type DocumentWorkspaceSharesSectionProps = {
  documentId: string;
  onCreateShare: () => void;
};

export function DocumentWorkspaceSharesSection({ documentId, onCreateShare }: DocumentWorkspaceSharesSectionProps) {
  const [links, setLinks] = useState<DocumentShareLink[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      const items = await listShareLinks(documentId);
      setLinks(items.filter((link) => link.isActive).slice(0, 5));
    } catch (error) {
      setLinks([]);
      setLoadError(error instanceof Error ? error.message : "Paylaşımlar yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <div className="space-y-3">
      <button
        type="button"
        className="inline-flex h-9 w-full items-center justify-center gap-2 rounded-lg border border-line bg-white text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand"
        onClick={onCreateShare}
      >
        <Share2 className="h-4 w-4" />
        Paylaşım linki oluştur
      </button>
      {isLoading ? <LoadingState label="Paylaşımlar yükleniyor..." className="py-4" /> : null}
      {!isLoading && loadError ? (
        <EmptyState
          title="Paylaşımlar yüklenemedi"
          description={loadError}
          action={
            <button type="button" className="text-sm font-semibold text-violet-brand" onClick={() => void load()}>
              Tekrar dene
            </button>
          }
        />
      ) : null}
      {!isLoading && !loadError && links.length === 0 ? <p className="text-sm text-slate-500">Aktif paylaşım linki yok.</p> : null}
      {!isLoading && !loadError && links.length > 0 ? (
        <ul className="space-y-2">
          {links.map((link) => (
            <li key={link.id} className="rounded-lg border border-line bg-slate-50 px-3 py-2 text-xs">
              <p className="font-semibold text-ink">{link.permission === "download" ? "Görüntüle + İndir" : "Yalnızca görüntüle"}</p>
              <p className="mt-1 text-slate-500">{link.viewCount} görüntüleme</p>
              <p className="mt-1 text-slate-400">{formatDateTime(link.createdAt)}</p>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
