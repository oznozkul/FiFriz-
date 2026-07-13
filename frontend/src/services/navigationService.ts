import { apiDelete, apiGet, apiPost } from "./apiClient";

export type PinItem = {
  id: string;
  entityType: string;
  entityId: string;
  label: string;
  path: string;
  createdAt: string;
};

export type RecentItem = {
  entityType: string;
  entityId: string;
  label: string;
  path: string;
  openedAt: string;
};

export type RecentResponse = {
  documents: RecentItem[];
  folders: RecentItem[];
  smartCollections: RecentItem[];
  integrations: RecentItem[];
  settingsPages: RecentItem[];
};

export function listPins() {
  return apiGet<PinItem[]>("/api/v1/pins", "Sabitlenenler yüklenemedi.");
}

export function createPin(input: { entityType: string; entityId: string; label: string; path: string }) {
  return apiPost<PinItem>("/api/v1/pins", input, "Sabitleme başarısız.");
}

export function deletePin(pinId: string) {
  return apiDelete(`/api/v1/pins/${encodeURIComponent(pinId)}`, "Sabitleme kaldırılamadı.");
}

export function listRecent() {
  return apiGet<RecentResponse>("/api/v1/recent", "Son açılanlar yüklenemedi.");
}

export function trackVisit(input: { entityType: string; entityId: string; label: string; path: string }) {
  return apiPost<void>("/api/v1/navigation/visited", input);
}
