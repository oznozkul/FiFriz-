import type { LucideIcon } from "lucide-react";
import type { TenantRole } from "../types/auth";

export type ContextMenuEntityType =
  | "document"
  | "folder"
  | "tag"
  | "smart-collection"
  | "comment"
  | "version"
  | "integration"
  | "explorer-background";

export type ContextMenuContext = {
  entityType: ContextMenuEntityType;
  entity: unknown;
  meta?: Record<string, unknown>;
};

export type ContextMenuItem = {
  id: string;
  label: string;
  icon?: LucideIcon;
  shortcut?: string;
  danger?: boolean;
  disabled?: boolean;
  divider?: boolean;
  children?: ContextMenuItem[];
  roles?: TenantRole[];
  action?: string;
};

export type ContextMenuActionDetail = {
  action: string;
  context: ContextMenuContext;
};

export type ContextMenuState = {
  x: number;
  y: number;
  items: ContextMenuItem[];
  context: ContextMenuContext;
};
