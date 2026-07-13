import { apiDelete, apiGet, apiPost, apiPut } from "./apiClient";

export type InternalSharePermission = "view" | "download" | "edit" | "manage";

export type DocumentInternalShare = {
  id: string;
  documentId: string;
  sharedWithUserId: string;
  sharedWithFullName: string;
  sharedWithEmail: string;
  sharedByUserId: string;
  sharedByFullName: string;
  permission: InternalSharePermission;
  createdAt: string;
  updatedAt: string;
};

export type SharedWithMeItem = {
  shareId: string;
  documentId: string;
  documentName: string;
  sharedByUserId: string;
  sharedByFullName: string;
  permission: InternalSharePermission;
  createdAt: string;
};

export type CreateInternalShareInput = {
  sharedWithUserId: string;
  permission: InternalSharePermission;
};

export function listInternalShares(documentId: string) {
  return apiGet<DocumentInternalShare[]>(`/api/v1/documents/${documentId}/internal-shares`, "Kişi paylaşımları yüklenemedi.");
}

export function createInternalShare(documentId: string, input: CreateInternalShareInput) {
  return apiPost<DocumentInternalShare>(`/api/v1/documents/${documentId}/internal-shares`, input, "Belge paylaşılamadı.");
}

export function updateInternalShare(documentId: string, shareId: string, permission: InternalSharePermission) {
  return apiPut<DocumentInternalShare>(
    `/api/v1/documents/${documentId}/internal-shares/${shareId}`,
    { permission },
    "Paylaşım yetkisi güncellenemedi.",
  );
}

export function deleteInternalShare(documentId: string, shareId: string) {
  return apiDelete(`/api/v1/documents/${documentId}/internal-shares/${shareId}`, "Paylaşım kaldırılamadı.");
}

export function listSharedWithMe() {
  return apiGet<SharedWithMeItem[]>("/api/v1/documents/shared-with-me", "Paylaşılan belgeler yüklenemedi.");
}

export const internalSharePermissionLabels: Record<InternalSharePermission, string> = {
  view: "Görüntüle",
  download: "İndir",
  edit: "Düzenle",
  manage: "Paylaşımı Yönet",
};

export const internalSharePermissionDescriptions: Record<InternalSharePermission, string> = {
  view: "Belgeyi görüntüleyebilir, önizleme yapabilir.",
  download: "Belgeyi görüntüleyebilir ve indirebilir.",
  edit: "Belgeyi görüntüleyebilir, indirebilir ve düzenleyebilir.",
  manage: "Belgeyi görüntüleyebilir, indirebilir, düzenleyebilir ve başka kullanıcılarla paylaşabilir.",
};
