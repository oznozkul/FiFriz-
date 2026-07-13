import { apiRequest } from "./apiClient";

export type AuditLog = {
  id: string;
  tenantId: string;
  userId: string | null;
  userEmail: string;
  userFullName: string;
  action: string;
  entityType: string;
  entityId: string | null;
  description: string;
  metadata: Record<string, unknown> | null;
  ipAddress: string;
  userAgent: string;
  createdAt: string;
};

export type AuditLogListResponse = {
  items: AuditLog[];
  totalCount: number;
  page: number;
  pageSize: number;
};

export type AuditLogListParams = {
  action?: string;
  entityType?: string;
  userId?: string;
  startDate?: string;
  endDate?: string;
  page?: number;
  pageSize?: number;
};

export function listAuditLogs(params: AuditLogListParams = {}) {
  const searchParams = new URLSearchParams();
  if (params.action) searchParams.set("action", params.action);
  if (params.entityType) searchParams.set("entityType", params.entityType);
  if (params.userId) searchParams.set("userId", params.userId);
  if (params.startDate) searchParams.set("startDate", params.startDate);
  if (params.endDate) searchParams.set("endDate", params.endDate);
  if (params.page) searchParams.set("page", String(params.page));
  if (params.pageSize) searchParams.set("pageSize", String(params.pageSize));

  const query = searchParams.toString();
  return apiRequest<AuditLogListResponse>(`/api/v1/audit-logs${query ? `?${query}` : ""}`);
}

export function getAuditLog(id: string) {
  return apiRequest<AuditLog>(`/api/v1/audit-logs/${id}`);
}

