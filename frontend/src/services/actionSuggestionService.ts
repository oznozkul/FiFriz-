import type { ActionSuggestion, ActionSuggestionAcceptResponse } from "../types/actionSuggestion";
import { apiRequest } from "./apiClient";

export function listActionSuggestions() {
  return apiRequest<ActionSuggestion[]>("/api/v1/ai/action-suggestions", undefined, "Aksiyon önerileri yüklenemedi.");
}

export function listDocumentActionSuggestions(documentId: string) {
  return apiRequest<ActionSuggestion[]>(
    `/api/v1/documents/${documentId}/ai/action-suggestions`,
    undefined,
    "Belge aksiyon önerileri yüklenemedi.",
  );
}

export function generateDocumentActionSuggestions(documentId: string) {
  return apiRequest<ActionSuggestion[]>(
    `/api/v1/documents/${documentId}/ai/action-suggestions/generate`,
    { method: "POST" },
    "Aksiyon önerileri oluşturulamadı.",
  );
}

export function acceptActionSuggestion(suggestionId: string) {
  return apiRequest<ActionSuggestionAcceptResponse>(
    `/api/v1/ai/action-suggestions/${suggestionId}/accept`,
    { method: "POST" },
    "Öneri kabul edilemedi.",
  );
}

export function rejectActionSuggestion(suggestionId: string) {
  return apiRequest<ActionSuggestion>(
    `/api/v1/ai/action-suggestions/${suggestionId}/reject`,
    { method: "POST" },
    "Öneri reddedilemedi.",
  );
}
