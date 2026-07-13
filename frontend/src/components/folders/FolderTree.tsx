import { ChevronRight, Folder as FolderIcon, GripVertical, Lock, Pencil, Plus, Star, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useAuth } from "../../services/AuthContext";
import { buildExplorerBackgroundMenuItems, buildFolderMenuItems, getExplorerBackgroundContext, getFolderMenuContext } from "../../contextMenu";
import { useContextMenu } from "../../hooks/useContextMenu";
import type { Folder } from "../../types/folder";
import type { TenantRole } from "../../types/auth";
import type { FolderLabels } from "../../hooks/useFolderLabels";
import { IconButton, PrimaryButton, StatusBadge } from "../common";
import { canDropFolderOn, findFolder, findFolderPath } from "./folderUtils";

type FolderTreeProps = {
  folders: Folder[];
  labels: FolderLabels;
  selectedFolderId: string | null;
  canManageFolders: boolean;
  pinnedFolderIds?: Set<string>;
  onSelect: (folder: Folder | null) => void;
  onCreateChild: (parent: Folder | null) => void;
  onRename: (folder: Folder) => void;
  onDelete: (folder: Folder) => void;
  onMove?: (folder: Folder, parentId: string | null) => void;
  onTogglePin?: (folder: Folder) => void;
};

export function FolderTree({
  folders,
  labels,
  selectedFolderId,
  canManageFolders,
  pinnedFolderIds,
  onSelect,
  onCreateChild,
  onRename,
  onDelete,
  onMove,
  onTogglePin,
}: FolderTreeProps) {
  const { activeTenant } = useAuth();
  const role = activeTenant?.role ?? "member";
  const { menu, bindContextMenu } = useContextMenu();
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [dropTargetId, setDropTargetId] = useState<string | null | undefined>(undefined);

  useEffect(() => {
    if (!selectedFolderId) return;
    const path = findFolderPath(folders, selectedFolderId);
    setExpandedIds((current) => {
      const next = new Set(current);
      for (const folder of path) {
        next.add(folder.id);
      }
      return next;
    });
  }, [selectedFolderId, folders]);

  const toggleExpanded = (folderId: string) => {
    setExpandedIds((current) => {
      const next = new Set(current);
      if (next.has(folderId)) next.delete(folderId);
      else next.add(folderId);
      return next;
    });
  };

  const handleDrop = (targetId: string | null) => {
    if (!draggingId || !onMove) return;
    if (!canDropFolderOn(draggingId, targetId, folders)) return;
    const dragged = findFolder(folders, draggingId);
    if (!dragged || dragged.isLocked) return;
    onMove(dragged, targetId);
    setDraggingId(null);
    setDropTargetId(undefined);
  };

  const rootSelected = selectedFolderId === null;
  const rootDropActive = dropTargetId === null && draggingId !== null;

  const rootMenuProps = bindContextMenu({
    items: buildExplorerBackgroundMenuItems({ role, newSiblingLabel: labels.newSiblingLabel }),
    context: getExplorerBackgroundContext({ currentFolderId: null }),
  });

  if (folders.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-line bg-slate-50/80 p-5 text-center">
        <FolderIcon className="mx-auto h-8 w-8 text-amber-400" fill="currentColor" />
        <p className="mt-3 text-sm font-bold text-ink">{labels.emptyTreeTitle}</p>
        <p className="mt-1 text-sm text-slate-500">{labels.emptyTreeDescription}</p>
        <PrimaryButton className="mt-4" onClick={() => onCreateChild(null)}>
          <Plus className="h-4 w-4" />
          {labels.newLabel}
        </PrimaryButton>
      </div>
    );
  }

  return (
    <>
      <div className="space-y-0.5">
        <div
          className={`relative flex h-10 w-full items-center rounded-lg transition ${
            rootSelected ? "bg-violet-soft ring-1 ring-violet-ring" : "hover:bg-violet-soft/40"
          } ${rootDropActive ? "bg-violet-soft/60 ring-2 ring-violet-brand" : ""}`}
          {...rootMenuProps}
        >
          {rootSelected ? <span className="absolute inset-y-1 left-0 w-0.5 rounded-full bg-violet-brand" /> : null}
          <button
            className={`flex min-w-0 flex-1 items-center gap-2 truncate px-3 text-left text-sm font-semibold ${
              rootSelected ? "text-violet-brand" : "text-slate-600"
            }`}
            type="button"
            onClick={() => onSelect(null)}
            onDragOver={(event) => {
              if (!draggingId) return;
              event.preventDefault();
              setDropTargetId(null);
            }}
            onDrop={(event) => {
              event.preventDefault();
              handleDrop(null);
            }}
          >
            <FolderIcon className="h-4 w-4 shrink-0 text-amber-400" fill="currentColor" />
            <span className="truncate">{labels.rootLabel}</span>
          </button>
          {rootDropActive ? (
            <span className="mr-2 shrink-0 text-[10px] font-semibold uppercase tracking-wide text-violet-brand">
              {labels.dragDropHint}
            </span>
          ) : null}
        </div>

        {folders.map((folder) => (
          <FolderTreeNode
            key={folder.id}
            folder={folder}
            folders={folders}
            labels={labels}
            role={role}
            depth={0}
            expandedIds={expandedIds}
            selectedFolderId={selectedFolderId}
            canManageFolders={canManageFolders}
            pinnedFolderIds={pinnedFolderIds}
            draggingId={draggingId}
            dropTargetId={dropTargetId}
            bindContextMenu={bindContextMenu}
            onToggleExpanded={toggleExpanded}
            onSelect={onSelect}
            onCreateChild={onCreateChild}
            onRename={onRename}
            onDelete={onDelete}
            onMove={onMove}
            onTogglePin={onTogglePin}
            onDragStart={setDraggingId}
            onDragEnd={() => {
              setDraggingId(null);
              setDropTargetId(undefined);
            }}
            onDropTarget={setDropTargetId}
            onDrop={handleDrop}
          />
        ))}
      </div>
      {menu}
    </>
  );
};

