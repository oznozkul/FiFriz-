import { ArrowUpRight } from "lucide-react";
import { Link } from "react-router-dom";
import { FileTypeIcon } from "../common";
import type { AiChatSource } from "../../types/aiChat";

type AiSourceCardProps = {
  source: AiChatSource;
  question?: string;
};

function documentLink(source: AiChatSource, question?: string) {
  const params = new URLSearchParams();
  if (question?.trim()) {
    params.set("q", question.trim());
  }
  const query = params.toString();
  return `/documents/${source.documentId}${query ? `?${query}` : ""}`;
}

function fileExtensionFromTitle(title: string) {
  const parts = title.split(".");
  if (parts.length < 2) return "file";
  return parts[parts.length - 1] || "file";
}

export function AiSourceCard({ source, question }: AiSourceCardProps) {
  const similarityPercent = Math.round(Math.max(0, Math.min(1, source.similarity)) * 100);
  const link = documentLink(source, question);

  return (
    <Link
      to={link}
      className="block rounded-lg border border-line bg-white p-3 shadow-sm transition hover:border-violet-ring hover:bg-violet-soft/30 focus:outline-none focus:ring-4 focus:ring-violet-ring/70"
    >
      <div className="flex items-start gap-3">
        <FileTypeIcon extension={fileExtensionFromTitle(source.documentTitle)} size="md" />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h4 className="truncate text-sm font-bold text-ink">{source.documentTitle}</h4>
            <span className="rounded-full bg-violet-soft px-2 py-0.5 text-xs font-semibold text-violet-brand">%{similarityPercent} eşleşme</span>
          </div>
          <p className="mt-1 line-clamp-2 text-xs leading-5 text-slate-500">{source.contentPreview}</p>
          <span className="mt-3 inline-flex items-center gap-1.5 text-xs font-semibold text-violet-brand">
            Belgeye git
            <ArrowUpRight className="h-3.5 w-3.5" />
          </span>
        </div>
      </div>
    </Link>
  );
}
