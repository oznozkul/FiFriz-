import { ChevronLeft, ChevronRight, Download, FileText, RefreshCw } from "lucide-react";
import { createRef, useEffect, useMemo, useRef, useState, type RefObject } from "react";
import { renderAsync } from "docx-preview";
import { ExcelSheetPreview } from "../documents/ExcelSheetPreview";
import { LoadingState, EmptyState, Spinner, Button } from "../common";
import type { ApiDocument } from "../../types/document";
import { parseXlsxWorkbook, type ExcelSheetData } from "../../utils/xlsxPreview";
function FilePreview({
  document,
  fileUrl,
  isFileLoading,
  fileError,
  onRetry,
  extractedText,
  isTextLoading,
  searchQuery,
  activeMatchIndex,
  activeSnippet,
}: {
  document: ApiDocument;
  fileUrl: string;
  isFileLoading: boolean;
  fileError: string | null;
  onRetry?: () => void;
  extractedText: string;
  isTextLoading: boolean;
  searchQuery: string;
  activeMatchIndex: number;
  activeSnippet?: HighlightResult["snippets"][number];
}) {
  const extension = document.file_extension.toLowerCase();

  if (isFileLoading) {
    return <LoadingState label="Dosya önizleme yükleniyor..." className="mx-auto max-w-[760px]" />;
  }

  if (fileError || !fileUrl) {
    return (
      <EmptyState
        title="Önizleme yüklenemedi"
        description={fileError ?? "Dosya önizleme için hazır değil."}
        action={
          onRetry ? (
            <Button variant="secondary" onClick={onRetry}>
              Tekrar dene
            </Button>
          ) : undefined
        }
      />
    );
  }

  if (extension === "pdf") {
    return (
      <iframe
        className="mx-auto min-h-[min(780px,70vh)] w-full max-w-[960px] rounded-xl border border-line bg-white shadow-panel"
        src={fileUrl}
        title={document.original_file_name}
      />
    );
  }

  if (["png", "jpg", "jpeg"].includes(extension)) {
    return (
      <div className="flex min-h-[620px] items-center justify-center">
        <img className="max-h-[760px] max-w-full rounded-lg border border-line bg-white object-contain shadow-panel" src={fileUrl} alt={document.original_file_name} />
      </div>
    );
  }

  if (extension === "docx") {
    return (
      <DocxFilePreview
        document={document}
        fileUrl={fileUrl}
        extractedText={extractedText}
        isTextLoading={isTextLoading}
        searchQuery={searchQuery}
        activeMatchIndex={activeMatchIndex}
        activeSnippet={activeSnippet}
      />
    );
  }

  if (extension === "xlsx") {
    return (
      <XlsxFilePreview
        document={document}
        fileUrl={fileUrl}
        extractedText={extractedText}
        isTextLoading={isTextLoading}
        searchQuery={searchQuery}
        activeMatchIndex={activeMatchIndex}
        activeSnippet={activeSnippet}
      />
    );
  }

  if (extension === "txt") {
    return (
      <TxtFilePreview
        document={document}
        fileUrl={fileUrl}
        extractedText={extractedText}
        isTextLoading={isTextLoading}
        searchQuery={searchQuery}
        activeMatchIndex={activeMatchIndex}
        activeSnippet={activeSnippet}
      />
    );
  }

  return (
    <div className="mx-auto max-w-[760px] rounded-xl border border-line bg-white px-8 py-10 text-center shadow-panel">
      <p className="text-lg font-bold text-ink">{document.original_file_name}</p>
      <p className="mt-2 text-sm text-slate-500">Bu dosya türü için önizleme desteklenmiyor. Aç veya İndir ile dosyayı görüntüleyebilirsiniz.</p>
    </div>
  );
}

