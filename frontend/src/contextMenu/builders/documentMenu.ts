import {
  Activity,
  CalendarPlus,
  Download,
  ExternalLink,
  FolderInput,
  GitBranch,
  History,
  Pencil,
  Share2,
  Sparkles,
  Star,
  Tag,
  Trash2,
} from "lucide-react";
import type { ApiDocument } from "../../types/document";
import type { TenantRole } from "../../types/auth";
import { ContextMenuActions } from "../actions";
import { filterMenuByRole } from "../permissions";
import type { ContextMenuItem } from "../types";

type DocumentMenuOptions = {
  document: ApiDocument;
  role: TenantRole;
  isTrash?: boolean;
};

export function buildDocumentMenuItems({ document, role, isTrash = false }: DocumentMenuOptions): ContextMenuItem[] {
  const isFavorite = document.isFavorite ?? false;

  if (isTrash) {
    return filterMenuByRole(
      [
        { id: "open", label: "Aç", icon: ExternalLink, action: ContextMenuActions.document.open },
        { id: "restore", label: "Geri Yükle", icon: History, action: ContextMenuActions.document.restore },
        { id: "divider-1", label: "", divider: true },
        {
          id: "delete",
          label: "Kalıcı Olarak Sil",
          icon: Trash2,
          danger: true,
          roles: ["owner", "admin"],
          action: ContextMenuActions.document.trash,
        },
      ],
      role,
    );
  }

  return filterMenuByRole(
    [
      { id: "open", label: "Aç", icon: ExternalLink, shortcut: "Enter", action: ContextMenuActions.document.open },
      { id: "download", label: "İndir", icon: Download, action: ContextMenuActions.document.download },
      {
        id: "favorite",
        label: isFavorite ? "Favorilerden Çıkar" : "Favorilere Ekle",
        icon: Star,
        action: ContextMenuActions.document.toggleFavorite,
      },
      { id: "divider-1", label: "", divider: true },
      { id: "tag", label: "Etiketle", icon: Tag, action: ContextMenuActions.document.tag },
      { id: "move", label: "Taşı", icon: FolderInput, roles: ["owner", "admin"], action: ContextMenuActions.document.move },
      { id: "share", label: "Paylaş", icon: Share2, roles: ["owner", "admin"], action: ContextMenuActions.document.share },
      { id: "rename", label: "Yeniden Adlandır", icon: Pencil, roles: ["owner", "admin"], action: ContextMenuActions.document.rename },
      { id: "divider-2", label: "", divider: true },
      { id: "versions", label: "Versiyonlar", icon: History, action: ContextMenuActions.document.versions },
      { id: "activity", label: "Aktivite", icon: Activity, roles: ["owner", "admin"], action: ContextMenuActions.document.activity },
      {
        id: "important-date",
        label: "Önemli Tarih Ekle",
        icon: CalendarPlus,
        action: ContextMenuActions.document.addImportantDate,
      },
      {
        id: "workflow",
        label: "Workflow Başlat",
        icon: GitBranch,
        roles: ["owner", "admin"],
        action: ContextMenuActions.document.startWorkflow,
      },
      {
        id: "ai",
        label: "Yapay Zeka",
        icon: Sparkles,
        children: [
          { id: "ai-summarize", label: "Özetle", action: ContextMenuActions.ai.summarize },
          { id: "ai-ask", label: "Soru Sor", action: ContextMenuActions.ai.ask },
          { id: "ai-keywords", label: "Anahtar Kelime", action: ContextMenuActions.ai.keywords },
        ],
      },
      { id: "divider-3", label: "", divider: true },
      {
        id: "trash",
        label: "Çöp Kutusuna Taşı",
        icon: Trash2,
        danger: true,
        roles: ["owner", "admin"],
        action: ContextMenuActions.document.trash,
      },
    ],
    role,
  );
}

export function getDocumentMenuContext(document: ApiDocument, meta?: Record<string, unknown>) {
  return { entityType: "document" as const, entity: document, meta };
}
