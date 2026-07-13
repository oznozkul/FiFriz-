import type { LucideIcon } from "lucide-react";
import type { WorkspacePanelSection } from "./documentWorkspaceUtils";

export type WorkspaceNavItem = {
  id: WorkspacePanelSection;
  label: string;
  group: "main" | "ai" | "advanced";
};

export const WORKSPACE_MAIN_NAV: WorkspaceNavItem[] = [
  { id: "general", label: "Genel", group: "main" },
  { id: "tags", label: "Etiketler", group: "main" },
  { id: "folder", label: "Klasör", group: "main" },
  { id: "shares", label: "Paylaşımlar", group: "main" },
  { id: "comments", label: "Yorumlar", group: "main" },
  { id: "versions", label: "Versiyonlar", group: "main" },
  { id: "dates", label: "Tarihler", group: "main" },
  { id: "workflow", label: "Workflow", group: "main" },
  { id: "activity", label: "Aktivite", group: "main" },
  { id: "related", label: "İlgili", group: "main" },
];

export const WORKSPACE_AI_NAV: WorkspaceNavItem[] = [
  { id: "summary", label: "AI Özeti", group: "ai" },
  { id: "aiMetadata", label: "Metadata", group: "ai" },
  { id: "actionSuggestions", label: "Öneriler", group: "ai" },
  { id: "suggestedTags", label: "Etiket AI", group: "ai" },
  { id: "ai", label: "Asistan", group: "ai" },
];

export const WORKSPACE_ADVANCED_NAV: WorkspaceNavItem[] = [
  { id: "health", label: "Sağlık", group: "advanced" },
  { id: "classification", label: "Tür", group: "advanced" },
  { id: "entities", label: "Varlıklar", group: "advanced" },
  { id: "aiProgress", label: "AI İlerleme", group: "advanced" },
];

export const WORKSPACE_ALL_NAV = [...WORKSPACE_MAIN_NAV, ...WORKSPACE_AI_NAV, ...WORKSPACE_ADVANCED_NAV];

export type WorkspaceNavItemWithIcon = WorkspaceNavItem & { icon?: LucideIcon };
