import { ArrowRight, ListChecks, Sparkles } from "lucide-react";
import { Link } from "react-router-dom";
import { acceptActionSuggestion, rejectActionSuggestion } from "../../services/actionSuggestionService";
import { toast } from "../../services/toast";
import type { ActionSuggestion } from "../../types/actionSuggestion";
import { ACTION_SUGGESTION_TYPE_LABELS, panelFromActionPayload, toolNameFromPayload } from "../../types/actionSuggestion";
import type { WorkspaceInsight, WorkspaceInsightPriority } from "../../types/workspaceInsights";
import { WORKSPACE_INSIGHT_PRIORITY_LABELS } from "../../types/workspaceInsights";
import { EmptyState, LoadingState } from "../common";
import { FileTypeIcon } from "../common/FileTypeIcon";
import { ActionSuggestionConfirmDialog } from "../documents/ActionSuggestionConfirmDialog";
import { useEffect, useState } from "react";

type WorkspaceInsightsPanelProps = {
  insights: WorkspaceInsight[];
  actionSuggestions?: ActionSuggestion[];
  isLoading?: boolean;
  isActionSuggestionsLoading?: boolean;
};

const priorityStyles: Record<WorkspaceInsightPriority, string> = {
  high: "border-red-200 bg-red-50 text-red-700",
  medium: "border-amber-200 bg-amber-50 text-amber-700",
  low: "border-violet-200 bg-violet-50 text-violet-700",
};

export function WorkspaceInsightsPanel({
  insights,
  actionSuggestions = [],
  isLoading = false,
  isActionSuggestionsLoading = false,
}: WorkspaceInsightsPanelProps) {
  const [localSuggestions, setLocalSuggestions] = useState(actionSuggestions);

  useEffect(() => {
    setLocalSuggestions(actionSuggestions);
  }, [actionSuggestions]);

  if (isLoading) {
    return <LoadingState variant="panel" label="AI içgörüleri yükleniyor..." />;
  }

  const hasInsights = insights.length > 0;
  const hasSuggestions = localSuggestions.length > 0;

  if (!hasInsights && !hasSuggestions && !isActionSuggestionsLoading) {
    return (
      <section className="panel p-5">
        <div className="mb-3 flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-violet-brand" />
          <h2 className="font-bold text-ink">AI İçgörüleri</h2>
        </div>
        <EmptyState title="Şu an öne çıkan içgörü yok" description="Belgeleriniz işlendikçe ve veriler biriktikçe burada öneriler görünecek." />
      </section>
    );
  }

  return (
    <section className="panel overflow-hidden">
      <div className="flex items-center gap-2 border-b border-line px-5 py-4">
        <Sparkles className="h-5 w-5 text-violet-brand" />
        <div>
          <h2 className="font-bold text-ink">AI İçgörüleri</h2>
          <p className="text-xs text-slate-500">Mevcut verilerinizden üretilen çalışma alanı önerileri</p>
        </div>
      </div>

      <div className="grid gap-4 p-5 lg:grid-cols-2">
        {hasInsights
          ? insights.map((insight) => <InsightCard key={`${insight.type}-${insight.title}`} insight={insight} />)
          : null}
      </div>

      {isActionSuggestionsLoading ? (
        <div className="border-t border-line px-5 py-4">
          <LoadingState label="Aksiyon önerileri yükleniyor..." />
        </div>
      ) : hasSuggestions ? (
        <div className="border-t border-line px-5 py-4">
          <div className="mb-3 flex items-center gap-2">
            <ListChecks className="h-4 w-4 text-violet-brand" />
            <h3 className="text-sm font-bold text-ink">AI Aksiyon Önerileri</h3>
          </div>
          <div className="grid gap-3 lg:grid-cols-2">
            {localSuggestions.slice(0, 6).map((suggestion) => (
              <ActionSuggestionCard
                key={suggestion.id}
                suggestion={suggestion}
                onResolved={(id) => setLocalSuggestions((current) => current.filter((item) => item.id !== id))}
              />
            ))}
          </div>
        </div>
      ) : null}
    </section>
  );
}

