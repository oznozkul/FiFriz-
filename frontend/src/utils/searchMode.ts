import type { SearchMode } from "../types/hybridSearch";

const SEARCH_MODE_STORAGE_KEY = "fifriz.searchMode";

export function getSearchMode(): SearchMode {
  if (typeof window === "undefined") return "smart";
  const stored = window.localStorage.getItem(SEARCH_MODE_STORAGE_KEY);
  return stored === "quick" ? "quick" : "smart";
}

export function setSearchMode(mode: SearchMode) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(SEARCH_MODE_STORAGE_KEY, mode);
}

export const SEARCH_MODE_LABELS: Record<SearchMode, string> = {
  quick: "Hızlı Arama",
  smart: "Akıllı Arama",
};
