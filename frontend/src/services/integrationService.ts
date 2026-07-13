import { apiRequest } from "./apiClient";

export type IntegrationProvider = string;

export type ProviderCapabilities = {
  supportsOAuth: boolean;
  supportsPassword: boolean;
  supportsWebhook: boolean;
  supportsRealtime: boolean;
  supportsAttachments: boolean;
  supportsFolders: boolean;
  supportsHistoryImport: boolean;
  supportsIncrementalSync: boolean;
  supportsScheduling: boolean;
  supportsSearch: boolean;
  supportsLabels: boolean;
  supportsMetadata: boolean;
};

export type IntegrationProviderDescriptor = {
  name: string;
  displayName: string;
  description: string;
  available: boolean;
  capabilities: ProviderCapabilities;
};

export type IntegrationStatus =
  | "disconnected"
  | "connected"
  | "sync_pending"
  | "syncing"
  | "sync_failed"
  | "sync_completed";

export type IntegrationAccount = {
  id: string;
  provider: IntegrationProvider;
  displayName: string;
  emailAddress: string | null;
  status: IntegrationStatus;
  settings: Record<string, unknown> | null;
  lastSyncAt: string | null;
  lastError: string | null;
  isActive: boolean;
  importedItemCount?: number;
  createdAt: string;
  updatedAt: string;
};

export type IntegrationSyncLog = {
  id: string;
  syncType: string;
  status: string;
  startedAt: string | null;
  finishedAt: string | null;
  totalItems: number;
  importedItems: number;
  skippedItems: number;
  failedItems: number;
  errorMessage: string | null;
  metadata: Record<string, unknown> | null;
  createdAt: string;
};

export type CreateIntegrationInput = {
  provider: IntegrationProvider;
  displayName: string;
  emailAddress?: string;
  settings?: Record<string, unknown>;
  credentials?: { password: string };
};

export type UpdateIntegrationInput = {
  displayName?: string;
  settings?: Record<string, unknown>;
  isActive?: boolean;
  credentials?: { password: string };
};

export function listIntegrationProviders() {
  return apiRequest<IntegrationProviderDescriptor[]>("/api/v1/integrations/providers", undefined, "Entegrasyon sağlayıcıları yüklenemedi.");
}

export function connectIntegrationProvider(provider: string) {
  return apiRequest<{ authUrl: string }>(`/api/v1/integrations/providers/${encodeURIComponent(provider)}/connect`, undefined, "Bağlantı başlatılamadı.");
}

export function checkIntegrationHealth(id: string) {
  return apiRequest<{ healthy: boolean; message: string; details?: Record<string, unknown> }>(`/api/v1/integrations/${id}/health`, undefined, "Sağlık kontrolü yapılamadı.");
}

export function listIntegrations() {
  return apiRequest<IntegrationAccount[]>("/api/v1/integrations", undefined, "Entegrasyonlar yüklenemedi.");
}

export function createIntegration(input: CreateIntegrationInput) {
  return apiRequest<IntegrationAccount>("/api/v1/integrations", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  }, "Entegrasyon oluşturulamadı.");
}

export function getIntegration(id: string) {
  return apiRequest<IntegrationAccount>(`/api/v1/integrations/${id}`, undefined, "Entegrasyon yüklenemedi.");
}

export function updateIntegration(id: string, input: UpdateIntegrationInput) {
  return apiRequest<IntegrationAccount>(`/api/v1/integrations/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  }, "Entegrasyon güncellenemedi.");
}

export function deleteIntegration(id: string) {
  return apiRequest<void>(`/api/v1/integrations/${id}`, { method: "DELETE" }, "Entegrasyon silinemedi.");
}

export function queueIntegrationSync(id: string) {
  return apiRequest<{ message: string }>(`/api/v1/integrations/${id}/sync`, { method: "POST" }, "Senkronizasyon başlatılamadı.");
}

export function listIntegrationSyncLogs(id: string) {
  return apiRequest<IntegrationSyncLog[]>(`/api/v1/integrations/${id}/sync-logs`, undefined, "Senkronizasyon kayıtları yüklenemedi.");
}

export function connectGmail() {
  return apiRequest<{ authUrl: string }>("/api/v1/integrations/gmail/connect", undefined, "Gmail bağlantısı başlatılamadı.");
}

export function connectOutlook() {
  return apiRequest<{ authUrl: string }>("/api/v1/integrations/outlook/connect", undefined, "Outlook bağlantısı başlatılamadı.");
}

export function connectGoogleDrive() {
  return apiRequest<{ authUrl: string }>("/api/v1/integrations/google-drive/connect", undefined, "Google Drive bağlantısı başlatılamadı.");
}

export const providerLabels: Record<string, string> = {
  gmail: "Gmail",
  outlook: "Outlook",
  imap: "IMAP Mail",
  google_drive: "Google Drive",
  onedrive: "OneDrive",
  one_drive: "OneDrive",
  dropbox: "Dropbox",
  manual_upload: "Manuel Yükleme",
};

export function resolveProviderDisplayName(
  provider: string,
  descriptors?: IntegrationProviderDescriptor[],
): string {
  const fromApi = descriptors?.find((item) => item.name === provider)?.displayName;
  if (fromApi) return fromApi;
  return providerLabels[provider] ?? provider;
}

export function integrationStatusLabel(status: IntegrationStatus) {
  const labels: Record<IntegrationStatus, string> = {
    disconnected: "Bağlı değil",
    connected: "Bağlı",
    sync_pending: "Kuyrukta",
    syncing: "Senkronize ediliyor",
    sync_failed: "Başarısız",
    sync_completed: "Tamamlandı",
  };
  return labels[status];
}

export function syncLogStatusLabel(status: string) {
  const labels: Record<string, string> = {
    pending: "Bekliyor",
    running: "Çalışıyor",
    completed: "Tamamlandı",
    failed: "Başarısız",
  };
  return labels[status] ?? status;
}