function InsightCard({ insight }: { insight: WorkspaceInsight }) {
  return (
    <article className="rounded-xl border border-line bg-white p-4 transition hover:border-violet-ring">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="font-semibold text-ink">{insight.title}</h3>
            <span className={`rounded-full border px-2 py-0.5 text-[11px] font-semibold ${priorityStyles[insight.priority]}`}>
              {WORKSPACE_INSIGHT_PRIORITY_LABELS[insight.priority]}
            </span>
          </div>
          <p className="mt-2 text-sm leading-6 text-slate-600">{insight.description}</p>
        </div>
      </div>

      {insight.relatedDocuments.length > 0 ? (
        <ul className="mt-3 space-y-2">
          {insight.relatedDocuments.slice(0, 3).map((document) => (
            <li key={document.documentId}>
              <Link
                to={`/documents/${document.documentId}`}
                className="flex items-center gap-2 rounded-lg border border-line px-3 py-2 text-sm transition hover:border-violet-ring hover:bg-violet-soft/40"
              >
                <FileTypeIcon extension={document.fileExtension ?? "file"} />
                <span className="truncate font-medium text-ink">{document.originalFileName}</span>
              </Link>
            </li>
          ))}
        </ul>
      ) : null}

      <div className="mt-4">
        <Link
          to={insight.actionUrl}
          className="inline-flex items-center gap-2 text-sm font-semibold text-violet-brand transition hover:text-violet-700"
        >
          {insight.actionLabel}
          <ArrowRight className="h-4 w-4" />
        </Link>
      </div>
    </article>
  );
}

function ActionSuggestionCard({ suggestion, onResolved }: { suggestion: ActionSuggestion; onResolved: (id: string) => void }) {
  const [busy, setBusy] = useState(false);
  const [pendingAccept, setPendingAccept] = useState(false);
  const documentId = suggestion.documentId ?? (typeof suggestion.actionPayload.documentId === "string" ? suggestion.actionPayload.documentId : null);
  const panel = panelFromActionPayload(suggestion.actionPayload);
  const targetUrl = documentId ? `/documents/${documentId}${panel ? `?panel=${panel === "aiMetadata" ? "ai-metadata" : panel}` : "?panel=action-suggestions"}` : "/documents";

  const runAccept = async () => {
    setBusy(true);
    try {
      const result = await acceptActionSuggestion(suggestion.id);
      onResolved(suggestion.id);
      toast.success(result.executed ? "Öneri uygulandı." : "Öneri kabul edildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Öneri kabul edilemedi.");
    } finally {
      setBusy(false);
      setPendingAccept(false);
    }
  };

  const handleAccept = () => {
    if (toolNameFromPayload(suggestion.actionPayload)) {
      setPendingAccept(true);
      return;
    }
    void runAccept();
  };

  const handleReject = async () => {
    setBusy(true);
    try {
      await rejectActionSuggestion(suggestion.id);
      onResolved(suggestion.id);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Öneri reddedilemedi.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <article className="rounded-xl border border-line bg-slate-50/70 p-4">
        <span className="rounded-full bg-violet-50 px-2 py-0.5 text-[11px] font-semibold text-violet-700">
          {ACTION_SUGGESTION_TYPE_LABELS[suggestion.suggestionType]}
        </span>
        <h4 className="mt-2 text-sm font-semibold text-ink">{suggestion.title}</h4>
        {suggestion.description ? <p className="mt-1 text-xs leading-5 text-slate-600">{suggestion.description}</p> : null}
        <div className="mt-3 flex flex-wrap gap-2 text-xs font-semibold">
          <button type="button" className="text-violet-brand disabled:opacity-60" disabled={busy} onClick={handleAccept}>
            Kabul Et
          </button>
          <button type="button" className="text-slate-500 disabled:opacity-60" disabled={busy} onClick={() => void handleReject()}>
            Reddet
          </button>
          <Link to={targetUrl} className="inline-flex items-center gap-1 text-violet-brand">
            Aksiyona git
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>
      </article>
      <ActionSuggestionConfirmDialog
        suggestion={pendingAccept ? suggestion : null}
        isSubmitting={busy}
        onCancel={() => setPendingAccept(false)}
        onConfirm={() => void runAccept()}
      />
    </>
  );
}
