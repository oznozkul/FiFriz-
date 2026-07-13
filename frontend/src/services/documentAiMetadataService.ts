import type { DocumentAIMetadata } from "../types/documentAiMetadata";
import { apiBaseUrl, apiPost, getAccessToken } from "./apiClient";

const metadataFallbackMessage = "Metadata oluşturulamadı. Lütfen tekrar deneyin.";

export async function getDocumentAIMetadata(documentId: string): Promise<DocumentAIMetadata | null> {
  const headers = new Headers();
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseUrl}/api/v1/documents/${documentId}/ai-metadata`, { headers });
  if (response.status === 401) {
    throw new Error("Oturum süresi doldu. Lütfen tekrar giriş yapın.");
  }

  const payload = (await response.json().catch(() => null)) as {
    success?: boolean;
    data?: DocumentAIMetadata | null;
    error?: { message?: string };
  } | null;

  if (!response.ok || payload?.success === false) {
    throw new Error(payload?.error?.message ?? "AI metadata yüklenemedi.");
  }

  return payload?.data ?? null;
}

export function generateDocumentAIMetadata(documentId: string) {
  return apiPost<DocumentAIMetadata>(`/api/v1/documents/${documentId}/ai-metadata/generate`, undefined, metadataFallbackMessage);
}
