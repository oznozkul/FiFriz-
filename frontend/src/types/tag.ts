export type TagStatus = "active" | "archived" | "hidden" | "deleted" | "locked";

export type Tag = {
  id: string;
  name: string;
  status?: string;
  lifecycleStatus?: TagStatus | string;
  isSystem?: boolean;
  color: string | null;
  icon?: string | null;
  documentCount?: number;
  createdAt: string;
  updatedAt: string;
};

export type TagBrief = {
  id: string;
  name: string;
  color: string | null;
  icon?: string | null;
};

export type TagAlias = {
  id: string;
  tagId: string;
  alias: string;
  createdAt: string;
};

export type TagStats = {
  tagId: string;
  documentCount: number;
  lastUsedAt?: string | null;
  createdAt: string;
  mostUsedFolders: Array<{ folderId: string | null; folderName: string; count: number }>;
  relatedTags: TagBrief[];
};

export type TagSearchResult = {
  id: string;
  name: string;
  status: string;
  lifecycleStatus: string;
  color: string | null;
  icon?: string | null;
  documentCount: number;
  aliasMatch: boolean;
};

export const tagColorOptions = [
  { value: "#7c3aed", label: "Mor" },
  { value: "#2563eb", label: "Mavi" },
  { value: "#059669", label: "Yesil" },
  { value: "#d97706", label: "Turuncu" },
  { value: "#e11d48", label: "Kirmizi" },
  { value: "#64748b", label: "Gri" },
] as const;

export const tagIconOptions = [
  { value: "tag", label: "Etiket" },
  { value: "file", label: "Dosya" },
  { value: "receipt", label: "Fatura" },
  { value: "scale", label: "Hukuk" },
  { value: "heart", label: "Saglik" },
  { value: "building", label: "Kurumsal" },
] as const;
