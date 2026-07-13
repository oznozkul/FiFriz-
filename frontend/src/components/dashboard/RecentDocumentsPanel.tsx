import { Link } from "react-router-dom";
import type { DashboardRecentDocument } from "../../services/dashboardService";
import { formatStorageSize } from "../../services/dashboardService";
import { FileTypeIcon } from "../common";
import { FavoriteStarButton } from "../documents/FavoriteStarButton";
import { formatDateTime } from "../documents/documentPresentation";
import { DashboardPanelCard } from "./DashboardPanelCard";

type RecentDocumentsPanelProps = {
  documents: DashboardRecentDocument[];
  favoriteLoadingId?: string | null;
  onToggleFavorite?: (document: DashboardRecentDocument) => void;
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
  onUpload?: () => void;
};

export function RecentDocumentsPanel({
  documents,
  favoriteLoadingId = null,
  onToggleFavorite,
  isLoading = false,
  error = null,
  onRetry,
  onUpload,
}: RecentDocumentsPanelProps) {
  return (
    <DashboardPanelCard
      title="Son yüklenen belgeler"
      subtitle="En son eklenen dosyalarınız"
      viewAllHref="/documents"
      isLoading={isLoading}
      error={error}
      onRetry={onRetry}
      isEmpty={!isLoading && !error && documents.length === 0}
      emptyTitle="Henüz belge yok"
      emptyDescription="İlk belgenizi yüklediğinizde burada görünecek."
      emptyAction={
        onUpload ? (
          <button type="button" className="text-sm font-semibold text-violet-brand hover:text-violet-hover" onClick={onUpload}>
            Belge yükle
          </button>
        ) : undefined
      }
    >
      <div className="divide-y divide-line">
        {documents.map((doc) => (
          <div key={doc.id} className="flex w-full items-center gap-3 px-5 py-3 transition hover:bg-violet-soft/35">
            {onToggleFavorite ? (
              <FavoriteStarButton isFavorite={doc.isFavorite} isLoading={favoriteLoadingId === doc.id} onToggle={() => onToggleFavorite(doc)} />
            ) : null}
            <Link to={`/documents/${doc.id}`} className="flex min-w-0 flex-1 items-center gap-3 focus:outline-none focus:ring-4 focus:ring-inset focus:ring-violet-ring/70">
              <FileTypeIcon extension={doc.fileExtension} />
              <span className="min-w-0 flex-1 truncate font-semibold text-ink">{doc.originalFileName}</span>
              <span className="hidden text-xs text-slate-500 sm:block">{formatDateTime(doc.createdAt)}</span>
              <span className="w-16 text-right text-xs text-slate-500">{formatStorageSize(doc.fileSize)}</span>
            </Link>
          </div>
        ))}
      </div>
    </DashboardPanelCard>
  );
}
