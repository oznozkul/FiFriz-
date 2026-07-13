import {
  ArrowRight,
  Clock,
  FileSearch,
  Pin,
  Search,
  Sparkles,
  User,
  X,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { aiActionCatalog } from "../../data/commandPaletteCatalog";
import { useDebounce } from "../../hooks/useDebounce";
import { useFolderLabels } from "../../hooks/useFolderLabels";
import {
  isPinnedAction,
  togglePinnedAction,
} from "../../services/commandPaletteStorage";
import { hybridSearchDocuments } from "../../services/hybridSearchService";
import { SearchModeToggle } from "../search/SearchModeToggle";
import { SearchMatchTypeBadge } from "../search/SearchMatchTypeBadge";
import { globalSearch } from "../../services/globalSearchService";
import type { GlobalSearchResponse } from "../../types/globalSearch";
import type { HybridSearchMatchType, SearchMode } from "../../types/hybridSearch";
import { getSearchMode } from "../../utils/searchMode";
import { Button } from "./Button";
import { buildEmptyStateItems, buildItemsFromResponse } from "./commandPaletteBuilders";
import { stripSnippetMarkup } from "../../services/documentService";
import { filterCatalogByQuery, getPaletteGroupOrder, highlightText, type PaletteItem } from "./commandPaletteUtils";
import { FileTypeIcon } from "./FileTypeIcon";
import { Spinner } from "./Spinner";

type CommandPaletteProps = {
  isOpen: boolean;
  onClose: () => void;
  onUploadClick: () => void;
};

function PaletteSkeleton() {
  return (
    <div className="space-y-1 p-2">
      {[1, 2, 3, 4, 5].map((row) => (
        <div key={row} className="flex items-start gap-3 px-4 py-3">
          <div className="h-9 w-9 shrink-0 animate-pulse rounded-lg bg-slate-100" />
          <div className="min-w-0 flex-1 space-y-2">
            <div className="h-4 w-3/5 animate-pulse rounded bg-slate-100" />
            <div className="h-3 w-2/5 animate-pulse rounded bg-slate-100" />
          </div>
        </div>
      ))}
    </div>
  );
}

function PaletteResultRow({
  item,
  flatIndex,
  isActive,
  trimmedQuery,
  onHover,
  onSelect,
  onTogglePin,
}: {
  item: PaletteItem;
  flatIndex: number;
  isActive: boolean;
  trimmedQuery: string;
  onHover: () => void;
  onSelect: () => void;
  onTogglePin: () => void;
}) {
  const Icon = item.icon;
  const pinned = item.actionId ? isPinnedAction(item.actionId) : false;

  return (
    <li>
      <button
        type="button"
        data-index={flatIndex}
        id={`palette-item-${flatIndex}`}
        role="option"
        aria-selected={isActive}
        className={`flex w-full items-start gap-3 px-4 py-3 text-left transition focus:outline-none ${
          isActive ? "bg-violet-soft ring-1 ring-inset ring-violet-ring" : "hover:bg-slate-50"
        }`}
        onMouseEnter={onHover}
        onClick={onSelect}
      >
        <span className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-50 text-slate-600 ring-1 ring-line">
          {item.extension ? <FileTypeIcon extension={item.extension} /> : Icon ? <Icon className="h-4 w-4" /> : <User className="h-4 w-4" />}
        </span>
        <span className="min-w-0 flex-1">
          <span className="flex flex-wrap items-center gap-2">
            <span className="truncate font-semibold text-ink">{highlightText(item.label, trimmedQuery)}</span>
            {item.matchType ? <SearchMatchTypeBadge matchType={item.matchType} /> : null}
            {item.extension ? (
              <span className="shrink-0 rounded bg-slate-100 px-2 py-0.5 text-[10px] font-bold uppercase text-slate-500">{item.extension}</span>
            ) : null}
            {item.kind === "search" ? (
              <span className="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-500">
                <Clock className="h-3 w-3" />
                Son arama
              </span>
            ) : null}
          </span>
          {item.subtitle ? <span className="mt-0.5 block truncate text-xs text-slate-500">{item.subtitle}</span> : null}
          {item.snippet ? (
            <span className="mt-1 line-clamp-2 block text-xs leading-5 text-slate-500">
              {highlightText(stripSnippetMarkup(item.snippet), trimmedQuery)}
            </span>
          ) : null}
        </span>
        {item.pinEligible && item.actionId ? (
          <span
            role="button"
            tabIndex={-1}
            aria-label={pinned ? "Sabitlemeyi kaldır" : "Sabitle"}
            className={`mt-1 rounded-lg p-1.5 text-slate-400 transition hover:bg-white hover:text-violet-brand focus:outline-none focus:ring-4 focus:ring-violet-ring/70 ${
              pinned ? "text-violet-brand" : ""
            }`}
            onClick={(event) => {
              event.stopPropagation();
              onTogglePin();
            }}
          >
            <Pin className={`h-4 w-4 ${pinned ? "fill-current" : ""}`} />
          </span>
        ) : null}
      </button>
    </li>
  );
}

export function CommandPalette({ isOpen, onClose, onUploadClick }: CommandPaletteProps) {
  const navigate = useNavigate();
  const folderLabels = useFolderLabels();
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<GlobalSearchResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const [pinnedRevision, setPinnedRevision] = useState(0);
  const [searchMode, setSearchMode] = useState<SearchMode>(() => getSearchMode());
  const [documentMatchTypes, setDocumentMatchTypes] = useState<Map<string, HybridSearchMatchType>>(new Map());
  const [retryToken, setRetryToken] = useState(0);
  const debouncedQuery = useDebounce(query, 280);

  const groupOrder = useMemo(() => getPaletteGroupOrder(folderLabels.plural), [folderLabels.plural]);

  useEffect(() => {
    if (!isOpen) {
      setQuery("");
      setResults(null);
      setError(null);
      setIsLoading(false);
      setActiveIndex(0);
      return;
    }

    const frame = window.requestAnimationFrame(() => {
      inputRef.current?.focus();
      inputRef.current?.select();
    });
    return () => window.cancelAnimationFrame(frame);
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;

    const trimmed = debouncedQuery.trim();
    if (!trimmed) {
      setIsLoading(true);
      setError(null);
      setDocumentMatchTypes(new Map());
      globalSearch("")
        .then((data) => setResults(data))
        .catch(() =>
          setResults({
            documents: [],
            folders: [],
            tags: [],
            collections: [],
            sharedDocuments: [],
            integrations: [],
            users: [],
            actions: [],
            settings: [],
          }),
        )
        .finally(() => setIsLoading(false));
      return;
    }

    let cancelled = false;
    setIsLoading(true);
    setError(null);
    setDocumentMatchTypes(new Map());

    const loadResults = async () => {
      try {
        const globalData = await globalSearch(trimmed);

        if (searchMode === "smart") {
          const hybridDocs = await hybridSearchDocuments(trimmed).catch(() => []);
          if (cancelled) return;

          const matchTypes = new Map(hybridDocs.map((doc) => [doc.documentId, doc.matchType]));
          setDocumentMatchTypes(matchTypes);

          const documents =
            hybridDocs.length > 0
              ? hybridDocs.map((doc) => ({
                  documentId: doc.documentId,
                  title: doc.title,
                  originalFileName: doc.originalFileName,
                  fileExtension: doc.fileExtension,
                  matchedSnippet: doc.matchedSnippet,
                  folderName: doc.folderName,
                }))
              : globalData.documents;

          setResults({ ...globalData, documents });
          return;
        }

        if (!cancelled) setResults(globalData);
      } catch (searchError) {
        if (!cancelled) {
          setError(searchError instanceof Error ? searchError.message : "Arama tamamlanamadı.");
          setResults(null);
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };

    void loadResults();

    return () => {
      cancelled = true;
    };
  }, [debouncedQuery, isOpen, retryToken, searchMode]);

  const items = useMemo(() => {
    if (!isOpen) return [];

    const trimmed = debouncedQuery.trim();
    const builderOptions = {
      folderGroupLabel: folderLabels.plural,
      folderSingularLower: folderLabels.singularLower,
      documentMatchTypes,
      extraAiActions: filterCatalogByQuery(aiActionCatalog, trimmed),
    };

    if (!trimmed) {
      const emptyItems = buildEmptyStateItems(setQuery, navigate, onUploadClick, onClose, pinnedRevision, "", folderLabels);
      const apiItems = results
        ? buildItemsFromResponse(results, "", navigate, onUploadClick, onClose, builderOptions).filter(
            (item) => item.group === "Aksiyonlar" || item.group === "Ayarlar",
          )
        : [];
      return [...emptyItems, ...apiItems];
    }

    if (!results) return [];
    return buildItemsFromResponse(results, trimmed, navigate, onUploadClick, onClose, builderOptions);
  }, [
    debouncedQuery,
    documentMatchTypes,
    folderLabels.plural,
    folderLabels.singularLower,
    folderLabels,
    isOpen,
    navigate,
    onClose,
    onUploadClick,
    pinnedRevision,
    results,
  ]);

  useEffect(() => {
    setActiveIndex(0);
  }, [debouncedQuery, items.length, searchMode]);

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        onClose();
        return;
      }

      if (items.length === 0) return;

      if (event.key === "ArrowDown") {
        event.preventDefault();
        setActiveIndex((index) => (index + 1) % items.length);
        return;
      }

      if (event.key === "ArrowUp") {
        event.preventDefault();
        setActiveIndex((index) => (index - 1 + items.length) % items.length);
        return;
      }

      if (event.key === "Home") {
        event.preventDefault();
        setActiveIndex(0);
        return;
      }

      if (event.key === "End") {
        event.preventDefault();
        setActiveIndex(items.length - 1);
        return;
      }

      if (event.key === "Enter" && items[activeIndex]) {
        event.preventDefault();
        items[activeIndex].onSelect();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [activeIndex, isOpen, items, onClose]);

  useEffect(() => {
    const activeElement = listRef.current?.querySelector(`[data-index="${activeIndex}"]`);
    activeElement?.scrollIntoView({ block: "nearest" });
  }, [activeIndex]);

  if (!isOpen) return null;

  const grouped = groupOrder
    .map((group) => ({
      group,
      items: items.filter((item) => item.group === group),
    }))
    .filter((section) => section.items.length > 0);

  const trimmedQuery = query.trim();
  const showEmpty = !isLoading && !error && items.length === 0 && trimmedQuery.length > 0;
  const showIdle = !isLoading && !error && items.length === 0 && trimmedQuery.length === 0;
  const showViewAll = trimmedQuery.length >= 2 && !isLoading && !error;

  return (
    <div
      className="fixed inset-0 z-[100] flex flex-col bg-slate-900/45 backdrop-blur-[2px] sm:items-center sm:justify-center sm:p-4"
      onMouseDown={onClose}
    >
      <div
        className="flex h-[100dvh] w-full flex-col overflow-hidden border-line bg-white sm:h-auto sm:max-h-[min(80dvh,720px)] sm:max-w-2xl sm:rounded-2xl sm:border sm:shadow-2xl"
        onMouseDown={(event) => event.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Komut paleti"
      >
        <div className="flex items-center gap-2 border-b border-line px-3 py-3 sm:gap-3 sm:px-4 sm:py-3.5">
          <Search className="h-5 w-5 shrink-0 text-violet-brand" />
          <input
            ref={inputRef}
            value={query}
            className="min-w-0 flex-1 bg-transparent text-base font-medium text-ink placeholder:text-slate-400 focus:outline-none"
            placeholder="Belge, klasör, etiket veya komut ara..."
            aria-label="Arama"
            aria-controls="command-palette-listbox"
            aria-activedescendant={items[activeIndex] ? `palette-item-${activeIndex}` : undefined}
            role="combobox"
            aria-expanded="true"
            aria-autocomplete="list"
            onChange={(event) => setQuery(event.target.value)}
          />
          <SearchModeToggle className="hidden shrink-0 sm:inline-flex" onChange={setSearchMode} />
          <button
            type="button"
            className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-slate-400 transition hover:bg-slate-50 hover:text-ink focus:outline-none focus:ring-4 focus:ring-violet-ring/70"
            aria-label="Kapat"
            onClick={onClose}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="border-t border-line px-3 py-2 sm:hidden">
          <SearchModeToggle onChange={setSearchMode} />
          {showViewAll ? (
            <button
              type="button"
              className="mt-2 inline-flex w-full items-center justify-center gap-1 rounded-lg border border-line bg-white py-2 text-sm font-semibold text-violet-brand"
              onClick={() => {
                onClose();
                navigate(`/search?q=${encodeURIComponent(trimmedQuery)}`);
              }}
            >
              Tüm sonuçları gör
              <ArrowRight className="h-3.5 w-3.5" />
            </button>
          ) : null}
        </div>

        <div
          ref={listRef}
          id="command-palette-listbox"
          role="listbox"
          className="min-h-0 flex-1 overflow-y-auto sm:max-h-[min(58vh,520px)]"
        >
          {isLoading ? (
            <PaletteSkeleton />
          ) : error ? (
            <div className="px-6 py-10 text-center">
              <p className="text-sm font-bold text-ink">Arama başarısız</p>
              <p className="mt-1 text-sm text-red-600">{error}</p>
              <Button variant="secondary" className="mt-4" onClick={() => setRetryToken((value) => value + 1)}>
                Tekrar dene
              </Button>
            </div>
          ) : showEmpty ? (
            <div className="flex flex-col items-center px-6 py-12 text-center">
              <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-violet-soft text-violet-brand">
                <FileSearch className="h-6 w-6" />
              </div>
              <p className="text-base font-bold text-ink">Sonuç bulunamadı</p>
              <p className="mt-2 max-w-sm text-sm text-slate-500">
                Farklı bir anahtar kelime deneyin veya bir komut adı yazın.
              </p>
            </div>
          ) : showIdle ? (
            <div className="flex flex-col items-center px-6 py-10 text-center">
              <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-violet-soft text-violet-brand">
                <Sparkles className="h-6 w-6" />
              </div>
              <p className="text-base font-bold text-ink">Hızlı erişim</p>
              <p className="mt-2 max-w-sm text-sm text-slate-500">
                Yazmaya başlayın veya aşağıdaki sabitlenen ve son kullanılan öğelerden seçin.
              </p>
            </div>
          ) : null}

          {!isLoading && !error && grouped.length > 0
            ? grouped.map((section) => (
                <section key={section.group} className="border-b border-line last:border-b-0">
                  <h3 className="sticky top-0 z-[1] bg-white/95 px-4 pb-1.5 pt-3 text-[11px] font-bold uppercase tracking-wide text-slate-400 backdrop-blur-sm">
                    {section.group}
                  </h3>
                  <ul role="group" aria-label={section.group}>
                    {section.items.map((item) => {
                      const flatIndex = items.indexOf(item);
                      return (
                        <PaletteResultRow
                          key={item.key}
                          item={item}
                          flatIndex={flatIndex}
                          isActive={flatIndex === activeIndex}
                          trimmedQuery={trimmedQuery}
                          onHover={() => setActiveIndex(flatIndex)}
                          onSelect={() => item.onSelect()}
                          onTogglePin={() => {
                            if (!item.actionId) return;
                            togglePinnedAction(item.actionId);
                            setPinnedRevision((value) => value + 1);
                          }}
                        />
                      );
                    })}
                  </ul>
                </section>
              ))
            : null}
        </div>

        <div className="hidden flex-wrap items-center justify-between gap-3 border-t border-line bg-slate-50/80 px-4 py-2.5 text-[11px] font-medium text-slate-500 sm:flex">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
            <span>↑↓ gezin</span>
            <span>Enter seç</span>
            <span>Home/End</span>
            <span>ESC kapat</span>
          </div>
          <div className="flex items-center gap-3">
            {showViewAll ? (
              <button
                type="button"
                className="inline-flex items-center gap-1 font-semibold text-violet-brand transition hover:text-violet-hover"
                onClick={() => {
                  onClose();
                  navigate(`/search?q=${encodeURIComponent(trimmedQuery)}`);
                }}
              >
                Tüm sonuçları gör
                <ArrowRight className="h-3.5 w-3.5" />
              </button>
            ) : (
              <span className="hidden sm:inline">Ctrl+K / Cmd+K</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
