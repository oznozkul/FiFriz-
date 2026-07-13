import type { AcceptSuggestedTagsResult, DocumentSuggestedTag } from "../types/documentSuggestedTag";
import type { TagBrief } from "../types/tag";
import { apiGet, apiPost } from "./apiClient";

export function getDocumentSuggestedTags(documentId: string) {
  return apiGet<DocumentSuggestedTag[]>(`/api/v1/documents/${documentId}/suggested-tags`, "AI etiket önerileri yüklenemedi.");
}

export function acceptSuggestedTag(documentId: string, suggestedTagId: string) {
  return apiPost<AcceptSuggestedTagsResult>(
    `/api/v1/documents/${documentId}/suggested-tags/${suggestedTagId}/accept`,
    undefined,
    "Etiket önerisi kabul edilemedi.",
  );
}

export function rejectSuggestedTag(documentId: string, suggestedTagId: string) {
  return apiPost<DocumentSuggestedTag>(
    `/api/v1/documents/${documentId}/suggested-tags/${suggestedTagId}/reject`,
    undefined,
    "Etiket önerisi reddedilemedi.",
  );
}

export function acceptAllSuggestedTags(documentId: string) {
  return apiPost<AcceptSuggestedTagsResult>(
    `/api/v1/documents/${documentId}/suggested-tags/accept-all`,
    undefined,
    "Etiket önerileri kabul edilemedi.",
  );
}
