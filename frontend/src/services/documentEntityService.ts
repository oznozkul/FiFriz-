import type { DocumentEntity } from "../types/documentEntity";
import { apiBaseUrl, apiPost, getAccessToken } from "./apiClient";

const entitiesFallbackMessage = "Varlıklar çıkarılamadı. Lütfen tekrar deneyin.";

export async function getDocumentEntities(documentId: string): Promise<DocumentEntity[]> {
  const headers = new Headers();
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseUrl}/api/v1/documents/${documentId}/entities`, { headers });
  if (response.status === 401) {
    throw new Error("Oturum süresi doldu. Lütfen tekrar giriş yapın.");
  }

  const payload = (await response.json().catch(() => null)) as {
    success?: boolean;
    data?: DocumentEntity[];
    error?: { message?: string };
  } | null;

  if (!response.ok || payload?.success === false) {
    throw new Error(payload?.error?.message ?? "AI varlıklar yüklenemedi.");
  }

  return payload?.data ?? [];
}

export function extractDocumentEntities(documentId: string) {
  return apiPost<DocumentEntity[]>(`/api/v1/documents/${documentId}/entities/extract-ai`, undefined, entitiesFallbackMessage);
}

export async function deleteDocumentEntity(documentId: string, entityId: string) {
  const headers = new Headers();
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseUrl}/api/v1/documents/${documentId}/entities/${entityId}`, {
    method: "DELETE",
    headers,
  });

  if (response.status === 401) {
    throw new Error("Oturum süresi doldu. Lütfen tekrar giriş yapın.");
  }

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? "Varlık silinemedi.");
  }
}