function TxtFilePreview({
  document,
  fileUrl,
  extractedText,
  isTextLoading,
  searchQuery,
}: {
  document: ApiDocument;
  fileUrl: string;
  extractedText: string;
  isTextLoading: boolean;
  searchQuery: string;
  activeMatchIndex: number;
  activeSnippet?: HighlightResult["snippets"][number];
}) {
  const [content, setContent] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const previewText = (extractedText.trim() || content).trim();
  const highlight = useMemo(() => buildHighlightParts(previewText, searchQuery), [previewText, searchQuery]);

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setIsLoading(true);
      setError(null);
      try {
        if (extractedText.trim()) {
          if (!cancelled) setContent(extractedText);
          return;
        }
        const response = await fetch(fileUrl);
        if (!response.ok) {
          throw new Error("TXT dosyası alınamadı.");
        }
        const text = await response.text();
        if (!cancelled) setContent(text);
      } catch (loadError) {
        if (!cancelled) {
          setError(loadError instanceof Error ? loadError.message : "TXT önizleme oluşturulamadı.");
          setContent("");
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };

    void load();

    return () => {
      cancelled = true;
    };
  }, [extractedText, fileUrl]);

  const showLoading = isLoading || (isTextLoading && !previewText);

  return (
    <div className="mx-auto max-w-[1040px] overflow-hidden rounded-lg border border-line bg-white shadow-panel">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-6 py-4">
        <div className="min-w-0">
          <p className="truncate text-lg font-bold text-ink">{document.original_file_name}</p>
          <p className="mt-1 text-sm text-slate-500">Metin dosyası önizlemesi</p>
        </div>
        <a className="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand" href={fileUrl} download>
          <Download className="h-4 w-4" />
          İndir
        </a>
      </div>
      <div className="min-h-[620px] bg-slate-50 p-5">
        {showLoading ? <LoadingState label="TXT önizleme yükleniyor..." /> : null}
        {error ? <p className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-semibold text-amber-800">{error}</p> : null}
        {!showLoading && !error && previewText ? (
          <div className="max-h-[760px] overflow-y-auto whitespace-pre-wrap rounded-lg border border-line bg-white p-5 text-left text-sm leading-7 text-slate-700">
            {highlight.matches.length > 0 ? (
              <HighlightedPlainText text={previewText} query={searchQuery} />
            ) : (
              previewText
            )}
          </div>
        ) : null}
      </div>
    </div>
  );
}

function HighlightedPlainText({ text, query }: { text: string; query: string }) {
  const terms = query
    .toLocaleLowerCase("tr-TR")
    .split(/[^\p{L}\p{N}]+/gu)
    .map((term) => term.trim())
    .filter((term) => term.length >= 2);

  if (terms.length === 0) {
    return <>{text}</>;
  }

  const pattern = terms.map(escapeRegExp).join("|");
  const regex = new RegExp(`(${pattern})[\\p{L}\\p{N}_-]*`, "giu");
  const parts: Array<{ text: string; hit: boolean }> = [];
  let lastIndex = 0;

  for (const match of text.matchAll(regex)) {
    const index = match.index ?? 0;
    if (index > lastIndex) {
      parts.push({ text: text.slice(lastIndex, index), hit: false });
    }
    parts.push({ text: match[0], hit: true });
    lastIndex = index + match[0].length;
  }

  if (lastIndex < text.length) {
    parts.push({ text: text.slice(lastIndex), hit: false });
  }

  return (
    <>
      {parts.map((part, index) =>
        part.hit ? (
          <mark key={index} className="rounded bg-yellow-300 px-1 font-bold text-amber-950">
            {part.text}
          </mark>
        ) : (
          <span key={index}>{part.text}</span>
        ),
      )}
    </>
  );
}

