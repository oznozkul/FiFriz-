import {
  Bot,
  CalendarClock,
  FileText,
  Layers,
  Link2,
  ScanSearch,
  Search,
  Sparkles,
  type LucideIcon,
} from "lucide-react";

export type AiChatMessage = {
  role: "user" | "assistant";
  content: string;
};

export const aiShowcaseConversation: AiChatMessage[] = [
  {
    role: "user",
    content: "Bu sözleşmenin bitiş tarihi nedir?",
  },
  {
    role: "assistant",
    content: "Bitiş tarihi 15 Mart 2027.\n\nSize 30 gün önceden hatırlatma oluşturabilirim.",
  },
  {
    role: "user",
    content: "Bu klasördeki faturaları özetle",
  },
  {
    role: "assistant",
    content: "Toplam 18 fatura bulundu.\n\nToplam ödeme:\n₺428.750",
  },
];

export type AiFloatingBadge = {
  label: string;
  icon: LucideIcon;
  className: string;
};

export const aiFloatingBadges: AiFloatingBadge[] = [
  {
    label: "12 belge analiz edildi",
    icon: FileText,
    className: "-left-2 top-10 sm:-left-6",
  },
  {
    label: "8 önemli tarih bulundu",
    icon: CalendarClock,
    className: "-right-1 top-24 sm:-right-5",
  },
  {
    label: "AI Summary hazır",
    icon: Sparkles,
    className: "-left-1 bottom-20 sm:-left-8",
  },
  {
    label: "4 yeni etiket önerildi",
    icon: Layers,
    className: "-right-2 bottom-10 sm:-right-6",
  },
];

export type AiCapability = {
  title: string;
  description: string;
  icon: LucideIcon;
};

export const aiCapabilities: AiCapability[] = [
  {
    title: "AI Summary",
    description: "Belgeleri saniyeler içinde özetler.",
    icon: Sparkles,
  },
  {
    title: "Smart Search",
    description: "Anlam bazlı arama yapar.",
    icon: Search,
  },
  {
    title: "Entity Extraction",
    description: "İsimleri, şirketleri ve tarihleri çıkarır.",
    icon: ScanSearch,
  },
  {
    title: "Automatic Classification",
    description: "Belgeleri otomatik kategorilere ayırır.",
    icon: Layers,
  },
  {
    title: "Important Dates",
    description: "Son tarihleri algılar.",
    icon: CalendarClock,
  },
  {
    title: "Related Documents",
    description: "Benzer belgeleri önerir.",
    icon: Link2,
  },
];

export const aiShowcaseBotIcon = Bot;
