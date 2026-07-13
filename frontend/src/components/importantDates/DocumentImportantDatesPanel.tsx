import { CalendarDays, Check, Loader2, Pencil, Plus, Sparkles, Trash2, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { ApiDocument } from "../../types/document";
import { EmptyState, LoadingState, PrimaryButton, Spinner, useConfirm } from "../common";
import { useImportantDateLabels } from "../../hooks/useImportantDateLabels";
import {
  acceptImportantDateSuggestion,
  completeDocumentImportantDate,
  createDocumentImportantDate,
  deleteDocumentImportantDate,
  extractAIImportantDates,
  listDocumentImportantDates,
  listImportantDateSuggestions,
  rejectImportantDateSuggestion,
  updateDocumentImportantDate,
} from "../../services/importantDateService";
import { useAuth } from "../../services/AuthContext";
import { toast } from "../../services/toast";
import type { ImportantDate, ImportantDateInput } from "../../types/importantDate";
import { importantDateTypeLabel } from "../../types/importantDate";
import { ImportantDateFormModal } from "./ImportantDateFormModal";

type DocumentImportantDatesPanelProps = {
  document: ApiDocument;
  refreshToken?: number;
};

function formatConfidence(confidence?: number | null) {
  if (confidence == null || Number.isNaN(confidence)) return null;
  return `${Math.round(confidence * 100)}%`;
}

export function DocumentImportantDatesPanel({ document, refreshToken = 0 }: DocumentImportantDatesPanelProps) {
  const labels = useImportantDateLabels();
  const { user, activeTenant } = useAuth();
  const confirm = useConfirm();
  const [items, setItems] = useState<ImportantDate[]>([]);
  const [suggestions, setSuggestions] = useState<ImportantDate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isExtracting, setIsExtracting] = useState(false);
  const [editing, setEditing] = useState<ImportantDate | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [actionId, setActionId] = useState<string | null>(null);

  const isPrivileged = activeTenant?.role === "owner" || activeTenant?.role === "admin";

  const fetchItems = async () => {
    setIsLoading(true);
    try {
      const [activeItems, suggestedItems] = await Promise.all([
        listDocumentImportantDates(document.id),
        listImportantDateSuggestions(document.id),
      ]);
      setItems(activeItems);
      setSuggestions(suggestedItems);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : labels.loadFailed);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchItems();
  }, [document.id, refreshToken]);

  const canEdit = isPrivileged || user?.id === document.uploaded_by_user_id;
  const canDelete = (item: ImportantDate) => item.createdByUserId === user?.id || isPrivileged;
  const canComplete = (item: ImportantDate) =>
    !item.isCompleted && (item.createdByUserId === user?.id || item.assignedToUserId === user?.id || isPrivileged);

  const handleSave = async (input: ImportantDateInput) => {
    setIsSaving(true);
    try {
      if (editing) {
        await updateDocumentImportantDate(document.id, editing.id, input);
        toast.success(labels.updatedToast);
      } else {
        await createDocumentImportantDate(document.id, input);
        toast.success(labels.createdToast);
      }
      setEditing(null);
      setIsCreating(false);
      await fetchItems();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : editing ? labels.updateFailed : labels.createFailed);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (item: ImportantDate) => {
    const shouldDelete = await confirm({
      title: "Önemli tarihi sil",
      message: labels.deleteConfirm(item.title),
      confirmLabel: "Sil",
      variant: "danger",
    });
    if (!shouldDelete) return;
    setActionId(item.id);
    try {
      await deleteDocumentImportantDate(document.id, item.id);
      toast.success(labels.deletedToast);
      await fetchItems();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : labels.deleteFailed);
    } finally {
      setActionId(null);
    }
  };

  const handleComplete = async (item: ImportantDate) => {
    setActionId(item.id);
    try {
      await completeDocumentImportantDate(document.id, item.id);
      toast.success(labels.completedToast);
      await fetchItems();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : labels.completeFailed);
    } finally {
      setActionId(null);
    }
  };

  const handleExtractAI = async () => {
    setIsExtracting(true);
    try {
      await extractAIImportantDates(document.id);
      toast.success("AI tarih önerileri oluşturuldu.");
      await fetchItems();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "AI tarih çıkarma başarısız oldu.");
    } finally {
      setIsExtracting(false);
    }
  };

  const handleAcceptSuggestion = async (item: ImportantDate) => {
    setActionId(item.id);
    try {
      await acceptImportantDateSuggestion(document.id, item.id);
      toast.success("Tarih önerisi kabul edildi.");
      await fetchItems();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Tarih önerisi kabul edilemedi.");
    } finally {
      setActionId(null);
    }
  };

  const handleRejectSuggestion = async (item: ImportantDate) => {
    setActionId(item.id);
    try {
      await rejectImportantDateSuggestion(document.id, item.id);
      toast.success("Tarih önerisi reddedildi.");
      await fetchItems();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Tarih önerisi reddedilemedi.");
    } finally {
      setActionId(null);
    }
  };

  return (
    <section className="flex flex-col">
      <div className="flex items-start justify-between gap-3 border-b border-line px-5 py-4">
        <div>
          <h2 className="flex items-center gap-2 text-base font-bold text-ink">
            <CalendarDays className="h-4 w-4 text-violet-brand" />
            {labels.panelTitle}
          </h2>
          <p className="mt-1 text-sm text-slate-500">{labels.panelSubtitle}</p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row">
          {canEdit ? (
            <button
              type="button"
              className="inline-flex h-9 items-center gap-2 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-violet-brand transition hover:border-violet-ring hover:bg-violet-soft disabled:opacity-60"
              disabled={isExtracting}
              onClick={() => void handleExtractAI()}
            >
              {isExtracting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Sparkles className="h-3.5 w-3.5" />}
              AI ile Tarihleri Bul
            </button>
          ) : null}
          {canEdit ? (
            <PrimaryButton className="h-9 px-3 text-xs" onClick={() => setIsCreating(true)}>
              <Plus className="h-3.5 w-3.5" />
              {labels.addLabel}
            </PrimaryButton>
          ) : null}
        </div>
      </div>

      <div className="space-y-6 p-5">
        {isLoading ? (
          <LoadingState variant="inline" label="Yükleniyor..." />
        ) : (
          <>
            <section>
              <h3 className="text-xs font-bold uppercase tracking-wide text-slate-400">Aktif Tarihler</h3>
              {items.length === 0 ? (
                <p className="mt-3 text-sm text-slate-500">Henüz aktif önemli tarih yok.</p>
              ) : (
                <div className="mt-3 space-y-3">
                  {items.map((item) => (
                    <ImportantDateCard
                      key={item.id}
                      item={item}
                      actionId={actionId}
                      labels={labels}
                      canEdit={canEdit}
                      canDelete={canDelete(item)}
                      canComplete={canComplete(item)}
                      onComplete={() => void handleComplete(item)}
                      onEdit={() => setEditing(item)}
                      onDelete={() => void handleDelete(item)}
                    />
                  ))}
                </div>
              )}
            </section>

            <section>
              <h3 className="text-xs font-bold uppercase tracking-wide text-slate-400">AI Önerileri</h3>
              {suggestions.length === 0 ? (
                <p className="mt-3 text-sm text-slate-500">Henüz AI tarih önerisi yok.</p>
              ) : (
                <div className="mt-3 space-y-3">
                  {suggestions.map((item) => {
                    const confidenceLabel = formatConfidence(item.confidence);
                    return (
                      <article key={item.id} className="rounded-lg border border-violet-200 bg-violet-50/40 p-4">
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0">
                            <div className="flex flex-wrap items-center gap-2">
                              <h3 className="font-semibold text-ink">{item.title}</h3>
                              <span className="rounded-full bg-violet-100 px-2 py-0.5 text-xs font-semibold text-violet-700">AI Önerisi</span>
                            </div>
                            <p className="mt-1 text-sm text-slate-600">{formatDisplayDate(item.date)}</p>
                            <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-500">
                              <span className="rounded-full bg-white px-2 py-1 font-semibold">{importantDateTypeLabel(item.type)}</span>
                              {confidenceLabel ? <span className="rounded-full bg-white px-2 py-1">Güven: {confidenceLabel}</span> : null}
                            </div>
                            {item.sourceText ? <p className="mt-2 text-sm italic text-slate-500">"{item.sourceText}"</p> : null}
                          </div>
                          {canEdit ? (
                            <div className="flex shrink-0 gap-2">
                              <button
                                type="button"
                                className="inline-flex h-8 items-center gap-1 rounded-lg border border-emerald-200 bg-emerald-50 px-2.5 text-xs font-semibold text-emerald-700 transition hover:bg-emerald-100 disabled:opacity-60"
                                disabled={actionId === item.id || isExtracting}
                                onClick={() => void handleAcceptSuggestion(item)}
                              >
                                {actionId === item.id ? <Spinner size="sm" /> : <Check className="h-3.5 w-3.5" />}
                                Kabul Et
                              </button>
                              <button
                                type="button"
                                className="inline-flex h-8 items-center gap-1 rounded-lg border border-line bg-white px-2.5 text-xs font-semibold text-slate-600 transition hover:border-rose-200 hover:bg-rose-50 hover:text-rose-600 disabled:opacity-60"
                                disabled={actionId === item.id || isExtracting}
                                onClick={() => void handleRejectSuggestion(item)}
                              >
                                <X className="h-3.5 w-3.5" />
                                Reddet
                              </button>
                            </div>
                          ) : null}
                        </div>
                      </article>
                    );
                  })}
                </div>
              )}
            </section>

            {items.length === 0 && suggestions.length === 0 ? (
              <EmptyState title={labels.emptyTitle} description={labels.emptyDescription} />
            ) : null}
          </>
        )}
      </div>

      {(isCreating || editing) && (
        <ImportantDateFormModal initial={editing} isSaving={isSaving} onClose={() => { setIsCreating(false); setEditing(null); }} onSave={(input) => void handleSave(input)} />
      )}
    </section>
  );
}