export const auditActionLabels: Record<string, string> = {
  document_uploaded: "Belge yüklendi",
  document_storage_uploaded: "Belge depolamaya yüklendi",
  document_storage_deleted: "Belge depolamadan silindi",
  document_storage_migrated: "Belge depolama taşındı",
  storage_migration_failed: "Depolama taşıma başarısız",
  document_viewed: "Belge görüntülendi",
  document_previewed: "Belge önizlendi",
  document_downloaded: "Belge indirildi",
  document_trashed: "Çöp kutusuna taşındı",
  document_restored: "Belge geri yüklendi",
  document_deleted: "Belge silindi",
  document_favorited: "Favoriye eklendi",
  document_unfavorited: "Favoriden çıkarıldı",
  document_reprocessed: "Metin yeniden işlendi",
  document_searched: "Belge araması",
  document_summary_generated: "Belge özeti oluşturuldu",
  document_summary_viewed: "Belge özeti görüntülendi",
  folder_created: "Klasör oluşturuldu",
  folder_updated: "Klasör güncellendi",
  folder_deleted: "Klasör silindi",
  folder_renamed: "Klasör yeniden adlandırıldı",
  folder_moved: "Klasör taşındı",
  folder_locked: "Klasör kilitlendi",
  folder_unlocked: "Klasör kilidi açıldı",
  folder_permission_changed: "Klasör izinleri değiştirildi",
  tag_created: "Etiket oluşturuldu",
  tag_updated: "Etiket güncellendi",
  tag_deleted: "Etiket silindi",
  tag_archived: "Etiket arşivlendi",
  tag_merged: "Etiket birleştirildi",
  tag_alias_created: "Etiket alias eklendi",
  tag_alias_deleted: "Etiket alias silindi",
  document_tags_updated: "Belge etiketleri güncellendi",
  share_link_created: "Paylaşım linki oluşturuldu",
  share_link_deleted: "Paylaşım linki silindi",
  share_link_opened: "Paylaşım linki açıldı",
  share_link_downloaded: "Paylaşım linkinden indirildi",
  share_link_password_verified: "Paylaşım şifresi doğrulandı",
  share_link_password_failed: "Paylaşım şifresi başarısız",
  share_link_shared_via_channel: "Paylaşım kanalı kullanıldı",
  share_created: "Paylaşım oluşturuldu",
  share_opened: "Paylaşım açıldı",
  share_downloaded: "Paylaşım indirildi",
  share_password_failed: "Paylaşım şifresi başarısız",
  share_expired: "Paylaşım süresi doldu",
  share_revoked: "Paylaşım iptal edildi",
  share_access_denied: "Paylaşım erişimi reddedildi",
  share_limit_reached: "Paylaşım erişim limitine ulaşıldı",
  document_internal_shared: "Belge tenant üyesiyle paylaşıldı",
  document_internal_share_updated: "İç paylaşım yetkisi güncellendi",
  document_internal_share_removed: "İç paylaşım kaldırıldı",
  shared_document_viewed: "Paylaşılan belge görüntülendi",
  shared_document_downloaded: "Paylaşılan belge indirildi",
  document_version_created: "Belge versiyonu oluşturuldu",
  document_version_restored: "Belge versiyonu geri yüklendi",
  document_version_viewed: "Belge versiyonları görüntülendi",
  document_comment_created: "Belge yorumu eklendi",
  document_comment_updated: "Belge yorumu güncellendi",
  document_comment_deleted: "Belge yorumu silindi",
  important_date_created: "Önemli tarih eklendi",
  important_date_updated: "Önemli tarih güncellendi",
  important_date_deleted: "Önemli tarih silindi",
  important_date_completed: "Önemli tarih tamamlandı",
  workflow_created: "Onay süreci oluşturuldu",
  workflow_step_approved: "Onay adımı onaylandı",
  workflow_step_rejected: "Onay adımı reddedildi",
  workflow_cancelled: "Onay süreci iptal edildi",
  workflow_completed: "Onay süreci tamamlandı",
  integration_created: "Entegrasyon oluşturuldu",
  integration_updated: "Entegrasyon güncellendi",
  integration_deleted: "Entegrasyon silindi",
  integration_sync_queued: "Entegrasyon senkronu kuyruğa alındı",
  integration_sync_completed: "Entegrasyon senkronu tamamlandı",
  integration_sync_failed: "Entegrasyon senkronu başarısız",
  integration_imap_connected: "IMAP bağlandı",
  integration_imap_sync_started: "IMAP senkronu başladı",
  integration_imap_email_imported: "IMAP mail gövdesi içe aktarıldı",
  integration_imap_attachment_imported: "IMAP eki içe aktarıldı",
  integration_imap_sync_completed: "IMAP senkronu tamamlandı",
  integration_imap_sync_failed: "IMAP senkronu başarısız",
  integration_gmail_connected: "Gmail bağlandı",
  integration_gmail_sync_started: "Gmail senkronu başladı",
  integration_gmail_email_imported: "Gmail mail gövdesi içe aktarıldı",
  integration_gmail_attachment_imported: "Gmail eki içe aktarıldı",
  integration_gmail_sync_completed: "Gmail senkronu tamamlandı",
  integration_gmail_sync_failed: "Gmail senkronu başarısız",
  integration_gmail_token_refresh_failed: "Gmail token yenileme başarısız",
  integration_outlook_connected: "Outlook bağlandı",
  integration_outlook_sync_started: "Outlook senkronu başladı",
  integration_outlook_email_imported: "Outlook mail gövdesi içe aktarıldı",
  integration_outlook_attachment_imported: "Outlook eki içe aktarıldı",
  integration_outlook_sync_completed: "Outlook senkronu tamamlandı",
  integration_outlook_sync_failed: "Outlook senkronu başarısız",
  integration_outlook_token_refresh_failed: "Outlook token yenileme başarısız",
  integration_google_drive_connected: "Google Drive bağlandı",
  integration_google_drive_sync_started: "Google Drive senkronu başladı",
  integration_google_drive_file_imported: "Google Drive dosyası içe aktarıldı",
  integration_google_drive_sync_completed: "Google Drive senkronu tamamlandı",
  integration_google_drive_sync_failed: "Google Drive senkronu başarısız",
  integration_onedrive_connected: "OneDrive bağlandı",
  integration_onedrive_sync_started: "OneDrive senkronu başladı",
  integration_onedrive_file_imported: "OneDrive dosyası içe aktarıldı",
  integration_onedrive_sync_completed: "OneDrive senkronu tamamlandı",
  integration_onedrive_sync_failed: "OneDrive senkronu başarısız",
  user_registered: "Kullanıcı kaydı",
  user_logged_in: "Giriş yapıldı",
  user_logged_out: "Çıkış yapıldı",
  password_reset_requested: "Şifre sıfırlama talebi",
  password_reset_completed: "Şifre sıfırlama tamamlandı",
  profile_updated: "Profil güncellendi",
  password_changed: "Şifre değiştirildi",
  ai_chat_session_created: "AI sohbet oturumu oluşturuldu",
  ai_chat_message_sent: "AI sohbet mesajı gönderildi",
  ai_chat_session_deleted: "AI sohbet oturumu silindi",
  ai_tags_generated: "AI etiket önerileri oluşturuldu",
  ai_tag_accepted: "AI etiket önerisi kabul edildi",
  ai_tag_rejected: "AI etiket önerisi reddedildi",
  ai_metadata_generated: "AI metadata oluşturuldu",
  ai_metadata_viewed: "AI metadata görüntülendi",
  ai_important_dates_extracted: "AI önemli tarihler çıkarıldı",
  ai_important_date_accepted: "AI önemli tarih önerisi kabul edildi",
  ai_important_date_rejected: "AI önemli tarih önerisi reddedildi",
  ai_document_classified: "Belge AI ile sınıflandırıldı",
  ai_document_classification_viewed: "AI sınıflandırma görüntülendi",
  ai_entities_extracted: "AI varlıklar çıkarıldı",
  ai_entity_deleted: "AI varlık silindi",
  ai_entities_viewed: "AI varlıklar görüntülendi",
  ai_related_documents_generated: "AI ilişkili belge önerileri oluşturuldu",
  ai_related_document_accepted: "AI ilişkili belge önerisi kabul edildi",
  ai_related_document_rejected: "AI ilişkili belge önerisi reddedildi",
  ai_action_suggestions_generated: "AI aksiyon önerileri oluşturuldu",
  ai_action_suggestion_accepted: "AI aksiyon önerisi kabul edildi",
  ai_action_suggestion_rejected: "AI aksiyon önerisi reddedildi",
  ai_tool_executed: "AI aracı çalıştırıldı",
  ai_tool_execution_failed: "AI aracı çalıştırılamadı",
  ai_agent_message_sent: "AI agent mesajı gönderildi",
  ai_agent_action_suggested: "AI agent aksiyon önerdi",
};

