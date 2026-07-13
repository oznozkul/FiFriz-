import type { LucideIcon } from "lucide-react";
import {
  Bell,
  Bot,
  FileArchive,
  FolderPlus,
  Layers,
  LayoutDashboard,
  MessageSquareText,
  Plug,
  Tag,
  Trash2,
  Upload,
  User,
} from "lucide-react";

export type CommandPaletteItem = {
  id: string;
  group: string;
  label: string;
  subtitle?: string;
  icon?: LucideIcon;
  path?: string;
  kind: "navigate" | "action" | "search";
  actionId?: string;
  pinEligible?: boolean;
};

export const pinnedActionCatalog: CommandPaletteItem[] = [
  { id: "upload-document", group: "Aksiyonlar", label: "Belge Yükle", icon: Upload, kind: "action", actionId: "upload-document", pinEligible: true },
  { id: "go-dashboard", group: "Aksiyonlar", label: "Dashboard'a Git", icon: LayoutDashboard, path: "/dashboard", kind: "navigate", actionId: "go-dashboard", pinEligible: true },
  { id: "go-integrations", group: "Aksiyonlar", label: "Entegrasyonlara Git", icon: Plug, path: "/integrations", kind: "navigate", actionId: "go-integrations", pinEligible: true },
  { id: "go-trash", group: "Aksiyonlar", label: "Çöp Kutusuna Git", icon: Trash2, path: "/trash", kind: "navigate", actionId: "go-trash", pinEligible: true },
  { id: "open-notifications", group: "Aksiyonlar", label: "Bildirimleri Aç", icon: Bell, path: "/notifications", kind: "navigate", actionId: "open-notifications", pinEligible: true },
  { id: "open-profile", group: "Aksiyonlar", label: "Profili Aç", icon: User, path: "/settings/profile", kind: "navigate", actionId: "open-profile", pinEligible: true },
  { id: "new-folder", group: "Aksiyonlar", label: "Yeni Klasör", icon: FolderPlus, path: "/folders", kind: "navigate", actionId: "new-folder", pinEligible: true },
  { id: "new-tag", group: "Aksiyonlar", label: "Yeni Etiket", icon: Tag, path: "/tags", kind: "navigate", actionId: "new-tag", pinEligible: true },
];

export const aiActionCatalog: CommandPaletteItem[] = [
  { id: "ai-assistant", group: "AI Aksiyonları", label: "AI Asistan", subtitle: "Belgelerinizle sohbet edin", icon: Bot, path: "/ai-assistant", kind: "navigate", actionId: "ai-assistant" },
  { id: "qa-history", group: "AI Aksiyonları", label: "Sohbet Geçmişi", subtitle: "Önceki AI konuşmaları", icon: MessageSquareText, path: "/qa-history", kind: "navigate", actionId: "qa-history" },
  { id: "summaries", group: "AI Aksiyonları", label: "Özetler", subtitle: "Belge özetlerine git", icon: FileArchive, path: "/summaries", kind: "navigate", actionId: "summaries" },
  { id: "go-collections", group: "AI Aksiyonları", label: "Akıllı Koleksiyonlar", subtitle: "Dinamik belge grupları", icon: Layers, path: "/collections", kind: "navigate", actionId: "go-collections" },
];

const AI_ACTION_IDS = new Set(aiActionCatalog.map((item) => item.actionId ?? item.id));

export function isAiActionId(actionId: string) {
  return AI_ACTION_IDS.has(actionId) || actionId.startsWith("ai-");
}

export function getPinnedActions(pinnedIds: string[]) {
  const map = new Map(pinnedActionCatalog.map((item) => [item.actionId ?? item.id, item]));
  return pinnedIds.map((id) => map.get(id)).filter((item): item is CommandPaletteItem => Boolean(item));
}
