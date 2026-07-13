import { useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { EmptyState, LoadingState, PageTitle, SearchInput } from "../components/common";
import { HybridSearchResultCard } from "../components/search/HybridSearchResultCard";
import { SearchModeToggle } from "../components/search/SearchModeToggle";
import { hybridSearchDocuments } from "../services/hybridSearchService";
import { searchDocuments, getSearchResultExtension, getSearchResultTitle, stripSnippetMarkup, type DocumentSearchResult } from "../services/documentService";
import { pushRecentSearch } from "../services/commandPaletteStorage";
import { toast } from "../services/toast";
import type { HybridSearchResult } from "../types/hybridSearch";
import type { SearchMode } from "../types/hybridSearch";
import { getSearchMode } from "../utils/searchMode";
import { ArrowRight } from "lucide-react";
import { Link } from "react-router-dom";
import { FileTypeIcon, PrimaryButton } from "../components/common";

export function SearchResultsPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const query = searchParams.get("q") ?? "";
  const [inputValue, setInputValue] = useState(query);
  const [searchMode, setSearchMode] = useState<SearchMode>(() => getSearchMode());
  const [quickResults, setQuickResults] = useState<DocumentSearchResult[]>([]);
  const [smartResults, setSmartResults] = useState<HybridSearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    setInputValue(query);
  }, [query]);

  useEffect(() => {
    const trimmedQuery = query.trim();
    if (trimmedQuery.length < 2) {
      setQuickResults([]);
      setSmartResults([]);
      setIsLoading(false);
      return;
    }

    let cancelled = false;
    setIsLoading(true);

    const searchPromise =
      searchMode === "smart" ? hybridSearchDocuments(trimmedQuery) : searchDocuments(trimmedQuery);

    searchPromise
      .then((items) => {
        if (cancelled) return;
        pushRecentSearch(trimmedQuery);
        if (searchMode === "smart") {
          setSmartResults(items as HybridSearchResult[]);
          setQuickResults([]);
        } else {
          setQuickResults(items as DocumentSearchResult[]);
          setSmartResults([]);
        }
      })
      .catch((searchError) => {
        if (!cancelled) {
          toast.error(searchError instanceof Error ? searchError.message : "Arama tamamlanamadı.");
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [query, searchMode]);

  const title = useMemo(() => {
    const trimmedQuery = query.trim();
    return trimmedQuery ? `"${trimmedQuery}" araması` : "Arama";
  }, [query]);

  const resultCount = searchMode === "smart" ? smartResults.length : quickResults.length;

  const handleSubmit = () => {
    const trimmedInput = inputValue.trim();
    if (trimmedInput.length < 2) return;
    navigate(`/search?q=${encodeURIComponent(trimmedInput)}`);
  };

  return (
    <div className="space-y-5">
      <PageTitle title="Arama Sonuçları" subtitle={title} />

      <section className="panel p-4">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
          <SearchInput
            value={inputValue}
            placeholder="Döküman içeriklerinde ara..."
            wrapperClassName="max-w-2xl flex-1"
            onChange={(event) => setInputValue(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                handleSubmit();
              }
            }}
          />
          <SearchModeToggle onChange={setSearchMode} />
        </div>
      </section>

      {query.trim().length > 0 && query.trim().length < 2 ? (
        <EmptyState title="En az 2 karakter yazın" description="Döküman içeriklerinde arama yapmak için sorgunuzu biraz daha uzatın." />
      ) : isLoading ? (
        <LoadingState label="Sonuçlar aranıyor..." />
      ) : resultCount > 0 ? (
        <section className="space-y-3">
          <div className="text-sm font-semibold text-slate-500">{resultCount} sonuç bulundu</div>
          {searchMode === "smart"
            ? smartResults.map((result) => <HybridSearchResultCard key={result.documentId} result={result} query={query} />)
            : quickResults.map((result) => (
                <article key={result.documentId} className="panel p-5 transition hover:border-violet-ring">
                  <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-3">
                        <FileTypeIcon extension={getSearchResultExtension(result)} size="lg" />
                        <div className="min-w-0">
                          <h2 className="truncate text-base font-bold text-ink">{getSearchResultTitle(result)}</h2>
                          <p className="text-xs font-bold uppercase text-slate-400">{getSearchResultExtension(result)}</p>
                        </div>
                      </div>
                      <p className="mt-4 text-sm leading-6 text-slate-600">{highlightSnippet(stripSnippetMarkup(result.matchedSnippet), query)}</p>
                    </div>
                    <Link to={`/documents/${result.documentId}?q=${encodeURIComponent(query.trim())}`}>
                      <PrimaryButton>
                        Belgeye Git
                        <ArrowRight className="h-4 w-4" />
                      </PrimaryButton>
                    </Link>
                  </div>
                </article>
              ))}
        </section>
      ) : query.trim().length >= 2 ? (
        <EmptyState title="Sonuç bulunamadı" description="Farklı bir kelimeyle yeniden arama yapabilirsiniz." />
      ) : (
        <EmptyState title="Arama yapın" description="Döküman içeriklerinde aramak için en az 2 karakter yazın." />
      )}
    </div>
  );
}

function highlightSnippet(snippet: string, query: string) {
  const term = query.trim();
  if (term.length < 2) return snippet;

  const index = snippet.toLocaleLowerCase("tr-TR").indexOf(term.toLocaleLowerCase("tr-TR"));
  if (index === -1) return snippet;

  return (
    <>
      {snippet.slice(0, index)}
      <mark className="rounded bg-violet-soft px-1 font-semibold text-violet-brand">{snippet.slice(index, index + term.length)}</mark>
      {snippet.slice(index + term.length)}
    </>
  );
}