export const auditDescriptionLabels: Record<string, string> = {
  "ai metadata generated": "AI metadata oluşturuldu",
  "ai metadata viewed": "AI metadata görüntülendi",
  "document summary generated": "Belge özeti oluşturuldu",
  "document summary viewed": "Belge özeti görüntülendi",
  "ai entities viewed": "AI varlıklar görüntülendi",
  "ai entities extracted": "AI varlıklar çıkarıldı",
  "ai entity deleted": "AI varlık silindi",
  "ai important dates extracted": "AI önemli tarihler çıkarıldı",
  "ai important date accepted": "AI önemli tarih önerisi kabul edildi",
  "ai important date rejected": "AI önemli tarih önerisi reddedildi",
  "ai document classified": "Belge AI ile sınıflandırıldı",
  "ai document classification viewed": "AI sınıflandırma görüntülendi",
  "ai tag suggestions generated": "AI etiket önerileri oluşturuldu",
  "ai suggested tag accepted": "AI etiket önerisi kabul edildi",
  "ai suggested tag rejected": "AI etiket önerisi reddedildi",
  "all ai suggested tags accepted": "Tüm AI etiket önerileri kabul edildi",
  "ai related documents generated": "AI ilişkili belge önerileri oluşturuldu",
  "ai related document accepted": "AI ilişkili belge önerisi kabul edildi",
  "ai related document rejected": "AI ilişkili belge önerisi reddedildi",
  "ai action suggestions generated": "AI aksiyon önerileri oluşturuldu",
  "ai action suggestion accepted": "AI aksiyon önerisi kabul edildi",
  "ai action suggestion rejected": "AI aksiyon önerisi reddedildi",
  "ai action suggestion updated": "AI aksiyon önerisi güncellendi",
  "ai tool executed": "AI aracı çalıştırıldı",
  "ai tool execution failed": "AI aracı çalıştırılamadı",
  "AI chat session created": "AI sohbet oturumu oluşturuldu",
  "AI chat message sent": "AI sohbet mesajı gönderildi",
  "AI chat session deleted": "AI sohbet oturumu silindi",
  "AI agent chat session created": "AI agent sohbet oturumu oluşturuldu",
  "AI agent message sent": "AI agent mesajı gönderildi",
  "AI agent action suggested": "AI agent aksiyon önerdi",
  "Belge tasindi": "Belge taşındı",
  "Belge etiketleri guncellendi": "Belge etiketleri güncellendi",
  "Belge yorumu eklendi": "Belge yorumu eklendi",
  "Belge yorumu silindi": "Belge yorumu silindi",
  "Onay sureci olusturuldu": "Onay süreci oluşturuldu",
  "Onay adimi onaylandi": "Onay adımı onaylandı",
  "Onay adimi reddedildi": "Onay adımı reddedildi",
};

