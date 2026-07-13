import { Pencil, Reply, Trash2 } from "lucide-react";
import type { DocumentComment } from "../../services/commentService";
import type { TenantRole } from "../../types/auth";
import { ContextMenuActions } from "../actions";
import { filterMenuByRole } from "../permissions";
import type { ContextMenuItem } from "../types";

type CommentMenuOptions = {
  comment: DocumentComment;
  role: TenantRole;
  canManage: boolean;
};

export function buildCommentMenuItems({ comment, role, canManage }: CommentMenuOptions): ContextMenuItem[] {
  void comment;
  return filterMenuByRole(
    [
      { id: "reply", label: "Yanıtla", icon: Reply, action: ContextMenuActions.comment.reply },
      { id: "edit", label: "Düzenle", icon: Pencil, disabled: !canManage, action: ContextMenuActions.comment.edit },
      { id: "divider-1", label: "", divider: true },
      { id: "delete", label: "Sil", icon: Trash2, danger: true, disabled: !canManage, action: ContextMenuActions.comment.delete },
    ],
    role,
  );
}

export function getCommentMenuContext(comment: DocumentComment, documentId: string) {
  return { entityType: "comment" as const, entity: comment, meta: { documentId } };
}
