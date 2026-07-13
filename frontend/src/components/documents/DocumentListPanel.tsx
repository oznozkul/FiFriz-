import { Link } from "react-router-dom";
import { mockDocuments } from "../../data/mockDocuments";
import { FileTypeIcon } from "../common";
import { formatDate } from "./documentPresentation";

type DocumentListPanelProps = {
  title: string;
  compact?: boolean;
};

export function DocumentListPanel({ title, compact }: DocumentListPanelProps) {
  const visibleDocuments = mockDocuments.slice(0, compact ? 5 : mockDocuments.length);

  return (
    <section className="panel overflow-hidden">
      <div className="flex items-center justify-between px-5 py-4">
        <h2 className="font-bold text-ink">{title}</h2>
      </div>
      <div className="divide-y divide-line">
        {visibleDocuments.map((doc) => (
          <Link key={doc.name} to={`/documents/${doc.id}`} className="flex w-full items-center gap-3 px-5 py-3 text-left transition hover:bg-violet-soft/35 focus:outline-none focus:ring-4 focus:ring-inset focus:ring-violet-ring/70">
            <FileTypeIcon extension={doc.fileType} />
            <span className="min-w-0 flex-1 truncate font-semibold text-ink">{doc.name}</span>
            <span className="hidden text-xs text-slate-500 sm:block">{formatDate(doc.uploadDate)}</span>
            <span className="w-16 text-right text-xs text-slate-500">{doc.size}</span>
          </Link>
        ))}
      </div>
      <Link className="inline-flex px-5 py-4 text-sm font-semibold text-violet-brand transition hover:text-violet-hover" to="/documents">
        Tümünü Gör →
      </Link>
    </section>
  );
}
