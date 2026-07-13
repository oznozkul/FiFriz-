import { apiRequest } from "./apiClient";
import type { ApiDocument } from "../types/document";

export type SmartCollection = {
  key: string;
  title: string;
  description: string;
  icon: string;
  count: number;
};

export function listSmartCollections() {
  return apiRequest<SmartCollection[]>("/api/v1/smart-collections", undefined, "Akıllı koleksiyonlar yüklenemedi.");
}

export function listSmartCollectionDocuments(key: string) {
  return apiRequest<ApiDocument[]>(`/api/v1/smart-collections/${encodeURIComponent(key)}/documents`, undefined, "Koleksiyon belgeleri yüklenemedi.");
}
