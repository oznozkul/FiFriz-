import { useEffect, useState } from "react";
import type { Tag, TagBrief } from "../../types/tag";
import { listTags, setDocumentTags } from "../../services/tagService";
import { toast } from "../../services/toast";
import { LoadingState } from "../common";
import { TagBadgeList } from "./TagBadge";

type DocumentTagsPanelProps = {
  documentId: string;
  selectedTags: TagBrief[];
  onUpdated: (tags: TagBrief[]) => void;
};

export function DocumentTagsPanel({ documentId, selectedTags, onUpdated }: DocumentTagsPanelProps) {
  const [allTags, setAllTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    listTags()
      .then((items) => {
        if (!cancelled) setAllTags(items);
      })
      .catch((loadError) => {
        if (!cancelled) toast.error(loadError instanceof Error ? loadError.message : "Etiketler yüklenemedi.");
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const selectedIds = new Set(selectedTags.map((tag) => tag.id));

  const handleToggle = async (tag: Tag) => {
    const nextIds = selectedIds.has(tag.id) ? selectedTags.filter((item) => item.id !== tag.id).map((item) => item.id) : [...selectedTags.map((item) => item.id), tag.id];

    setIsSaving(true);
    try {
      const updated = (await setDocumentTags(documentId, nextIds)) as { tags?: TagBrief[] };
      onUpdated(updated.tags ?? []);
      toast.success("Etiketler güncellendi.");
    } catch (saveError) {
      toast.error(saveError instanceof Error ? saveError.message : "Etiketler güncellenemedi.");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <section className="panel overflow-hidden">
      <div className="border-b border-line px-5 py-4">
        <h2 className="text-base font-bold text-ink">Etiketler</h2>
        <p className="mt-1 text-sm text-slate-500">Belgeye tenant etiketleri atayın.</p>
      </div>

      <div className="space-y-4 px-5 py-4">
        <TagBadgeList tags={selectedTags} emptyLabel="Henüz etiket atanmadı." />

        {isLoading ? (
          <LoadingState variant="inline" label="Etiketler yükleniyor..." />
        ) : allTags.length === 0 ? (
          <p className="text-sm text-slate-500">Henüz etiket yok. Etiketler sayfasından oluşturabilirsiniz.</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {allTags.map((tag) => {
              const isSelected = selectedIds.has(tag.id);
              const color = tag.color || "#7c3aed";
              return (
                <button
                  key={tag.id}
                  className={`rounded-full border px-3 py-1.5 text-xs font-semibold transition ${isSelected ? "border-transparent text-white" : "border-line bg-white text-slate-600 hover:border-violet-brand hover:text-violet-brand"}`}
                  style={isSelected ? { backgroundColor: color } : undefined}
                  type="button"
                  disabled={isSaving}
                  onClick={() => void handleToggle(tag)}
                >
                  {tag.name}
                </button>
              );
            })}
          </div>
        )}
      </div>
    </section>
  );
}
