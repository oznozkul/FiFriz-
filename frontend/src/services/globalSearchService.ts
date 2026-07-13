import { apiGet } from "./apiClient";
import type { GlobalSearchResponse } from "../types/globalSearch";

const emptyGlobalSearchResponse: GlobalSearchResponse = {
  documents: [],
  folders: [],
  tags: [],
  collections: [],
  sharedDocuments: [],
  integrations: [],
  users: [],
  actions: [],
  settings: [],
};

export function normalizeGlobalSearchResponse(data: Partial<GlobalSearchResponse> | null | undefined): GlobalSearchResponse {
  if (!data) return { ...emptyGlobalSearchResponse };

  return {
    documents: Array.isArray(data.documents) ? data.documents : [],
    folders: Array.isArray(data.folders) ? data.folders : [],
    tags: Array.isArray(data.tags) ? data.tags : [],
    collections: Array.isArray(data.collections) ? data.collections : [],
    sharedDocuments: Array.isArray(data.sharedDocuments) ? data.sharedDocuments : [],
    integrations: Array.isArray(data.integrations) ? data.integrations : [],
    users: Array.isArray(data.users) ? data.users : [],
    actions: Array.isArray(data.actions) ? data.actions : [],
    settings: Array.isArray(data.settings) ? data.settings : [],
  };
}

export function globalSearch(query: string) {
  const params = new URLSearchParams();
  if (query.trim()) {
    params.set("q", query.trim());
  }
  const suffix = params.toString() ? `?${params.toString()}` : "";
  return apiGet<GlobalSearchResponse>(`/api/v1/search/global${suffix}`, "Global arama tamamlanamadı.").then(normalizeGlobalSearchResponse);
}
