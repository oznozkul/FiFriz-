import { Archive, Eye, Merge, Pencil, Plus, RefreshCw, Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Button, IconButton, PageState, PageTitle, PrimaryButton, StatusBadge, useConfirm } from "../components/common";
import { TagBadge, TagDetailModal } from "../components/tags";
import {
  buildTagMenuItems,
  ContextMenuActions,
  getTagMenuContext,
  useContextMenuActionHandler,
} from "../contextMenu";
import { useContextMenu } from "../hooks/useContextMenu";
import { createTag, deleteTag, listTags, mergeTag, updateTag } from "../services/tagService";
import { useAuth } from "../services/AuthContext";
import { toast } from "../services/toast";
import type { Tag } from "../types/tag";
import { tagColorOptions, tagIconOptions } from "../types/tag";

export function TagsPage() {
  const { activeTenant } = useAuth();
  const role = activeTenant?.role ?? "member";
  const canManageTags = activeTenant?.role === "owner" || activeTenant?.role === "admin";
  const canCreateTags = true;
  const { menu, bindContextMenu } = useContextMenu();
  const confirm = useConfirm();
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [modalMode, setModalMode] = useState<"create" | "edit" | null>(null);
  const [editingTag, setEditingTag] = useState<Tag | null>(null);
  const [detailTag, setDetailTag] = useState<Tag | null>(null);
  const [mergeSource, setMergeSource] = useState<Tag | null>(null);
  const [mergeTargetId, setMergeTargetId] = useState("");
  const [name, setName] = useState("");
  const [color, setColor] = useState<string>(tagColorOptions[0].value);
  const [icon, setIcon] = useState<string>(tagIconOptions[0].value);
  const [status, setStatus] = useState("active");
  const [isSaving, setIsSaving] = useState(false);

  const fetchTags = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setTags(await listTags());
    } catch (fetchError) {
      const message = fetchError instanceof Error ? fetchError.message : "Etiketler yüklenemedi.";
      setLoadError(message);
      setTags([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchTags();
  }, []);

  const openCreateModal = () => {
    setEditingTag(null);
    setName("");
    setColor(tagColorOptions[0].value);
    setIcon(tagIconOptions[0].value);
    setStatus("active");
    setModalMode("create");
  };

  const openEditModal = (tag: Tag) => {
    setEditingTag(tag);
    setName(tag.name);
    setColor(tag.color || tagColorOptions[0].value);
    setIcon(tag.icon || tagIconOptions[0].value);
    setStatus(tag.status || "active");
    setModalMode("edit");
  };

  const closeModal = () => {
    if (isSaving) return;
    setModalMode(null);
    setEditingTag(null);
  };

  const handleSave = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) return;

    setIsSaving(true);
    try {
      if (modalMode === "edit" && editingTag) {
        await updateTag(editingTag.id, { name: trimmedName, color, icon, status: canManageTags ? status : editingTag.status });
        toast.success("Etiket güncellendi.");
      } else {
        await createTag({ name: trimmedName, color, icon });
        toast.success("Etiket oluşturuldu.");
      }
      closeModal();
      await fetchTags();
    } catch (saveError) {
      toast.error(saveError instanceof Error ? saveError.message : "Etiket kaydedilemedi.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleArchive = async (tag: Tag) => {
    try {
      await updateTag(tag.id, { name: tag.name, color: tag.color || undefined, status: "archived", icon: tag.icon });
      await fetchTags();
      toast.success("Etiket arşivlendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Etiket arşivlenemedi.");
    }
  };

  const handleDelete = async (tag: Tag) => {
    if (tag.isSystem) {
      toast.error("Sistem etiketi silinemez.");
      return;
    }
    const shouldDelete = await confirm({
      title: "Etiketi sil",
      message: `${tag.name} etiketi silinsin mi?`,
      confirmLabel: "Sil",
      variant: "danger",
    });
    if (!shouldDelete) return;
    try {
      await deleteTag(tag.id);
      await fetchTags();
      toast.success("Etiket silindi.");
    } catch (deleteError) {
      toast.error(deleteError instanceof Error ? deleteError.message : "Etiket silinemedi.");
    }
  };

  const handleMerge = async () => {
    if (!mergeSource || !mergeTargetId) return;
    setIsSaving(true);
    try {
      await mergeTag(mergeSource.id, mergeTargetId);
      setMergeSource(null);
      setMergeTargetId("");
      setDetailTag(null);
      await fetchTags();
      toast.success("Etiketler birleştirildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Etiket birleştirilemedi.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleContextMenuAction = useCallback(
    (detail: { action: string; context: { entityType: string; entity: unknown } }) => {
      if (detail.context.entityType !== "tag") return;
      const tag = detail.context.entity as Tag;

      switch (detail.action) {
        case ContextMenuActions.tag.edit:
        case ContextMenuActions.tag.changeColor:
          openEditModal(tag);
          return;
        case ContextMenuActions.tag.delete:
          void handleDelete(tag);
          return;
        case ContextMenuActions.tag.merge:
          setMergeSource(tag);
          return;
      }
    },
    [],
  );

  useContextMenuActionHandler(handleContextMenuAction);

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title="Etiketler" subtitle="Tenant içinde belgelerinizi gruplamak için etiketler oluşturun." />
        <div className="flex gap-2">
          <IconButton aria-label="Yenile" onClick={() => void fetchTags()}>
            <RefreshCw className="h-4 w-4" />
          </IconButton>
          {canCreateTags ? (
            <PrimaryButton onClick={openCreateModal}>
              <Plus className="h-4 w-4" />
              Yeni Etiket
            </PrimaryButton>
          ) : null}
        </div>
      </div>

      <PageState
        isLoading={isLoading}
        error={loadError}
        isEmpty={tags.length === 0}
        loadingLabel="Etiketler yükleniyor..."
        loadingVariant="table"
        errorTitle="Etiketler yüklenemedi"
        onRetry={() => void fetchTags()}
        emptyTitle="Henüz etiket yok"
        emptyDescription="İlk etiketinizi oluşturduğunuzda burada listelenecek."
        emptyAction={
          canCreateTags ? (
            <PrimaryButton onClick={openCreateModal}>
              <Plus className="h-4 w-4" />
              Yeni Etiket
            </PrimaryButton>
          ) : undefined
        }
      >
        <section className="panel overflow-hidden rounded-2xl">
          <div className="divide-y divide-line">
            {tags.map((tag) => (
              <div
                key={tag.id}
                className="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-center sm:justify-between"
                {...bindContextMenu({
                  items: buildTagMenuItems({ tag, role, canManageTags }),
                  context: getTagMenuContext(tag),
                })}
              >
                <button type="button" className="flex min-w-0 flex-1 items-center gap-3 text-left" onClick={() => setDetailTag(tag)}>
                  <TagBadge tag={tag} />
                  <StatusBadge status={tag.lifecycleStatus ?? tag.status ?? "active"} />
                  <span className="text-xs text-slate-400">{tag.documentCount ?? 0} belge</span>
                  <span className="text-xs text-slate-400">{formatDate(tag.createdAt)}</span>
                </button>
                <div className="flex flex-wrap gap-2">
                  <button className="inline-flex h-9 items-center gap-1 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600 transition hover:bg-violet-soft/40" type="button" onClick={() => setDetailTag(tag)}>
                    <Eye className="h-4 w-4" />
                    Detay
                  </button>
                  {canManageTags && !tag.isSystem ? (
                    <>
                      <button className="inline-flex h-9 items-center gap-1 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600 transition hover:bg-violet-soft/40" type="button" onClick={() => openEditModal(tag)}>
                        <Pencil className="h-4 w-4" />
                        Düzenle
                      </button>
                      <button className="inline-flex h-9 items-center gap-1 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600 transition hover:bg-violet-soft/40" type="button" onClick={() => void handleArchive(tag)}>
                        <Archive className="h-4 w-4" />
                        Arşivle
                      </button>
                      <button className="inline-flex h-9 items-center gap-1 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600 transition hover:bg-violet-soft/40" type="button" onClick={() => setMergeSource(tag)}>
                        <Merge className="h-4 w-4" />
                        Birleştir
                      </button>
                      <button className="inline-flex h-9 items-center gap-1 rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-600 transition hover:bg-red-50" type="button" onClick={() => void handleDelete(tag)}>
                        <Trash2 className="h-4 w-4" />
                        Sil
                      </button>
                    </>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        </section>
      </PageState>

      {menu}

      {modalMode && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="panel w-full max-w-md p-5">
            <h2 className="text-lg font-bold text-ink">{modalMode === "edit" ? "Etiketi Düzenle" : "Yeni Etiket"}</h2>
            <div className="mt-4 space-y-4">
              <label className="block space-y-1 text-sm">
                <span className="font-semibold text-slate-500">Etiket adı</span>
                <input className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-700 outline-none focus:border-violet-brand" value={name} onChange={(event) => setName(event.target.value)} />
              </label>
              {canManageTags && modalMode === "edit" ? (
                <label className="block space-y-1 text-sm">
                  <span className="font-semibold text-slate-500">Durum</span>
                  <select className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-700" value={status} onChange={(event) => setStatus(event.target.value)}>
                    <option value="active">Aktif</option>
                    <option value="archived">Arşivlendi</option>
                    <option value="hidden">Gizli</option>
                  </select>
                </label>
              ) : null}
              <div className="space-y-2 text-sm">
                <span className="font-semibold text-slate-500">Renk</span>
                <div className="flex flex-wrap gap-2">
                  {tagColorOptions.map((option) => (
                    <button
                      key={option.value}
                      className={`h-9 w-9 rounded-full border-2 transition ${color === option.value ? "border-ink scale-110" : "border-transparent"}`}
                      style={{ backgroundColor: option.value }}
                      type="button"
                      aria-label={option.label}
                      onClick={() => setColor(option.value)}
                    />
                  ))}
                </div>
              </div>
              <div className="space-y-2 text-sm">
                <span className="font-semibold text-slate-500">İkon</span>
                <div className="flex flex-wrap gap-2">
                  {tagIconOptions.map((option) => (
                    <button
                      key={option.value}
                      className={`rounded-lg border px-3 py-2 text-xs font-semibold ${icon === option.value ? "border-violet-brand bg-violet-soft text-violet-brand" : "border-line text-slate-600"}`}
                      type="button"
                      onClick={() => setIcon(option.value)}
                    >
                      {option.label}
                    </button>
                  ))}
                </div>
              </div>
            </div>
            <div className="mt-5 flex justify-end gap-2">
              <button className="inline-flex h-10 items-center rounded-lg border border-line px-4 text-sm font-semibold text-slate-600" type="button" onClick={closeModal} disabled={isSaving}>
                İptal
              </button>
              <PrimaryButton onClick={() => void handleSave()} disabled={isSaving || !name.trim()}>
                {isSaving ? "Kaydediliyor..." : "Kaydet"}
              </PrimaryButton>
            </div>
          </div>
        </div>
      )}

      {detailTag ? (
        <TagDetailModal
          tag={detailTag}
          allTags={tags}
          canManageTags={canManageTags}
          onClose={() => setDetailTag(null)}
          onUpdated={() => void fetchTags()}
          onMerge={(tag) => {
            setDetailTag(null);
            setMergeSource(tag);
          }}
        />
      ) : null}

      {mergeSource ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="panel w-full max-w-md p-5">
            <h2 className="text-lg font-bold text-ink">"{mergeSource.name}" birleştir</h2>
            <label className="mt-4 block space-y-1 text-sm">
              <span className="font-semibold text-slate-500">Hedef etiket</span>
              <select className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-700" value={mergeTargetId} onChange={(event) => setMergeTargetId(event.target.value)}>
                <option value="">Seçin</option>
                {tags
                  .filter((tag) => tag.id !== mergeSource.id)
                  .map((tag) => (
                    <option key={tag.id} value={tag.id}>
                      {tag.name}
                    </option>
                  ))}
              </select>
            </label>
            <div className="mt-5 flex justify-end gap-2">
              <Button variant="secondary" onClick={() => setMergeSource(null)}>
                İptal
              </Button>
              <PrimaryButton disabled={!mergeTargetId || isSaving} onClick={() => void handleMerge()}>
                Birleştir
              </PrimaryButton>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}
