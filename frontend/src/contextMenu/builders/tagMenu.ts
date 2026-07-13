import { GitMerge, Palette, Pencil, Trash2 } from "lucide-react";
import type { Tag } from "../../types/tag";
import type { TenantRole } from "../../types/auth";
import { ContextMenuActions } from "../actions";
import { filterMenuByRole } from "../permissions";
import type { ContextMenuItem } from "../types";

type TagMenuOptions = {
  tag: Tag;
  role: TenantRole;
  canManageTags: boolean;
};

export function buildTagMenuItems({ tag, role, canManageTags }: TagMenuOptions): ContextMenuItem[] {
  return filterMenuByRole(
    [
      { id: "edit", label: "Düzenle", icon: Pencil, disabled: !canManageTags, roles: ["owner", "admin"], action: ContextMenuActions.tag.edit },
      { id: "color", label: "Renk Değiştir", icon: Palette, disabled: !canManageTags, roles: ["owner", "admin"], action: ContextMenuActions.tag.changeColor },
      { id: "merge", label: "Birleştir", icon: GitMerge, disabled: !canManageTags, roles: ["owner", "admin"], action: ContextMenuActions.tag.merge },
      { id: "divider-1", label: "", divider: true },
      {
        id: "delete",
        label: "Sil",
        icon: Trash2,
        danger: true,
        disabled: !canManageTags,
        roles: ["owner", "admin"],
        action: ContextMenuActions.tag.delete,
      },
    ],
    role,
  );
}

export function getTagMenuContext(tag: Tag) {
  return { entityType: "tag" as const, entity: tag };
}
