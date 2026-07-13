import { apiBaseUrl, apiDelete, apiGet, apiPost } from "./apiClient";

export type SharePermission = "view" | "download";
export type ShareChannel = "whatsapp" | "telegram" | "email" | "copy" | "native";
export type ShareType = "public_link" | "email_invite" | "internal_user" | "internal_workspace";
export type ShareStatus = "active" | "expired" | "revoked" | "disabled";
export type ExpirationPreset = "never" | "1d" | "7d" | "30d" | "custom";

export type DocumentShareLink = {
  id: string;
  documentId: string;
  shareType: ShareType;
  status: ShareStatus;
  shareUrl: string;
  token: string;
  permission: SharePermission;
  expirationPreset?: ExpirationPreset | string;
  expiresAt: string | null;
  maxViewCount: number | null;
  viewCount: number;
  downloadCount: number;
  allowDownload: boolean;
  allowPrint: boolean;
  allowCopyText: boolean;
  allowedDomains?: string[];
  watermarkEnabled: boolean;
  isActive: boolean;
  hasPassword: boolean;
  invitedEmail?: string | null;
  createdAt: string;
};

export type ShareLinkAnalytics = {
  shareLinkId: string;
  documentId: string;
  openCount: number;
  downloadCount: number;
  lastOpenedAt: string | null;
  lastDownloadedAt: string | null;
  lastIpAddress: string | null;
  lastUserAgent: string | null;
  createdByUserId: string;
  createdByName: string;
  status: ShareStatus;
};

export type CreateShareLinkInput = {
  shareType?: ShareType;
  permission: SharePermission;
  password?: string;
  expirationPreset?: ExpirationPreset | string;
  expiresAt?: string | null;
  maxViewCount?: number | null;
  allowDownload?: boolean;
  allowPrint?: boolean;
  allowCopyText?: boolean;
  allowedDomains?: string[];
  watermarkEnabled?: boolean;
  watermarkTemplate?: string;
  invitedEmail?: string;
};

export type BulkCreateShareLinksInput = CreateShareLinkInput & {
  documentIds: string[];
};

export type PublicShareInfo = {
  documentName: string;
  fileExtension: string;
  fileSize: number;
  previewUrl?: string;
  allowDownload: boolean;
  allowPrint: boolean;
  allowCopyText: boolean;
  watermarkEnabled: boolean;
  status: ShareStatus;
  passwordRequired: boolean;
  domainRestricted: boolean;
  canAccess: boolean;
  expired: boolean;
  maxViewsReached: boolean;
};

type ApiEnvelope<T> = {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
};

async function parsePublicResponse<T>(response: Response, fallbackMessage: string): Promise<T> {
  const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | null;
  if (!response.ok || payload?.success === false) {
    throw new Error(payload?.error?.message ?? fallbackMessage);
  }
  return payload?.data as T;
}

export function createShareLink(documentId: string, input: CreateShareLinkInput) {
  return apiPost<DocumentShareLink>(`/api/v1/documents/${documentId}/share-links`, input, "Paylaşım linki oluşturulamadı.");
}

export function bulkCreateShareLinks(input: BulkCreateShareLinksInput) {
  return apiPost<DocumentShareLink[]>("/api/v1/documents/bulk-share-links", input, "Toplu paylaşım oluşturulamadı.");
}

export function listShareLinks(documentId: string) {
  return apiGet<DocumentShareLink[]>(`/api/v1/documents/${documentId}/share-links`, "Paylaşım linkleri yüklenemedi.");
}

export function deleteShareLink(documentId: string, shareLinkId: string) {
  return apiDelete(`/api/v1/documents/${documentId}/share-links/${shareLinkId}`, "Paylaşım linki silinemedi.");
}

export function revokeShareLink(documentId: string, shareLinkId: string) {
  return apiPost(`/api/v1/documents/${documentId}/share-links/${shareLinkId}/revoke`, {}, "Paylaşım linki iptal edilemedi.");
}

export function getShareLinkAnalytics(shareLinkId: string) {
  return apiGet<ShareLinkAnalytics>(`/api/v1/documents/share-links/${shareLinkId}/analytics`, "Paylaşım analitiği yüklenemedi.");
}

export function getPublicShare(token: string, accessToken?: string, visitorEmail?: string) {
  const headers: Record<string, string> = {};
  if (accessToken) headers["X-Share-Access-Token"] = accessToken;
  if (visitorEmail) headers["X-Share-Visitor-Email"] = visitorEmail;

  return fetch(`${apiBaseUrl}/api/v1/public/share/${encodeURIComponent(token)}`, { headers }).then((response) =>
    parsePublicResponse<PublicShareInfo>(response, "Paylaşım linki açılamadı."),
  );
}

export function verifySharePassword(token: string, password: string, visitorEmail?: string) {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (visitorEmail) headers["X-Share-Visitor-Email"] = visitorEmail;

  return fetch(`${apiBaseUrl}/api/v1/public/share/${encodeURIComponent(token)}/verify-password`, {
    method: "POST",
    headers,
    body: JSON.stringify({ password }),
  }).then((response) =>
    parsePublicResponse<{ accessToken: string; expiresIn: number }>(response, "Şifre doğrulanamadı."),
  );
}

export function getPublicShareDownloadUrl(token: string, accessToken?: string, download = false, visitorEmail?: string) {
  const params = new URLSearchParams();
  if (accessToken) params.set("accessToken", accessToken);
  if (download) params.set("download", "1");
  if (visitorEmail) params.set("visitorEmail", visitorEmail);
  const query = params.toString();
  return `${apiBaseUrl}/api/v1/public/share/${encodeURIComponent(token)}/download${query ? `?${query}` : ""}`;
}

export function formatShareFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.max(bytes / 1024, 1).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export const expirationPresetLabels: Record<ExpirationPreset, string> = {
  never: "Süresiz",
  "1d": "1 Gün",
  "7d": "7 Gün",
  "30d": "30 Gün",
  custom: "Özel Tarih",
};

export const shareStatusLabels: Record<ShareStatus, string> = {
  active: "Aktif",
  expired: "Süresi Doldu",
  revoked: "İptal Edildi",
  disabled: "Devre Dışı",
};

export async function logShareChannel(shareLinkId: string, channel: ShareChannel) {
  try {
    await apiPost(`/api/v1/documents/share-links/${shareLinkId}/channel-log`, { channel });
  } catch (error) {
    console.warn("Share channel log failed", error);
  }
}
