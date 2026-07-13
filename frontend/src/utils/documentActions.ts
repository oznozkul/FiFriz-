import type { ApiDocument } from "../types/document";
import { apiBaseUrl, getAccessToken } from "../services/apiClient";

export async function downloadDocumentFile(document: ApiDocument) {
  const token = getAccessToken();
  const response = await fetch(`${apiBaseUrl}${document.fileUrl}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });

  if (!response.ok) {
    throw new Error("Dosya indirilemedi.");
  }

  const blob = await response.blob();
  const objectUrl = URL.createObjectURL(blob);
  const anchor = window.document.createElement("a");
  anchor.href = objectUrl;
  anchor.download = document.original_file_name;
  anchor.click();
  URL.revokeObjectURL(objectUrl);
}

export function workspacePanelPath(documentId: string, panel: string, searchQuery?: string) {
  const params = new URLSearchParams();
  if (searchQuery?.trim()) {
    params.set("q", searchQuery.trim());
  }
  params.set("panel", panel);
  const query = params.toString();
  return `/documents/${documentId}${query ? `?${query}` : ""}`;
}
