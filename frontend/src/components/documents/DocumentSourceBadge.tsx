import type { ApiDocument } from "../../types/document";

type DocumentSourceBadgeProps = {
  document: Pick<ApiDocument, "sourceType" | "sourceMetadata">;
  compact?: boolean;
};

export function DocumentSourceBadge({ document, compact = false }: DocumentSourceBadgeProps) {
  if (!document.sourceType || document.sourceType === "manual_upload") {
    return null;
  }

  const label =
    document.sourceType === "imap"
      ? "IMAP"
      : document.sourceType === "gmail"
        ? "Gmail"
      : document.sourceType === "outlook"
        ? "Outlook"
        : document.sourceType === "google_drive"
          ? "Google Drive"
          : document.sourceType === "onedrive"
            ? "OneDrive"
          : document.sourceType.toUpperCase();
  const metadata = document.sourceMetadata ?? {};
  const from = typeof metadata.from === "string" ? metadata.from : null;
  const subject = typeof metadata.subject === "string" ? metadata.subject : null;
  const date = typeof metadata.date === "string" ? metadata.date : null;

  if (compact) {
    return <span className="rounded-full border border-sky-200 bg-sky-50 px-2 py-0.5 text-[11px] font-semibold text-sky-700">{label}</span>;
  }

  return (
    <div className="rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-xs text-sky-900">
      <div className="font-semibold">{label} kaynagi</div>
      {from ? <p className="mt-1">From: {from}</p> : null}
      {subject ? <p className="mt-1">Subject: {subject}</p> : null}
      {date ? <p className="mt-1">Date: {date}</p> : null}
    </div>
  );
}
