import { apiRequest } from "./apiClient";

export type DocumentSearchResult = {
  documentId: string;
  title?: string;
  originalFileName?: string;
  fileExtension?: string;
  matchedSnippet: string;
  createdAt?: string;
};

export function searchDocuments(query: string) {
  return apiRequest<DocumentSearchResult[]>(`/api/v1/documents/search?q=${encodeURIComponent(query)}`, undefined, "Arama su anda tamamlanamadı.");
}

export function getSearchResultTitle(result: DocumentSearchResult) {
  return result.originalFileName ?? result.title ?? "Belge";
}

export function getSearchResultExtension(result: DocumentSearchResult) {
  const title = getSearchResultTitle(result);
  return result.fileExtension ?? title.split(".").pop() ?? "file";
}

export function stripSnippetMarkup(value: string) {
  return value.replace(/<\/?b>/gi, "");
}
