import { apiRequest } from "./apiClient";
import type { ImportantDate, ImportantDateInput, ImportantDateListParams } from "../types/importantDate";

function buildQuery(params: ImportantDateListParams = {}) {
  const searchParams = new URLSearchParams();
  if (params.upcomingOnly) searchParams.set("upcomingOnly", "true");
  if (params.completed !== undefined) searchParams.set("completed", String(params.completed));
  if (params.assignedToMe) searchParams.set("assignedToMe", "true");
  if (params.from) searchParams.set("from", params.from);
  if (params.to) searchParams.set("to", params.to);
  if (params.type) searchParams.set("type", params.type);
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export function listImportantDates(params: ImportantDateListParams = {}) {
  return apiRequest<ImportantDate[]>(`/api/v1/important-dates${buildQuery(params)}`);
}

export function listDocumentImportantDates(documentId: string) {
  return apiRequest<ImportantDate[]>(`/api/v1/documents/${documentId}/important-dates`);
}

export function createDocumentImportantDate(documentId: string, input: ImportantDateInput) {
  return apiRequest<ImportantDate>(`/api/v1/documents/${documentId}/important-dates`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function updateDocumentImportantDate(documentId: string, dateId: string, input: ImportantDateInput) {
  return apiRequest<ImportantDate>(`/api/v1/documents/${documentId}/important-dates/${dateId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function deleteDocumentImportantDate(documentId: string, dateId: string) {
  return apiRequest<void>(`/api/v1/documents/${documentId}/important-dates/${dateId}`, {
    method: "DELETE",
  });
}

export function completeDocumentImportantDate(documentId: string, dateId: string) {
  return apiRequest<ImportantDate>(`/api/v1/documents/${documentId}/important-dates/${dateId}/complete`, {
    method: "PATCH",
  });
}

export function listImportantDateSuggestions(documentId: string) {
  return apiRequest<ImportantDate[]>(`/api/v1/documents/${documentId}/important-dates/suggestions`);
}

export function extractAIImportantDates(documentId: string) {
  return apiRequest<ImportantDate[]>(`/api/v1/documents/${documentId}/important-dates/extract-ai`, {
    method: "POST",
  });
}

export function acceptImportantDateSuggestion(documentId: string, suggestionId: string) {
  return apiRequest<ImportantDate>(`/api/v1/documents/${documentId}/important-dates/suggestions/${suggestionId}/accept`, {
    method: "POST",
  });
}

export function rejectImportantDateSuggestion(documentId: string, suggestionId: string) {
  return apiRequest<ImportantDate>(`/api/v1/documents/${documentId}/important-dates/suggestions/${suggestionId}/reject`, {
    method: "POST",
  });
}
