import { FolderPlus, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  FolderCreateModal,
  FolderExplorer,
  FolderMetadataPanel,
  FolderMoveModal,
  type ExplorerViewMode,
  FolderRenameModal,
  FolderTree,
  findFolder,
  findFolderPath,
  getChildFolders,
} from "../components/folders";
import { ErrorState, IconButton, PageLoader, PageTitle, PrimaryButton, useConfirm } from "../components/common";
import { listDocuments } from "../data/documentsApi";
import { ContextMenuActions, useContextMenuActionHandler } from "../contextMenu";
import { useFolderLabels } from "../hooks/useFolderLabels";
import { useDocumentFavorite } from "../hooks/useDocumentFavorite";
import { useDocumentMenuActions } from "../hooks/useDocumentMenuActions";
import {
  createFolder,
  deleteFolder,
  listFolderDocuments,
  listFolders,
  lockFolder,
  moveFolder,
  unlockFolder,
  updateFolder,
} from "../services/folderService";
import { createPin, deletePin, listPins } from "../services/navigationService";
import { useAuth } from "../services/AuthContext";
import { toast } from "../services/toast";
import type { ApiDocument } from "../types/document";
import type { Folder } from "../types/folder";

const FOLDER_COLORS = ["#8b5cf6", "#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#ec4899"];