function DocxFilePreview({
  document,
  fileUrl,
  extractedText,
  isTextLoading,
  searchQuery,
  activeMatchIndex,
  activeSnippet,
}: {
  document: ApiDocument;
  fileUrl: string;
  extractedText: string;
  isTextLoading: boolean;
  searchQuery: string;
  activeMatchIndex: number;
  activeSnippet?: HighlightResult["snippets"][number];
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [isPreviewLoading, setIsPreviewLoading] = useState(true);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [isPreviewReady, setIsPreviewReady] = useState(false);
  const previewText = extractedText.trim();

  useEffect(() => {
    let cancelled = false;
    const container = containerRef.current;
    if (!container) return;

    const renderDocument = async () => {
      setIsPreviewLoading(true);
      setIsPreviewReady(false);
      setPreviewError(null);
      container.innerHTML = "";

      try {
        const response = await fetch(fileUrl);
        if (!response.ok) {
          throw new Error("DOCX dosyası alınamadı.");
        }

        const blob = await response.blob();
        if (cancelled) return;

        await renderAsync(blob, container, undefined, {
          className: "docx-render",
          inWrapper: true,
          breakPages: true,
          ignoreLastRenderedPageBreak: true,
          renderHeaders: true,
          renderFooters: true,
          renderFootnotes: true,
          renderEndnotes: true,
          renderComments: false,
          renderChanges: false,
        });

        if (!cancelled) {
          setIsPreviewLoading(false);
          setIsPreviewReady(true);
        }
      } catch (error) {
        if (!cancelled) {
          container.innerHTML = "";
          setPreviewError(error instanceof Error ? error.message : "DOCX önizleme oluşturulamadı.");
          setIsPreviewLoading(false);
          setIsPreviewReady(false);
        }
      }
    };

    void renderDocument();

    return () => {
      cancelled = true;
      container.innerHTML = "";
    };
  }, [fileUrl]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container || !isPreviewReady) return;

    clearDocxPreviewHighlights(container);
    const highlightedCount = highlightPreviewMatches(container, searchQuery, "docx-search-hit");
    if (highlightedCount === 0) return;

    scrollPreviewMatchIntoView(container, activeMatchIndex, activeSnippet, "docx-search-hit", "docx-search-hit-active");
  }, [activeMatchIndex, activeSnippet, isPreviewReady, searchQuery]);

  return (
    <div className="mx-auto max-w-[1040px] overflow-hidden rounded-lg border border-line bg-white shadow-panel">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-6 py-4">
        <div className="min-w-0">
          <p className="truncate text-lg font-bold text-ink">{document.original_file_name}</p>
          <p className="mt-1 text-sm text-slate-500">DOCX önizleme Word görünümüne yakın HTML olarak render edilir.</p>
        </div>
        <a className="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand" href={fileUrl} download>
          <Download className="h-4 w-4" />
          İndir
        </a>
      </div>
      <div className="relative min-h-[620px] bg-slate-200/70">
        {isPreviewLoading && (
          <div className="absolute inset-x-0 top-0 z-10 border-b border-line bg-white/95 px-5 py-3 text-sm font-semibold text-slate-500">
            DOCX önizleme yükleniyor...
          </div>
        )}
        {previewError ? (
          <div className="p-5">
            <p className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-semibold text-amber-800">{previewError}</p>
            {isTextLoading ? (
              <p className="mt-4 text-center text-sm font-semibold text-slate-500">Çıkarılmış metin yükleniyor...</p>
            ) : previewText ? (
              <div className="mt-4 max-h-[620px] overflow-y-auto whitespace-pre-wrap rounded-lg border border-line bg-white p-5 text-left text-sm leading-7 text-slate-700">
                {previewText}
              </div>
            ) : null}
          </div>
        ) : null}
        <div ref={containerRef} className="docx-preview-surface overflow-auto p-4 sm:p-6" />
      </div>
    </div>
  );
}

type ParsedXlsxSheet = ExcelSheetData;

