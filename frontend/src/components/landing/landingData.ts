import {
  Bot,
  ClipboardCheck,
  CloudUpload,
  FolderKanban,
  Globe,
  History,
  Plug,
  Search,
  Shield,
  Sparkles,
  Zap,
  type LucideIcon,
} from "lucide-react";

export const landingNavLinks = [
  { label: "Ürün", href: "#urun" },
  { label: "Özellikler", href: "#ozellikler" },
  { label: "Çözümler", href: "#cozumler" },
  { label: "Fiyatlandırma", href: "#fiyatlandirma" },
  { label: "Hakkımızda", href: "#hakkimizda" },
  { label: "İletişim", href: "#iletisim" },
] as const;

export type LandingFeature = {
  title: string;
  description: string;
  icon: LucideIcon;
  highlights: readonly [string, string];
};

export const landingFeatures: LandingFeature[] = [
  {
    title: "Yapay Zeka Destekli",
    description: "Belgelerinizi özetler, sınıflandırır ve anlamlandırır.",
    icon: Sparkles,
    highlights: ["AI Summary", "Metadata Extraction"],
  },
  {
    title: "Güvenli ve Uyumlu",
    description: "Enterprise seviyesinde güvenlik.",
    icon: Shield,
    highlights: ["RBAC", "KVKK Ready"],
  },
  {
    title: "Kolay Entegrasyon",
    description: "Google Drive, Gmail, Outlook, Logo ERP vb.",
    icon: Plug,
    highlights: ["Plugin Marketplace", "Cloud & Mail Sync"],
  },
  {
    title: "Akıllı Arama",
    description: "Kelime ve anlam bazlı arama.",
    icon: Search,
    highlights: ["Hybrid Search", "Semantic Match"],
  },
  {
    title: "Workflow & Onay",
    description: "Belge süreçlerini dijitalleştirin.",
    icon: ClipboardCheck,
    highlights: ["Approval Engine", "Multi-step Flow"],
  },
  {
    title: "Versiyon Takibi",
    description: "Tüm belge geçmişini koruyun.",
    icon: History,
    highlights: ["Version History", "Change Audit"],
  },
];

export type LandingStep = {
  title: string;
  description: string;
  icon: LucideIcon;
};

export const landingSteps: LandingStep[] = [
  {
    title: "Yükle",
    description: "PDF, Word ve görsellerinizi tek tıkla sisteme aktarın.",
    icon: CloudUpload,
  },
  {
    title: "AI İşler",
    description: "Yapay zeka metni çıkarır, etiketler ve özetler oluşturur.",
    icon: Bot,
  },
  {
    title: "Düzenle",
    description: "Klasör, kategori ve etiketlerle belgelerinizi organize edin.",
    icon: FolderKanban,
  },
  {
    title: "Hızlan",
    description: "Arama, onay ve entegrasyonlarla iş akışınızı hızlandırın.",
    icon: Zap,
  },
];

export const landingStats = [
  { value: "1M+", label: "Yönetilen Belge" },
  { value: "1.000+", label: "Mutlu Kurum" },
  { value: "%99,9", label: "Sistem Uptime" },
  { value: "50+", label: "Entegrasyon" },
] as const;

export const footerProductLinks = [
  { label: "Özellikler", href: "#ozellikler" },
  { label: "Entegrasyonlar", href: "#cozumler" },
  { label: "Fiyatlandırma", href: "#fiyatlandirma" },
  { label: "Güvenlik", href: "#ozellikler" },
  { label: "API", href: "#iletisim" },
] as const;

export const footerResourceLinks = [
  { label: "Dokümantasyon", href: "#iletisim" },
  { label: "Blog", href: "#iletisim" },
  { label: "Yardım Merkezi", href: "#iletisim" },
  { label: "Durum Sayfası", href: "#iletisim" },
] as const;

export const footerCompanyLinks = [
  { label: "Hakkımızda", href: "#hakkimizda" },
  { label: "Kariyer", href: "#iletisim" },
  { label: "İletişim", href: "#iletisim" },
  { label: "Basın", href: "#iletisim" },
] as const;

export const footerLegalLinks = [
  { label: "Gizlilik Politikası", href: "#iletisim" },
  { label: "Kullanım Şartları", href: "#iletisim" },
  { label: "KVKK", href: "#iletisim" },
  { label: "Çerez Politikası", href: "#iletisim" },
] as const;

export const socialLinks = [
  { label: "LinkedIn", href: "https://linkedin.com", icon: Globe },
  { label: "Twitter", href: "https://twitter.com", icon: Globe },
  { label: "GitHub", href: "https://github.com", icon: Globe },
] as const;
