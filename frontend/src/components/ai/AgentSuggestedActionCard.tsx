import { useState } from "react";
import { acceptActionSuggestion, rejectActionSuggestion } from "../../services/actionSuggestionService";
import { toast } from "../../services/toast";
import type { ActionSuggestion } from "../../types/actionSuggestion";
import type { AgentSuggestedAction } from "../../types/agentChat";
import { ActionSuggestionConfirmDialog } from "../documents/ActionSuggestionConfirmDialog";

type AgentSuggestedActionCardProps = {
  action: AgentSuggestedAction;
  onResolved: (suggestionId: string) => void;
};

function toActionSuggestion(action: AgentSuggestedAction): ActionSuggestion {
  return {
    id: action.suggestionId,
    suggestionType: "add_tag",
    title: action.title,
    description: action.description,
    actionPayload: action.actionPayload,
    status: "pending",
    createdAt: "",
    updatedAt: "",
  };
}

export function AgentSuggestedActionCard({ action, onResolved }: AgentSuggestedActionCardProps) {
  const [busy, setBusy] = useState(false);
  const [pendingAccept, setPendingAccept] = useState(false);

  const runAccept = async () => {
    setBusy(true);
    try {
      const result = await acceptActionSuggestion(action.suggestionId);
      onResolved(action.suggestionId);
      toast.success(result.executed ? "İşlem uygulandı." : "Öneri kabul edildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "İşlem uygulanamadı.");
    } finally {
      setBusy(false);
      setPendingAccept(false);
    }
  };

  const handleReject = async () => {
    setBusy(true);
    try {
      await rejectActionSuggestion(action.suggestionId);
      onResolved(action.suggestionId);
      toast.info("Öneri reddedildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Öneri reddedilemedi.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <article className="rounded-xl border border-violet-200 bg-violet-50/50 p-4">
        <p className="text-[11px] font-semibold uppercase tracking-wide text-violet-brand">Önerilen aksiyon</p>
        <h4 className="mt-1 text-sm font-semibold text-ink">{action.title}</h4>
        {action.description ? <p className="mt-1 text-xs leading-5 text-slate-600">{action.description}</p> : null}
        <p className="mt-2 text-[11px] text-slate-500">Onayınız olmadan çalıştırılmaz.</p>
        <div className="mt-3 flex flex-wrap gap-2 text-xs font-semibold">
          <button
            type="button"
            className="rounded-lg bg-violet-brand px-3 py-1.5 text-white disabled:opacity-60"
            disabled={busy}
            onClick={() => setPendingAccept(true)}
          >
            Onayla
          </button>
          <button type="button" className="text-slate-500 disabled:opacity-60" disabled={busy} onClick={() => void handleReject()}>
            Reddet
          </button>
        </div>
      </article>
      <ActionSuggestionConfirmDialog
        suggestion={pendingAccept ? toActionSuggestion(action) : null}
        isSubmitting={busy}
        onCancel={() => setPendingAccept(false)}
        onConfirm={() => void runAccept()}
      />
    </>
  );
}
