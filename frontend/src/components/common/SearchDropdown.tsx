import type { DocumentSearchResult } from "../../services/documentService";
import { getSearchResultExtension, getSearchResultTitle, stripSnippetMarkup } from "../../services/documentService";
import { FileTypeIcon } from "./FileTypeIcon";
import { Spinner } from "./Spinner";

type SearchDropdownProps = {
  query: string;
  results: DocumentSearchResult[];
  isLoading: boolean;
  error?: string | null;
  onSelect: (result: DocumentSearchResult) => void;
};

export function SearchDropdown({ query, results, isLoading, error, onSelect }: SearchDropdownProps) {
  const trimmedQuery = query.trim();

  return (
    <div className="absolute left-0 right-0 top-12 z-50 overflow-hidden rounded-lg border border-line bg-white shadow-panel">
      {trimmedQuery.length > 0 && trimmedQuery.length < 2 ? (
        <div className="px-4 py-3 text-sm font-medium text-slate-500">En az 2 karakter yazın.</div>
      ) : isLoading ? (
        <div className="flex items-center gap-2 px-4 py-3 text-sm font-medium text-slate-500">
          <Spinner size="sm" label="Aranıyor..." />
        </div>
      ) : error ? (
        <div className="px-4 py-3 text-sm font-medium text-red-600">{error}</div>
      ) : results.length > 0 ? (
        <div className="max-h-[420px] divide-y divide-line overflow-y-auto">
          {results.map((result) => (
            <button key={result.documentId} className="flex w-full items-start gap-3 px-4 py-3 text-left transition hover:bg-violet-soft/40 focus:outline-none focus:ring-4 focus:ring-inset focus:ring-violet-ring/70" type="button" onMouseDown={(event) => event.preventDefault()} onClick={() => onSelect(result)}>
              <span className="mt-0.5 shrink-0">
                <FileTypeIcon extension={getSearchResultExtension(result)} />
              </span>
              <span className="min-w-0 flex-1">
                <span className="flex items-center gap-2">
                  <span className="truncate font-semibold text-ink">{getSearchResultTitle(result)}</span>
                  <span className="shrink-0 rounded bg-slate-100 px-2 py-0.5 text-[11px] font-bold uppercase text-slate-500">{getSearchResultExtension(result)}</span>
                </span>
                <span className="mt-1 line-clamp-2 block text-xs leading-5 text-slate-500">{stripSnippetMarkup(result.matchedSnippet)}</span>
              </span>
            </button>
          ))}
        </div>
      ) : trimmedQuery.length >= 2 ? (
        <div className="px-4 py-3 text-sm font-medium text-slate-500">Sonuç bulunamadı.</div>
      ) : null}
    </div>
  );
}
