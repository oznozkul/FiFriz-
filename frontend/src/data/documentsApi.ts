import type { ApiDocument, ApiResponse, DocumentStatus } from "../types/document";
import { apiRequest } from "../services/apiClient";

export function listDocuments(folderId?: string, tagId?: string) {
  const searchParams = new URLSearchParams();
  if (folderId) searchParams.set("folderId", folderId);
  if (tagId) searchParams.set("tagId", tagId);
  const query = searchParams.toString();
  return apiRequest<ApiDocument[]>(`/api/v1/documents${query ? `?${query}` : ""}`);
}

export function listTrashDocuments() {
  return apiRequest<ApiDocument[]>("/api/v1/documents/trash");
}

export function listFavoriteDocuments() {
  return apiRequest<ApiDocument[]>("/api/v1/documents/favorites");
}

export function favoriteDocument(id: string) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${id}/favorite`, {
    method: "PATCH",
  });
}

export function unfavoriteDocument(id: string) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${id}/unfavorite`, {
    method: "PATCH",
  });
}

export function getDocument(id: string) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${id}`);
}

export function getDocumentWorkspace(id: string) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${id}/workspace`);
}

export function getDocumentText(id: string) {
  return apiRequest<{ document_id: string; full_text: string; created_at: string }>(`/api/v1/documents/${id}/text`);
}

export function reprocessDocumentText(id: string) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${id}/reprocess-text`, {
    method: "POST",
  });
}

export function updateDocument(id: string, input: { original_file_name: string; status?: DocumentStatus; folder_id?: string }) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${id}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });
}

export function deleteDocument(id: string) {
  return apiRequest<void>(`/api/v1/documents/${id}`, {
    method: "DELETE",
  });
}

export function restoreDocument(id: string) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${id}/restore`, {
    method: "POST",
  });
}

export function permanentlyDeleteDocument(id: string) {
  return apiRequest<void>(`/api/v1/documents/${id}/permanent`, {
    method: "DELETE",
  });
}

export function setDocumentTags(documentId: string, tagIds: string[]) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${documentId}/tags`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ tagIds }),
  });
}
