import { FolderIcon, FolderPlus, Lock, Palette, Pencil, RefreshCw, Share2, Shapes, Trash2, Unlock } from "lucide-react";
import type { Folder } from "../../types/folder";
import type { TenantRole } from "../../types/auth";
import { ContextMenuActions } from "../actions";
import { filterMenuByRole } from "../permissions";
import type { ContextMenuItem } from "../types";

type FolderMenuOptions = {
  folder: Folder;
  role: TenantRole;
  canManageFolders: boolean;
  newChildLabel?: string;
};

export function buildFolderMenuItems({
  folder,
  role,
  canManageFolders,
  newChildLabel = "Yeni Alt Klasör",
}: FolderMenuOptions): ContextMenuItem[] {
  const canModify = canManageFolders && !folder.isSystem;

  return filterMenuByRole(
    [
      { id: "open", label: "Aç", icon: FolderIcon, action: ContextMenuActions.folder.open },
      { id: "new", label: newChildLabel, icon: FolderPlus, roles: ["owner", "admin"], action: ContextMenuActions.folder.newSubfolder },
      { id: "divider-1", label: "", divider: true },
      { id: "rename", label: "Yeniden Adlandır", icon: Pencil, disabled: !canModify, roles: ["owner", "admin"], action: ContextMenuActions.folder.rename },
      { id: "share", label: "Paylaş", icon: Share2, disabled: !canModify, roles: ["owner", "admin"], action: ContextMenuActions.folder.share },
      { id: "move", label: "Taşı", icon: FolderPlus, disabled: !canModify, roles: ["owner", "admin"], action: ContextMenuActions.folder.move },
      {
        id: "lock",
        label: "Kilitle",
        icon: Lock,
        disabled: !canModify || folder.isLocked,
        roles: ["owner", "admin"],
        action: ContextMenuActions.folder.lock,
      },
      {
        id: "unlock",
        label: "Kilidi Aç",
        icon: Unlock,
        disabled: !canManageFolders || !folder.isLocked,
        roles: ["owner", "admin"],
        action: ContextMenuActions.folder.unlock,
      },
      { id: "color", label: "Renk Değiştir", icon: Palette, disabled: !canModify, roles: ["owner", "admin"], action: ContextMenuActions.folder.changeColor },
      { id: "icon", label: "İkon Değiştir", icon: Shapes, disabled: !canModify, roles: ["owner", "admin"], action: ContextMenuActions.folder.changeIcon },
      { id: "divider-2", label: "", divider: true },
      {
        id: "delete",
        label: "Sil",
        icon: Trash2,
        danger: true,
        disabled: !canModify,
        roles: ["owner", "admin"],
        action: ContextMenuActions.folder.delete,
      },
    ],
    role,
  );
}

type ExplorerBackgroundMenuOptions = {
  role: TenantRole;
  newSiblingLabel?: string;
};

export function buildExplorerBackgroundMenuItems({
  role,
  newSiblingLabel = "Yeni Klasör",
}: ExplorerBackgroundMenuOptions): ContextMenuItem[] {
  return filterMenuByRole(
    [
      { id: "new", label: newSiblingLabel, icon: FolderPlus, roles: ["owner", "admin"], action: ContextMenuActions.folder.newSibling },
      { id: "refresh", label: "Yenile", icon: RefreshCw, action: ContextMenuActions.folder.refresh },
    ],
    role,
  );
}

export function getFolderMenuContext(folder: Folder, meta?: Record<string, unknown>) {
  return { entityType: "folder" as const, entity: folder, meta };
}

export function getExplorerBackgroundContext(meta?: Record<string, unknown>) {
  return { entityType: "explorer-background" as const, entity: null, meta };
}
