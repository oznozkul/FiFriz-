import { CalendarDays, ExternalLink, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { IconButton, PageState, PageTitle } from "../components/common";
import { useImportantDateLabels } from "../hooks/useImportantDateLabels";
import { listImportantDates } from "../services/importantDateService";
import { toast } from "../services/toast";
import type { ImportantDate, ImportantDateType } from "../types/importantDate";
import { IMPORTANT_DATE_TYPE_OPTIONS, importantDateTypeLabel } from "../types/importantDate";

type FilterMode = "upcoming" | "completed" | "assigned" | "all";

export function ImportantDatesPage() {
  const labels = useImportantDateLabels();
  const [items, setItems] = useState<ImportantDate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [filterMode, setFilterMode] = useState<FilterMode>("upcoming");
  const [typeFilter, setTypeFilter] = useState<ImportantDateType | "">("");
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");

  const fetchItems = useCallback(async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      const params =
        filterMode === "upcoming"
          ? { upcomingOnly: true, type: typeFilter || undefined, from: fromDate || undefined, to: toDate || undefined }
          : filterMode === "completed"
            ? { completed: true, type: typeFilter || undefined, from: fromDate || undefined, to: toDate || undefined }
            : filterMode === "assigned"
              ? { assignedToMe: true, type: typeFilter || undefined, from: fromDate || undefined, to: toDate || undefined }
              : { type: typeFilter || undefined, from: fromDate || undefined, to: toDate || undefined };

      setItems(await listImportantDates(params));
    } catch (error) {
      const message = error instanceof Error ? error.message : labels.loadFailed;
      setLoadError(message);
      setItems([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, [filterMode, typeFilter, fromDate, toDate, labels.loadFailed]);

  useEffect(() => {
    void fetchItems();
  }, [fetchItems]);

  const sectionTitle = useMemo(() => {
    switch (filterMode) {
      case "completed":
        return labels.completedTitle;
      case "assigned":
        return labels.assignedTitle;
      case "all":
        return labels.pageTitle;
      default:
        return labels.upcomingTitle;
    }
  }, [filterMode, labels]);

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title={labels.pageTitle} subtitle={labels.pageSubtitle} />
        <IconButton aria-label="Yenile" onClick={() => void fetchItems()}>
          <RefreshCw className="h-4 w-4" />
        </IconButton>
      </div>

      <section className="panel p-4">
        <div className="flex flex-wrap gap-2">
          {(
            [
              { id: "upcoming", label: labels.upcomingTitle },
              { id: "completed", label: labels.completedTitle },
              { id: "assigned", label: labels.assignedTitle },
              { id: "all", label: "Tümü" },
            ] as const
          ).map((tab) => (
            <button
              key={tab.id}
              className={`rounded-lg px-3 py-2 text-sm font-semibold transition ${filterMode === tab.id ? "bg-violet-soft text-violet-brand" : "text-slate-500 hover:bg-slate-50 hover:text-ink"}`}
              type="button"
              onClick={() => setFilterMode(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          <label className="space-y-2">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Tür</span>
            <select className="h-10 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" value={typeFilter} onChange={(e) => setTypeFilter(e.target.value as ImportantDateType | "")}>
              <option value="">Tümü</option>
              {IMPORTANT_DATE_TYPE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="space-y-2">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Başlangıç</span>
            <input className="h-10 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" type="date" value={fromDate} onChange={(e) => setFromDate(e.target.value)} />
          </label>
          <label className="space-y-2">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Bitiş</span>
            <input className="h-10 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" type="date" value={toDate} onChange={(e) => setToDate(e.target.value)} />
          </label>
        </div>
      </section>

      <section className="panel overflow-hidden">
        <div className="border-b border-line px-5 py-4">
          <h2 className="flex items-center gap-2 text-base font-bold text-ink">
            <CalendarDays className="h-4 w-4 text-violet-brand" />
            {sectionTitle}
          </h2>
        </div>

        <PageState
          isLoading={isLoading}
          error={loadError}
          isEmpty={items.length === 0}
          variant="inline"
          loadingLabel="Önemli tarihler yükleniyor..."
          errorTitle="Önemli tarihler yüklenemedi"
          onRetry={() => void fetchItems()}
          emptyTitle={labels.emptyTitle}
          emptyDescription={labels.emptyDescription}
          emptyIcon={CalendarDays}
        >
          <div className="divide-y divide-line">
            {items.map((item) => (
              <article key={item.id} className="flex flex-col gap-3 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-semibold text-ink">{item.title}</h3>
                    {item.isCompleted && <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700">Tamamlandı</span>}
                    {!item.isCompleted && isOverdue(item.date) && <span className="rounded-full bg-red-100 px-2 py-0.5 text-xs font-semibold text-red-700">Gecikmiş</span>}
                  </div>
                  <p className="mt-1 text-sm text-slate-600">{formatDisplayDate(item.date)}</p>
                  <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-500">
                    <span className="rounded-full bg-slate-50 px-2 py-1 font-semibold">{importantDateTypeLabel(item.type)}</span>
                    {item.documentName && <span className="rounded-full bg-slate-50 px-2 py-1">{item.documentName}</span>}
                    {item.assignedToUserName && (
                      <span className="rounded-full bg-slate-50 px-2 py-1">
                        {labels.assignedLabel}: {item.assignedToUserName}
                      </span>
                    )}
                    {item.reminderDaysBefore != null && <span className="rounded-full bg-slate-50 px-2 py-1">{labels.reminderLabel}: {item.reminderDaysBefore} gün önce</span>}
                  </div>
                </div>
                <Link className="inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded-lg border border-line bg-white px-4 text-sm font-semibold text-violet-brand transition hover:border-violet-brand hover:bg-violet-soft/40" to={`/documents/${item.documentId}`}>
                  <ExternalLink className="h-4 w-4" />
                  {labels.goToDocument}
                </Link>
              </article>
            ))}
          </div>
        </PageState>
      </section>
    </div>
  );
}

function formatDisplayDate(value: string) {
  const parsed = new Date(`${value}T00:00:00`);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleDateString("tr-TR", { day: "numeric", month: "long", year: "numeric" });
}

function isOverdue(value: string) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const date = new Date(`${value}T00:00:00`);
  return date < today;
}
