import { Check, Loader2, Sparkles, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { PrimaryButton } from "../common";
import {
  acceptAllSuggestedTags,
  acceptSuggestedTag,
  getDocumentSuggestedTags,
  rejectSuggestedTag,
} from "../../services/documentSuggestedTagService";
import type { DocumentSuggestedTag } from "../../types/documentSuggestedTag";
import type { TagBrief } from "../../types/tag";

type DocumentSuggestedTagsPanelProps = {
  documentId: string;
  refreshToken?: number;
  onTagsUpdated?: (tags: TagBrief[]) => void;
};

function formatConfidence(confidence?: number | null) {
  if (confidence == null || Number.isNaN(confidence)) {
    return null;
  }
  return `${Math.round(confidence * 100)}%`;
}

export function DocumentSuggestedTagsPanel({ documentId, refreshToken = 0, onTagsUpdated }: DocumentSuggestedTagsPanelProps) {
  const [suggestions, setSuggestions] = useState<DocumentSuggestedTag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [isAcceptingAll, setIsAcceptingAll] = useState(false);

  const loadSuggestions = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await getDocumentSuggestedTags(documentId);
      setSuggestions(result);
    } catch (loadError) {
      setSuggestions([]);
      setError(loadError instanceof Error ? loadError.message : "AI etiket önerileri yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void loadSuggestions();
  }, [documentId, loadSuggestions, refreshToken]);

  const handleAccept = async (suggestion: DocumentSuggestedTag) => {
    setBusyId(suggestion.id);
    setError(null);
    try {
      const result = await acceptSuggestedTag(documentId, suggestion.id);
      setSuggestions((current) => current.filter((item) => item.id !== suggestion.id));
      onTagsUpdated?.(result.tags);
    } catch (acceptError) {
      setError(acceptError instanceof Error ? acceptError.message : "Etiket önerisi kabul edilemedi.");
    } finally {
      setBusyId(null);
    }
  };

  const handleReject = async (suggestion: DocumentSuggestedTag) => {
    setBusyId(suggestion.id);
    setError(null);
    try {
      await rejectSuggestedTag(documentId, suggestion.id);
      setSuggestions((current) => current.filter((item) => item.id !== suggestion.id));
    } catch (rejectError) {
      setError(rejectError instanceof Error ? rejectError.message : "Etiket önerisi reddedilemedi.");
    } finally {
      setBusyId(null);
    }
  };

  const handleAcceptAll = async () => {
    setIsAcceptingAll(true);
    setError(null);
    try {
      const result = await acceptAllSuggestedTags(documentId);
      setSuggestions([]);
      onTagsUpdated?.(result.tags);
    } catch (acceptAllError) {
      setError(acceptAllError instanceof Error ? acceptAllError.message : "Etiket önerileri kabul edilemedi.");
    } finally {
      setIsAcceptingAll(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        AI etiket önerileri yükleniyor...
      </div>
    );
  }

  if (suggestions.length === 0) {
    return (
      <div className="space-y-2">
        <p className="text-sm text-slate-500">Henüz AI etiket önerisi yok.</p>
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <p className="text-xs font-bold uppercase tracking-wide text-slate-400">AI Etiket Önerileri</p>
        <PrimaryButton className="h-8 px-3 text-xs" disabled={isAcceptingAll || busyId !== null} onClick={() => void handleAcceptAll()}>
          {isAcceptingAll ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Sparkles className="h-3.5 w-3.5" />}
          Tümünü Kabul Et
        </PrimaryButton>
      </div>

      <ul className="space-y-2">
        {suggestions.map((suggestion) => {
          const confidenceLabel = formatConfidence(suggestion.confidence);
          const isBusy = busyId === suggestion.id || isAcceptingAll;
          return (
            <li key={suggestion.id} className="rounded-lg border border-line bg-slate-50/70 px-3 py-3">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-sm font-semibold text-ink">{suggestion.tagName}</p>
                  {confidenceLabel ? <p className="mt-1 text-xs text-slate-500">Güven: {confidenceLabel}</p> : null}
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <button
                    type="button"
                    className="inline-flex h-8 items-center gap-1 rounded-lg border border-emerald-200 bg-emerald-50 px-2.5 text-xs font-semibold text-emerald-700 transition hover:bg-emerald-100 disabled:opacity-60"
                    disabled={isBusy}
                    onClick={() => void handleAccept(suggestion)}
                  >
                    {busyId === suggestion.id ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Check className="h-3.5 w-3.5" />}
                    Kabul Et
                  </button>
                  <button
                    type="button"
                    className="inline-flex h-8 items-center gap-1 rounded-lg border border-line bg-white px-2.5 text-xs font-semibold text-slate-600 transition hover:border-rose-200 hover:bg-rose-50 hover:text-rose-600 disabled:opacity-60"
                    disabled={isBusy}
                    onClick={() => void handleReject(suggestion)}
                  >
                    <X className="h-3.5 w-3.5" />
                    Reddet
                  </button>
                </div>
              </div>
            </li>
          );
        })}
      </ul>

      {error ? <p className="text-sm text-rose-600">{error}</p> : null}
    </div>
  );
}
