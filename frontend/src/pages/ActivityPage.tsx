import { ChevronLeft, ChevronRight, RefreshCw } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { EmptyState, IconButton, LoadingState, PageTitle, PrimaryButton } from "../components/common";
import { auditActionLabels, auditEntityTypeLabels, formatAuditAction, formatAuditDescription, listAuditLogs, type AuditLog } from "../services/auditService";
import { toast } from "../services/toast";

const actionOptions = Object.entries(auditActionLabels).map(([value, label]) => ({ value, label }));
const entityTypeOptions = Object.entries(auditEntityTypeLabels).map(([value, label]) => ({ value, label }));

export function ActivityPage() {
  const [items, setItems] = useState<AuditLog[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [action, setAction] = useState("");
  const [entityType, setEntityType] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  const totalPages = useMemo(() => Math.max(1, Math.ceil(totalCount / pageSize)), [totalCount, pageSize]);

  const fetchLogs = async (targetPage = page) => {
    setIsLoading(true);
    try {
      const response = await listAuditLogs({
        action: action || undefined,
        entityType: entityType || undefined,
        startDate: startDate ? new Date(startDate).toISOString() : undefined,
        endDate: endDate ? new Date(`${endDate}T23:59:59`).toISOString() : undefined,
        page: targetPage,
        pageSize,
      });
      setItems(response.items);
      setTotalCount(response.totalCount);
      setPage(response.page);
    } catch (fetchError) {
      toast.error(fetchError instanceof Error ? fetchError.message : "Aktivite kayıtları yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchLogs(1);
  }, [action, entityType, startDate, endDate]);

  const handleApplyFilters = () => {
    setPage(1);
    void fetchLogs(1);
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title="Aktivite Geçmişi" subtitle="Tenant içindeki önemli kullanıcı işlemlerini izleyin." />
        <IconButton aria-label="Yenile" onClick={() => void fetchLogs(page)}>
          <RefreshCw className="h-4 w-4" />
        </IconButton>
      </div>

      <section className="panel p-4">
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <label className="space-y-1 text-sm">
            <span className="font-semibold text-slate-500">İşlem turu</span>
            <select className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-600 outline-none focus:border-violet-brand" value={action} onChange={(event) => setAction(event.target.value)}>
              <option value="">Tümü</option>
              {actionOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-1 text-sm">
            <span className="font-semibold text-slate-500">Varlık türü</span>
            <select className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-600 outline-none focus:border-violet-brand" value={entityType} onChange={(event) => setEntityType(event.target.value)}>
              <option value="">Tümü</option>
              {entityTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-1 text-sm">
            <span className="font-semibold text-slate-500">Başlangıç tarihi</span>
            <input className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-600 outline-none focus:border-violet-brand" type="date" value={startDate} onChange={(event) => setStartDate(event.target.value)} />
          </label>

          <label className="space-y-1 text-sm">
            <span className="font-semibold text-slate-500">Bitis tarihi</span>
            <input className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-600 outline-none focus:border-violet-brand" type="date" value={endDate} onChange={(event) => setEndDate(event.target.value)} />
          </label>
        </div>
        <div className="mt-3">
          <PrimaryButton type="button" onClick={handleApplyFilters}>
            Filtreleri uygula
          </PrimaryButton>
        </div>
      </section>

      {isLoading ? (
        <LoadingState label="Aktivite kayıtları yükleniyor..." />
      ) : items.length === 0 ? (
        <EmptyState title="Henüz aktivite kayıdı yok." description="Önemli işlemler burada listelenecek." />
      ) : (
        <section className="panel overflow-hidden">
          <div className="hidden overflow-x-auto md:block">
            <table className="min-w-full text-sm">
              <thead className="border-b border-line bg-slate-50/80 text-left text-xs font-semibold uppercase tracking-wide text-slate-400">
                <tr>
                  <th className="px-5 py-3">İşlem</th>
                  <th className="px-5 py-3">Kullanıcı</th>
                  <th className="px-5 py-3">Açıklama</th>
                  <th className="px-5 py-3">Tarih</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {items.map((item) => (
                  <tr key={item.id} className="hover:bg-violet-soft/30">
                    <td className="px-5 py-4 font-semibold text-ink">{formatAuditAction(item.action)}</td>
                    <td className="px-5 py-4 text-slate-600">
                      <div className="font-semibold text-ink">{item.userFullName || "Sistem"}</div>
                      <div className="text-xs text-slate-500">{item.userEmail || "-"}</div>
                    </td>
                    <td className="px-5 py-4 text-slate-600">{formatAuditDescription(item)}</td>
                    <td className="whitespace-nowrap px-5 py-4 text-slate-500">{formatDateTime(item.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="divide-y divide-line md:hidden">
            {items.map((item) => (
              <article key={item.id} className="space-y-2 px-4 py-4">
                <p className="text-sm font-bold text-ink">{formatAuditAction(item.action)}</p>
                <p className="text-sm text-slate-600">{formatAuditDescription(item)}</p>
                <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-slate-500">
                  <span className="font-semibold text-slate-700">{item.userFullName || "Sistem"}</span>
                  <span>{formatDateTime(item.createdAt)}</span>
                </div>
              </article>
            ))}
          </div>

          <div className="flex flex-col gap-3 border-t border-line px-5 py-4 text-sm text-slate-500 sm:flex-row sm:items-center sm:justify-between">
            <span>
              Toplam {totalCount} kayit · Sayfa {page}/{totalPages}
            </span>
            <div className="flex items-center gap-2">
              <button className="inline-flex h-9 items-center gap-1 rounded-lg border border-line px-3 font-semibold text-slate-600 transition hover:bg-violet-soft/40 disabled:opacity-50" type="button" disabled={page <= 1} onClick={() => void fetchLogs(page - 1)}>
                <ChevronLeft className="h-4 w-4" />
                Önceki
              </button>
              <button className="inline-flex h-9 items-center gap-1 rounded-lg border border-line px-3 font-semibold text-slate-600 transition hover:bg-violet-soft/40 disabled:opacity-50" type="button" disabled={page >= totalPages} onClick={() => void fetchLogs(page + 1)}>
                Sonraki
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </div>
        </section>
      )}
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
