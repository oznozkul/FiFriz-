import { ChevronDown, ChevronUp, Loader2, RefreshCw, Sparkles } from "lucide-react";
import { useCallback, useEffect, useState, type ReactNode } from "react";
import { PrimaryButton } from "../common";
import { generateDocumentAIMetadata, getDocumentAIMetadata } from "../../services/documentAiMetadataService";
import { DOCUMENT_TYPE_LABELS, type DocumentAIMetadata } from "../../types/documentAiMetadata";

type DocumentAiMetadataPanelProps = {
  documentId: string;
  refreshToken?: number;
};

function documentTypeLabel(value?: string | null) {
  if (!value) return "Belirtilmedi";
  return DOCUMENT_TYPE_LABELS[value] ?? value;
}

function formatConfidence(confidence?: number | null) {
  if (confidence == null || Number.isNaN(confidence)) {
    return null;
  }
  return `${Math.round(confidence * 100)}%`;
}

function MetadataSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div>
      <h4 className="text-xs font-bold uppercase tracking-wide text-slate-400">{title}</h4>
      <div className="mt-2">{children}</div>
    </div>
  );
}

export function DocumentAiMetadataPanel({ documentId, refreshToken = 0 }: DocumentAiMetadataPanelProps) {
  const [metadata, setMetadata] = useState<DocumentAIMetadata | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showRawJson, setShowRawJson] = useState(false);

  const loadMetadata = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await getDocumentAIMetadata(documentId);
      setMetadata(result);
    } catch (loadError) {
      setMetadata(null);
      setError(loadError instanceof Error ? loadError.message : "AI metadata yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void loadMetadata();
  }, [documentId, loadMetadata, refreshToken]);

  const handleGenerate = async () => {
    setIsGenerating(true);
    setError(null);
    try {
      const result = await generateDocumentAIMetadata(documentId);
      setMetadata(result);
    } catch (generateError) {
      setError(generateError instanceof Error ? generateError.message : "Metadata oluşturulamadı. Lütfen tekrar deneyin.");
    } finally {
      setIsGenerating(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        AI metadata yükleniyor...
      </div>
    );
  }

  if (isGenerating) {
    return (
      <div className="rounded-lg border border-violet-200 bg-violet-50/60 px-4 py-6 text-center">
        <Loader2 className="mx-auto h-6 w-6 animate-spin text-violet-brand" />
        <p className="mt-3 text-sm font-semibold text-violet-brand">FiFriz metadata çıkarıyor...</p>
        <p className="mt-1 text-xs text-slate-500">Bu işlem birkaç dakika sürebilir.</p>
      </div>
    );
  }

  if (!metadata) {
    return (
      <div className="space-y-4">
        <p className="text-sm text-slate-500">Bu belge için AI metadata henüz oluşturulmadı.</p>
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        <PrimaryButton className="h-10" onClick={() => void handleGenerate()}>
          <Sparkles className="h-4 w-4" />
          Metadata Oluştur
        </PrimaryButton>
      </div>
    );
  }

  const content = metadata.metadata;
  const confidenceLabel = formatConfidence(metadata.confidence ?? content.confidence);

  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-line bg-slate-50/70 p-4 space-y-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Belge Türü</p>
          <p className="mt-1 text-sm font-semibold text-ink">{documentTypeLabel(metadata.documentType ?? content.documentType)}</p>
        </div>
        {content.title ? (
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Başlık</p>
            <p className="mt-1 text-sm text-slate-700">{content.title}</p>
          </div>
        ) : null}
        {confidenceLabel ? (
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Güven</p>
            <p className="mt-1 text-sm text-slate-700">{confidenceLabel}</p>
          </div>
        ) : null}
        {content.summary ? (
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Özet</p>
            <p className="mt-1 text-sm leading-6 text-slate-700">{content.summary}</p>
          </div>
        ) : null}
      </div>

      {content.parties.length > 0 ? (
        <MetadataSection title="Taraflar">
          <ul className="space-y-2">
            {content.parties.map((party) => (
              <li key={`${party.name}-${party.role ?? ""}`} className="rounded-lg border border-line bg-white px-3 py-2 text-sm">
                <p className="font-semibold text-ink">{party.name}</p>
                {party.role ? <p className="text-slate-500">{party.role}</p> : null}
              </li>
            ))}
          </ul>
        </MetadataSection>
      ) : null}

      {content.dates.length > 0 ? (
        <MetadataSection title="Tarihler">
          <ul className="space-y-2">
            {content.dates.map((item) => (
              <li key={`${item.label}-${item.value}`} className="flex items-center justify-between gap-3 rounded-lg border border-line bg-white px-3 py-2 text-sm">
                <span className="text-slate-600">{item.label}</span>
                <span className="font-semibold text-ink">{item.value}</span>
              </li>
            ))}
          </ul>
        </MetadataSection>
      ) : null}

      {content.amounts.length > 0 ? (
        <MetadataSection title="Tutarlar">
          <ul className="space-y-2">
            {content.amounts.map((item) => (
              <li key={`${item.label}-${item.value}-${item.currency ?? ""}`} className="flex items-center justify-between gap-3 rounded-lg border border-line bg-white px-3 py-2 text-sm">
                <span className="text-slate-600">{item.label}</span>
                <span className="font-semibold text-ink">
                  {item.value.toLocaleString("tr-TR")} {item.currency ?? ""}
                </span>
              </li>
            ))}
          </ul>
        </MetadataSection>
      ) : null}

      {content.identifiers.length > 0 ? (
        <MetadataSection title="Kimlik / Referans Numaraları">
          <ul className="space-y-2">
            {content.identifiers.map((item) => (
              <li key={`${item.label}-${item.value}`} className="rounded-lg border border-line bg-white px-3 py-2 text-sm">
                <p className="text-slate-500">{item.label}</p>
                <p className="font-semibold text-ink break-all">{item.value}</p>
              </li>
            ))}
          </ul>
        </MetadataSection>
      ) : null}

      {content.addresses.length > 0 ? (
        <MetadataSection title="Adresler">
          <ul className="space-y-2">
            {content.addresses.map((address) => (
              <li key={address} className="rounded-lg border border-line bg-white px-3 py-2 text-sm text-slate-700">
                {address}
              </li>
            ))}
          </ul>
        </MetadataSection>
      ) : null}

      {content.keywords.length > 0 ? (
        <MetadataSection title="Anahtar Kelimeler">
          <div className="flex flex-wrap gap-2">
            {content.keywords.map((keyword) => (
              <span key={keyword} className="rounded-full border border-line bg-white px-3 py-1 text-xs font-semibold text-slate-600">
                {keyword}
              </span>
            ))}
          </div>
        </MetadataSection>
      ) : null}

      <button
        type="button"
        className="inline-flex h-9 items-center gap-2 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-slate-600 transition hover:border-violet-ring hover:text-violet-brand"
        onClick={() => setShowRawJson((value) => !value)}
      >
        {showRawJson ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
        {showRawJson ? "Ham JSON Gizle" : "Ham JSON Göster"}
      </button>

      {showRawJson ? (
        <pre className="max-h-64 overflow-auto rounded-lg border border-line bg-slate-900 p-3 text-xs text-slate-100">
          {JSON.stringify(metadata.rawJson ?? metadata.metadata, null, 2)}
        </pre>
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
