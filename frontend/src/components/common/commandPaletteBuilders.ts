import { Folder, Plug, Search, Settings, Share2, Sparkles, Tag, Users } from "lucide-react";
import type { NavigateFunction } from "react-router-dom";
import { aiActionCatalog, getPinnedActions, isAiActionId, type CommandPaletteItem } from "../../data/commandPaletteCatalog";
import type { FolderLabels } from "../../hooks/useFolderLabels";
import {
  getPinnedActionIds,
  getRecentDocuments,
  getRecentSearches,
  pushRecentDocument,
  pushRecentSearch,
} from "../../services/commandPaletteStorage";
import { stripSnippetMarkup } from "../../services/documentService";
import type { GlobalSearchResponse } from "../../types/globalSearch";
import type { HybridSearchMatchType } from "../../types/hybridSearch";
import { formatDate } from "../documents/documentPresentation";
import { smartCollectionIcon } from "../smartCollections/smartCollectionIcons";
import { filterCatalogByQuery, type PaletteItem } from "./commandPaletteUtils";

function getDocumentTitle(doc: GlobalSearchResponse["documents"][number]) {
  return doc.originalFileName ?? doc.title ?? "Belge";
}

function getDocumentExtension(doc: GlobalSearchResponse["documents"][number]) {
  const title = getDocumentTitle(doc);
  return doc.fileExtension ?? title.split(".").pop() ?? "file";
}

function catalogToPaletteItem(
  item: CommandPaletteItem,
  group: string,
  onSelect: () => void,
): PaletteItem {
  return {
    key: `catalog-${item.id}`,
    group,
    label: item.label,
    subtitle: item.subtitle,
    icon: item.icon,
    kind: item.kind,
    path: item.path,
    actionId: item.actionId,
    pinEligible: item.pinEligible,
    onSelect,
  };
}

export function buildItemsFromResponse(
  data: GlobalSearchResponse,
  query: string,
  navigate: NavigateFunction,
  onUploadClick: () => void,
  onClose: () => void,
  options: {
    folderGroupLabel: string;
    folderSingularLower: string;
    documentMatchTypes?: Map<string, HybridSearchMatchType>;
    extraAiActions?: CommandPaletteItem[];
  },
): PaletteItem[] {
  const items: PaletteItem[] = [];
  const { folderGroupLabel, folderSingularLower, documentMatchTypes, extraAiActions = [] } = options;

  const closeAndNavigate = (path: string) => {
    onClose();
    navigate(path);
  };

  const closeAndUpload = () => {
    onClose();
    onUploadClick();
  };

  for (const doc of data.documents) {
    const title = getDocumentTitle(doc);
    const extension = getDocumentExtension(doc);
    const matchType = documentMatchTypes?.get(doc.documentId);
    items.push({
      key: `doc-${doc.documentId}`,
      group: "Belgeler",
      label: title,
      subtitle:
        [
          doc.folderName ? `${folderSingularLower}: ${doc.folderName}` : null,
          doc.updatedAt ? formatDate(doc.updatedAt) : null,
        ]
          .filter(Boolean)
          .join(" · ") || undefined,
      snippet: doc.matchedSnippet,
      extension,
      matchType,
      kind: "navigate",
      path: `/documents/${doc.documentId}`,
      onSelect: () => {
        pushRecentSearch(query);
        pushRecentDocument({ id: doc.documentId, title, extension });
        const querySuffix = query.trim() ? `?q=${encodeURIComponent(query.trim())}` : "";
        closeAndNavigate(`/documents/${doc.documentId}${querySuffix}`);
      },
    });
  }

  for (const shared of data.sharedDocuments) {
    items.push({
      key: `shared-${shared.shareId}`,
      group: "Paylaşılan Belgeler",
      label: shared.documentName,
      subtitle: `${shared.sharedByFullName} · ${formatDate(shared.createdAt)}`,
      icon: Share2,
      kind: "navigate",
      path: `/documents/${shared.documentId}`,
      onSelect: () => {
        pushRecentSearch(query);
        closeAndNavigate(`/documents/${shared.documentId}`);
      },
    });
  }

  for (const folder of data.folders) {
    items.push({
      key: `folder-${folder.id}`,
      group: folderGroupLabel,
      label: folder.name,
      subtitle: `${folder.documentCount} belge`,
      icon: Folder,
      kind: "navigate",
      path: `/folders?folderId=${folder.id}`,
      onSelect: () => {
        pushRecentSearch(query);
        closeAndNavigate(`/folders?folderId=${folder.id}`);
      },
    });
  }

  for (const tag of data.tags) {
    items.push({
      key: `tag-${tag.id}`,
      group: "Etiketler",
      label: tag.name,
      subtitle: `${tag.usageCount} kullanım`,
      icon: Tag,
      kind: "navigate",
      path: "/tags",
      onSelect: () => {
        pushRecentSearch(query);
        closeAndNavigate("/tags");
      },
    });
  }

  for (const collection of data.collections) {
    items.push({
      key: `collection-${collection.key}`,
      group: "Akıllı Koleksiyonlar",
      label: collection.title,
      subtitle: collection.description,
      icon: smartCollectionIcon(collection.icon),
      kind: "navigate",
      path: `/collections/${collection.key}`,
      onSelect: () => {
        pushRecentSearch(query);
        closeAndNavigate(`/collections/${collection.key}`);
      },
    });
  }

  for (const integration of data.integrations) {
    items.push({
      key: `integration-${integration.id}`,
      group: "Entegrasyonlar",
      label: integration.displayName,
      subtitle: integration.provider,
      icon: Plug,
      kind: "navigate",
      path: "/integrations",
      onSelect: () => {
        pushRecentSearch(query);
        closeAndNavigate("/integrations");
      },
    });
  }

  for (const user of data.users) {
    items.push({
      key: `user-${user.userId}`,
      group: "Kullanıcılar",
      label: user.fullName,
      subtitle: `${user.email} · ${user.role}`,
      icon: Users,
      kind: "navigate",
      path: "/settings/members",
      onSelect: () => {
        pushRecentSearch(query);
        closeAndNavigate("/settings/members");
      },
    });
  }

  for (const action of data.actions) {
    const isUpload = action.id === "upload-document";
    const isAI = isAiActionId(action.id);
    items.push({
      key: `action-${action.id}`,
      group: isAI ? "AI Aksiyonları" : "Aksiyonlar",
      label: action.label,
      icon: isAI ? Sparkles : undefined,
      kind: isUpload ? "action" : "navigate",
      path: action.path,
      actionId: action.id,
      pinEligible: !isAI,
      onSelect: () => {
        if (query.trim()) pushRecentSearch(query);
        if (isUpload) {
          closeAndUpload();
          return;
        }
        if (action.path) closeAndNavigate(action.path);
      },
    });
  }

  for (const aiAction of extraAiActions) {
    items.push(
      catalogToPaletteItem(aiAction, "AI Aksiyonları", () => {
        if (query.trim()) pushRecentSearch(query);
        if (aiAction.path) closeAndNavigate(aiAction.path);
      }),
    );
  }

  for (const setting of data.settings) {
    items.push({
      key: `setting-${setting.id}`,
      group: "Ayarlar",
      label: setting.label,
      icon: Settings,
      kind: "navigate",
      path: setting.path,
      onSelect: () => {
        if (query.trim()) pushRecentSearch(query);
        closeAndNavigate(setting.path);
      },
    });
  }

  return items;
}

