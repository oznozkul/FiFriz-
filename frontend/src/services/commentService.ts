import { apiRequest } from "./apiClient";

export type DocumentComment = {
  id: string;
  documentId: string;
  userId: string;
  userFullName: string;
  parentId: string | null;
  content: string;
  isEdited: boolean;
  createdAt: string;
  updatedAt: string;
  replies: DocumentComment[];
};

export function listDocumentComments(documentId: string) {
  return apiRequest<DocumentComment[]>(`/api/v1/documents/${documentId}/comments`, undefined, "Yorumlar yüklenemedi.");
}

export function createDocumentComment(documentId: string, content: string, parentId?: string) {
  return apiRequest<DocumentComment>(
    `/api/v1/documents/${documentId}/comments`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        content,
        parentId: parentId ?? null,
      }),
    },
    "Yorum eklenemedi.",
  );
}

export function updateDocumentComment(documentId: string, commentId: string, content: string) {
  return apiRequest<DocumentComment>(
    `/api/v1/documents/${documentId}/comments/${commentId}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content }),
    },
    "Yorum güncellenemedi.",
  );
}

export function deleteDocumentComment(documentId: string, commentId: string) {
  return apiRequest<void>(
    `/api/v1/documents/${documentId}/comments/${commentId}`,
    { method: "DELETE" },
    "Yorum silinemedi.",
  );
}
