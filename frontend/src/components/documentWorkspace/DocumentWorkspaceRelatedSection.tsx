import { Check, ExternalLink, Link2, Loader2, Plus, RefreshCw, Search, Sparkles, Trash2, X } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { FileTypeIcon, LoadingState, PrimaryButton, Spinner } from "../common";
import { getSearchResultExtension, getSearchResultTitle, searchDocuments } from "../../services/documentService";
import {
  acceptRelatedSuggestion,
  formatSimilarityPercent,
  generateRelatedSuggestions,
  listRelatedSuggestions,
  rejectRelatedSuggestion,
  type DocumentRelatedSuggestion,
} from "../../services/relatedSuggestionService";
import {
  createDocumentRelationship,
  deleteDocumentRelationship,
  listDocumentRelationships,
  relationshipTypeLabels,
  relationshipTypes,
  type DocumentRelationshipEdge,
  type RelationshipType,
} from "../../services/relationshipService";
import { toast } from "../../services/toast";
import type { ApiDocument } from "../../types/document";

type DocumentWorkspaceRelatedSectionProps = {
  document: ApiDocument;
  refreshToken?: number;
};

type SearchCandidate = {
  documentId: string;
  title: string;
  extension: string;
};

export function DocumentWorkspaceRelatedSection({ document, refreshToken = 0 }: DocumentWorkspaceRelatedSectionProps) {
  const [items, setItems] = useState<DocumentRelationshipEdge[]>([]);
  const [suggestions, setSuggestions] = useState<DocumentRelatedSuggestion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isComposerOpen, setIsComposerOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchCandidate[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [selectedTarget, setSelectedTarget] = useState<SearchCandidate | null>(null);
  const [relationshipType, setRelationshipType] = useState<RelationshipType>("related");
  const [isSaving, setIsSaving] = useState(false);
  const [removingId, setRemovingId] = useState<string | null>(null);
  const [busySuggestionId, setBusySuggestionId] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [relationships, relatedSuggestions] = await Promise.all([
        listDocumentRelationships(document.id),
        listRelatedSuggestions(document.id),
      ]);
      setItems(relationships.items);
      setSuggestions(relatedSuggestions);
    } catch (error) {
      setItems([]);
      setSuggestions([]);
      toast.error(error instanceof Error ? error.message : "İlişkili belgeler yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [document.id]);

  useEffect(() => {
    void loadData();
  }, [loadData, refreshToken]);

  useEffect(() => {
    const trimmed = searchQuery.trim();
    if (trimmed.length < 2) {
      setSearchResults([]);
      return;
    }

    let cancelled = false;
    const timer = window.setTimeout(() => {
      setIsSearching(true);
      searchDocuments(trimmed)
        .then((results) => {
          if (cancelled) return;
          setSearchResults(
            results
              .filter((result) => result.documentId !== document.id)
              .slice(0, 8)
              .map((result) => ({
                documentId: result.documentId,
                title: getSearchResultTitle(result),
                extension: getSearchResultExtension(result),
              })),
          );
        })
        .catch(() => {
          if (!cancelled) setSearchResults([]);
        })
        .finally(() => {
          if (!cancelled) setIsSearching(false);
        });
    }, 300);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [document.id, searchQuery]);

  const linkedDocumentIds = useMemo(() => new Set(items.map((item) => item.relatedDocument.id)), [items]);

  const handleGenerateSuggestions = async () => {
    setIsGenerating(true);
    try {
      const result = await generateRelatedSuggestions(document.id);
      setSuggestions(result);
      toast.success("AI ilişki önerileri güncellendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "AI ilişki önerileri oluşturulamadı.");
    } finally {
      setIsGenerating(false);
    }
  };

  const handleAcceptSuggestion = async (suggestion: DocumentRelatedSuggestion) => {
    setBusySuggestionId(suggestion.id);
    try {
      const created = await acceptRelatedSuggestion(document.id, suggestion.id);
      setSuggestions((current) => current.filter((item) => item.id !== suggestion.id));
      setItems((current) => [created, ...current.filter((item) => item.id !== created.id)]);
      toast.success("İlişki önerisi kabul edildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "İlişki önerisi kabul edilemedi.");
    } finally {
      setBusySuggestionId(null);
    }
  };

  const handleRejectSuggestion = async (suggestion: DocumentRelatedSuggestion) => {
    setBusySuggestionId(suggestion.id);
    try {
      await rejectRelatedSuggestion(document.id, suggestion.id);
      setSuggestions((current) => current.filter((item) => item.id !== suggestion.id));
      toast.success("İlişki önerisi reddedildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "İlişki önerisi reddedilemedi.");
    } finally {
      setBusySuggestionId(null);
    }
  };

  const handleCreate = async () => {
    if (!selectedTarget) return;

    setIsSaving(true);
    try {
      const created = await createDocumentRelationship(document.id, {
        targetDocumentId: selectedTarget.documentId,
        relationshipType,
      });
      setItems((current) => [created, ...current]);
      setIsComposerOpen(false);
      setSearchQuery("");
      setSearchResults([]);
      setSelectedTarget(null);
      setRelationshipType("related");
      toast.success("Belge ilişkisi eklendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Belge ilişkisi oluşturulamadı.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleRemove = async (relationshipId: string) => {
    setRemovingId(relationshipId);
    try {
      await deleteDocumentRelationship(relationshipId);
      setItems((current) => current.filter((item) => item.id !== relationshipId));
      toast.success("İlişki kaldırıldı.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "İlişki kaldırılamadı.");
    } finally {
      setRemovingId(null);
    }
  };

  if (isLoading) {
    return <LoadingState label="İlişkili belgeler yükleniyor..." className="py-6" />;
  }

  if (isGenerating) {
    return (
      <div className="flex items-center gap-2 py-6 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        AI ilişkili belgeler aranıyor...
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <PrimaryButton type="button" className="w-full" onClick={() => void handleGenerateSuggestions()}>
        <RefreshCw className="h-4 w-4" />
        AI ile İlgili Belgeleri Bul
      </PrimaryButton>

      <section className="space-y-3">
        <h4 className="text-xs font-bold uppercase tracking-wide text-slate-400">Onaylanmış İlişkiler</h4>

        <button
          type="button"
          className="inline-flex h-9 w-full items-center justify-center gap-2 rounded-lg border border-line bg-white text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand"
          onClick={() => setIsComposerOpen((current) => !current)}
        >
          <Plus className="h-4 w-4" />
          İlişki ekle
        </button>

        {isComposerOpen ? (
          <div className="space-y-3 rounded-lg border border-line bg-slate-50 p-3">
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm font-bold text-ink">Belge ilişkilendir</p>
              <button type="button" className="text-slate-400 hover:text-slate-600" aria-label="Kapat" onClick={() => setIsComposerOpen(false)}>
                <X className="h-4 w-4" />
              </button>
            </div>

            <label className="block space-y-1 text-sm">
              <span className="font-semibold text-slate-500">İlişki türü</span>
              <select
                className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-600 outline-none focus:border-violet-brand"
                value={relationshipType}
                onChange={(event) => setRelationshipType(event.target.value as RelationshipType)}
              >
                {relationshipTypes.map((type) => (
                  <option key={type} value={type}>
                    {relationshipTypeLabels[type]}
                  </option>
                ))}
              </select>
            </label>

            <label className="block space-y-1 text-sm">
              <span className="font-semibold text-slate-500">Belge ara</span>
              <div className="relative">
                <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
                <input
                  className="h-10 w-full rounded-lg border border-line bg-white pl-9 pr-3 text-sm font-semibold text-slate-700 outline-none focus:border-violet-brand"
                  placeholder="Belge adı ile ara..."
                  value={searchQuery}
                  onChange={(event) => {
                    setSearchQuery(event.target.value);
                    setSelectedTarget(null);
                  }}
                />
              </div>
            </label>

            {isSearching ? (
              <div className="flex items-center gap-2 text-sm text-slate-500">
                <Spinner size="sm" />
                Aranıyor...
              </div>
            ) : null}

            {!isSearching && searchQuery.trim().length >= 2 && searchResults.length === 0 ? (
              <p className="text-sm text-slate-500">Eşleşen belge bulunamadı.</p>
            ) : null}

            {searchResults.length > 0 ? (
              <ul className="max-h-44 space-y-1 overflow-y-auto">
                {searchResults.map((result) => {
                  const isSelected = selectedTarget?.documentId === result.documentId;
                  const isLinked = linkedDocumentIds.has(result.documentId);
                  return (
                    <li key={result.documentId}>
                      <button
                        type="button"
                        disabled={isLinked}
                        className={`flex w-full items-center gap-2 rounded-lg border px-3 py-2 text-left text-sm transition ${
                          isSelected ? "border-violet-brand bg-violet-50" : "border-line bg-white hover:border-violet-brand"
                        } ${isLinked ? "cursor-not-allowed opacity-50" : ""}`}
                        onClick={() => setSelectedTarget(result)}
                      >
                        <FileTypeIcon extension={result.extension} />
                        <span className="truncate font-semibold text-ink">{result.title}</span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            ) : null}

            <PrimaryButton
              type="button"
              className="w-full"
              disabled={!selectedTarget || isSaving}
              onClick={() => void handleCreate()}
            >
              {isSaving ? <Spinner size="sm" /> : <Link2 className="h-4 w-4" />}
              İlişkilendir
            </PrimaryButton>
          </div>
        ) : null}

        {items.length === 0 ? (
          <p className="text-sm text-slate-500">Henüz ilişkilendirilmiş belge yok.</p>
        ) : (
          <ul className="space-y-2">
            {items.map((item) => (
              <li key={item.id} className="rounded-lg border border-line bg-white p-3">
                <div className="flex items-start gap-2">
                  <FileTypeIcon extension={item.relatedDocument.fileExtension} />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-bold text-ink">{item.relatedDocument.originalFileName}</p>
                    <p className="mt-1 text-xs font-semibold text-violet-brand">
                      {relationshipTypeLabels[item.relationshipType] ?? item.relationshipType}
                      <span className="text-slate-400"> · {item.direction === "outgoing" ? "Giden" : "Gelen"}</span>
                    </p>
                  </div>
                </div>
                <div className="mt-3 flex gap-2">
                  <Link
                    className="inline-flex h-8 flex-1 items-center justify-center gap-1.5 rounded-lg border border-line bg-slate-50 text-xs font-semibold text-slate-700 transition hover:border-violet-brand hover:text-violet-brand"
                    to={`/documents/${item.relatedDocument.id}`}
                  >
                    <ExternalLink className="h-3.5 w-3.5" />
                    Aç
                  </Link>
                  <button
                    type="button"
                    className="inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-red-600 transition hover:border-red-300 hover:bg-red-50 disabled:opacity-60"
                    disabled={removingId === item.id}
                    onClick={() => void handleRemove(item.id)}
                  >
                    {removingId === item.id ? <Spinner size="sm" /> : <Trash2 className="h-3.5 w-3.5" />}
                    Kaldır
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="space-y-3">
        <h4 className="text-xs font-bold uppercase tracking-wide text-slate-400">AI Önerileri</h4>

        {suggestions.length === 0 ? (
          <div className="rounded-xl border border-dashed border-line bg-slate-50/80 px-4 py-5 text-center">
            <Sparkles className="mx-auto h-5 w-5 text-violet-brand" />
            <p className="mt-2 text-sm text-slate-600">Henüz AI ilişki önerisi yok.</p>
          </div>
        ) : (
          <ul className="space-y-2">
            {suggestions.map((suggestion) => {
              const similarity = formatSimilarityPercent(suggestion.similarity);
              const type = (suggestion.relationshipType as RelationshipType) ?? "related";
              return (
                <li key={suggestion.id} className="rounded-lg border border-violet-100 bg-violet-50/40 p-3">
                  <div className="flex items-start gap-2">
                    <FileTypeIcon extension={suggestion.targetDocument.fileExtension} />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-bold text-ink">{suggestion.targetDocument.originalFileName}</p>
                      <p className="mt-1 text-xs font-semibold text-violet-brand">
                        {relationshipTypeLabels[type] ?? type}
                        {similarity ? <span className="text-slate-500"> · Benzerlik {similarity}</span> : null}
                      </p>
                      {suggestion.reason ? <p className="mt-2 text-xs text-slate-600">{suggestion.reason}</p> : null}
                    </div>
                  </div>
                  <div className="mt-3 flex flex-wrap gap-2">
                    <Link
                      className="inline-flex h-8 flex-1 min-w-[96px] items-center justify-center gap-1.5 rounded-lg border border-line bg-white text-xs font-semibold text-slate-700 transition hover:border-violet-brand hover:text-violet-brand"
                      to={`/documents/${suggestion.targetDocument.id}`}
                    >
                      <ExternalLink className="h-3.5 w-3.5" />
                      Belgeye Git
                    </Link>
                    <button
                      type="button"
                      className="inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border border-emerald-200 bg-emerald-50 px-3 text-xs font-semibold text-emerald-700 transition hover:bg-emerald-100 disabled:opacity-60"
                      disabled={busySuggestionId === suggestion.id}
                      onClick={() => void handleAcceptSuggestion(suggestion)}
                    >
                      {busySuggestionId === suggestion.id ? <Spinner size="sm" /> : <Check className="h-3.5 w-3.5" />}
                      Kabul Et
                    </button>
                    <button
                      type="button"
                      className="inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-slate-600 transition hover:border-red-300 hover:bg-red-50 hover:text-red-600 disabled:opacity-60"
                      disabled={busySuggestionId === suggestion.id}
                      onClick={() => void handleRejectSuggestion(suggestion)}
                    >
                      {busySuggestionId === suggestion.id ? <Spinner size="sm" /> : <X className="h-3.5 w-3.5" />}
                      Reddet
                    </button>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </section>
    </div>
  );
}
