import { Plug, RefreshCw, Settings, Trash2 } from "lucide-react";
import type { IntegrationAccount } from "../../services/integrationService";
import type { TenantRole } from "../../types/auth";
import { ContextMenuActions } from "../actions";
import { filterMenuByRole } from "../permissions";
import type { ContextMenuItem } from "../types";

type IntegrationMenuOptions = {
  integration: IntegrationAccount;
  role: TenantRole;
};

export function buildIntegrationMenuItems({ integration, role }: IntegrationMenuOptions): ContextMenuItem[] {
  void integration;
  return filterMenuByRole(
    [
      { id: "sync", label: "Şimdi Senkronize Et", icon: RefreshCw, action: ContextMenuActions.integration.sync },
      { id: "settings", label: "Ayarlar", icon: Settings, action: ContextMenuActions.integration.settings },
      { id: "divider-1", label: "", divider: true },
      { id: "disable", label: "Devre Dışı Bırak", icon: Plug, roles: ["owner", "admin"], action: ContextMenuActions.integration.disable },
      {
        id: "delete",
        label: "Sil",
        icon: Trash2,
        danger: true,
        roles: ["owner", "admin"],
        action: ContextMenuActions.integration.delete,
      },
    ],
    role,
  );
}

export function getIntegrationMenuContext(integration: IntegrationAccount) {
  return { entityType: "integration" as const, entity: integration };
}
