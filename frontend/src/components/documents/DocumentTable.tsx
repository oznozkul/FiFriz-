import { useNavigate } from "react-router-dom";
import type { ApiDocument } from "../../types/document";
import { buildDocumentMenuItems, getDocumentMenuContext } from "../../contextMenu";
import { useContextMenu } from "../../hooks/useContextMenu";
import { useAuth } from "../../services/AuthContext";
import { DocumentListPanel, EmptyState, ErrorState, TablePagination, TableSkeleton } from "../common";
import { DocumentRow, DOCUMENT_TABLE_GRID_CLASS } from "./DocumentRow";
import { useFolderLabels } from "../../hooks/useFolderLabels";

const GRID_CLASS = DOCUMENT_TABLE_GRID_CLASS;

type DocumentTableProps = {
  documents: ApiDocument[];
  isLoading?: boolean;
  onEdit: (document: ApiDocument) => void;
  onDelete: (document: ApiDocument) => void;
  onRestore?: (document: ApiDocument) => void;
  onToggleFavorite?: (document: ApiDocument) => void;
  favoriteLoadingId?: string | null;
  selectedIds: string[];
  onSelect: (documentId: string, selected: boolean) => void;
  onSelectAll: (selected: boolean) => void;
  isTrash?: boolean;
  navigationQuery?: string;
  page?: number;
  pageSize?: number;
  totalCount?: number;
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  error?: string | null;
  onRetry?: () => void;
  emptyTitle?: string;
  emptyDescription?: string;
};

export function DocumentTable({
  documents,
  isLoading = false,
  onEdit,
  onDelete,
  onRestore,
  onToggleFavorite,
  favoriteLoadingId = null,
  selectedIds,
  onSelect,
  onSelectAll,
  isTrash = false,
  navigationQuery = "",
  page = 1,
  pageSize = 20,
  totalCount,
  onPageChange,
  onPageSizeChange,
  error = null,
  onRetry,
  emptyTitle = "Belge bulunamadı",
  emptyDescription = "Belgeler yüklendiğinde veya filtre değiştiğinde burada listelenecek.",
}: DocumentTableProps) {
  const navigate = useNavigate();
  const { activeTenant } = useAuth();
  const folderLabels = useFolderLabels();
  const role = activeTenant?.role ?? "member";
  const { menu, bindContextMenu } = useContextMenu();
  const allSelected = documents.length > 0 && documents.every((document) => selectedIds.includes(document.id));
  const getDocumentPath = (documentId: string) => {
    const trimmedQuery = navigationQuery.trim();
    return trimmedQuery.length >= 2 ? `/documents/${documentId}?q=${encodeURIComponent(trimmedQuery)}` : `/documents/${documentId}`;
  };

  if (isLoading) {
    return (
      <DocumentListPanel>
        <TableSkeleton rows={5} />
      </DocumentListPanel>
    );
  }

  if (error) {
    return <ErrorState title="Belgeler yüklenemedi" description={error} onRetry={onRetry} />;
  }

  if (documents.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  return (
    <>
      <DocumentListPanel
        footer={
          onPageChange && onPageSizeChange ? (
            <TablePagination
              page={page}
              pageSize={pageSize}
              totalCount={totalCount ?? documents.length}
              itemLabel="belge"
              onPageChange={onPageChange}
              onPageSizeChange={onPageSizeChange}
            />
          ) : undefined
        }
      >
        <div className={`${GRID_CLASS} hidden border-b border-line px-5 py-3 text-xs font-semibold uppercase tracking-wide text-slate-400 md:grid`}>
          <span>
            <input
              className="h-4 w-4 rounded border-line accent-violet-brand focus:ring-violet-ring"
              type="checkbox"
              aria-label="Tümünü seç"
              checked={allSelected}
              onChange={(event) => onSelectAll(event.target.checked)}
            />
          </span>
          <span>Ad</span>
          <span>{folderLabels.singular}</span>
          <span>Durum</span>
          <span>Boyut</span>
          <span>Yüklenme Tarihi</span>
          <span />
        </div>
        <div className="divide-y divide-line">
          {documents.map((doc) => (
            <DocumentRow
              key={doc.id}
              document={doc}
              onOpen={(document) => navigate(getDocumentPath(document.id))}
              onEdit={onEdit}
              onDelete={onDelete}
              onRestore={onRestore}
              onToggleFavorite={onToggleFavorite}
              favoriteLoadingId={favoriteLoadingId}
              isTrash={isTrash}
              isSelected={selectedIds.includes(doc.id)}
              onSelect={onSelect}
              contextMenuProps={bindContextMenu({
                items: buildDocumentMenuItems({ document: doc, role, isTrash }),
                context: getDocumentMenuContext(doc),
              })}
            />
          ))}
        </div>
      </DocumentListPanel>
      {menu}
    </>
  );
}
