import { Check, ClipboardCheck, ExternalLink, X } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { PageState, PageTitle, PrimaryButton, Spinner, useConfirm } from "../components/common";
import { WorkflowStatusBadge } from "../components/documents/WorkflowCreateModal";
import { approveWorkflow, listPendingApprovals, rejectWorkflow, type PendingApproval } from "../services/workflowService";
import { toast } from "../services/toast";

export function PendingApprovalsPage() {
  const [items, setItems] = useState<PendingApproval[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const confirm = useConfirm();
  const [actingId, setActingId] = useState<string | null>(null);
  const [note, setNote] = useState("");
  const [selected, setSelected] = useState<PendingApproval | null>(null);

  const fetchItems = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setItems(await listPendingApprovals());
    } catch (error) {
      const message = error instanceof Error ? error.message : "Bekleyen onaylar yüklenemedi.";
      setLoadError(message);
      setItems([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchItems();
  }, []);

  const handleApprove = async (item: PendingApproval) => {
    const shouldApprove = await confirm({
      title: "Onayla",
      message: "Bu onay isteğini onaylamak istediğinize emin misiniz?",
      confirmLabel: "Onayla",
    });
    if (!shouldApprove) return;
    setActingId(item.workflowId);
    try {
      await approveWorkflow(item.workflowId, note.trim() || undefined);
      setNote("");
      setSelected(null);
      await fetchItems();
      toast.success("Onaylandı.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Onay işlemi başarısız.");
    } finally {
      setActingId(null);
    }
  };

  const handleReject = async (item: PendingApproval) => {
    const shouldReject = await confirm({
      title: "Reddet",
      message: "Bu onay isteğini reddetmek istediğinize emin misiniz?",
      confirmLabel: "Reddet",
      variant: "danger",
    });
    if (!shouldReject) return;
    setActingId(item.workflowId);
    try {
      await rejectWorkflow(item.workflowId, note.trim() || undefined);
      setNote("");
      setSelected(null);
      await fetchItems();
      toast.success("Reddedildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Red işlemi başarısız.");
    } finally {
      setActingId(null);
    }
  };

  return (
    <div className="space-y-4">
      <PageTitle title="Onaylarım" subtitle="Sizden beklenen onay adımlarını yönetin." />

      <PageState
        isLoading={isLoading}
        error={loadError}
        isEmpty={(items ?? []).length === 0}
        loadingLabel="Bekleyen onaylar yükleniyor..."
        errorTitle="Onaylar yüklenemedi"
        onRetry={() => void fetchItems()}
        emptyTitle="Bekleyen onayın yok"
        emptyDescription="Size atanmış aktif onay adımı bulunmuyor."
        emptyIcon={ClipboardCheck}
      >
        <section className="panel divide-y divide-line overflow-hidden rounded-2xl">
          {(items ?? []).map((item) => (
            <div key={`${item.workflowId}-${item.stepId}`} className="space-y-4 px-5 py-4">
              <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="text-sm font-bold text-ink">{item.workflowTitle}</h3>
                    <WorkflowStatusBadge status="pending" label={`Adım ${item.stepOrder}`} />
                  </div>
                  <p className="mt-1 text-sm text-slate-700">{item.documentName}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    {item.createdByFullName || "Kullanıcı"} · {formatDateTime(item.createdAt)}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Link className="inline-flex h-10 items-center gap-2 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600 hover:border-violet-brand hover:text-violet-brand" to={`/documents/${item.documentId}`}>
                    <ExternalLink className="h-4 w-4" />
                    Detaya Git
                  </Link>
                  <button type="button" className="inline-flex h-10 items-center gap-2 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600 hover:border-violet-brand hover:text-violet-brand" onClick={() => setSelected(selected?.workflowId === item.workflowId ? null : item)}>
                    İşlem Yap
                  </button>
                </div>
              </div>

              {selected?.workflowId === item.workflowId ? (
                <div className="rounded-lg border border-line bg-slate-50 p-4">
                  <label className="block space-y-2 text-sm">
                    <span className="font-semibold text-slate-600">Not (opsiyonel)</span>
                    <textarea className="min-h-20 w-full rounded-lg border border-line bg-white px-3 py-2 outline-none focus:border-violet-brand" value={note} onChange={(e) => setNote(e.target.value)} />
                  </label>
                  <div className="mt-3 flex flex-wrap gap-2">
                    <PrimaryButton className="h-10 bg-emerald-600 hover:bg-emerald-700" disabled={actingId === item.workflowId} onClick={() => void handleApprove(item)}>
                      {actingId === item.workflowId ? <Spinner size="sm" className="text-white" /> : <Check className="h-4 w-4" />}
                      Onayla
                    </PrimaryButton>
                    <button type="button" className="inline-flex h-10 items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 text-sm font-semibold text-red-700 disabled:opacity-60" disabled={actingId === item.workflowId} onClick={() => void handleReject(item)}>
                      <X className="h-4 w-4" />
                      Reddet
                    </button>
                  </div>
                </div>
              ) : null}
            </div>
          ))}
        </section>
      </PageState>
    </div>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}
