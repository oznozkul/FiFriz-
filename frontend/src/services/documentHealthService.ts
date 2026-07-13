import type { ApiDocument, DocumentHealth } from "../types/document";
import { apiGet, apiPost } from "./apiClient";

export function getDocumentHealth(documentId: string) {
  return apiGet<DocumentHealth>(`/api/v1/documents/${documentId}/health`, "Belge sağlık durumu yüklenemedi.");
}

export function lockDocument(documentId: string) {
  return apiPost<ApiDocument>(`/api/v1/documents/${documentId}/lock`, {}, "Belge kilitlenemedi.");
}

export function unlockDocument(documentId: string) {
  return apiPost<ApiDocument>(`/api/v1/documents/${documentId}/unlock`, {}, "Belge kilidi açılamadı.");
}
