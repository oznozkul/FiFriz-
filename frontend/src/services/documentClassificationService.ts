import type { DocumentClassification } from "../types/documentClassification";
import { apiBaseUrl, apiPost, getAccessToken } from "./apiClient";

const classificationFallbackMessage = "Belge sınıflandırılamadı. Lütfen tekrar deneyin.";

export async function getDocumentClassification(documentId: string): Promise<DocumentClassification | null> {
  const headers = new Headers();
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseUrl}/api/v1/documents/${documentId}/classification`, { headers });
  if (response.status === 401) {
    throw new Error("Oturum süresi doldu. Lütfen tekrar giriş yapın.");
  }

  const payload = (await response.json().catch(() => null)) as {
    success?: boolean;
    data?: DocumentClassification | null;
    error?: { message?: string };
  } | null;

  if (!response.ok || payload?.success === false) {
    throw new Error(payload?.error?.message ?? "Belge sınıflandırması yüklenemedi.");
  }

  return payload?.data ?? null;
}

export function generateDocumentClassification(documentId: string) {
  return apiPost<DocumentClassification>(`/api/v1/documents/${documentId}/classification/generate`, undefined, classificationFallbackMessage);
}
