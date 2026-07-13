import { Loader2, RefreshCw, Sparkles } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { PrimaryButton } from "../common";
import { notifyOnboardingStep } from "../dashboard/onboarding/onboardingStorage";
import { generateDocumentSummary, getDocumentSummary } from "../../services/documentSummaryService";
import type { DocumentSummary } from "../../types/documentSummary";

type DocumentSummaryPanelProps = {
  documentId: string;
  refreshToken?: number;
};

export function DocumentSummaryPanel({ documentId, refreshToken = 0 }: DocumentSummaryPanelProps) {
  const [summary, setSummary] = useState<DocumentSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(false);

  const loadSummary = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setNotFound(false);
    try {
      const result = await getDocumentSummary(documentId);
      if (!result) {
        setSummary(null);
        setNotFound(true);
        return;
      }
      setSummary(result);
      setNotFound(false);
    } catch (loadError) {
      setSummary(null);
      const message = loadError instanceof Error ? loadError.message : "Belge özeti yüklenemedi.";
      setError(message);
      setNotFound(false);
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void loadSummary();
  }, [documentId, loadSummary, refreshToken]);

  const handleGenerate = async () => {
    setIsGenerating(true);
    setError(null);
    try {
      const result = await generateDocumentSummary(documentId);
      setSummary(result);
      setNotFound(false);
      notifyOnboardingStep("ai-summary");
    } catch (generateError) {
      setError(generateError instanceof Error ? generateError.message : "Özet oluşturulamadı. Lütfen tekrar deneyin.");
    } finally {
      setIsGenerating(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-6 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        Özet yükleniyor...
      </div>
    );
  }

  if (isGenerating) {
    return (
      <div className="rounded-lg border border-violet-200 bg-violet-50/60 px-4 py-6 text-center">
        <Loader2 className="mx-auto h-6 w-6 animate-spin text-violet-brand" />
        <p className="mt-3 text-sm font-semibold text-violet-brand">FiFriz belgeyi analiz ediyor...</p>
        <p className="mt-1 text-xs text-slate-500">Bu işlem birkaç dakika sürebilir.</p>
      </div>
    );
  }

  if (notFound || !summary) {
    return (
      <div className="space-y-4">
        <p className="text-sm text-slate-500">Bu belge için henüz özet oluşturulmadı.</p>
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        <PrimaryButton className="h-10" onClick={() => void handleGenerate()}>
          <Sparkles className="h-4 w-4" />
          Özet Oluştur
        </PrimaryButton>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-line bg-slate-50/70 p-4">
        <p className="text-sm leading-6 text-slate-700">{summary.summary}</p>
      </div>

      {summary.keyPoints.length > 0 ? (
        <div>
          <h4 className="text-xs font-bold uppercase tracking-wide text-slate-400">Önemli maddeler</h4>
          <ul className="mt-2 space-y-2">
            {summary.keyPoints.map((point) => (
              <li key={point} className="flex gap-2 text-sm text-slate-700">
                <span className="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-violet-brand" />
                <span>{point}</span>
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {error ? <p className="text-sm text-rose-600">{error}</p> : null}

      <button
        type="button"
        className="inline-flex h-9 items-center gap-2 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-violet-brand transition hover:border-violet-ring hover:bg-violet-soft"
        onClick={() => void handleGenerate()}
      >
        <RefreshCw className="h-3.5 w-3.5" />
        Yeniden Oluştur
      </button>
    </div>
  );
}
