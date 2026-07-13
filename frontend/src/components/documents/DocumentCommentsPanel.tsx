import { MessageSquare, Pencil, Reply, Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import type { ApiDocument } from "../../types/document";
import { EmptyState, LoadingState, PrimaryButton, Spinner, useConfirm } from "../common";
import {
  buildCommentMenuItems,
  ContextMenuActions,
  getCommentMenuContext,
  useContextMenuActionHandler,
} from "../../contextMenu";
import { useContextMenu } from "../../hooks/useContextMenu";
import {
  createDocumentComment,
  deleteDocumentComment,
  listDocumentComments,
  updateDocumentComment,
  type DocumentComment,
} from "../../services/commentService";
import { useAuth } from "../../services/AuthContext";
import { toast } from "../../services/toast";

type DocumentCommentsPanelProps = {
  document: ApiDocument;
};

export function DocumentCommentsPanel({ document }: DocumentCommentsPanelProps) {
  const { user, activeTenant } = useAuth();
  const role = activeTenant?.role ?? "member";
  const confirm = useConfirm();
  const { menu, bindContextMenu } = useContextMenu();
  const [comments, setComments] = useState<DocumentComment[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [newComment, setNewComment] = useState("");
  const [replyingToId, setReplyingToId] = useState<string | null>(null);
  const [replyContent, setReplyContent] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editContent, setEditContent] = useState("");
  const [submittingId, setSubmittingId] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  const canManageComment = (comment: DocumentComment) => {
    if (!user) return false;
    const isPrivileged = activeTenant?.role === "owner" || activeTenant?.role === "admin";
    return comment.userId === user.id || isPrivileged;
  };

  const fetchComments = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setComments(await listDocumentComments(document.id));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Yorumlar yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchComments();
  }, [document.id]);

  const handleCreate = async (content: string, parentId?: string) => {
    const trimmed = content.trim();
    if (!trimmed) {
      toast.error("Yorum içeriği boş olamaz.");
      return;
    }

    setIsCreating(true);
    try {
      await createDocumentComment(document.id, trimmed, parentId);
      await fetchComments();
      setNewComment("");
      setReplyContent("");
      setReplyingToId(null);
      toast.success(parentId ? "Cevap eklendi." : "Yorum eklendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Yorum eklenemedi.");
    } finally {
      setIsCreating(false);
    }
  };

  const handleUpdate = async (commentId: string) => {
    const trimmed = editContent.trim();
    if (!trimmed) {
      toast.error("Yorum içeriği boş olamaz.");
      return;
    }

    setSubmittingId(commentId);
    try {
      await updateDocumentComment(document.id, commentId, trimmed);
      await fetchComments();
      setEditingId(null);
      setEditContent("");
      toast.success("Yorum güncellendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Yorum güncellenemedi.");
    } finally {
      setSubmittingId(null);
    }
  };

  const handleDelete = async (comment: DocumentComment) => {
    const shouldDelete = await confirm({
      title: "Yorumu sil",
      message: "Bu yorumu silmek istediğinize emin misiniz?",
      confirmLabel: "Sil",
      variant: "danger",
    });
    if (!shouldDelete) return;

    setSubmittingId(comment.id);
    try {
      await deleteDocumentComment(document.id, comment.id);
      await fetchComments();
      toast.success("Yorum silindi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Yorum silinemedi.");
    } finally {
      setSubmittingId(null);
    }
  };

  const handleContextMenuAction = useCallback(
    (detail: { action: string; context: { entityType: string; entity: unknown } }) => {
      if (detail.context.entityType !== "comment") return;
      const comment = detail.context.entity as DocumentComment;

      switch (detail.action) {
        case ContextMenuActions.comment.reply:
          setReplyingToId(comment.id);
          setReplyContent("");
          setEditingId(null);
          return;
        case ContextMenuActions.comment.edit:
          setEditingId(comment.id);
          setEditContent(comment.content);
          setReplyingToId(null);
          return;
        case ContextMenuActions.comment.delete:
          void handleDelete(comment);
          return;
      }
    },
    [],
  );

  useContextMenuActionHandler(handleContextMenuAction);

  const getCommentMenuProps = (comment: DocumentComment) =>
    bindContextMenu({
      items: buildCommentMenuItems({ comment, role, canManage: canManageComment(comment) }),
      context: getCommentMenuContext(comment, document.id),
    });

  return (
    <section className="flex flex-col">
      <div className="border-b border-line px-5 py-4">
        <h2 className="flex items-center gap-2 text-base font-bold text-ink">
          <MessageSquare className="h-4 w-4 text-violet-brand" />
          Yorumlar
        </h2>
        <p className="mt-1 text-sm text-slate-500">Belge üzerinde ekip içi notlar</p>
      </div>

      <div className="border-b border-line p-5">
        <label className="block space-y-2 text-sm">
          <span className="font-semibold text-slate-600">Yeni yorum</span>
          <textarea
            className="min-h-24 w-full rounded-lg border border-line px-3 py-2 outline-none focus:border-violet-brand"
            placeholder="Yorumunu yaz..."
            value={newComment}
            onChange={(event) => setNewComment(event.target.value)}
          />
        </label>
        <div className="mt-3 flex justify-end">
          <PrimaryButton disabled={isCreating || !newComment.trim()} onClick={() => void handleCreate(newComment)}>
            {isCreating ? <Spinner size="sm" label="Gönderiliyor..." className="text-white" /> : "Yorum Ekle"}
          </PrimaryButton>
        </div>
      </div>

      {isLoading ? (
        <div className="p-5">
          <LoadingState label="Yorumlar yükleniyor..." />
        </div>
      ) : loadError ? (
        <div className="p-5">
          <EmptyState title="Yorumlar yüklenemedi" description={loadError} action={<PrimaryButton onClick={() => void fetchComments()}>Tekrar Dene</PrimaryButton>} />
        </div>
      ) : comments.length === 0 ? (
        <div className="p-5">
          <EmptyState title="Bu belge için henüz yorum yok" description="İlk yorumu sen ekle." />
        </div>
      ) : (
        <div className="divide-y divide-line">
          {comments.map((comment) => (
            <CommentThread
              key={comment.id}
              comment={comment}
              depth={0}
              replyingToId={replyingToId}
              replyContent={replyContent}
              editingId={editingId}
              editContent={editContent}
              submittingId={submittingId}
              isCreating={isCreating}
              canManage={canManageComment}
              onReply={(id) => {
                setReplyingToId(id);
                setReplyContent("");
                setEditingId(null);
              }}
              onCancelReply={() => {
                setReplyingToId(null);
                setReplyContent("");
              }}
              onReplyContentChange={setReplyContent}
              onSubmitReply={(parentId) => void handleCreate(replyContent, parentId)}
              onEdit={(comment) => {
                setEditingId(comment.id);
                setEditContent(comment.content);
                setReplyingToId(null);
              }}
              onCancelEdit={() => {
                setEditingId(null);
                setEditContent("");
              }}
              onEditContentChange={setEditContent}
              onSubmitEdit={(commentId) => void handleUpdate(commentId)}
              onDelete={(comment) => void handleDelete(comment)}
              getCommentMenuProps={getCommentMenuProps}
            />
          ))}
        </div>
      )}
      {menu}
    </section>
  );
}