function ImportantDateCard({
  item,
  actionId,
  labels,
  canEdit,
  canDelete,
  canComplete,
  onComplete,
  onEdit,
  onDelete,
}: {
  item: ImportantDate;
  actionId: string | null;
  labels: ReturnType<typeof useImportantDateLabels>;
  canEdit: boolean;
  canDelete: boolean;
  canComplete: boolean;
  onComplete: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <article className="rounded-lg border border-line bg-slate-50 p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="font-semibold text-ink">{item.title}</h3>
            {item.isCompleted ? (
              <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700">Tamamlandı</span>
            ) : null}
          </div>
          <p className="mt-1 text-sm text-slate-600">{formatDisplayDate(item.date)}</p>
          <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-500">
            <span className="rounded-full bg-white px-2 py-1 font-semibold">{importantDateTypeLabel(item.type)}</span>
            {item.reminderDaysBefore != null ? <span className="rounded-full bg-white px-2 py-1">{labels.reminderLabel}: {item.reminderDaysBefore} gün önce</span> : null}
            {item.assignedToUserName ? (
              <span className="rounded-full bg-white px-2 py-1">
                {labels.assignedLabel}: {item.assignedToUserName}
              </span>
            ) : null}
          </div>
          {item.note ? <p className="mt-2 text-sm text-slate-500">{item.note}</p> : null}
        </div>
        <div className="flex shrink-0 gap-1">
          {canComplete ? (
            <button className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-emerald-600 transition hover:bg-emerald-50" type="button" aria-label="Tamamla" disabled={actionId === item.id} onClick={onComplete}>
              {actionId === item.id ? <Spinner size="sm" /> : <Check className="h-4 w-4" />}
            </button>
          ) : null}
          {canEdit ? (
            <button className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-500 transition hover:bg-white" type="button" aria-label="Düzenle" onClick={onEdit}>
              <Pencil className="h-4 w-4" />
            </button>
          ) : null}
          {canDelete ? (
            <button className="inline-flex h-8 w-8 items-center justify-center rounded-lg text-red-500 transition hover:bg-red-50" type="button" aria-label="Sil" disabled={actionId === item.id} onClick={onDelete}>
              <Trash2 className="h-4 w-4" />
            </button>
          ) : null}
        </div>
      </div>
    </article>
  );
}

function formatDisplayDate(value: string) {
  const parsed = new Date(`${value}T00:00:00`);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleDateString("tr-TR", { day: "numeric", month: "long", year: "numeric" });
}
