import { apiPost, apiRequest } from "./apiClient";
import type { ApiDocument } from "../types/document";
import type {
  Folder,
  FolderInput,
  FolderMetadata,
  FolderPermissions,
  FolderSearchResult,
} from "../types/folder";

export type UpdateFolderPermissionsInput = {
  inheritPermissions?: boolean;
  permissions?: Array<{ userId: string; permission: string }>;
};

export function listFolders() {
  return apiRequest<Folder[]>("/api/v1/folders");
}

export function getFolder(id: string) {
  return apiRequest<Folder>(`/api/v1/folders/${id}`);
}

export function searchFolders(query: string) {
  const params = new URLSearchParams({ q: query });
  return apiRequest<FolderSearchResult[]>(`/api/v1/folders/search?${params.toString()}`);
}

export function getFolderMetadata(id: string) {
  return apiRequest<FolderMetadata>(`/api/v1/folders/${id}/metadata`);
}

export function getFolderBreadcrumb(id: string) {
  return apiRequest<Folder[]>(`/api/v1/folders/${id}/breadcrumb`);
}

export function getFolderPermissions(id: string) {
  return apiRequest<FolderPermissions>(`/api/v1/folders/${id}/permissions`);
}

export function updateFolderPermissions(id: string, input: UpdateFolderPermissionsInput) {
  return apiRequest<FolderPermissions>(`/api/v1/folders/${id}/permissions`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function createFolder(input: FolderInput) {
  return apiRequest<Folder>("/api/v1/folders", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function updateFolder(id: string, input: FolderInput) {
  return apiRequest<Folder>(`/api/v1/folders/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function moveFolder(id: string, parentId: string | null) {
  return apiPost<Folder>(`/api/v1/folders/${id}/move`, { parentId: parentId ?? "" }, "Klasör taşınamadı.");
}

export function lockFolder(id: string) {
  return apiPost<Folder>(`/api/v1/folders/${id}/lock`, {}, "Klasör kilitlenemedi.");
}

export function unlockFolder(id: string) {
  return apiPost<Folder>(`/api/v1/folders/${id}/unlock`, {}, "Klasör kilidi açılamadı.");
}

export function restoreFolder(id: string) {
  return apiPost<Folder>(`/api/v1/folders/${id}/restore`, {}, "Klasör geri yüklenemedi.");
}

export function deleteFolder(id: string) {
  return apiRequest<void>(`/api/v1/folders/${id}`, {
    method: "DELETE",
  });
}

export function listFolderDocuments(id: string) {
  return apiRequest<ApiDocument[]>(`/api/v1/folders/${id}/documents`);
}
