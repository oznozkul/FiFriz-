import { ArrowDown, ArrowUp, ClipboardCheck, Plus, Trash2, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { TenantMember } from "../../services/tenantService";
import { listTenantMembers } from "../../services/tenantService";
import { PrimaryButton, Spinner, ModalFooter, ModalShell } from "../common";
import { StatusBadge } from "../common/StatusBadge";

export type WorkflowStepDraft = {
  id: string;
  approverUserId: string;
  stepOrder: number;
};

type WorkflowCreateModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (input: { title: string; description: string; steps: { approverUserId: string; stepOrder: number }[] }) => Promise<void>;
};

export function WorkflowCreateModal({ isOpen, onClose, onSubmit }: WorkflowCreateModalProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [steps, setSteps] = useState<WorkflowStepDraft[]>([{ id: crypto.randomUUID(), approverUserId: "", stepOrder: 1 }]);
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [isLoadingMembers, setIsLoadingMembers] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    setIsLoadingMembers(true);
    listTenantMembers()
      .then(setMembers)
      .finally(() => setIsLoadingMembers(false));
  }, [isOpen]);

  if (!isOpen) return null;

  const addStep = () => {
    setSteps((current) => [...current, { id: crypto.randomUUID(), approverUserId: "", stepOrder: current.length + 1 }]);
  };

  const removeStep = (id: string) => {
    setSteps((current) => current.filter((step) => step.id !== id).map((step, index) => ({ ...step, stepOrder: index + 1 })));
  };

  const moveStep = (index: number, direction: -1 | 1) => {
    setSteps((current) => {
      const next = [...current];
      const target = index + direction;
      if (target < 0 || target >= next.length) return current;
      [next[index], next[target]] = [next[target], next[index]];
      return next.map((step, stepIndex) => ({ ...step, stepOrder: stepIndex + 1 }));
    });
  };

  const handleSubmit = async () => {
    if (!title.trim() || steps.some((step) => !step.approverUserId)) return;
    setIsSubmitting(true);
    try {
      await onSubmit({
        title: title.trim(),
        description: description.trim(),
        steps: steps.map((step) => ({ approverUserId: step.approverUserId, stepOrder: step.stepOrder })),
      });
      setTitle("");
      setDescription("");
      setSteps([{ id: crypto.randomUUID(), approverUserId: "", stepOrder: 1 }]);
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <ModalShell onClose={onClose} size="2xl" zIndexClass="z-[95]" ariaLabel="Onay süreci oluştur">
      <div className="flex items-center justify-between border-b border-line px-4 py-4 sm:px-5">
        <h3 className="flex items-center gap-2 text-lg font-bold text-ink">
          <ClipboardCheck className="h-5 w-5 text-violet-brand" />
          Onay Süreci Oluştur
        </h3>
        <button type="button" className="rounded-lg p-2 text-slate-500 hover:bg-slate-100" onClick={onClose}>
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="min-h-0 flex-1 space-y-4 overflow-y-auto p-4 sm:p-5">
          <label className="block space-y-2 text-sm">
            <span className="font-semibold text-slate-600">Baslik</span>
            <input className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Örnek: Sözleşme onay süreci" />
          </label>
          <label className="block space-y-2 text-sm">
            <span className="font-semibold text-slate-600">Açıklama (opsiyonel)</span>
            <textarea className="min-h-20 w-full rounded-lg border border-line px-3 py-2 outline-none focus:border-violet-brand" value={description} onChange={(e) => setDescription(e.target.value)} />
          </label>

          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-semibold text-slate-600">Onay adımlari</span>
              <button type="button" className="inline-flex items-center gap-1 text-sm font-semibold text-violet-brand" onClick={addStep}>
                <Plus className="h-4 w-4" />
                Adım ekle
              </button>
            </div>
            {isLoadingMembers ? <p className="text-sm text-slate-500">Üyeler yükleniyor...</p> : null}
            {steps.map((step, index) => (
              <div key={step.id} className="flex flex-col gap-2 rounded-lg border border-line bg-slate-50 p-3 sm:flex-row sm:items-center">
                <span className="w-8 text-sm font-bold text-violet-brand">{step.stepOrder}</span>
                <select className="h-10 min-w-0 flex-1 rounded-lg border border-line bg-white px-3 text-sm outline-none focus:border-violet-brand" value={step.approverUserId} onChange={(e) => setSteps((current) => current.map((item) => (item.id === step.id ? { ...item, approverUserId: e.target.value } : item)))}>
                  <option value="">Onaylayıcı seç...</option>
                  {members.map((member) => (
                    <option key={member.userId} value={member.userId}>
                      {member.fullName} ({member.email})
                    </option>
                  ))}
                </select>
                <div className="flex items-center gap-1 self-end sm:self-auto">
                <button type="button" className="rounded-lg p-2 text-slate-500 hover:bg-white disabled:opacity-40" disabled={index === 0} onClick={() => moveStep(index, -1)}>
                  <ArrowUp className="h-4 w-4" />
                </button>
                <button type="button" className="rounded-lg p-2 text-slate-500 hover:bg-white disabled:opacity-40" disabled={index === steps.length - 1} onClick={() => moveStep(index, 1)}>
                  <ArrowDown className="h-4 w-4" />
                </button>
                <button type="button" className="rounded-lg p-2 text-slate-500 hover:bg-white disabled:opacity-40" disabled={steps.length === 1} onClick={() => removeStep(step.id)}>
                  <Trash2 className="h-4 w-4" />
                </button>
                </div>
              </div>
            ))}
          </div>
      </div>

      <ModalFooter>
          <button type="button" className="h-10 w-full rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-600 transition hover:border-violet-ring hover:bg-violet-soft/40 sm:w-auto" onClick={onClose}>
            İptal
          </button>
          <PrimaryButton className="w-full sm:w-auto" disabled={isSubmitting || !title.trim() || steps.some((s) => !s.approverUserId)} onClick={() => void handleSubmit()}>
            {isSubmitting ? <Spinner size="sm" label="Oluşturuluyor..." className="text-white" /> : "Oluştur"}
          </PrimaryButton>
      </ModalFooter>
    </ModalShell>
  );
}

export function WorkflowStatusBadge({ status, label }: { status: string; label: string }) {
  return <StatusBadge status={status} label={label} />;
}
