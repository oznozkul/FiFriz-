import { Download, Eye, RotateCcw } from "lucide-react";
import type { DocumentVersion } from "../../services/versionService";
import type { TenantRole } from "../../types/auth";
import { ContextMenuActions } from "../actions";
import { filterMenuByRole } from "../permissions";
import type { ContextMenuItem } from "../types";

type VersionMenuOptions = {
  version: DocumentVersion;
  role: TenantRole;
};

export function buildVersionMenuItems({ version, role }: VersionMenuOptions): ContextMenuItem[] {
  return filterMenuByRole(
    [
      { id: "preview", label: "Önizle", icon: Eye, action: ContextMenuActions.version.preview },
      { id: "download", label: "İndir", icon: Download, action: ContextMenuActions.version.download },
      {
        id: "restore",
        label: "Geri Yükle",
        icon: RotateCcw,
        disabled: version.isCurrent,
        roles: ["owner", "admin"],
        action: ContextMenuActions.version.restore,
      },
    ],
    role,
  );
}

export function getVersionMenuContext(version: DocumentVersion, documentId: string) {
  return { entityType: "version" as const, entity: version, meta: { documentId } };
}