type FolderTreeNodeProps = {
  folder: Folder;
  folders: Folder[];
  labels: FolderLabels;
  role: TenantRole;
  depth: number;
  expandedIds: Set<string>;
  selectedFolderId: string | null;
  canManageFolders: boolean;
  pinnedFolderIds?: Set<string>;
  draggingId: string | null;
  dropTargetId: string | null | undefined;
  bindContextMenu: ReturnType<typeof useContextMenu>["bindContextMenu"];
  onToggleExpanded: (folderId: string) => void;
  onSelect: (folder: Folder | null) => void;
  onCreateChild: (parent: Folder | null) => void;
  onRename: (folder: Folder) => void;
  onDelete: (folder: Folder) => void;
  onMove?: (folder: Folder, parentId: string | null) => void;
  onTogglePin?: (folder: Folder) => void;
  onDragStart: (folderId: string) => void;
  onDragEnd: () => void;
  onDropTarget: (folderId: string | null | undefined) => void;
  onDrop: (targetId: string | null) => void;
};

function FolderTreeNode({
  folder,
  folders,
  labels,
  role,
  depth,
  expandedIds,
  selectedFolderId,
  canManageFolders,
  pinnedFolderIds,
  draggingId,
  dropTargetId,
  bindContextMenu,
  onToggleExpanded,
  onSelect,
  onCreateChild,
  onRename,
  onDelete,
  onMove,
  onTogglePin,
  onDragStart,
  onDragEnd,
  onDropTarget,
  onDrop,
}: FolderTreeNodeProps) {
  const hasChildren = (folder.children?.length ?? 0) > 0;
  const expanded = expandedIds.has(folder.id);
  const selected = selectedFolderId === folder.id;
  const canModify = canManageFolders && !folder.isSystem && !folder.isLocked;
  const canDrag = canModify && Boolean(onMove);
  const isPinned = pinnedFolderIds?.has(folder.id);
  const lifecycle = folder.lifecycleStatus ?? folder.status ?? "active";
  const folderColor = folder.color ?? "#fbbf24";
  const isDropTarget = dropTargetId === folder.id && draggingId !== null && draggingId !== folder.id;
  const isDragging = draggingId === folder.id;

  const menuProps = bindContextMenu({
    items: buildFolderMenuItems({ folder, role, canManageFolders, newChildLabel: labels.newChildLabel }),
    context: getFolderMenuContext(folder),
  });

  return (
    <div className="relative" style={{ paddingLeft: `${depth * 16}px` }}>
      {depth > 0 ? (
        <span
          className="pointer-events-none absolute bottom-0 left-[7px] top-0 w-px bg-line"
          aria-hidden="true"
        />
      ) : null}
      <div
        className={`group relative flex h-10 items-center gap-0.5 rounded-lg transition ${
          selected ? "bg-violet-soft ring-1 ring-violet-ring" : "hover:bg-violet-soft/40"
        } ${isDropTarget ? "bg-violet-soft/70 ring-2 ring-violet-brand" : ""} ${isDragging ? "opacity-45" : ""}`}
        draggable={canDrag}
        onDragStart={(event) => {
          if (!canDrag) return;
          event.dataTransfer.setData("text/plain", folder.id);
          event.dataTransfer.effectAllowed = "move";
          onDragStart(folder.id);
        }}
        onDragEnd={onDragEnd}
        onDragOver={(event) => {
          if (!draggingId || draggingId === folder.id) return;
          if (!canDropFolderOn(draggingId, folder.id, folders)) return;
          event.preventDefault();
          onDropTarget(folder.id);
        }}
        onDrop={(event) => {
          event.preventDefault();
          onDrop(folder.id);
        }}
        {...menuProps}
      >
        {selected ? <span className="absolute inset-y-1.5 left-0 w-0.5 rounded-full bg-violet-brand" /> : null}

        <button
          className="inline-flex h-8 w-7 shrink-0 items-center justify-center rounded-md text-slate-400 transition hover:bg-white/70 hover:text-ink"
          type="button"
          aria-label={expanded ? labels.collapseAria : labels.expandAria}
          aria-expanded={hasChildren ? expanded : undefined}
          onClick={(event) => {
            event.stopPropagation();
            if (hasChildren) onToggleExpanded(folder.id);
          }}
        >
          {hasChildren ? (
            <ChevronRight className={`h-4 w-4 transition-transform ${expanded ? "rotate-90" : ""}`} />
          ) : (
            <span className="inline-block h-4 w-4" />
          )}
        </button>

        {canDrag ? (
          <GripVertical className="h-3.5 w-3.5 shrink-0 cursor-grab text-slate-300 opacity-0 transition group-hover:opacity-100 active:cursor-grabbing" />
        ) : (
          <span className="w-3.5 shrink-0" />
        )}

        <button
          className={`flex min-w-0 flex-1 items-center gap-2 truncate pr-1 text-left text-sm font-semibold ${
            selected ? "text-violet-brand" : "text-slate-600"
          }`}
          type="button"
          onClick={() => onSelect(folder)}
        >
          <FolderIcon className="h-4 w-4 shrink-0" style={{ color: folderColor }} fill="currentColor" />
          <span className="truncate">{folder.name}</span>
          {isPinned ? <Star className="h-3 w-3 shrink-0 fill-amber-400 text-amber-400" /> : null}
          {folder.isLocked ? <Lock className="h-3 w-3 shrink-0 text-violet-brand" /> : null}
          {lifecycle !== "active" ? <StatusBadge status={lifecycle} className="scale-90" /> : null}
        </button>

        {isDropTarget ? (
          <span className="mr-1 hidden shrink-0 text-[10px] font-semibold uppercase tracking-wide text-violet-brand sm:inline">
            {labels.dragDropHint}
          </span>
        ) : null}

        <span className="hidden items-center gap-0.5 pr-1 group-hover:flex" onClick={(event) => event.stopPropagation()}>
          {onTogglePin ? (
            <IconButton
              className={`h-7 w-7 border-0 ${isPinned ? "text-amber-500" : ""}`}
              aria-label="Sabitle"
              onClick={() => onTogglePin(folder)}
            >
              <Star className={`h-3.5 w-3.5 ${isPinned ? "fill-current" : ""}`} />
            </IconButton>
          ) : null}
          <IconButton className="h-7 w-7 border-0" aria-label={labels.addChildAria} onClick={() => onCreateChild(folder)}>
            <Plus className="h-3.5 w-3.5" />
          </IconButton>
          {canModify && (
            <>
              <IconButton className="h-7 w-7 border-0" aria-label={labels.renameAria} onClick={() => onRename(folder)}>
                <Pencil className="h-3.5 w-3.5" />
              </IconButton>
              <IconButton
                className="h-7 w-7 border-0 text-red-500 hover:bg-red-50"
                aria-label={labels.deleteAria}
                onClick={() => onDelete(folder)}
              >
                <Trash2 className="h-3.5 w-3.5" />
              </IconButton>
            </>
          )}
        </span>
      </div>

      {hasChildren && expanded
        ? folder.children?.map((child) => (
            <FolderTreeNode
              key={child.id}
              folder={child}
              folders={folders}
              labels={labels}
              role={role}
              depth={depth + 1}
              expandedIds={expandedIds}
              selectedFolderId={selectedFolderId}
              canManageFolders={canManageFolders}
              pinnedFolderIds={pinnedFolderIds}
              draggingId={draggingId}
              dropTargetId={dropTargetId}
              bindContextMenu={bindContextMenu}
              onToggleExpanded={onToggleExpanded}
              onSelect={onSelect}
              onCreateChild={onCreateChild}
              onRename={onRename}
              onDelete={onDelete}
              onMove={onMove}
              onTogglePin={onTogglePin}
              onDragStart={onDragStart}
              onDragEnd={onDragEnd}
              onDropTarget={onDropTarget}
              onDrop={onDrop}
            />
          ))
        : null}
    </div>
  );
}
