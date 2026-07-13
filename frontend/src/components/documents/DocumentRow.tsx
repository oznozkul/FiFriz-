import { Eye, Pencil, RotateCcw, Trash2 } from "lucide-react";
import type { KeyboardEvent } from "react";
import type { ApiDocument } from "../../types/document";
import type { ContextMenuTriggerBindings } from "../../hooks/useContextMenu";
import { useFolderLabels } from "../../hooks/useFolderLabels";
import { FileTypeIcon, IconButton, StatusBadge } from "../common";
import { formatDateTime } from "../documentWorkspace/documentWorkspaceUtils";
import { TagBadgeList } from "../tags";
import { FavoriteStarButton } from "./FavoriteStarButton";
import { DocumentSourceBadge } from "./DocumentSourceBadge";
import { DocumentClassificationBadge } from "./DocumentClassificationBadge";

type DocumentRowProps = {
  document: ApiDocument;
  onOpen: (document: ApiDocument) => void;
  onEdit: (document: ApiDocument) => void;
  onDelete: (document: ApiDocument) => void;
  onRestore?: (document: ApiDocument) => void;
  onToggleFavorite?: (document: ApiDocument) => void;
  favoriteLoadingId?: string | null;
  isSelected: boolean;
  isTrash?: boolean;
  onSelect: (documentId: string, selected: boolean) => void;
  contextMenuProps?: ContextMenuTriggerBindings;
};

const GRID_CLASS = "grid grid-cols-[36px_minmax(240px,1fr)_150px_140px_110px_170px_120px]";

export function DocumentRow({
  document,
  onOpen,
  onEdit,
  onDelete,
  onRestore,
  onToggleFavorite,
  favoriteLoadingId = null,
  isSelected,
  isTrash = false,
  onSelect,
  contextMenuProps,
}: DocumentRowProps) {
  const folderLabels = useFolderLabels();

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    contextMenuProps?.onKeyDown?.(event);
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      onOpen(document);
    }
  };

  const actionButtons = (
    <span className="flex justify-end gap-1" onClick={(event) => event.stopPropagation()}>
      <IconButton className="h-8 w-8 border-0" aria-label="Aç" onClick={() => onOpen(document)}>
        <Eye className="h-4 w-4" />
      </IconButton>
      {isTrash ? (
        <IconButton
          className="h-8 w-8 border-0 text-emerald-600 hover:border-emerald-200 hover:bg-emerald-50 hover:text-emerald-700"
          aria-label="Geri yükle"
          onClick={() => onRestore?.(document)}
        >
          <RotateCcw className="h-4 w-4" />
        </IconButton>
      ) : (
        <IconButton className="h-8 w-8 border-0" aria-label="Düzenle" onClick={() => onEdit(document)}>
          <Pencil className="h-4 w-4" />
        </IconButton>
      )}
      <IconButton
        className="h-8 w-8 border-0 text-red-500 hover:border-red-200 hover:bg-red-50 hover:text-red-600"
        aria-label="Sil"
        onClick={() => onDelete(document)}
      >
        <Trash2 className="h-4 w-4" />
      </IconButton>
    </span>
  );

  const nameBlock = (
    <div className="flex min-w-0 items-center gap-2">
      {!isTrash && onToggleFavorite ? (
        <span onClick={(event) => event.stopPropagation()}>
          <FavoriteStarButton
            isFavorite={document.isFavorite ?? false}
            isLoading={favoriteLoadingId === document.id}
            onToggle={() => onToggleFavorite(document)}
          />
        </span>
      ) : null}
      <FileTypeIcon extension={document.file_extension} />
      <div className="min-w-0">
        <span className="block truncate font-semibold text-ink">{document.original_file_name}</span>
        {document.aiDocumentType ? (
          <div className="mt-1">
            <DocumentClassificationBadge document={document} compact />
          </div>
        ) : null}
        {document.sourceType && document.sourceType !== "manual_upload" ? (
          <div className="mt-1">
            <DocumentSourceBadge document={document} compact />
          </div>
        ) : null}
        {!isTrash && document.tags && document.tags.length > 0 ? (
          <div className="mt-1">
            <TagBadgeList tags={document.tags} />
          </div>
        ) : null}
      </div>
    </div>
  );

  const contextHandlers = {
    onContextMenu: contextMenuProps?.onContextMenu,
    onTouchStart: contextMenuProps?.onTouchStart,
    onTouchMove: contextMenuProps?.onTouchMove,
    onTouchEnd: contextMenuProps?.onTouchEnd,
    onTouchCancel: contextMenuProps?.onTouchCancel,
  };

  return (
    <>
      <div
        className={`${GRID_CLASS} hidden w-full cursor-pointer items-center px-5 py-4 text-left transition hover:bg-slate-50/80 focus-within:ring-4 focus-within:ring-inset focus-within:ring-violet-ring/70 md:grid`}
        role="row"
        tabIndex={0}
        onClick={() => onOpen(document)}
        onKeyDown={handleKeyDown}
        {...contextHandlers}
      >
        <span onClick={(event) => event.stopPropagation()}>
          <input
            className="h-4 w-4 rounded border-line accent-violet-brand focus:ring-violet-ring"
            type="checkbox"
            aria-label={`${document.original_file_name} seç`}
            checked={isSelected}
            onChange={(event) => onSelect(document.id, event.target.checked)}
          />
        </span>
        {nameBlock}
        <span className="truncate text-slate-500">{document.folder_name || folderLabels.noneLabel}</span>
        <StatusBadge status={document.lifecycleStatus ?? document.status} />
        <span className="text-slate-500">{formatFileSize(document.file_size)}</span>
        <span className="text-slate-500">{formatDateTime(document.created_at)}</span>
        {actionButtons}
      </div>

      <div
        className="cursor-pointer bg-white px-4 py-4 transition hover:bg-violet-soft/30 focus-within:ring-4 focus-within:ring-inset focus-within:ring-violet-ring/70 md:hidden"
        role="row"
        tabIndex={0}
        onClick={() => onOpen(document)}
        onKeyDown={handleKeyDown}
        {...contextHandlers}
      >
        <div className="flex items-start gap-3">
          <span onClick={(event) => event.stopPropagation()}>
            <input
              className="mt-1 h-4 w-4 rounded border-line accent-violet-brand focus:ring-violet-ring"
              type="checkbox"
              aria-label={`${document.original_file_name} seç`}
              checked={isSelected}
              onChange={(event) => onSelect(document.id, event.target.checked)}
            />
          </span>
          <div className="min-w-0 flex-1">{nameBlock}</div>
          {actionButtons}
        </div>
        <div className="mt-3 grid gap-2 border-t border-line pt-3 text-sm">
          <div className="flex items-center justify-between gap-3">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">{folderLabels.singular}</span>
            <span className="truncate text-slate-600">{document.folder_name || folderLabels.noneLabel}</span>
          </div>
          <div className="flex items-center justify-between gap-3">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Durum</span>
            <StatusBadge status={document.lifecycleStatus ?? document.status} />
          </div>
          <div className="flex items-center justify-between gap-3">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Boyut</span>
            <span className="text-slate-600">{formatFileSize(document.file_size)}</span>
          </div>
          <div className="flex items-center justify-between gap-3">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Yüklenme</span>
            <span className="text-slate-600">{formatDateTime(document.created_at)}</span>
          </div>
        </div>
      </div>
    </>
  );
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(bytes / 1024, 1).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export { GRID_CLASS as DOCUMENT_TABLE_GRID_CLASS };
