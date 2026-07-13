import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { AppModal, Button, LoadingState, PrimaryButton, StatusBadge } from "../common";
import { TagBadge } from "./TagBadge";
import { createTagAlias, deleteTagAlias, getTagStats, listDocumentsByTag, listTagAliases } from "../../services/tagService";
import type { ApiDocument } from "../../types/document";
import type { Tag, TagAlias, TagStats } from "../../types/tag";
import { toast } from "../../services/toast";

type TagDetailModalProps = {
  tag: Tag;
  allTags: Tag[];
  canManageTags: boolean;
  onClose: () => void;
  onUpdated: () => void;
  onMerge: (tag: Tag) => void;
};

export function TagDetailModal({ tag, allTags, canManageTags, onClose, onUpdated, onMerge }: TagDetailModalProps) {
  const navigate = useNavigate();
  const [stats, setStats] = useState<TagStats | null>(null);
  const [aliases, setAliases] = useState<TagAlias[]>([]);
  const [documents, setDocuments] = useState<ApiDocument[]>([]);
  const [aliasInput, setAliasInput] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setIsLoading(true);
    void Promise.all([getTagStats(tag.id), listTagAliases(tag.id), listDocumentsByTag(tag.id)])
      .then(([statsData, aliasData, docs]) => {
        if (!active) return;
        setStats(statsData);
        setAliases(aliasData);
        setDocuments(Array.isArray(docs) ? docs : []);
      })
      .catch((error) => {
        toast.error(error instanceof Error ? error.message : "Etiket detayları yüklenemedi.");
      })
      .finally(() => {
        if (active) setIsLoading(false);
      });
    return () => {
      active = false;
    };
  }, [tag.id]);

  const handleAddAlias = async () => {
    const trimmed = aliasInput.trim();
    if (!trimmed) return;
    try {
      await createTagAlias(tag.id, trimmed);
      setAliasInput("");
      setAliases(await listTagAliases(tag.id));
      onUpdated();
      toast.success("Alias eklendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Alias eklenemedi.");
    }
  };

  const handleDeleteAlias = async (alias: TagAlias) => {
    try {
      await deleteTagAlias(tag.id, alias.id);
      setAliases((current) => current.filter((item) => item.id !== alias.id));
      onUpdated();
      toast.success("Alias silindi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Alias silinemedi.");
    }
  };

  return (
    <AppModal title={tag.name} onClose={onClose} size="lg">
      {isLoading ? (
        <LoadingState variant="inline" label="Etiket detayları yükleniyor..." />
      ) : (
        <div className="space-y-5">
          <div className="flex flex-wrap items-center gap-3">
            <TagBadge tag={tag} />
            <StatusBadge status={tag.lifecycleStatus ?? tag.status ?? "active"} />
            {tag.isSystem ? <span className="text-xs font-semibold text-violet-brand">Sistem etiketi</span> : null}
          </div>

          {stats ? (
            <dl className="grid gap-3 text-sm sm:grid-cols-2">
              <InfoItem label="Kullanım" value={String(stats.documentCount)} />
              <InfoItem label="Son kullanım" value={stats.lastUsedAt ? new Date(stats.lastUsedAt).toLocaleString("tr-TR") : "-"} />
              <InfoItem label="Oluşturulma" value={new Date(stats.createdAt).toLocaleString("tr-TR")} />
            </dl>
          ) : null}

          <section className="space-y-2">
            <h3 className="text-sm font-bold text-ink">Aliaslar</h3>
            <div className="flex flex-wrap gap-2">
              {aliases.length === 0 ? <p className="text-sm text-slate-500">Henüz alias yok.</p> : null}
              {aliases.map((alias) => (
                <span key={alias.id} className="inline-flex items-center gap-2 rounded-full border border-line bg-slate-50 px-3 py-1 text-sm font-semibold text-slate-600">
                  {alias.alias}
                  {canManageTags ? (
                    <button type="button" className="text-rose-500" onClick={() => void handleDeleteAlias(alias)}>
                      ×
                    </button>
                  ) : null}
                </span>
              ))}
            </div>
            {canManageTags ? (
              <div className="flex gap-2">
                <input
                  className="h-10 flex-1 rounded-lg border border-line px-3 text-sm font-semibold"
                  placeholder="Yeni alias"
                  value={aliasInput}
                  onChange={(event) => setAliasInput(event.target.value)}
                />
                <Button variant="secondary" onClick={() => void handleAddAlias()}>
                  Ekle
                </Button>
              </div>
            ) : null}
          </section>

          {stats?.relatedTags?.length ? (
            <section className="space-y-2">
              <h3 className="text-sm font-bold text-ink">İlişkili etiketler</h3>
              <div className="flex flex-wrap gap-2">
                {stats.relatedTags.map((related) => (
                  <TagBadge key={related.id} tag={related} />
                ))}
              </div>
            </section>
          ) : null}

          <section className="space-y-2">
            <h3 className="text-sm font-bold text-ink">Bağlı belgeler</h3>
            {documents.length === 0 ? (
              <p className="text-sm text-slate-500">Bu etikete bağlı belge yok.</p>
            ) : (
              <ul className="divide-y divide-line rounded-lg border border-line">
                {documents.slice(0, 8).map((document) => (
                  <li key={document.id}>
                    <button
                      type="button"
                      className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-semibold text-ink hover:bg-violet-soft/30"
                      onClick={() => navigate(`/documents/${document.id}`)}
                    >
                      <span className="truncate">{document.original_file_name}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>

          {canManageTags && !tag.isSystem ? (
            <div className="flex justify-end gap-2 border-t border-line pt-4">
              <Button variant="secondary" onClick={() => onMerge(tag)}>
                Birleştir
              </Button>
            </div>
          ) : null}
        </div>
      )}
    </AppModal>
  );
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-line bg-slate-50 px-3 py-2">
      <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</dt>
      <dd className="mt-1 font-semibold text-ink">{value}</dd>
    </div>
  );
}
