import type { DocumentSummary } from "../types/documentSummary";
import { apiBaseUrl, getAccessToken } from "./apiClient";
import { apiPost } from "./apiClient";

const summaryFallbackMessage = "Özet oluşturulamadı. Lütfen tekrar deneyin.";

export async function getDocumentSummary(documentId: string): Promise<DocumentSummary | null> {
  const headers = new Headers();
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseUrl}/api/v1/documents/${documentId}/summary`, { headers });
  if (response.status === 401) {
    throw new Error("Oturum süresi doldu. Lütfen tekrar giriş yapın.");
  }

  const payload = (await response.json().catch(() => null)) as {
    success?: boolean;
    data?: DocumentSummary | null;
    error?: { message?: string };
  } | null;

  if (!response.ok || payload?.success === false) {
    throw new Error(payload?.error?.message ?? "Belge özeti yüklenemedi.");
  }

  return payload?.data ?? null;
}

export function generateDocumentSummary(documentId: string) {
  return apiPost<DocumentSummary>(`/api/v1/documents/${documentId}/summary/generate`, undefined, summaryFallbackMessage);
}
