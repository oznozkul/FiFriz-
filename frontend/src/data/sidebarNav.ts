import {
  Bot,
  CalendarDays,
  ClipboardCheck,
  FileArchive,
  FileText,
  Folder,
  History,
  LayoutDashboard,
  Layers,
  MessageSquareText,
  Plug,
  Settings,
  Share2,
  Star,
  Tag,
  Trash2,
} from "lucide-react";
import type { NavGroup } from "../types";

export const dashboardNavItem = {
  path: "/dashboard",
  label: "Dashboard",
  icon: LayoutDashboard,
  end: true,
};

export const documentsNavGroup: NavGroup = {
  label: "Belgeler",
  items: [
    { path: "/documents", label: "Belgelerim", icon: FileText },
    { path: "/collections", label: "Akıllı Koleksiyonlar", icon: Layers },
    { path: "/shared-with-me", label: "Benimle Paylaşılanlar", icon: Share2 },
    { path: "/favorites", label: "Favoriler", icon: Star },
    { path: "/trash", label: "Çöp Kutusu", icon: Trash2 },
  ],
};

export const organizationNavGroup: NavGroup = {
  label: "Organizasyon",
  items: [
    { path: "/folders", label: "Klasörler", icon: Folder },
    { path: "/tags", label: "Etiketler", icon: Tag },
  ],
};

export const workflowsNavGroup: NavGroup = {
  label: "İş Akışları",
  items: [
    { path: "/approvals", label: "Onaylar", icon: ClipboardCheck },
    { path: "/activity", label: "Aktivite", icon: History, roles: ["owner", "admin"] },
    { path: "/important-dates", label: "Önemli Tarihler", icon: CalendarDays },
  ],
};

export const integrationsNavItem = {
  path: "/integrations",
  label: "Entegrasyonlar",
  icon: Plug,
  end: true,
};

export const aiNavGroup: NavGroup = {
  label: "Yapay Zeka",
  items: [
    { path: "/ai-assistant", label: "AI Asistan", icon: Bot },
    { path: "/qa-history", label: "Sohbet Geçmişi", icon: MessageSquareText },
    { path: "/summaries", label: "Özetler", icon: FileArchive },
  ],
};

export const settingsNavItem = {
  path: "/settings",
  label: "Ayarlar",
  icon: Settings,
};

export const productNavGroups: NavGroup[] = [documentsNavGroup, organizationNavGroup, workflowsNavGroup];
