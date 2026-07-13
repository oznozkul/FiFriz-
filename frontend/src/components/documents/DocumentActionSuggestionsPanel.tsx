import { ArrowRight, Check, Loader2, Sparkles, WandSparkles, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { PrimaryButton } from "../common";
import { ActionSuggestionConfirmDialog } from "./ActionSuggestionConfirmDialog";
import {
  acceptActionSuggestion,
  generateDocumentActionSuggestions,
  listDocumentActionSuggestions,
  rejectActionSuggestion,
} from "../../services/actionSuggestionService";
import { toast } from "../../services/toast";
import {
  ACTION_SUGGESTION_TYPE_LABELS,
  confirmationMessageFromPayload,
  panelFromActionPayload,
  toolNameFromPayload,
  type ActionSuggestion,
} from "../../types/actionSuggestion";
import type { WorkspacePanelSection } from "../documentWorkspace/documentWorkspaceUtils";

type DocumentActionSuggestionsPanelProps = {
  documentId: string;
  refreshToken?: number;
  onFocusSection?: (section: WorkspacePanelSection) => void;
  onActionExecuted?: () => void;
};

function mapPanel(panel: ReturnType<typeof panelFromActionPayload>): WorkspacePanelSection | null {
  if (!panel) return null;
  return panel;
}

export function DocumentActionSuggestionsPanel({
  documentId,
  refreshToken = 0,
  onFocusSection,
  onActionExecuted,
}: DocumentActionSuggestionsPanelProps) {
  const [suggestions, setSuggestions] = useState<ActionSuggestion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [pendingAccept, setPendingAccept] = useState<ActionSuggestion | null>(null);

  const loadSuggestions = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const items = await listDocumentActionSuggestions(documentId);
      if (items.length === 0) {
        const generated = await generateDocumentActionSuggestions(documentId);
        setSuggestions(generated);
      } else {
        setSuggestions(items);
      }
    } catch (loadError) {
      setSuggestions([]);
      setError(loadError instanceof Error ? loadError.message : "AI önerileri yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void loadSuggestions();
  }, [documentId, loadSuggestions, refreshToken]);

  const handleGenerate = async () => {
    setIsGenerating(true);
    setError(null);
    try {
      setSuggestions(await generateDocumentActionSuggestions(documentId));
    } catch (generateError) {
      setError(generateError instanceof Error ? generateError.message : "Öneriler oluşturulamadı.");
    } finally {
      setIsGenerating(false);
    }
  };

  const goToAction = (suggestion: ActionSuggestion) => {
    const panel = mapPanel(panelFromActionPayload(suggestion.actionPayload));
    if (panel) onFocusSection?.(panel);
  };

  const handleAcceptConfirmed = async (suggestion: ActionSuggestion) => {
    setBusyId(suggestion.id);
    setError(null);
    try {
      const result = await acceptActionSuggestion(suggestion.id);
      setSuggestions((current) => current.filter((item) => item.id !== suggestion.id));
      setPendingAccept(null);
      if (result.executed) {
        toast.success("Öneri uygulandı.");
        onActionExecuted?.();
      } else {
        toast.success("Öneri kabul edildi.");
      }
      goToAction(suggestion);
    } catch (acceptError) {
      const message = acceptError instanceof Error ? acceptError.message : "Öneri kabul edilemedi.";
      setError(message);
      toast.error(message);
    } finally {
      setBusyId(null);
    }
  };

  const handleAcceptClick = (suggestion: ActionSuggestion) => {
    if (toolNameFromPayload(suggestion.actionPayload)) {
      setPendingAccept(suggestion);
      return;
    }
    void handleAcceptConfirmed(suggestion);
  };

  const handleReject = async (suggestion: ActionSuggestion) => {
    setBusyId(suggestion.id);
    setError(null);
    try {
      await rejectActionSuggestion(suggestion.id);
      setSuggestions((current) => current.filter((item) => item.id !== suggestion.id));
    } catch (rejectError) {
      setError(rejectError instanceof Error ? rejectError.message : "Öneri reddedilemedi.");
    } finally {
      setBusyId(null);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 px-1 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin" />
        AI önerileri yükleniyor...
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-2">
        <p className="text-xs text-slate-500">Mevcut verilerden üretilen uygulanabilir öneriler</p>
        <button
          type="button"
          className="inline-flex items-center gap-1 text-xs font-semibold text-violet-brand disabled:opacity-60"
          disabled={isGenerating}
          onClick={() => void handleGenerate()}
        >
          {isGenerating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <WandSparkles className="h-3.5 w-3.5" />}
          Yenile
        </button>
      </div>

      {error ? <p className="text-sm font-medium text-red-600">{error}</p> : null}

      {suggestions.length === 0 ? (
        <p className="text-sm text-slate-500">Bu belge için şu an aksiyon önerisi yok.</p>
      ) : (
        <ul className="space-y-3">
          {suggestions.map((suggestion) => {
            const busy = busyId === suggestion.id;
            const toolName = toolNameFromPayload(suggestion.actionPayload);
            return (
              <li key={suggestion.id} className="rounded-xl border border-line bg-slate-50/70 p-3">
                <div className="flex flex-wrap items-center gap-2">
                  <Sparkles className="h-4 w-4 text-violet-brand" />
                  <span className="rounded-full bg-violet-50 px-2 py-0.5 text-[11px] font-semibold text-violet-700">
                    {ACTION_SUGGESTION_TYPE_LABELS[suggestion.suggestionType]}
                  </span>
                  {toolName ? (
                    <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-semibold uppercase text-slate-500">{toolName}</span>
                  ) : null}
                </div>
                <h4 className="mt-2 text-sm font-semibold text-ink">{suggestion.title}</h4>
                {suggestion.description ? <p className="mt-1 text-xs leading-5 text-slate-600">{suggestion.description}</p> : null}
                {toolName ? (
                  <p className="mt-2 text-xs text-slate-500">{confirmationMessageFromPayload(suggestion.actionPayload, suggestion.title)}</p>
                ) : null}

                <div className="mt-3 flex flex-wrap gap-2">
                  <PrimaryButton className="h-8 px-3 text-xs" disabled={busy} onClick={() => handleAcceptClick(suggestion)}>
                    {busy ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
                    Kabul Et
                  </PrimaryButton>
                  <button
                    type="button"
                    className="inline-flex h-8 items-center gap-1 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-slate-600 disabled:opacity-60"
                    disabled={busy}
                    onClick={() => void handleReject(suggestion)}
                  >
                    <X className="h-3.5 w-3.5" />
                    Reddet
                  </button>
                  <button
                    type="button"
                    className="inline-flex h-8 items-center gap-1 rounded-lg px-2 text-xs font-semibold text-violet-brand"
                    onClick={() => goToAction(suggestion)}
                  >
                    İlgili aksiyona git
                    <ArrowRight className="h-3.5 w-3.5" />
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
      )}

      <ActionSuggestionConfirmDialog
        suggestion={pendingAccept}
        isSubmitting={busyId === pendingAccept?.id}
        onCancel={() => setPendingAccept(null)}
        onConfirm={() => {
          if (pendingAccept) void handleAcceptConfirmed(pendingAccept);
        }}
      />
    </div>
  );
}
