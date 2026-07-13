import { apiDelete, apiGet, apiPost } from "./apiClient";

export const relationshipTypes = [
  "related",
  "attachment",
  "parent",
  "child",
  "replacement",
  "reference",
  "invoice_of",
  "contract_of",
  "identity_of",
  "custom",
] as const;

export type RelationshipType = (typeof relationshipTypes)[number];

export type RelatedDocumentNode = {
  id: string;
  originalFileName: string;
  fileExtension: string;
  mimeType: string;
  folderId: string | null;
  folderName: string;
};

export type DocumentRelationshipEdge = {
  id: string;
  relationshipType: RelationshipType;
  direction: "outgoing" | "incoming";
  sourceDocumentId: string;
  targetDocumentId: string;
  relatedDocument: RelatedDocumentNode;
  createdByUserId: string;
  createdAt: string;
};

export type DocumentRelationshipGraphSummary = {
  totalCount: number;
  outgoingCount: number;
  incomingCount: number;
  byType: Record<string, number>;
};

export type DocumentRelationshipsResponse = {
  documentId: string;
  items: DocumentRelationshipEdge[];
  summary: DocumentRelationshipGraphSummary;
};

export type CreateDocumentRelationshipInput = {
  targetDocumentId: string;
  relationshipType: RelationshipType;
};

export const relationshipTypeLabels: Record<RelationshipType, string> = {
  related: "İlişkili",
  attachment: "Ek",
  parent: "Üst belge",
  child: "Alt belge",
  replacement: "Yerine geçen",
  reference: "Referans",
  invoice_of: "Faturası",
  contract_of: "Sözleşmesi",
  identity_of: "Kimliği",
  custom: "Özel",
};

export function listDocumentRelationships(documentId: string) {
  return apiGet<DocumentRelationshipsResponse>(`/api/v1/documents/${documentId}/relationships`, "İlişkili belgeler yüklenemedi.");
}

export function createDocumentRelationship(documentId: string, input: CreateDocumentRelationshipInput) {
  return apiPost<DocumentRelationshipEdge>(`/api/v1/documents/${documentId}/relationships`, input, "Belge ilişkisi oluşturulamadı.");
}

export function deleteDocumentRelationship(relationshipId: string) {
  return apiDelete(`/api/v1/document-relationships/${relationshipId}`, "Belge ilişkisi kaldırılamadı.");
}
