import { notifyOnboardingStep } from "../components/dashboard/onboarding/onboardingStorage";

const recentSearchesKey = "fifriz_command_palette_recent_searches";
const recentDocumentsKey = "fifriz_command_palette_recent_documents";
const pinnedActionsKey = "fifriz_command_palette_pinned";

const defaultPinned = ["upload-document", "go-dashboard", "go-integrations"];

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

function writeJSON<T>(key: string, value: T) {
  localStorage.setItem(key, JSON.stringify(value));
}

export type RecentDocument = {
  id: string;
  title: string;
  extension?: string;
  openedAt: string;
};

export function getRecentSearches() {
  return readJSON<string[]>(recentSearchesKey, []);
}

export function pushRecentSearch(query: string) {
  const trimmed = query.trim();
  if (trimmed.length < 2) return;
  const items = getRecentSearches().filter((item) => item !== trimmed);
  items.unshift(trimmed);
  writeJSON(recentSearchesKey, items.slice(0, 8));
  notifyOnboardingStep("search");
}

export function getRecentDocuments() {
  return readJSON<RecentDocument[]>(recentDocumentsKey, []);
}

export function pushRecentDocument(document: Omit<RecentDocument, "openedAt">) {
  const items = getRecentDocuments().filter((item) => item.id !== document.id);
  items.unshift({ ...document, openedAt: new Date().toISOString() });
  writeJSON(recentDocumentsKey, items.slice(0, 8));
}

export function getPinnedActionIds() {
  const stored = readJSON<string[]>(pinnedActionsKey, []);
  if (stored.length === 0) return defaultPinned;
  return stored;
}

export function togglePinnedAction(actionId: string) {
  const current = new Set(getPinnedActionIds());
  if (current.has(actionId)) {
    current.delete(actionId);
  } else {
    current.add(actionId);
  }
  writeJSON(pinnedActionsKey, Array.from(current));
}

export function isPinnedAction(actionId: string) {
  return getPinnedActionIds().includes(actionId);
}
