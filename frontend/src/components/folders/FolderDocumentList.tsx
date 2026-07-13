import { useNavigate } from "react-router-dom";
import type { ApiDocument } from "../../types/document";
import { EmptyState, FileTypeIcon, StatusBadge, TableRowSkeleton } from "../common";
import { FavoriteStarButton } from "../documents/FavoriteStarButton";
import { formatDateTime, formatFileSize } from "../documentWorkspace/documentWorkspaceUtils";

type FolderDocumentListProps = {
  documents: ApiDocument[];
  isLoading: boolean;
  favoriteLoadingId?: string | null;
  onToggleFavorite?: (document: ApiDocument) => void;
};

export function FolderDocumentList({ documents, isLoading, favoriteLoadingId = null, onToggleFavorite }: FolderDocumentListProps) {
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="divide-y divide-line">
        {[1, 2, 3, 4].map((item) => (
          <TableRowSkeleton key={item} />
        ))}
      </div>
    );
  }

  if (documents.length === 0) {
    return <EmptyState title="Bu klasörde belge yok" description="Belge yüklerken bu klasörü seçersen burada görünecek." />;
  }

  return (
    <div className="divide-y divide-line">
      {documents.map((document) => (
        <div key={document.id} className="flex w-full items-center gap-3 px-5 py-4 transition hover:bg-violet-soft/35">
          {onToggleFavorite ? (
            <FavoriteStarButton isFavorite={document.isFavorite ?? false} isLoading={favoriteLoadingId === document.id} onToggle={() => onToggleFavorite(document)} />
          ) : null}
          <button className="flex min-w-0 flex-1 items-center gap-3 text-left" type="button" onClick={() => navigate(`/documents/${document.id}`)}>
            <FileTypeIcon extension={document.file_extension} />
            <div className="min-w-0 flex-1">
              <p className="truncate font-semibold text-ink">{document.original_file_name}</p>
              <p className="mt-1 text-xs font-semibold text-slate-500">
                {formatFileSize(document.file_size)} · {formatDateTime(document.created_at)}
              </p>
            </div>
            <StatusBadge status={document.status} />
          </button>
        </div>
      ))}
    </div>
  );
}
