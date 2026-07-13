import type { DocumentRelationshipEdge, RelatedDocumentNode, RelationshipType } from "./relationshipService";
import { apiPost } from "./apiClient";
import { apiGet } from "./apiClient";

export type DocumentRelatedSuggestion = {
  id: string;
  sourceDocumentId: string;
  targetDocumentId: string;
  relationshipType?: RelationshipType | string | null;
  reason?: string | null;
  similarity?: number | null;
  status: "suggested" | "accepted" | "rejected";
  generatedBy: string;
  targetDocument: RelatedDocumentNode;
  createdAt: string;
  updatedAt: string;
};

export function listRelatedSuggestions(documentId: string) {
  return apiGet<DocumentRelatedSuggestion[]>(`/api/v1/documents/${documentId}/related-suggestions`, "AI ilişki önerileri yüklenemedi.");
}

export function generateRelatedSuggestions(documentId: string) {
  return apiPost<DocumentRelatedSuggestion[]>(
    `/api/v1/documents/${documentId}/related-suggestions/generate`,
    undefined,
    "AI ilişki önerileri oluşturulamadı.",
  );
}

export function acceptRelatedSuggestion(documentId: string, suggestionId: string) {
  return apiPost<DocumentRelationshipEdge>(
    `/api/v1/documents/${documentId}/related-suggestions/${suggestionId}/accept`,
    undefined,
    "İlişki önerisi kabul edilemedi.",
  );
}

export function rejectRelatedSuggestion(documentId: string, suggestionId: string) {
  return apiPost<DocumentRelatedSuggestion>(
    `/api/v1/documents/${documentId}/related-suggestions/${suggestionId}/reject`,
    undefined,
    "İlişki önerisi reddedilemedi.",
  );
}

export function formatSimilarityPercent(similarity?: number | null) {
  if (similarity == null || Number.isNaN(similarity)) {
    return null;
  }
  return `${Math.round(similarity * 100)}%`;
}
