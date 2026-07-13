import type { HybridSearchResult } from "../types/hybridSearch";
import { apiRequest } from "./apiClient";

export function hybridSearchDocuments(query: string) {
  return apiRequest<HybridSearchResult[]>(
    `/api/v1/documents/hybrid-search?q=${encodeURIComponent(query)}`,
    undefined,
    "Akıllı arama tamamlanamadı.",
  );
}
