export type RouteMeta = {
  label: string;
  section?: string;
  parentPath?: string;
};

export const routeMetadata: Record<string, RouteMeta> = {
  "/dashboard": { label: "Dashboard" },
  "/documents": { label: "Belgelerim", section: "Belgeler" },
  "/collections": { label: "Akıllı Koleksiyonlar", section: "Belgeler" },
  "/favorites": { label: "Favoriler", section: "Belgeler" },
  "/shared-with-me": { label: "Benimle Paylaşılanlar", section: "Belgeler" },
  "/trash": { label: "Çöp Kutusu", section: "Belgeler" },
  "/folders": { label: "Klasörler", section: "Organizasyon" },
  "/tags": { label: "Etiketler", section: "Organizasyon" },
  "/approvals": { label: "Onaylar", section: "İş Akışları" },
  "/activity": { label: "Aktivite", section: "İş Akışları" },
  "/important-dates": { label: "Önemli Tarihler", section: "İş Akışları" },
  "/integrations": { label: "Entegrasyonlar", section: "Entegrasyonlar" },
  "/ai-assistant": { label: "AI Asistan", section: "Yapay Zeka" },
  "/qa-history": { label: "Sohbet Geçmişi", section: "Yapay Zeka" },
  "/summaries": { label: "Özetler", section: "Yapay Zeka" },
  "/notifications": { label: "Bildirimler" },
  "/search": { label: "Arama Sonuçları", section: "Belgeler" },
  "/recent": { label: "Son Açılanlar" },
  "/settings": { label: "Ayarlar" },
  "/settings/profile": { label: "Profil", section: "Ayarlar", parentPath: "/settings" },
  "/settings/security": { label: "Güvenlik", section: "Ayarlar", parentPath: "/settings" },
  "/settings/workspace": { label: "Workspace", section: "Ayarlar", parentPath: "/settings" },
  "/settings/members": { label: "Üyeler", section: "Ayarlar", parentPath: "/settings" },
  "/settings/notifications": { label: "Bildirimler", section: "Ayarlar", parentPath: "/settings" },
  "/settings/roles": { label: "Roller & Yetkiler", section: "Ayarlar", parentPath: "/settings" },
  "/settings/storage": { label: "Depolama", section: "Ayarlar", parentPath: "/settings" },
  "/settings/appearance": { label: "Görünüm", section: "Ayarlar", parentPath: "/settings" },
  "/settings/plan": { label: "Plan", section: "Ayarlar", parentPath: "/settings" },
  "/settings/usage": { label: "Kullanım", section: "Ayarlar", parentPath: "/settings" },
  "/settings/billing": { label: "Faturalandırma", section: "Ayarlar", parentPath: "/settings" },
};

export const smartCollectionTitles: Record<string, string> = {
  favorites: "Favoriler",
  "shared-with-me": "Paylaşılanlar",
  "recent-uploads": "Son Yüklenenler",
  "expiring-soon": "Yakında Sona Erenler",
  "pending-approval": "Onay Bekleyenler",
  "large-files": "Büyük Dosyalar",
  "email-imports": "E-posta İçe Aktarımları",
  "cloud-sync": "Bulut Senkronu",
};
