import { Check, ClipboardCheck, ExternalLink, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { ApiDocument } from "../../types/document";
import { useAuth } from "../../services/AuthContext";
import { EmptyState, LoadingState, PrimaryButton, Spinner, useConfirm } from "../common";
import { notifyOnboardingStep } from "../dashboard/onboarding/onboardingStorage";
import { WorkflowCreateModal, WorkflowStatusBadge } from "./WorkflowCreateModal";
import {
  approveWorkflow,
  cancelWorkflow,
  createDocumentWorkflow,
  listDocumentWorkflows,
  rejectWorkflow,
  stepStatusLabel,
  workflowStatusLabel,
  type Workflow,
} from "../../services/workflowService";
import { toast } from "../../services/toast";

type DocumentWorkflowPanelProps = {
  document: ApiDocument;
};

export function DocumentWorkflowPanel({ document }: DocumentWorkflowPanelProps) {
  const { user } = useAuth();
  const confirm = useConfirm();
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isCancellingId, setIsCancellingId] = useState<string | null>(null);
  const [actingWorkflowId, setActingWorkflowId] = useState<string | null>(null);
  const [note, setNote] = useState("");

  const activeWorkflow = useMemo(() => workflows.find((item) => item.status === "pending"), [workflows]);
  const myCurrentStep = useMemo(() => {
    if (!user || !activeWorkflow) return null;
    return activeWorkflow.steps.find((step) => step.isCurrent && step.approverUserId === user.id) ?? null;
  }, [activeWorkflow, user]);

  const fetchWorkflows = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setWorkflows(await listDocumentWorkflows(document.id));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Onay süreçleri yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchWorkflows();
  }, [document.id]);

  const handleCreate = async (input: { title: string; description: string; steps: { approverUserId: string; stepOrder: number }[] }) => {
    try {
      await createDocumentWorkflow(document.id, {
        title: input.title,
        description: input.description || undefined,
        steps: input.steps,
      });
      await fetchWorkflows();
      toast.success("Onay süreci oluşturuldu.");
      notifyOnboardingStep("start-workflow");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Onay süreci oluşturulamadı.");
      throw error;
    }
  };

  const handleCancel = async (workflow: Workflow) => {
    const shouldCancel = await confirm({
      title: "Onay sürecini iptal et",
      message: "Bu onay sürecini iptal etmek istediğinize emin misiniz?",
      confirmLabel: "İptal et",
      variant: "danger",
    });
    if (!shouldCancel) return;
    setIsCancellingId(workflow.id);
    try {
      await cancelWorkflow(workflow.id);
      await fetchWorkflows();
      toast.success("Onay süreci iptal edildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Onay süreci iptal edilemedi.");
    } finally {
      setIsCancellingId(null);
    }
  };

  const handleApprove = async (workflowId: string) => {
    const shouldApprove = await confirm({
      title: "Onayla",
      message: "Bu onay isteğini onaylamak istediğinize emin misiniz?",
      confirmLabel: "Onayla",
    });
    if (!shouldApprove) return;
    setActingWorkflowId(workflowId);
    try {
      await approveWorkflow(workflowId, note.trim() || undefined);
      setNote("");
      await fetchWorkflows();
      toast.success("Onaylandı.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Onay işlemi başarısız.");
    } finally {
      setActingWorkflowId(null);
    }
  };

  const handleReject = async (workflowId: string) => {
    const shouldReject = await confirm({
      title: "Reddet",
      message: "Bu onay isteğini reddetmek istediğinize emin misiniz?",
      confirmLabel: "Reddet",
      variant: "danger",
    });
    if (!shouldReject) return;
    setActingWorkflowId(workflowId);
    try {
      await rejectWorkflow(workflowId, note.trim() || undefined);
      setNote("");
      await fetchWorkflows();
      toast.success("Reddedildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Red işlemi başarısız.");
    } finally {
      setActingWorkflowId(null);
    }
  };

  return (
    <section className="flex flex-col">
      <div className="flex items-center justify-between border-b border-line px-5 py-4">
        <div>
          <h2 className="flex items-center gap-2 text-base font-bold text-ink">
            <ClipboardCheck className="h-4 w-4 text-violet-brand" />
            Onay Süreçi
          </h2>
          <p className="mt-1 text-sm text-slate-500">
            Belge durumu: <WorkflowStatusBadge status={document.approvalStatus ?? "none"} label={approvalStatusLabel(document.approvalStatus)} />
          </p>
        </div>
        <PrimaryButton className="h-9 px-3 text-sm" disabled={!!activeWorkflow} onClick={() => setIsModalOpen(true)}>
          Süreç Oluştur
        </PrimaryButton>
      </div>

      {isLoading ? (
        <div className="p-5"><LoadingState label="Onay süreçleri yükleniyor..." /></div>
      ) : loadError ? (
        <div className="p-5"><EmptyState title="Onay süreçleri yüklenemedi" description={loadError} action={<PrimaryButton onClick={() => void fetchWorkflows()}>Tekrar Dene</PrimaryButton>} /></div>
      ) : workflows.length === 0 ? (
        <div className="p-5"><EmptyState title="Bu belge için onay süreci yok." description="İlk onay sürecini oluşturarak başlayabilirsin." /></div>
      ) : (
        <div className="divide-y divide-line">
          {myCurrentStep && activeWorkflow ? (
            <div className="space-y-3 border-b border-line bg-violet-soft/20 px-5 py-4">
              <p className="text-sm font-semibold text-ink">Sizden onay bekleniyor (Adım {myCurrentStep.stepOrder})</p>
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">Not (opsiyonel)</span>
                <textarea className="min-h-20 w-full rounded-lg border border-line bg-white px-3 py-2 outline-none focus:border-violet-brand" value={note} onChange={(e) => setNote(e.target.value)} />
              </label>
              <div className="flex flex-wrap gap-2">
                <PrimaryButton className="h-10 bg-emerald-600 hover:bg-emerald-700" disabled={actingWorkflowId === activeWorkflow.id} onClick={() => void handleApprove(activeWorkflow.id)}>
                  {actingWorkflowId === activeWorkflow.id ? <Spinner size="sm" className="text-white" /> : <Check className="h-4 w-4" />}
                  Onayla
                </PrimaryButton>
                <button type="button" className="inline-flex h-10 items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 text-sm font-semibold text-red-700 disabled:opacity-60" disabled={actingWorkflowId === activeWorkflow.id} onClick={() => void handleReject(activeWorkflow.id)}>
                  <X className="h-4 w-4" />
                  Reddet
                </button>
              </div>
            </div>
          ) : null}
          {workflows.map((workflow) => (
            <div key={workflow.id} className="space-y-4 px-5 py-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="text-sm font-bold text-ink">{workflow.title}</h3>
                    <WorkflowStatusBadge status={workflow.status} label={workflowStatusLabel(workflow.status)} />
                  </div>
                  <p className="mt-1 text-xs text-slate-500">
                    {workflow.createdByFullName || "Kullanıcı"} · {formatDateTime(workflow.createdAt)}
                  </p>
                  {workflow.description ? <p className="mt-2 text-sm text-slate-600">{workflow.description}</p> : null}
                </div>
                {workflow.status === "pending" ? (
                  <button type="button" className="text-sm font-semibold text-slate-500 hover:text-red-600 disabled:opacity-60" disabled={isCancellingId === workflow.id} onClick={() => void handleCancel(workflow)}>
                    İptal
                  </button>
                ) : null}
              </div>
              <div className="space-y-2">
                {workflow.steps.map((step) => (
                  <div key={step.id} className={`rounded-lg border px-4 py-3 ${step.isCurrent ? "border-violet-brand bg-violet-soft/30" : "border-line bg-white"}`}>
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-sm font-bold text-ink">Adım {step.stepOrder}</span>
                      <WorkflowStatusBadge status={step.status} label={stepStatusLabel(step.status)} />
                      {step.isCurrent ? <span className="rounded-full bg-violet-soft px-2 py-0.5 text-xs font-semibold text-violet-brand">Aktif</span> : null}
                    </div>
                    <p className="mt-1 text-sm text-slate-700">{step.approverFullName || "Kullanıcı"}</p>
                    {step.note ? <p className="mt-2 text-sm text-slate-600">{step.note}</p> : null}
                    {step.decidedAt ? <p className="mt-1 text-xs text-slate-400">{formatDateTime(step.decidedAt)}</p> : null}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      <WorkflowCreateModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} onSubmit={handleCreate} />
    </section>
  );
}

function approvalStatusLabel(status?: string) {
  switch (status) {
    case "pending": return "Onay bekliyor";
    case "approved": return "Onaylandı";
    case "rejected": return "Reddedildi";
    default: return "Yok";
  }
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}
