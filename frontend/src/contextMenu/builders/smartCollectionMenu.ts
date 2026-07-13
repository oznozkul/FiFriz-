import { Link2, Pin, RefreshCw } from "lucide-react";
import type { SmartCollection } from "../../services/smartCollectionService";
import type { TenantRole } from "../../types/auth";
import { ContextMenuActions } from "../actions";
import { filterMenuByRole } from "../permissions";
import type { ContextMenuItem } from "../types";

type SmartCollectionMenuOptions = {
  collection: SmartCollection;
  role: TenantRole;
};

export function buildSmartCollectionMenuItems({ collection, role }: SmartCollectionMenuOptions): ContextMenuItem[] {
  void collection;
  return filterMenuByRole(
    [
      { id: "refresh", label: "Yenile", icon: RefreshCw, action: ContextMenuActions.smartCollection.refresh },
      { id: "pin", label: "Pinle", icon: Pin, action: ContextMenuActions.smartCollection.pin },
      { id: "shortcut", label: "Kısayol Oluştur", icon: Link2, action: ContextMenuActions.smartCollection.shortcut },
    ],
    role,
  );
}

export function getSmartCollectionMenuContext(collection: SmartCollection) {
  return { entityType: "smart-collection" as const, entity: collection };
}
