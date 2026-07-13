import { apiRequest } from "./apiClient";

export type DashboardRecentDocument = {
  id: string;
  originalFileName: string;
  fileExtension: string;
  fileSize: number;
  status: string;
  createdAt: string;
  folderName: string | null;
  isFavorite: boolean;
};

export type DashboardRecentActivity = {
  id: string;
  action: string;
  entityType: string;
  description: string;
  createdAt: string;
};

export type DashboardDocumentTypeStat = {
  extension: string;
  count: number;
};

export type DashboardTopFolder = {
  folderId: string;
  folderName: string;
  documentCount: number;
};

export type DashboardImportantDateItem = {
  id: string;
  documentId: string;
  title: string;
  date: string;
  type?: string | null;
  isCompleted: boolean;
  documentName: string;
  assignedToUserId?: string | null;
};

export type DashboardSummary = {
  totalDocuments: number;
  totalFolders: number;
  favoriteDocuments: number;
  trashedDocuments: number;
  totalStorageBytes: number;
  documentsUploadedThisMonth: number;
  recentDocuments: DashboardRecentDocument[];
  recentActivities: DashboardRecentActivity[];
  documentTypeStats: DashboardDocumentTypeStat[];
  topFolders: DashboardTopFolder[];
  upcomingImportantDates: number;
  overdueImportantDates: number;
  assignedImportantDates: number;
  upcomingImportantDateItems: DashboardImportantDateItem[];
};

export function getDashboardSummary() {
  return apiRequest<DashboardSummary>("/api/v1/dashboard/summary");
}

export function formatStorageSize(bytes: number) {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  if (bytes < 1024 * 1024 * 1024) {
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}