function XlsxFilePreview({
  document,
  fileUrl,
  extractedText,
  isTextLoading,
  searchQuery,
  activeMatchIndex,
  activeSnippet,
}: {
  document: ApiDocument;
  fileUrl: string;
  extractedText: string;
  isTextLoading: boolean;
  searchQuery: string;
  activeMatchIndex: number;
  activeSnippet?: HighlightResult["snippets"][number];
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [sheets, setSheets] = useState<ParsedXlsxSheet[]>([]);
  const [activeSheetIndex, setActiveSheetIndex] = useState(0);
  const [isPreviewLoading, setIsPreviewLoading] = useState(true);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [isPreviewReady, setIsPreviewReady] = useState(false);
  const previewText = extractedText.trim();
  const activeSheet = sheets[activeSheetIndex];

  useEffect(() => {
    let cancelled = false;

    const loadWorkbook = async () => {
      setIsPreviewLoading(true);
      setIsPreviewReady(false);
      setPreviewError(null);
      setSheets([]);
      setActiveSheetIndex(0);

      try {
        const response = await fetch(fileUrl);
        if (!response.ok) {
          throw new Error("XLSX dosyası alınamadı.");
        }

        const buffer = await response.arrayBuffer();
        if (cancelled) return;

        const parsedSheets = parseXlsxWorkbook(buffer);
        if (parsedSheets.length === 0) {
          throw new Error("XLSX dosyasında sayfa bulunamadı.");
        }

        if (!cancelled) {
          setSheets(parsedSheets);
          setIsPreviewLoading(false);
          setIsPreviewReady(true);
        }
      } catch (error) {
        if (!cancelled) {
          setPreviewError(error instanceof Error ? error.message : "XLSX önizleme oluşturulamadı.");
          setIsPreviewLoading(false);
          setIsPreviewReady(false);
        }
      }
    };

    void loadWorkbook();

    return () => {
      cancelled = true;
    };
  }, [fileUrl]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container || !isPreviewReady || !activeSheet) return;

    clearPreviewHighlights(container, "xlsx-search-hit");
    const highlightedCount = highlightPreviewMatches(container, searchQuery, "xlsx-search-hit");
    if (highlightedCount === 0) return;

    scrollPreviewMatchIntoView(container, activeMatchIndex, activeSnippet, "xlsx-search-hit", "xlsx-search-hit-active");
  }, [activeMatchIndex, activeSheet, activeSnippet, isPreviewReady, searchQuery]);

  return (
    <div className="excel-workbook mx-auto max-w-[1280px] overflow-hidden rounded-lg border border-[#ababab] bg-[#f3f3f3] shadow-panel">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#d4d4d4] bg-white px-5 py-3">
        <div className="min-w-0">
          <p className="truncate text-base font-bold text-ink">{document.original_file_name}</p>
          <p className="mt-0.5 text-xs text-slate-500">Excel görünümlü tablo önizleme</p>
        </div>
        <a className="inline-flex h-8 items-center justify-center gap-2 rounded border border-[#d4d4d4] bg-[#f8f8f8] px-3 text-xs font-semibold text-slate-700 transition hover:bg-white" href={fileUrl} download>
          <Download className="h-3.5 w-3.5" />
          İndir
        </a>
      </div>

      <div className="excel-name-bar border-b border-[#d4d4d4] bg-[#e9e9e9] px-3 py-1.5 text-[11px] font-semibold text-[#444]">
        {activeSheet ? activeSheet.name : "Sayfa"}
      </div>

      <div className="relative min-h-[620px] bg-[#e6e6e6]">
        {isPreviewLoading && (
          <div className="absolute inset-x-0 top-0 z-20 border-b border-[#d4d4d4] bg-white/95 px-5 py-3 text-sm font-semibold text-slate-500">
            Excel önizleme yükleniyor...
          </div>
        )}
        {previewError ? (
          <div className="p-5">
            <p className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-semibold text-amber-800">{previewError}</p>
            {isTextLoading ? (
              <p className="mt-4 text-center text-sm font-semibold text-slate-500">Çıkarılmış metin yükleniyor...</p>
            ) : previewText ? (
              <div className="mt-4 max-h-[620px] overflow-y-auto whitespace-pre-wrap rounded-lg border border-line bg-white p-5 text-left text-sm leading-7 text-slate-700">
                {previewText}
              </div>
            ) : null}
          </div>
        ) : null}

        {!previewError && activeSheet ? (
          <div ref={containerRef} className="xlsx-preview-surface p-3">
            <ExcelSheetPreview sheet={activeSheet} />
          </div>
        ) : null}
      </div>

      {sheets.length > 0 && (
        <div className="excel-sheet-tabs">
          <div className="excel-sheet-tabs-scroll">
            {sheets.map((sheet, index) => (
              <button
                key={`${sheet.name}-${index}`}
                className={`excel-sheet-tab ${index === activeSheetIndex ? "excel-sheet-tab-active" : ""}`}
                type="button"
                onClick={() => setActiveSheetIndex(index)}
              >
                {sheet.name}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

type HighlightResult = {
  matches: string[];
  snippets: Array<{
    before: string;
    match: string;
    after: string;
  }>;
};

function SearchMatchesPanel({
  query,
  isLoading,
  hasExtractedText,
  highlight,
  matchRefs,
  activeMatchIndex,
  onPrevious,
  onNext,
  onSelect,
}: {
  query: string;
  isLoading: boolean;
  hasExtractedText: boolean;
  highlight: HighlightResult;
  matchRefs: RefObject<HTMLElement | null>[];
  activeMatchIndex: number;
  onPrevious: () => void;
  onNext: () => void;
  onSelect: (index: number) => void;
}) {
  return (
    <section className="panel overflow-hidden">
      <div className="border-b border-line bg-amber-50 px-5 py-4">
        <h2 className="text-base font-bold text-amber-950">Arama Eşleşmeleri</h2>
        <p className="mt-1 text-sm font-semibold text-amber-900">
          {query} {highlight.matches.length > 0 ? `(${activeMatchIndex + 1}/${highlight.matches.length})` : ""}
        </p>
      </div>
      <div className="p-4">
        {isLoading ? (
          <Spinner size="sm" label="Eşleşmeler yükleniyor..." />
        ) : !hasExtractedText ? (
          <p className="text-sm leading-6 text-slate-500">Bu belge için çıkarılmış metin bulunmadığı için eşleşme paneli gösterilemiyor.</p>
        ) : highlight.matches.length === 0 ? (
          <p className="text-sm leading-6 text-slate-500">Aranan ifade çıkarılmış metinde bulunamadı.</p>
        ) : (
          <>
            <div className="mb-3 flex gap-2">
              <button className="inline-flex h-9 flex-1 items-center justify-center gap-2 rounded-lg border border-amber-200 bg-white px-3 text-xs font-bold text-amber-900 transition hover:bg-amber-100" type="button" onClick={onPrevious}>
                <ChevronLeft className="h-4 w-4" />
                Önceki
              </button>
              <button className="inline-flex h-9 flex-1 items-center justify-center gap-2 rounded-lg border border-amber-200 bg-white px-3 text-xs font-bold text-amber-900 transition hover:bg-amber-100" type="button" onClick={onNext}>
                Sonraki
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
            <div className="max-h-[420px] space-y-3 overflow-y-auto pr-1">
              {highlight.snippets.map((snippet, index) => (
                <button
                  key={`${snippet.before}-${index}`}
                  className={`w-full rounded-lg border p-3 text-left text-sm leading-6 transition ${index === activeMatchIndex ? "border-yellow-400 bg-yellow-50" : "border-line bg-white hover:border-yellow-300 hover:bg-amber-50/60"}`}
                  type="button"
                  onClick={() => {
                    onSelect(index);
                    matchRefs[index]?.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
                  }}
                >
                  <span className="text-slate-500">{snippet.before}</span>
                  <mark ref={matchRefs[index]} className="rounded bg-yellow-300 px-1 font-bold text-amber-950">
                    {snippet.match}
                  </mark>
                  <span className="text-slate-500">{snippet.after}</span>
                </button>
              ))}
            </div>
          </>
        )}
      </div>
    </section>
  );
}

function ExtractedTextPanel({
  fileTypeLabel,
  text,
  isLoading,
  isReprocessing,
  hasExtractedText,
  error,
  onReprocess,
}: {
  fileTypeLabel: string;
  text: string;
  isLoading: boolean;
  isReprocessing: boolean;
  hasExtractedText: boolean;
  error: string | null;
  onReprocess: () => void;
}) {
  const previewText = text.trim();

  return (
    <section className="panel overflow-hidden">
      <div className="border-b border-line px-5 py-4">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-base font-bold text-ink">Çıkarılmış Metin</h2>
            <p className="mt-1 text-sm text-slate-500">{hasExtractedText ? `${fileTypeLabel} içeriği aramaya dahil.` : "Bu belge için metin henüz yok."}</p>
          </div>
          <span className={`inline-flex h-8 items-center gap-2 rounded-lg px-3 text-xs font-bold ${hasExtractedText ? "bg-emerald-50 text-emerald-700" : "bg-amber-50 text-amber-800"}`}>
            <FileText className="h-4 w-4" />
            {hasExtractedText ? "Hazir" : "Bekliyor"}
          </span>
        </div>
      </div>
      <div className="space-y-3 p-4">
        {error && <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm font-semibold text-red-700">{error}</p>}
        {isLoading ? (
          <Spinner size="sm" label="Metin yükleniyor..." />
        ) : previewText ? (
          <div className="max-h-[280px] overflow-y-auto rounded-lg border border-line bg-slate-50 p-3 text-sm leading-6 text-slate-700">
            {previewText.slice(0, 2500)}
            {previewText.length > 2500 ? <span className="text-slate-400"> ...</span> : null}
          </div>
        ) : (
          <p className="text-sm leading-6 text-slate-500">Çıkarılmış metin görünmüyor. Belgeyi yeniden işleyerek mevcut {fileTypeLabel} dosyasından metin çıkarabilirsin.</p>
        )}
        <button
          className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-line bg-white text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand disabled:cursor-not-allowed disabled:opacity-60"
          type="button"
          onClick={onReprocess}
          disabled={isReprocessing}
        >
          <RefreshCw className={`h-4 w-4 ${isReprocessing ? "animate-spin" : ""}`} />
          {isReprocessing ? "İşleniyor..." : "Metni Yeniden İşle"}
        </button>
      </div>
    </section>
  );
}

function buildHighlightParts(text: string, query: string): HighlightResult {
  const terms = query
    .toLocaleLowerCase("tr-TR")
    .split(/[^\p{L}\p{N}]+/gu)
    .map((term) => term.trim())
    .filter((term) => term.length >= 2);

  if (!text || terms.length === 0) {
    return { matches: [], snippets: [] };
  }

  const pattern = terms.map(escapeRegExp).join("|");
  const regex = new RegExp(`(${pattern})[\\p{L}\\p{N}_-]*`, "giu");
  const matches: string[] = [];
  const snippets: HighlightResult["snippets"] = [];

  for (const match of text.matchAll(regex)) {
    const index = match.index ?? 0;
    const matchText = match[0];
    matches.push(matchText);
    snippets.push(buildSnippet(text, index, matchText.length));
  }

  return { matches, snippets };
}

function clearPreviewHighlights(container: HTMLElement, hitClass: string) {
  const hits = Array.from(container.querySelectorAll<HTMLElement>(`.${hitClass}`));
  for (const hit of hits) {
    hit.replaceWith(document.createTextNode(hit.textContent ?? ""));
  }
  container.normalize();
}

function clearDocxPreviewHighlights(container: HTMLElement) {
  clearPreviewHighlights(container, "docx-search-hit");
}

function highlightPreviewMatches(container: HTMLElement, query: string, hitClass: string) {
  const terms = extractSearchTerms(query);
  if (terms.length === 0) return 0;

  const regex = new RegExp(`(${terms.map(escapeRegExp).join("|")})[\\p{L}\\p{N}_-]*`, "giu");
  const textNodes: Text[] = [];
  const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, {
    acceptNode(node) {
      const parent = node.parentElement;
      if (!parent) return NodeFilter.FILTER_REJECT;
      if (["STYLE", "SCRIPT"].includes(parent.tagName)) return NodeFilter.FILTER_REJECT;
      if (parent.closest(`.${hitClass}`)) return NodeFilter.FILTER_REJECT;
      if (!node.textContent?.trim()) return NodeFilter.FILTER_REJECT;
      return NodeFilter.FILTER_ACCEPT;
    },
  });

  let currentNode = walker.nextNode();
  while (currentNode) {
    textNodes.push(currentNode as Text);
    currentNode = walker.nextNode();
  }

  let matchIndex = 0;
  for (const node of textNodes) {
    const text = node.textContent ?? "";
    regex.lastIndex = 0;

    let lastIndex = 0;
    let hasMatch = false;
    const fragment = document.createDocumentFragment();

    for (const match of text.matchAll(regex)) {
      const index = match.index ?? 0;
      const matchText = match[0];
      if (!matchText) continue;

      if (index > lastIndex) {
        fragment.appendChild(document.createTextNode(text.slice(lastIndex, index)));
      }

      const mark = document.createElement("mark");
      mark.className = hitClass;
      mark.dataset.previewMatchIndex = String(matchIndex);
      mark.textContent = matchText;
      fragment.appendChild(mark);

      matchIndex += 1;
      lastIndex = index + matchText.length;
      hasMatch = true;
    }

    if (!hasMatch) continue;

    if (lastIndex < text.length) {
      fragment.appendChild(document.createTextNode(text.slice(lastIndex)));
    }

    node.replaceWith(fragment);
  }

  return matchIndex;
}

function scrollPreviewMatchIntoView(
  container: HTMLElement,
  activeMatchIndex: number,
  activeSnippet: HighlightResult["snippets"][number] | undefined,
  hitClass: string,
  activeHitClass: string,
) {
  const hits = Array.from(container.querySelectorAll<HTMLElement>(`.${hitClass}`));
  if (hits.length === 0) return;

  const targetHit = findBestPreviewHit(hits, activeMatchIndex, activeSnippet);
  for (const hit of hits) {
    hit.classList.remove(activeHitClass);
  }

  targetHit.classList.add(activeHitClass);
  targetHit.scrollIntoView({ behavior: "smooth", block: "center", inline: "nearest" });
}

function findBestPreviewHit(hits: HTMLElement[], activeMatchIndex: number, activeSnippet?: HighlightResult["snippets"][number]) {
  const fallbackIndex = Math.min(Math.max(activeMatchIndex, 0), hits.length - 1);
  const fallbackHit = hits[fallbackIndex];

  if (!activeSnippet) return fallbackHit;

  const keywords = extractSnippetKeywords(activeSnippet);
  if (keywords.length === 0) return fallbackHit;

  let bestHit = fallbackHit;
  let bestScore = -1;

  for (const hit of hits) {
    const context = normalizePreviewText(getHitContext(hit));
    let score = 0;
    for (const keyword of keywords) {
      if (context.includes(keyword)) {
        score += keyword.length;
      }
    }

    if (score > bestScore) {
      bestScore = score;
      bestHit = hit;
    }
  }

  return bestScore > 0 ? bestHit : fallbackHit;
}

function getHitContext(hit: HTMLElement) {
  const contextElement = hit.closest("p,td,th,li,section") ?? hit.parentElement;
  const text = contextElement?.textContent ?? hit.textContent ?? "";
  const index = text.indexOf(hit.textContent ?? "");
  if (index < 0) return text;

  return text.slice(Math.max(0, index - 180), index + 240);
}

function extractSnippetKeywords(snippet: HighlightResult["snippets"][number]) {
  const text = `${snippet.before} ${snippet.match} ${snippet.after}`;
  const stopWords = new Set(["icin", "olan", "ve", "ile", "bir", "bu", "the", "and"]);

  return normalizePreviewText(text)
    .split(/[^a-z0-9_]+/g)
    .map((word: string) => word.trim())
    .filter((word: string) => word.length >= 4 && !stopWords.has(word))
    .slice(0, 16);
}

function normalizePreviewText(value: string) {
  return value
    .toLocaleLowerCase("tr-TR")
    .replace(/ç/g, "c")
    .replace(/ğ/g, "g")
    .replace(/ı/g, "i")
    .replace(/i̇/g, "i")
    .replace(/ö/g, "o")
    .replace(/ş/g, "s")
    .replace(/ü/g, "u")
    .replace(/â/g, "a")
    .replace(/î/g, "i")
    .replace(/û/g, "u");
}

function extractSearchTerms(query: string) {
  return query
    .toLocaleLowerCase("tr-TR")
    .split(/[^\p{L}\p{N}]+/gu)
    .map((term) => term.trim())
    .filter((term) => term.length >= 2);
}

function buildSnippet(text: string, matchStart: number, matchLength: number) {
  const radius = 90;
  const snippetStart = Math.max(0, matchStart - radius);
  const snippetEnd = Math.min(text.length, matchStart + matchLength + radius);
  const beforePrefix = snippetStart > 0 ? "... " : "";
  const afterSuffix = snippetEnd < text.length ? " ..." : "";

  return {
    before: `${beforePrefix}${text.slice(snippetStart, matchStart)}`,
    match: text.slice(matchStart, matchStart + matchLength),
    after: `${text.slice(matchStart + matchLength, snippetEnd)}${afterSuffix}`,
  };
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function MetadataItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="px-5 py-4">
      <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</dt>
      <dd className="mt-1 break-words font-semibold text-slate-700">{value}</dd>
    </div>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(bytes / 1024, 1).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
export { FilePreview, SearchMatchesPanel, ExtractedTextPanel, buildHighlightParts };
export type { HighlightResult };
