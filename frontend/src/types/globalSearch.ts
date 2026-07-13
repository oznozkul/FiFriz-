export type GlobalSearchDocument = {
  documentId: string;
  title?: string;
  originalFileName?: string;
  fileExtension?: string;
  matchedSnippet?: string;
  folderId?: string;
  folderName?: string;
  updatedAt?: string;
};

export type GlobalSearchFolder = {
  id: string;
  name: string;
  documentCount: number;
};

export type GlobalSearchTag = {
  id: string;
  name: string;
  usageCount: number;
};

export type GlobalSearchCollection = {
  key: string;
  title: string;
  description: string;
  icon: string;
};

export type GlobalSearchSharedDocument = {
  shareId: string;
  documentId: string;
  documentName: string;
  sharedByFullName: string;
  createdAt: string;
};

export type GlobalSearchIntegration = {
  id: string;
  provider: string;
  displayName: string;
  status: string;
};

export type GlobalSearchUser = {
  userId: string;
  fullName: string;
  email: string;
  role: string;
};

export type GlobalSearchAction = {
  id: string;
  label: string;
  path?: string;
  kind: string;
};

export type GlobalSearchSetting = {
  id: string;
  label: string;
  path: string;
};

export type GlobalSearchResponse = {
  documents: GlobalSearchDocument[];
  folders: GlobalSearchFolder[];
  tags: GlobalSearchTag[];
  collections: GlobalSearchCollection[];
  sharedDocuments: GlobalSearchSharedDocument[];
  integrations: GlobalSearchIntegration[];
  users: GlobalSearchUser[];
  actions: GlobalSearchAction[];
  settings: GlobalSearchSetting[];
};