export function FoldersPage() {
  const labels = useFolderLabels();
  const { activeTenant } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const canManageFolders = activeTenant?.role === "owner" || activeTenant?.role === "admin";
  const [folders, setFolders] = useState<Folder[]>([]);
  const [currentFolderId, setCurrentFolderId] = useState<string | null>(searchParams.get("folderId"));
  const [documents, setDocuments] = useState<ApiDocument[]>([]);
  const [pins, setPins] = useState<Array<{ id: string; entityId: string }>>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [foldersLoadError, setFoldersLoadError] = useState<string | null>(null);
  const [contentsLoadError, setContentsLoadError] = useState<string | null>(null);
  const [isContentsLoading, setIsContentsLoading] = useState(false);
  const confirm = useConfirm();
  const [isSaving, setIsSaving] = useState(false);
  const [createParent, setCreateParent] = useState<Folder | null | undefined>(undefined);
  const [renamingFolder, setRenamingFolder] = useState<Folder | null>(null);
  const [movingFolder, setMovingFolder] = useState<Folder | null>(null);
  const [viewMode, setViewMode] = useState<ExplorerViewMode>("list");

  const pinnedFolderIds = useMemo(() => new Set(pins.map((pin) => pin.entityId)), [pins]);
  const currentFolder = useMemo(() => (currentFolderId ? findFolder(folders, currentFolderId) : null), [folders, currentFolderId]);
  const currentPath = useMemo(() => (currentFolderId ? findFolderPath(folders, currentFolderId) : []), [folders, currentFolderId]);
  const childFolders = useMemo(() => getChildFolders(folders, currentFolder), [folders, currentFolder]);

  const { favoriteLoadingId, toggleFavorite } = useDocumentFavorite({
    onUpdated: (updatedDocument) => {
      setDocuments((current) => current.map((item) => (item.id === updatedDocument.id ? updatedDocument : item)));
    },
  });

  const { handleContextMenuAction: handleDocumentMenuAction, modals: documentActionModals } = useDocumentMenuActions({
    onToggleFavorite: toggleFavorite,
    onDocumentUpdated: (updated) => {
      setDocuments((current) => current.map((item) => (item.id === updated.id ? updated : item)));
    },
    onDocumentRemoved: (removed) => {
      setDocuments((current) => current.filter((item) => item.id !== removed.id));
      void fetchContents();
    },
  });

  const fetchFolders = useCallback(async () => {
    setIsLoading(true);
    setFoldersLoadError(null);
    try {
      const items = await listFolders();
      setFolders(items);
      if (currentFolderId && !findFolder(items, currentFolderId)) {
        setCurrentFolderId(null);
        setDocuments([]);
      }
    } catch (fetchError) {
      const message = fetchError instanceof Error ? fetchError.message : labels.loadFailed;
      setFoldersLoadError(message);
      setFolders([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, [currentFolderId, labels.loadFailed]);

  const fetchPins = useCallback(async () => {
    try {
      const items = await listPins();
      setPins(items.filter((item) => item.entityType === "folder").map((item) => ({ id: item.id, entityId: item.entityId })));
    } catch {
      setPins([]);
    }
  }, []);

  const fetchContents = useCallback(async () => {
    setIsContentsLoading(true);
    setContentsLoadError(null);
    try {
      const items = currentFolderId ? await listFolderDocuments(currentFolderId) : await listDocuments("root");
      setDocuments(items);
    } catch (fetchError) {
      const message = fetchError instanceof Error ? fetchError.message : labels.contentLoadFailed;
      setContentsLoadError(message);
      setDocuments([]);
      toast.error(message);
    } finally {
      setIsContentsLoading(false);
    }
  }, [currentFolderId, labels.contentLoadFailed]);

  useEffect(() => {
    void fetchFolders();
    void fetchPins();
  }, []);

  useEffect(() => {
    void fetchContents();
  }, [fetchContents]);

  useEffect(() => {
    const folderId = searchParams.get("folderId");
    setCurrentFolderId(folderId);
  }, [searchParams]);

  const handleNavigate = (folder: Folder | null) => {
    const nextId = folder?.id ?? null;
    setCurrentFolderId(nextId);
    if (nextId) {
      setSearchParams({ folderId: nextId });
    } else {
      setSearchParams({});
    }
  };

  const handleCreate = async (input: { name: string; description: string; parentId?: string }) => {
    setIsSaving(true);
    try {
      const created = await createFolder(input);
      setCreateParent(undefined);
      await fetchFolders();
      handleNavigate(created);
      toast.success(labels.createdToast);
    } catch (createError) {
      toast.error(createError instanceof Error ? createError.message : labels.createFailed);
    } finally {
      setIsSaving(false);
    }
  };

  const handleRename = async (input: { name: string; description: string }) => {
    if (!renamingFolder) return;
    setIsSaving(true);
    try {
      await updateFolder(renamingFolder.id, input);
      setRenamingFolder(null);
      await fetchFolders();
      toast.success(labels.updatedToast);
    } catch (renameError) {
      toast.error(renameError instanceof Error ? renameError.message : labels.updateFailed);
    } finally {
      setIsSaving(false);
    }
  };

  const handleMove = async (folder: Folder, parentId: string | null) => {
    setIsSaving(true);
    try {
      await moveFolder(folder.id, parentId);
      setMovingFolder(null);
      await fetchFolders();
      toast.success(labels.moveSuccessToast);
    } catch (moveError) {
      toast.error(moveError instanceof Error ? moveError.message : labels.moveFailedToast);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (folder: Folder) => {
    if (folder.isSystem) {
      toast.error(labels.cannotDeleteSystem);
      return;
    }
    if (folder.isLocked) {
      toast.error(labels.lockedDeleteError);
      return;
    }
    const shouldDelete = await confirm({
      title: labels.deleteConfirmTitle,
      message: labels.deleteConfirm(folder.name),
      confirmLabel: "Sil",
      variant: "danger",
    });
    if (!shouldDelete) return;
    try {
      await deleteFolder(folder.id);
      if (currentFolderId === folder.id) {
        handleNavigate(folder.parentId ? findFolder(folders, folder.parentId) : null);
      }
      await fetchFolders();
      await fetchContents();
      toast.success(labels.deletedToast);
    } catch (deleteError) {
      toast.error(deleteError instanceof Error ? deleteError.message : labels.deleteFailed);
    }
  };

  const handleTogglePin = async (folder: Folder) => {
    const existing = pins.find((pin) => pin.entityId === folder.id);
    try {
      if (existing) {
        await deletePin(existing.id);
        toast.info(labels.pinRemovedToast);
      } else {
        await createPin({ entityType: "folder", entityId: folder.id, label: folder.name, path: "/folders" });
        toast.success(labels.pinAddedToast);
      }
      await fetchPins();
    } catch (pinError) {
      toast.error(pinError instanceof Error ? pinError.message : "Sabitleme işlemi başarısız.");
    }
  };

  const handleChangeColor = async (folder: Folder) => {
    const currentIndex = FOLDER_COLORS.findIndex((color) => color === folder.color);
    const nextColor = FOLDER_COLORS[(currentIndex + 1) % FOLDER_COLORS.length];
    try {
      await updateFolder(folder.id, { name: folder.name, description: folder.description, color: nextColor });
      await fetchFolders();
      toast.success(labels.colorUpdatedToast);
    } catch (colorError) {
      toast.error(colorError instanceof Error ? colorError.message : "Renk güncellenemedi.");
    }
  };

  const handleLockToggle = async (folder: Folder, lock: boolean) => {
    try {
      if (lock) await lockFolder(folder.id);
      else await unlockFolder(folder.id);
      await fetchFolders();
      toast.success(lock ? labels.lockSuccessToast : labels.unlockSuccessToast);
    } catch (lockError) {
      toast.error(lockError instanceof Error ? lockError.message : "Kilit işlemi başarısız.");
    }
  };

  const handleRefresh = () => {
    void fetchFolders();
    void fetchContents();
    toast.info(labels.refreshToast);
  };

  const handleContextMenuAction = useCallback(
    (detail: { action: string; context: { entityType: string; entity: unknown; meta?: Record<string, unknown> } }) => {
      const { action, context } = detail;

      if (context.entityType === "folder") {
        const folder = context.entity as Folder;
        switch (action) {
          case ContextMenuActions.folder.open:
            handleNavigate(folder);
            return;
          case ContextMenuActions.folder.newSubfolder:
            setCreateParent(folder);
            return;
          case ContextMenuActions.folder.rename:
            setRenamingFolder(folder);
            return;
          case ContextMenuActions.folder.delete:
            void handleDelete(folder);
            return;
          case ContextMenuActions.folder.move:
            setMovingFolder(folder);
            return;
          case ContextMenuActions.folder.changeColor:
            void handleChangeColor(folder);
            return;
          case ContextMenuActions.folder.changeIcon:
            void updateFolder(folder.id, { name: folder.name, description: folder.description, icon: folder.icon ? null : "folder" })
              .then(() => fetchFolders())
              .then(() => toast.success(labels.iconUpdatedToast))
              .catch((error) => toast.error(error instanceof Error ? error.message : "İkon güncellenemedi."));
            return;
          case ContextMenuActions.folder.lock:
            void handleLockToggle(folder, true);
            return;
          case ContextMenuActions.folder.unlock:
            void handleLockToggle(folder, false);
            return;
          case ContextMenuActions.folder.share:
            toast.info("Paylaşım belge paylaşım akışı üzerinden yönetilir.");
            return;
        }
      }

      if (context.entityType === "explorer-background") {
        switch (action) {
          case ContextMenuActions.folder.newSibling:
            setCreateParent(currentFolder);
            return;
          case ContextMenuActions.folder.refresh:
            handleRefresh();
            return;
        }
      }

      if (context.entityType === "document") {
        handleDocumentMenuAction(detail);
      }
    },
    [currentFolder, fetchContents, folders, handleDocumentMenuAction, labels.iconUpdatedToast],
  );

  useContextMenuActionHandler(handleContextMenuAction);

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title={labels.pageTitle} subtitle={labels.pageSubtitle} />
        <div className="flex gap-2">
          <IconButton aria-label="Yenile" onClick={handleRefresh}>
            <RefreshCw className="h-4 w-4" />
          </IconButton>
          <PrimaryButton onClick={() => setCreateParent(currentFolder)}>
            <FolderPlus className="h-4 w-4" />
            {labels.newLabel}
          </PrimaryButton>
        </div>
      </div>

      <div className="grid min-h-[660px] gap-4 lg:grid-cols-[minmax(240px,300px)_minmax(0,1fr)]">
        <aside className="flex max-h-[760px] flex-col rounded-2xl border border-line bg-white p-4 shadow-panel">
          <div className="mb-3 flex items-center justify-between gap-2">
            <div>
              <h2 className="text-sm font-bold text-ink">{labels.treeTitle}</h2>
              <p className="text-xs text-slate-500">{labels.plural}</p>
            </div>
            <IconButton className="h-8 w-8" aria-label={labels.addRootAria} onClick={() => setCreateParent(null)}>
              <FolderPlus className="h-4 w-4" />
            </IconButton>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto pr-1">
            {isLoading ? (
              <PageLoader label={labels.loadingLabel} variant="spinner" className="border-0 bg-slate-50 shadow-none" />
            ) : foldersLoadError ? (
              <ErrorState
                title={labels.loadFoldersErrorTitle}
                description={foldersLoadError}
                onRetry={() => void fetchFolders()}
                variant="inline"
              />
            ) : (
              <FolderTree
                folders={folders}
                labels={labels}
                selectedFolderId={currentFolderId}
                canManageFolders={canManageFolders}
                pinnedFolderIds={pinnedFolderIds}
                onSelect={handleNavigate}
                onCreateChild={(parent) => setCreateParent(parent)}
                onRename={setRenamingFolder}
                onDelete={(folder) => void handleDelete(folder)}
                onMove={(folder, parentId) => void handleMove(folder, parentId)}
                onTogglePin={(folder) => void handleTogglePin(folder)}
              />
            )}
          </div>
        </aside>

        <div className="space-y-4">
          <FolderExplorer
            folders={folders}
            labels={labels}
            currentFolder={currentFolder}
            currentPath={currentPath}
            childFolders={childFolders}
            documents={documents}
            isLoading={isContentsLoading}
            loadError={contentsLoadError}
            onRetry={() => void fetchContents()}
            canManageFolders={canManageFolders}
            favoriteLoadingId={favoriteLoadingId}
            viewMode={viewMode}
            onViewModeChange={setViewMode}
            onNavigate={handleNavigate}
            onCreateSubfolder={(parent) => setCreateParent(parent)}
            onRefresh={handleRefresh}
          />
          {currentFolder ? (
            <section className="rounded-2xl border border-line bg-white p-4 shadow-panel">
              <h3 className="mb-3 text-sm font-bold text-ink">{labels.metadataTitle}</h3>
              <FolderMetadataPanel folderId={currentFolder.id} />
            </section>
          ) : null}
        </div>
      </div>

      {createParent !== undefined && (
        <FolderCreateModal parent={createParent} isSaving={isSaving} onClose={() => setCreateParent(undefined)} onCreate={(input) => void handleCreate(input)} />
      )}
      {renamingFolder && !renamingFolder.isSystem && (
        <FolderRenameModal folder={renamingFolder} isSaving={isSaving} onClose={() => setRenamingFolder(null)} onSave={(input) => void handleRename(input)} />
      )}
      {movingFolder && <FolderMoveModal folder={movingFolder} folders={folders} isSaving={isSaving} onClose={() => setMovingFolder(null)} onMove={(parentId) => void handleMove(movingFolder, parentId)} />}
      {documentActionModals}
    </div>
  );
}
