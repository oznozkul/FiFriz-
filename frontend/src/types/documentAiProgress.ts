export type DocumentAiStepProgress = {
  stepName: string;
  status: "pending" | "running" | "completed" | "failed" | "retrying" | "skipped";
  startedAt?: string | null;
  completedAt?: string | null;
  error?: string | null;
  retryCount: number;
};

export const AI_PIPELINE_STEP_LABELS: Record<string, string> = {
  extract_text: "Text",
  chunk_document: "Chunk",
  generate_embedding: "Embedding",
  generate_summary: "Summary",
  classify_document: "Belge Türü",
  extract_metadata: "Metadata",
  extract_dates: "Dates",
  generate_tags: "Tags",
  extract_entities: "Varlıklar",
  link_related_documents: "İlişkiler",
  completed: "Completed",
};
