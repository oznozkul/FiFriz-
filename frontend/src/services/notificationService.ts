import { apiRequest } from "./apiClient";

export type NotificationItem = {
  id: string;
  type: string;
  title: string;
  message: string;
  entityType: string | null;
  entityId: string | null;
  isRead: boolean;
  readAt: string | null;
  metadata: Record<string, unknown> | null;
  createdAt: string;
};

export type NotificationListResponse = {
  items: NotificationItem[];
  totalCount: number;
  page: number;
  pageSize: number;
};

export type NotificationListParams = {
  unreadOnly?: boolean;
  page?: number;
  pageSize?: number;
};

export const notificationTypeLabels: Record<string, string> = {
  document_shared_with_you: "Belge paylaşımı",
  public_share_link_created: "Paylaşım linki",
  document_comment_added: "Yeni yorum",
  document_comment_replied: "Yorum cevabı",
  document_version_uploaded: "Yeni versiyon",
  document_version_restored: "Versiyon geri yüklendi",
  document_favorited: "Favori",
  document_moved_to_trash: "Çöp kutusu",
  document_restored: "Belge geri yüklendi",
  folder_created: "Klasör",
  document_processing_failed: "Belge işleme başarısız",
  document_processed: "Belge işlendi",
  workflow_created: "Onay süreci",
  workflow_approval_requested: "Onay isteği",
  workflow_approved: "Onaylandı",
  workflow_rejected: "Reddedildi",
  workflow_cancelled: "Onay iptal",
  workflow_completed: "Onay tamamlandı",
  integration_sync_completed: "Entegrasyon senkronu",
  integration_sync_failed: "Entegrasyon hatası",
  onedrive_sync_completed: "OneDrive senkronu",
  onedrive_sync_failed: "OneDrive hatası",
};

export function listNotifications(params: NotificationListParams = {}) {
  const searchParams = new URLSearchParams();
  if (params.unreadOnly) searchParams.set("unreadOnly", "true");
  if (params.page) searchParams.set("page", String(params.page));
  if (params.pageSize) searchParams.set("pageSize", String(params.pageSize));

  const query = searchParams.toString();
  return apiRequest<NotificationListResponse>(`/api/v1/notifications${query ? `?${query}` : ""}`, undefined, "Bildirimler yüklenemedi.");
}

export function getUnreadNotificationCount() {
  return apiRequest<{ count: number }>("/api/v1/notifications/unread-count", undefined, "Okunmamış bildirim sayısı alınamadı.");
}

export function markNotificationRead(id: string) {
  return apiRequest<NotificationItem>(`/api/v1/notifications/${id}/read`, { method: "PATCH" }, "Bildirim okundu olarak işaretlenemedi.");
}

export function markAllNotificationsRead() {
  return apiRequest<void>("/api/v1/notifications/read-all", { method: "PATCH" }, "Bildirimler okundu olarak işaretlenemedi.");
}

export function getNotificationLink(notification: NotificationItem) {
  const documentId =
    (notification.entityType === "document" && notification.entityId) ||
    (typeof notification.metadata?.documentId === "string" ? notification.metadata.documentId : null);

  if (documentId) {
    return `/documents/${documentId}`;
  }

  if (notification.entityType === "workflow" && notification.entityId) {
    const workflowDocumentId = typeof notification.metadata?.documentId === "string" ? notification.metadata.documentId : null;
    if (workflowDocumentId) {
      return `/documents/${workflowDocumentId}`;
    }
    return "/approvals";
  }

  if (notification.entityType === "integration" && notification.entityId) {
    return "/integrations";
  }

  if (notification.entityType === "folder" && notification.entityId) {
    return `/folders`;
  }

  return null;
}

export function formatNotificationTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