type CommentThreadProps = {
  comment: DocumentComment;
  depth: number;
  replyingToId: string | null;
  replyContent: string;
  editingId: string | null;
  editContent: string;
  submittingId: string | null;
  isCreating: boolean;
  canManage: (comment: DocumentComment) => boolean;
  onReply: (id: string) => void;
  onCancelReply: () => void;
  onReplyContentChange: (value: string) => void;
  onSubmitReply: (parentId: string) => void;
  onEdit: (comment: DocumentComment) => void;
  onCancelEdit: () => void;
  onEditContentChange: (value: string) => void;
  onSubmitEdit: (commentId: string) => void;
  onDelete: (comment: DocumentComment) => void;
  getCommentMenuProps: (comment: DocumentComment) => ReturnType<ReturnType<typeof useContextMenu>["bindContextMenu"]>;
};

function CommentThread({
  comment,
  depth,
  replyingToId,
  replyContent,
  editingId,
  editContent,
  submittingId,
  isCreating,
  canManage,
  onReply,
  onCancelReply,
  onReplyContentChange,
  onSubmitReply,
  onEdit,
  onCancelEdit,
  onEditContentChange,
  onSubmitEdit,
  onDelete,
  getCommentMenuProps,
}: CommentThreadProps) {
  const isEditing = editingId === comment.id;
  const isReplying = replyingToId === comment.id;
  const isBusy = submittingId === comment.id;
  const menuProps = getCommentMenuProps(comment);

  return (
    <div className={depth > 0 ? "border-t border-line/70" : ""}>
      <div
        className="space-y-3 px-5 py-4"
        style={{ paddingLeft: `${20 + depth * 16}px` }}
        onContextMenu={menuProps.onContextMenu}
        onTouchStart={menuProps.onTouchStart}
        onTouchMove={menuProps.onTouchMove}
        onTouchEnd={menuProps.onTouchEnd}
        onTouchCancel={menuProps.onTouchCancel}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-sm font-bold text-ink">{comment.userFullName || "Kullanıcı"}</span>
              {comment.isEdited ? (
                <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-500">Düzenlendi</span>
              ) : null}
            </div>
            <p className="mt-1 text-xs text-slate-500">{formatDateTime(comment.createdAt)}</p>
            {isEditing ? (
              <div className="mt-3 space-y-2">
                <textarea
                  className="min-h-20 w-full rounded-lg border border-line px-3 py-2 text-sm outline-none focus:border-violet-brand"
                  value={editContent}
                  onChange={(event) => onEditContentChange(event.target.value)}
                />
                <div className="flex justify-end gap-2">
                  <button type="button" className="h-9 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600" onClick={onCancelEdit}>
                    İptal
                  </button>
                  <PrimaryButton className="h-9 px-3 text-sm" disabled={isBusy || !editContent.trim()} onClick={() => onSubmitEdit(comment.id)}>
                    {isBusy ? <Spinner size="sm" /> : "Kaydet"}
                  </PrimaryButton>
                </div>
              </div>
            ) : (
              <p className="mt-2 whitespace-pre-wrap text-sm text-slate-700">{comment.content}</p>
            )}
          </div>
          {!isEditing ? (
            <div className="flex shrink-0 items-center gap-1">
              <button
                type="button"
                className="inline-flex h-8 items-center gap-1 rounded-lg px-2 text-xs font-semibold text-slate-500 transition hover:bg-violet-soft hover:text-violet-brand"
                onClick={() => onReply(comment.id)}
              >
                <Reply className="h-3.5 w-3.5" />
                Cevapla
              </button>
              {canManage(comment) ? (
                <>
                  <button
                    type="button"
                    className="inline-flex h-8 items-center gap-1 rounded-lg px-2 text-xs font-semibold text-slate-500 transition hover:bg-violet-soft hover:text-violet-brand"
                    onClick={() => onEdit(comment)}
                  >
                    <Pencil className="h-3.5 w-3.5" />
                    Düzenle
                  </button>
                  <button
                    type="button"
                    className="inline-flex h-8 items-center gap-1 rounded-lg px-2 text-xs font-semibold text-slate-500 transition hover:bg-red-50 hover:text-red-600 disabled:opacity-60"
                    disabled={isBusy}
                    onClick={() => onDelete(comment)}
                  >
                    {isBusy ? <Spinner size="sm" /> : <Trash2 className="h-3.5 w-3.5" />}
                    Sil
                  </button>
                </>
              ) : null}
            </div>
          ) : null}
        </div>

        {isReplying ? (
          <div className="space-y-2 rounded-lg border border-line bg-slate-50 p-3">
            <textarea
              className="min-h-20 w-full rounded-lg border border-line bg-white px-3 py-2 text-sm outline-none focus:border-violet-brand"
              placeholder="Cevabini yaz..."
              value={replyContent}
              onChange={(event) => onReplyContentChange(event.target.value)}
            />
            <div className="flex justify-end gap-2">
              <button type="button" className="h-9 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600" onClick={onCancelReply}>
                İptal
              </button>
              <PrimaryButton className="h-9 px-3 text-sm" disabled={isCreating || !replyContent.trim()} onClick={() => onSubmitReply(comment.id)}>
                {isCreating ? <Spinner size="sm" /> : "Cevap Gönder"}
              </PrimaryButton>
            </div>
          </div>
        ) : null}
      </div>

      {comment.replies.map((reply) => (
        <CommentThread
          key={reply.id}
          comment={reply}
          depth={depth + 1}
          replyingToId={replyingToId}
          replyContent={replyContent}
          editingId={editingId}
          editContent={editContent}
          submittingId={submittingId}
          isCreating={isCreating}
          canManage={canManage}
          onReply={onReply}
          onCancelReply={onCancelReply}
          onReplyContentChange={onReplyContentChange}
          onSubmitReply={onSubmitReply}
          onEdit={onEdit}
          onCancelEdit={onCancelEdit}
          onEditContentChange={onEditContentChange}
          onSubmitEdit={onSubmitEdit}
          onDelete={onDelete}
          getCommentMenuProps={getCommentMenuProps}
        />
      ))}
    </div>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
