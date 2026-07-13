import { Folder as FolderIcon, FolderPlus, Grid3X3, List, RefreshCw } from "lucide-react";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../../services/AuthContext";
import {
  buildDocumentMenuItems,
  buildExplorerBackgroundMenuItems,
  buildFolderMenuItems,
  ContextMenuActions,
  getDocumentMenuContext,
  getExplorerBackgroundContext,
  getFolderMenuContext,
} from "../../contextMenu";
import { publishContextMenuAction } from "../../contextMenu/events";
import { useContextMenu } from "../../hooks/useContextMenu";
import type { FolderLabels } from "../../hooks/useFolderLabels";
import type { ApiDocument } from "../../types/document";
import type { Folder } from "../../types/folder";
import { Button, EmptyState, ErrorState, FileTypeIcon, IconButton, PrimaryButton, SearchInput, TableSkeleton } from "../common";
import { FavoriteStarButton } from "../documents/FavoriteStarButton";
import { formatDate, formatDateTime } from "../documents/documentPresentation";
import { FolderBreadcrumb } from "./FolderBreadcrumb";

export type ExplorerViewMode = "list" | "grid";
export type ExplorerContentFilter = "all" | "folders" | "documents";

type FolderExplorerProps = {
  folders: Folder[];
  labels: FolderLabels;
  currentFolder: Folder | null;
  currentPath: Folder[];
  childFolders: Folder[];
  documents: ApiDocument[];
  isLoading: boolean;
  loadError?: string | null;
  onRetry?: () => void;
  canManageFolders: boolean;
  favoriteLoadingId: string | null;
  viewMode: ExplorerViewMode;
  onViewModeChange: (mode: ExplorerViewMode) => void;
  onNavigate: (folder: Folder | null) => void;
  onCreateSubfolder: (parent: Folder | null) => void;
  onRefresh: () => void;
};

