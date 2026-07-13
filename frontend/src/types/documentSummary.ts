export type DocumentSummary = {
  documentId: string;
  summary: string;
  keyPoints: string[];
  createdAt: string;
  updatedAt: string;
  model?: string | null;
};
