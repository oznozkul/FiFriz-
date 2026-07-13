export { statusBadgeTone, statusLabel } from "../common/StatusBadge";

export function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

export function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(bytes / 1024, 1).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function sourceLabel(sourceType?: string) {
  if (!sourceType || sourceType === "manual_upload") return "Manuel yükleme";
  return sourceType.replace(/_/g, " ");
}

export type WorkspacePanelSection =
  | "general"
  | "health"
  | "tags"
  | "suggestedTags"
  | "folder"
  | "shares"
  | "workflow"
  | "comments"
  | "versions"
  | "dates"
  | "activity"
  | "related"
  | "aiMetadata"
  | "classification"
  | "entities"
  | "summary"
  | "aiProgress"
  | "ai"
  | "actionSuggestions";
