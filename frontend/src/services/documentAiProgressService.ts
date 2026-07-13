import type { DocumentAiStepProgress } from "../types/documentAiProgress";
import { apiGet } from "./apiClient";

export function getDocumentAiProgress(documentId: string) {
  return apiGet<DocumentAiStepProgress[]>(`/api/v1/documents/${documentId}/ai-progress`, "AI ilerlemesi yüklenemedi.");
}
