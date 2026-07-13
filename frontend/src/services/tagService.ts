import { apiPost, apiRequest } from "./apiClient";
import type { Tag, TagAlias, TagSearchResult, TagStats } from "../types/tag";

export function listTags() {
  return apiRequest<Tag[]>("/api/v1/tags");
}

export function getTag(id: string) {
  return apiRequest<Tag>(`/api/v1/tags/${id}`);
}

export function searchTags(query: string) {
  const params = new URLSearchParams({ q: query });
  return apiRequest<TagSearchResult[]>(`/api/v1/tags/search?${params.toString()}`);
}

export function getTagStats(id: string) {
  return apiRequest<TagStats>(`/api/v1/tags/${id}/stats`);
}

export function listTagAliases(tagId: string) {
  return apiRequest<TagAlias[]>(`/api/v1/tags/${tagId}/aliases`);
}

export function createTagAlias(tagId: string, alias: string) {
  return apiPost<TagAlias>(`/api/v1/tags/${tagId}/aliases`, { alias }, "Alias eklenemedi.");
}

export function deleteTagAlias(tagId: string, aliasId: string) {
  return apiRequest<void>(`/api/v1/tags/${tagId}/aliases/${aliasId}`, { method: "DELETE" });
}

export function createTag(input: { name: string; color?: string; icon?: string | null }) {
  return apiRequest<Tag>("/api/v1/tags", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function updateTag(id: string, input: { name: string; color?: string; status?: string; icon?: string | null }) {
  return apiRequest<Tag>(`/api/v1/tags/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function mergeTag(id: string, targetTagId: string) {
  return apiPost<Tag>(`/api/v1/tags/${id}/merge`, { targetTagId }, "Etiket birleştirilemedi.");
}

export function deleteTag(id: string) {
  return apiRequest<void>(`/api/v1/tags/${id}`, {
    method: "DELETE",
  });
}

export function setDocumentTags(documentId: string, tagIds: string[]) {
  return apiRequest(`/api/v1/documents/${documentId}/tags`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ tagIds }),
  });
}

export function listDocumentsByTag(tagId: string) {
  return apiRequest(`/api/v1/documents/by-tag/${tagId}`);
}
