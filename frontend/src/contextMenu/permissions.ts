import type { ContextMenuItem } from "./types";
import type { TenantRole } from "../types/auth";

export function isPrivilegedRole(role: TenantRole) {
  return role === "owner" || role === "admin";
}

export function canManageWorkspace(role: TenantRole) {
  return isPrivilegedRole(role);
}

export function filterMenuByRole(items: ContextMenuItem[], role: TenantRole): ContextMenuItem[] {
  const filtered: ContextMenuItem[] = [];

  for (const item of items) {
    if (item.divider) {
      if (filtered.length > 0 && !filtered[filtered.length - 1]?.divider) {
        filtered.push(item);
      }
      continue;
    }

    if (item.roles && !item.roles.includes(role)) {
      continue;
    }

    const nextItem = item.children
      ? { ...item, children: filterMenuByRole(item.children, role) }
      : item;

    if (nextItem.children && nextItem.children.length === 0) {
      continue;
    }

    filtered.push(nextItem);
  }

  while (filtered.length > 0 && filtered[filtered.length - 1]?.divider) {
    filtered.pop();
  }

  return filtered;
}
