import { ArrowRight } from "lucide-react";
import { Link } from "react-router-dom";
import { AI_DOCUMENT_TYPE_LABELS } from "../../types/documentClassification";
import type { HybridSearchResult } from "../../types/hybridSearch";
import { FileTypeIcon, PrimaryButton } from "../common";
import { TagBadgeList } from "../tags";
import { SearchMatchTypeBadge } from "./SearchMatchTypeBadge";
import { stripSnippetMarkup } from "../../services/documentService";

type HybridSearchResultCardProps = {
  result: HybridSearchResult;
  query: string;
};

export function HybridSearchResultCard({ result, query }: HybridSearchResultCardProps) {
  const typeLabel = result.aiDocumentType ? AI_DOCUMENT_TYPE_LABELS[result.aiDocumentType] ?? result.aiDocumentType : null;

  return (
    <article className="panel p-5 transition hover:border-violet-ring">
      <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <FileTypeIcon extension={result.fileExtension} size="lg" />
            <div className="min-w-0 flex-1">
              <h2 className="truncate text-base font-bold text-ink">{result.originalFileName}</h2>
              <div className="mt-1 flex flex-wrap items-center gap-2">
                <SearchMatchTypeBadge matchType={result.matchType} />
                <span className="text-xs font-semibold text-slate-400">{Math.round(result.score * 100)}% skor</span>
              </div>
            </div>
          </div>

          <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-slate-500">
            {result.folderName ? <span>Klasör: {result.folderName}</span> : null}
            {typeLabel ? <span>· {typeLabel}</span> : null}
          </div>

          {result.tags.length > 0 ? (
            <div className="mt-2">
              <TagBadgeList tags={result.tags} />
            </div>
          ) : null}

          {result.matchedSnippet ? (
            <p className="mt-4 text-sm leading-6 text-slate-600">{stripSnippetMarkup(result.matchedSnippet)}</p>
          ) : null}

          {result.matchedChunks.length > 0 ? (
            <div className="mt-3 space-y-1">
              {result.matchedChunks.map((chunk, index) => (
                <p key={`${result.documentId}-chunk-${index}`} className="text-xs leading-5 text-slate-500">
                  {stripSnippetMarkup(chunk)}
                </p>
              ))}
            </div>
          ) : null}

          {result.matchedEntities.length > 0 ? (
            <div className="mt-3 flex flex-wrap gap-2">
              {result.matchedEntities.map((entity) => (
                <span key={`${entity.entityType}-${entity.value}`} className="rounded-full bg-slate-100 px-2 py-1 text-xs font-semibold text-slate-600">
                  {entity.label ? `${entity.label}: ` : ""}
                  {entity.value}
                </span>
              ))}
            </div>
          ) : null}
        </div>

        <Link to={`/documents/${result.documentId}?q=${encodeURIComponent(query.trim())}`}>
          <PrimaryButton>
            Belgeye Git
            <ArrowRight className="h-4 w-4" />
          </PrimaryButton>
        </Link>
      </div>
    </article>
  );
}
