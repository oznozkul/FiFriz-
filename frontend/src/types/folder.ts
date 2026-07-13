export type FolderStatus = "active" | "archived" | "hidden" | "locked" | "deleted";

export type Folder = {
  id: string;
  tenant_id: string;
  parentId: string | null;
  name: string;
  description: string;
  status?: string;
  lifecycleStatus?: FolderStatus | string;
  isSystem?: boolean;
  isLocked?: boolean;
  lockedAt?: string | null;
  lockedByUserId?: string | null;
  inheritPermissions?: boolean;
  icon?: string | null;
  color?: string | null;
  lastActivityAt?: string | null;
  lastModifiedByUserId?: string | null;
  created_by_user_id: string;
  created_at: string;
  updated_at: string;
  children?: Folder[];
};

export type FolderInput = {
  name: string;
  parentId?: string;
  description?: string;
  status?: string;
  icon?: string | null;
  color?: string | null;
};

export type FolderMetadata = {
  folderId: string;
  documentCount: number;
  subFolderCount: number;
  totalSize: number;
  storageUsage: number;
  lastActivityAt?: string | null;
  lastModifiedByUserId?: string | null;
  aiIndexed: boolean;
};

export type FolderSearchResult = {
  id: string;
  name: string;
  description: string;
  parentId: string | null;
  status: string;
  lifecycleStatus: string;
  documentCount: number;
  subFolderCount: number;
  aiClassMatch: boolean;
};

export type FolderPermissionEntry = {
  userId: string;
  permission: "view" | "edit" | "manage";
};

export type FolderPermissions = {
  folderId: string;
  inheritPermissions: boolean;
  permissions: FolderPermissionEntry[];
};
