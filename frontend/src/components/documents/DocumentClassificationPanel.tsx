import { Loader2, RefreshCw, Sparkles } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { PrimaryButton } from "../common";
import { generateDocumentClassification, getDocumentClassification } from "../../services/documentClassificationService";
import { aiDocumentTypeLabel, type DocumentClassification } from "../../types/documentClassification";

type DocumentClassificationPanelProps = {
  documentId: string;
  refreshToken?: number;
  onClassificationUpdated?: (classification: DocumentClassification | null) => void;
};

function formatConfidence(confidence?: number | null) {
  if (confidence == null || Number.isNaN(confidence)) {
    return null;
  }
  return `${Math.round(confidence * 100)}%`;
}

export function DocumentClassificationPanel({
  documentId,
  refreshToken = 0,
  onClassificationUpdated,
}: DocumentClassificationPanelProps) {
  const [classification, setClassification] = useState<DocumentClassification | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadClassification = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await getDocumentClassification(documentId);
      setClassification(result);
      onClassificationUpdated?.(result);
    } catch (loadError) {
      setClassification(null);
      setError(loadError instanceof Error ? loadError.message : "Belge sınıflandırması yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId, onClassificationUpdated]);

  useEffect(() => {
    void loadClassification();
  }, [documentId, loadClassification, refreshToken]);

  const handleGenerate = async () => {
    setIsGenerating(true);
    setError(null);
    try {
      const result = await generateDocumentClassification(documentId);
      setClassification(result);
      onClassificationUpdated?.(result);
    } catch (generateError) {
      setError(generateError instanceof Error ? generateError.message : "Belge sınıflandırılamadı. Lütfen tekrar deneyin.");
    } finally {
      setIsGenerating(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        Belge türü yükleniyor...
      </div>
    );
  }

  if (isGenerating) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        AI belge türü analiz ediliyor...
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {error ? <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p> : null}

      {classification ? (
        <dl className="space-y-3 text-sm">
          <div>
            <dt className="text-xs font-bold uppercase tracking-wide text-slate-400">AI Belge Türü</dt>
            <dd className="mt-1 font-semibold text-ink">{aiDocumentTypeLabel(classification.documentType)}</dd>
          </div>
          {formatConfidence(classification.confidence) ? (
            <div>
              <dt className="text-xs font-bold uppercase tracking-wide text-slate-400">Güven</dt>
              <dd className="mt-1 font-semibold text-slate-700">{formatConfidence(classification.confidence)}</dd>
            </div>
          ) : null}
          {classification.reason ? (
            <div>
              <dt className="text-xs font-bold uppercase tracking-wide text-slate-400">Gerekçe</dt>
              <dd className="mt-1 text-slate-600">{classification.reason}</dd>
            </div>
          ) : null}
        </dl>
      ) : (
        <div className="rounded-xl border border-dashed border-line bg-slate-50/80 px-4 py-5 text-center">
          <Sparkles className="mx-auto h-5 w-5 text-violet-brand" />
          <p className="mt-2 text-sm text-slate-600">Bu belge henüz AI ile sınıflandırılmadı.</p>
        </div>
      )}

      <PrimaryButton type="button" className="w-full" onClick={() => void handleGenerate()}>
        <RefreshCw className="h-4 w-4" />
        {classification ? "Yeniden Sınıflandır" : "Belge Türünü Bul"}
      </PrimaryButton>
    </div>
  );
}
