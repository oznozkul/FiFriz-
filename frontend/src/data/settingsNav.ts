import { Bell, Building2, CreditCard, HardDrive, Shield, User, Users } from "lucide-react";
import type { LucideIcon } from "lucide-react";

export type SettingsNavItem = {
  to: string;
  label: string;
  icon: LucideIcon;
  end?: boolean;
};

export const settingsNavItems: SettingsNavItem[] = [
  { to: "/settings/profile", label: "Profil", icon: User, end: true },
  { to: "/settings/security", label: "Güvenlik", icon: Shield, end: true },
  { to: "/settings/workspace", label: "Workspace", icon: Building2, end: true },
  { to: "/settings/members", label: "Üyeler", icon: Users, end: true },
  { to: "/settings/notifications", label: "Bildirimler", icon: Bell, end: true },
  { to: "/settings/storage", label: "Depolama", icon: HardDrive, end: true },
  { to: "/settings/billing", label: "Faturalandırma", icon: CreditCard, end: true },
];

/** @deprecated Use settingsNavItems — kept for breadcrumb lookup compatibility */
export const settingsNavGroups = [{ label: "Ayarlar", items: settingsNavItems }];