export function FolderExplorer({
  labels,
  currentFolder,
  currentPath,
  childFolders,
  documents,
  isLoading,
  loadError = null,
  onRetry,
  canManageFolders,
  favoriteLoadingId,
  viewMode,
  onViewModeChange,
  onNavigate,
  onCreateSubfolder,
  onRefresh,
}: FolderExplorerProps) {
  const navigate = useNavigate();
  const { activeTenant } = useAuth();
  const role = activeTenant?.role ?? "member";
  const { menu, bindContextMenu } = useContextMenu();
  const [searchQuery, setSearchQuery] = useState("");
  const [contentFilter, setContentFilter] = useState<ExplorerContentFilter>("all");

  const normalizedQuery = searchQuery.trim().toLowerCase();

  const filteredFolders = useMemo(() => {
    if (contentFilter === "documents") return [];
    if (!normalizedQuery) return childFolders;
    return childFolders.filter((folder) => folder.name.toLowerCase().includes(normalizedQuery));
  }, [childFolders, contentFilter, normalizedQuery]);

  const filteredDocuments = useMemo(() => {
    if (contentFilter === "folders") return [];
    if (!normalizedQuery) return documents;
    return documents.filter((document) => document.original_file_name.toLowerCase().includes(normalizedQuery));
  }, [contentFilter, documents, normalizedQuery]);

  const folderMenuProps = (folder: Folder) =>
    bindContextMenu({
      items: buildFolderMenuItems({ folder, role, canManageFolders, newChildLabel: labels.newChildLabel }),
      context: getFolderMenuContext(folder),
    });

  const documentMenuProps = (document: ApiDocument) =>
    bindContextMenu({
      items: buildDocumentMenuItems({ document, role }),
      context: getDocumentMenuContext(document),
    });

  const backgroundMenuProps = bindContextMenu({
    items: buildExplorerBackgroundMenuItems({ role, newSiblingLabel: labels.newSiblingLabel }),
    context: getExplorerBackgroundContext({ currentFolderId: currentFolder?.id ?? null }),
  });

  const isEmpty = !isLoading && filteredFolders.length === 0 && filteredDocuments.length === 0;
  const hasActiveFilter = normalizedQuery.length > 0 || contentFilter !== "all";

  return (
    <section className="flex min-h-[560px] flex-col overflow-hidden rounded-2xl border border-line bg-white shadow-panel">
      <div className="space-y-3 border-b border-line px-4 py-3">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <FolderBreadcrumb path={currentPath} onSelect={onNavigate} showRoot />
          <div className="flex flex-wrap items-center gap-2">
            <PrimaryButton className="h-9 px-3 text-xs sm:text-sm" onClick={() => onCreateSubfolder(currentFolder)}>
              <FolderPlus className="h-4 w-4" />
              {labels.newLabel}
            </PrimaryButton>
            <IconButton aria-label="Yenile" onClick={onRefresh}>
              <RefreshCw className="h-4 w-4" />
            </IconButton>
            <div className="inline-flex rounded-lg border border-line bg-white p-0.5">
              <button
                className={`inline-flex h-8 w-8 items-center justify-center rounded-md transition focus:outline-none focus:ring-4 focus:ring-violet-ring/70 ${
                  viewMode === "list" ? "bg-violet-soft text-violet-brand" : "text-slate-500 hover:text-ink"
                }`}
                type="button"
                aria-label="Liste görünümü"
                onClick={() => onViewModeChange("list")}
              >
                <List className="h-4 w-4" />
              </button>
              <button
                className={`inline-flex h-8 w-8 items-center justify-center rounded-md transition focus:outline-none focus:ring-4 focus:ring-violet-ring/70 ${
                  viewMode === "grid" ? "bg-violet-soft text-violet-brand" : "text-slate-500 hover:text-ink"
                }`}
                type="button"
                aria-label="İkon görünümü"
                onClick={() => onViewModeChange("grid")}
              >
                <Grid3X3 className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <SearchInput
            wrapperClassName="flex-1"
            placeholder={labels.searchPlaceholder}
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
          />
          <div className="flex flex-wrap gap-1.5">
            {(
              [
                { id: "all" as const, label: labels.filterAll },
                { id: "folders" as const, label: labels.filterFoldersOnly },
                { id: "documents" as const, label: labels.filterDocumentsOnly },
              ] as const
            ).map((filter) => (
              <button
                key={filter.id}
                type="button"
                className={`rounded-lg px-2.5 py-1.5 text-xs font-semibold transition focus:outline-none focus:ring-4 focus:ring-violet-ring/70 ${
                  contentFilter === filter.id
                    ? "bg-violet-soft text-violet-brand ring-1 ring-violet-ring"
                    : "border border-line bg-white text-slate-600 hover:bg-violet-soft/40"
                }`}
                onClick={() => setContentFilter(filter.id)}
              >
                {filter.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto" {...backgroundMenuProps}>
        {loadError ? (
          <ErrorState title="İçerik yüklenemedi" description={loadError} onRetry={onRetry} variant="inline" />
        ) : isLoading ? (
          <TableSkeleton rows={5} />
        ) : isEmpty ? (
          <EmptyState
            variant="inline"
              title={
                hasActiveFilter
                  ? "Sonuç bulunamadı"
                  : currentFolder
                    ? labels.emptyCurrentTitle
                    : labels.emptyRootTitle
              }
              description={
                hasActiveFilter
                  ? "Arama veya filtreyi değiştirmeyi deneyin."
                  : currentFolder
                    ? labels.emptyCurrentDescription
                    : labels.emptyRootDescription
              }
              action={
                !hasActiveFilter ? (
                  <PrimaryButton onClick={() => onCreateSubfolder(currentFolder)}>
                    <FolderPlus className="h-4 w-4" />
                    {labels.newLabel}
                  </PrimaryButton>
                ) : (
                  <Button variant="secondary" onClick={() => { setSearchQuery(""); setContentFilter("all"); }}>
                    Filtreleri temizle
                  </Button>
                )
              }
            />
        ) : viewMode === "grid" ? (
          <div className="grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-5">
            {filteredFolders.map((folder) => (
              <ExplorerGridItem
                key={folder.id}
                title={folder.name}
                subtitle={labels.itemTypeLabel}
                icon={<FolderIcon className="h-10 w-10" style={{ color: folder.color ?? "#fbbf24" }} fill="currentColor" />}
                onOpen={() => onNavigate(folder)}
                menuProps={folderMenuProps(folder)}
              />
            ))}
            {filteredDocuments.map((document) => (
              <ExplorerGridItem
                key={document.id}
                title={document.original_file_name}
                subtitle={formatFileSize(document.file_size)}
                icon={<FileTypeIcon extension={document.file_extension} size="xl" />}
                onOpen={() => navigate(`/documents/${document.id}`)}
                menuProps={documentMenuProps(document)}
              />
            ))}
          </div>
        ) : (
          <div className="min-w-[720px]">
            <div className="grid grid-cols-[minmax(220px,1fr)_140px_120px_100px] border-b border-line px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-400">
              <span>{labels.columnName}</span>
              <span>{labels.columnDate}</span>
              <span>{labels.columnType}</span>
              <span className="text-right">{labels.columnSize}</span>
            </div>
            <div className="divide-y divide-line">
              {filteredFolders.map((folder) => (
                <div
                  key={folder.id}
                  className="grid cursor-pointer grid-cols-[minmax(220px,1fr)_140px_120px_100px] items-center px-4 py-3 transition hover:bg-violet-soft/35"
                  onClick={() => onNavigate(folder)}
                  onDoubleClick={() => onNavigate(folder)}
                  {...folderMenuProps(folder)}
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <FolderIcon className="h-5 w-5 shrink-0" style={{ color: folder.color ?? "#fbbf24" }} fill="currentColor" />
                    <span className="truncate font-semibold text-ink">{folder.name}</span>
                  </div>
                  <span className="text-sm text-slate-500">{formatDate(folder.updated_at)}</span>
                  <span className="text-sm text-slate-500">{labels.itemTypeLabel}</span>
                  <span className="text-right text-sm text-slate-400">-</span>
                </div>
              ))}
              {filteredDocuments.map((document) => (
                <div
                  key={document.id}
                  className="grid cursor-pointer grid-cols-[minmax(220px,1fr)_140px_120px_100px] items-center px-4 py-3 transition hover:bg-violet-soft/35"
                  onClick={() => navigate(`/documents/${document.id}`)}
                  onDoubleClick={() => navigate(`/documents/${document.id}`)}
                  {...documentMenuProps(document)}
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <span onClick={(event) => event.stopPropagation()}>
                      <FavoriteStarButton
                        isFavorite={document.isFavorite ?? false}
                        isLoading={favoriteLoadingId === document.id}
                        onToggle={() =>
                          publishContextMenuAction({
                            action: ContextMenuActions.document.toggleFavorite,
                            context: getDocumentMenuContext(document),
                          })
                        }
                      />
                    </span>
                    <FileTypeIcon extension={document.file_extension} />
                    <span className="truncate font-semibold text-ink">{document.original_file_name}</span>
                  </div>
                  <span className="text-sm text-slate-500">{formatDateTime(document.created_at)}</span>
                  <span className="text-sm uppercase text-slate-500">{document.file_extension}</span>
                  <span className="text-right text-sm text-slate-500">{formatFileSize(document.file_size)}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {!isLoading && !loadError && (
        <div className="border-t border-line px-4 py-2 text-xs font-semibold text-slate-500">
          {labels.footerSummary(filteredFolders.length, filteredDocuments.length)}
        </div>
      )}

      {menu}
    </section>
  );
}

function ExplorerGridItem({
  title,
  subtitle,
  icon,
  onOpen,
  menuProps,
}: {
  title: string;
  subtitle: string;
  icon: React.ReactNode;
  onOpen: () => void;
  menuProps: ReturnType<ReturnType<typeof useContextMenu>["bindContextMenu"]>;
}) {
  return (
    <div
      className="flex cursor-pointer flex-col items-center gap-2 rounded-xl border border-transparent p-4 text-center transition hover:border-violet-ring hover:bg-violet-soft/30 focus:outline-none focus:ring-4 focus:ring-violet-ring/70"
      role="button"
      tabIndex={0}
      onClick={onOpen}
      onDoubleClick={onOpen}
      {...menuProps}
      onKeyDown={(event) => {
        menuProps.onKeyDown?.(event);
        if (event.key === "Enter") onOpen();
      }}
    >
      <div className="flex h-14 w-14 items-center justify-center">{icon}</div>
      <div className="w-full min-w-0">
        <p className="truncate text-sm font-semibold text-ink">{title}</p>
        <p className="mt-0.5 truncate text-xs text-slate-500">{subtitle}</p>
      </div>
    </div>
  );
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(bytes / 1024, 1).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
