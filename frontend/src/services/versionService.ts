import type { ApiDocument } from "../types/document";
import { apiBaseUrl, apiRequest } from "./apiClient";

export type DocumentVersion = {
  id: string;
  documentId: string;
  versionNumber: number;
  originalFileName: string;
  fileExtension: string;
  mimeType: string;
  fileSize: number;
  uploadedByUserId: string;
  uploadedByFullName: string;
  changeNote: string | null;
  isCurrent: boolean;
  createdAt: string;
};

export function listDocumentVersions(documentId: string) {
  return apiRequest<DocumentVersion[]>(`/api/v1/documents/${documentId}/versions`, undefined, "Versiyonlar yüklenemedi.");
}

export function uploadDocumentVersion(documentId: string, file: File, changeNote?: string) {
  const formData = new FormData();
  formData.append("file", file);
  if (changeNote?.trim()) formData.append("changeNote", changeNote.trim());

  return apiRequest<DocumentVersion>(`/api/v1/documents/${documentId}/versions`, {
    method: "POST",
    body: formData,
  }, "Yeni versiyon yüklenemedi.");
}

export function restoreDocumentVersion(documentId: string, versionId: string) {
  return apiRequest<ApiDocument>(`/api/v1/documents/${documentId}/versions/${versionId}/restore`, {
    method: "POST",
  }, "Versiyon geri yüklenemedi.");
}

export function formatVersionFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(bytes / 1024, 1).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export { apiBaseUrl };
