import type { TagBrief } from "./tag";

export type HybridSearchMatchType = "full_text" | "semantic" | "metadata" | "entity" | "hybrid";

export type HybridSearchMatchedEntity = {
  entityType: string;
  value: string;
  label?: string | null;
};

export type HybridSearchResult = {
  documentId: string;
  originalFileName: string;
  title: string;
  fileExtension: string;
  folderName: string;
  tags: TagBrief[];
  sourceType: string;
  aiDocumentType?: string | null;
  matchedSnippet: string;
  matchedChunks: string[];
  matchedEntities: HybridSearchMatchedEntity[];
  score: number;
  matchType: HybridSearchMatchType;
};

export const HYBRID_MATCH_TYPE_LABELS: Record<HybridSearchMatchType, string> = {
  full_text: "Kelime eşleşmesi",
  semantic: "Anlam bazlı",
  metadata: "Metadata",
  entity: "Varlık",
  hybrid: "Karma eşleşme",
};

export type SearchMode = "quick" | "smart";