export function formatAuditAction(action: string): string {
  return auditActionLabels[action] ?? action;
}

export function formatAuditDescription(log: Pick<AuditLog, "action" | "description">): string {
  const trimmed = (log.description ?? "").trim();
  if (!trimmed) {
    return auditActionLabels[log.action] ?? "-";
  }
  if (auditDescriptionLabels[trimmed]) {
    return auditDescriptionLabels[trimmed];
  }
  if (looksLikeContextualDescription(trimmed, log.action)) {
    return normalizeAsciiTurkishDescription(trimmed);
  }
  return auditActionLabels[log.action] ?? trimmed;
}

function looksLikeContextualDescription(description: string, action: string): boolean {
  if (description.includes(":")) {
    return true;
  }
  const englishActionPhrase = action.replace(/_/g, " ").toLowerCase();
  if (description.toLowerCase() === englishActionPhrase) {
    return false;
  }
  return /belge|klasor|klasör|etiket|kullanici|kullanıcı|onay|entegrasyon|yorum|profil|sifre|şifre|favori|paylas|paylaş|arama|versiyon|yuklendi|yüklendi|guncellendi|güncellendi|olusturuldu|oluşturuldu|silindi|tamamlandi|tamamlandı|giris|giriş|tasindi|taşındı/i.test(
    description,
  );
}

function normalizeAsciiTurkishDescription(description: string): string {
  return description
    .replace(/\byuklendi\b/gi, "yüklendi")
    .replace(/\bguncellendi\b/gi, "güncellendi")
    .replace(/\bolusturuldu\b/gi, "oluşturuldu")
    .replace(/\bsilindi\b/gi, "silindi")
    .replace(/\btamamlandi\b/gi, "tamamlandı")
    .replace(/\bgiris\b/gi, "giriş")
    .replace(/\btasindi\b/gi, "taşındı")
    .replace(/\bKlasor\b/g, "Klasör")
    .replace(/\bklasor\b/g, "klasör")
    .replace(/\bEtiket\b/g, "Etiket")
    .replace(/\bKullanici\b/g, "Kullanıcı")
    .replace(/\bkullanici\b/g, "kullanıcı")
    .replace(/\bSifre\b/g, "Şifre")
    .replace(/\bsifre\b/g, "şifre")
    .replace(/\bOnay sureci\b/g, "Onay süreci")
    .replace(/\bOnay adimi\b/g, "Onay adımı")
    .replace(/\bpaylasildi\b/gi, "paylaşıldı")
    .replace(/\bcikarildi\b/gi, "çıkarıldı")
    .replace(/\barcamasi\b/gi, "araması")
    .replace(/\bdegistirildi\b/gi, "değiştirildi");
}

export const auditEntityTypeLabels: Record<string, string> = {
  document: "Belge",
  folder: "Klasör",
  tag: "Etiket",
  user: "Kullanıcı",
  auth: "Kimlik",
  search: "Arama",
  share_link: "Paylaşım linki",
  document_share: "Belge paylaşımı",
  document_version: "Belge versiyonu",
  document_comment: "Belge yorumu",
  important_date: "Önemli tarih",
  workflow: "Onay süreci",
  integration: "Entegrasyon",
  ai_chat_session: "AI sohbet oturumu",
};
