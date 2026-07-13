import type { TagBrief } from "./tag";

export type DocumentSuggestedTag = {
  id: string;
  documentId: string;
  tagName: string;
  confidence?: number | null;
  status: "pending" | "accepted" | "rejected";
  generatedBy: string;
  createdAt: string;
  updatedAt: string;
};

export type AcceptSuggestedTagsResult = {
  accepted: DocumentSuggestedTag[];
  tags: TagBrief[];
};