export function buildEmptyStateItems(
  setQuery: (value: string) => void,
  navigate: NavigateFunction,
  onUploadClick: () => void,
  onClose: () => void,
  pinnedRevision: number,
  query: string,
  folderLabels: FolderLabels,
): PaletteItem[] {
  void pinnedRevision;
  const items: PaletteItem[] = [];

  const closeAndNavigate = (path: string) => {
    onClose();
    navigate(path);
  };

  const closeAndUpload = () => {
    onClose();
    onUploadClick();
  };

  for (const pinned of getPinnedActions(getPinnedActionIds())) {
    const label = pinned.actionId === "new-folder" ? folderLabels.newLabel : pinned.label;
    items.push(
      catalogToPaletteItem({ ...pinned, label }, "Sabitlenenler", () => {
        if (pinned.kind === "action" && pinned.actionId === "upload-document") {
          closeAndUpload();
          return;
        }
        if (pinned.path) closeAndNavigate(pinned.path);
      }),
    );
  }

  for (const recent of getRecentDocuments()) {
    items.push({
      key: `recent-doc-${recent.id}`,
      group: "Son Açılan Belgeler",
      label: recent.title,
      subtitle: recent.extension ? recent.extension.toUpperCase() : undefined,
      extension: recent.extension,
      kind: "navigate",
      path: `/documents/${recent.id}`,
      onSelect: () => closeAndNavigate(`/documents/${recent.id}`),
    });
  }

  for (const recentSearch of getRecentSearches()) {
    items.push({
      key: `recent-search-${recentSearch}`,
      group: "Son Aramalar",
      label: recentSearch,
      icon: Search,
      kind: "search",
      onSelect: () => setQuery(recentSearch),
    });
  }

  for (const aiAction of filterCatalogByQuery(aiActionCatalog, query)) {
    items.push(
      catalogToPaletteItem(aiAction, "AI Aksiyonları", () => {
        if (aiAction.path) closeAndNavigate(aiAction.path);
      }),
    );
  }

  return items;
}

export function stripPaletteSnippet(snippet?: string) {
  if (!snippet) return undefined;
  return stripSnippetMarkup(snippet);
}

// re-export for palette row
export { stripSnippetMarkup };
