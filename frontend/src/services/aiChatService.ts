import type { AiChatRequest, AiChatResponse, AiChatSessionDetail, AiChatSessionSummary } from "../types/aiChat";
import { apiDelete, apiGet, apiPost } from "./apiClient";

const chatFallbackMessage = "AI şu anda cevap veremiyor. Lütfen tekrar deneyin.";

export function sendAiChatMessage(request: AiChatRequest) {
  return apiPost<AiChatResponse>(
    "/api/v1/ai/chat",
    {
      question: request.question,
      sessionId: request.sessionId ?? null,
      documentId: request.documentId ?? null,
      folderId: request.folderId ?? null,
      tagIds: request.tagIds ?? [],
      collectionKey: request.collectionKey ?? null,
    },
    chatFallbackMessage,
  );
}

export function listAiChatSessions(scopeType?: string, scopeId?: string) {
  const params = new URLSearchParams();
  if (scopeType) params.set("scopeType", scopeType);
  if (scopeId) params.set("scopeId", scopeId);
  const query = params.toString();
  return apiGet<AiChatSessionSummary[]>(
    `/api/v1/ai/chat-sessions${query ? `?${query}` : ""}`,
    "Sohbet geçmişi yüklenemedi.",
  );
}

export function getAiChatSession(sessionId: string) {
  return apiGet<AiChatSessionDetail>(`/api/v1/ai/chat-sessions/${sessionId}`, "Sohbet yüklenemedi.");
}

export function deleteAiChatSession(sessionId: string) {
  return apiDelete(`/api/v1/ai/chat-sessions/${sessionId}`, "Sohbet silinemedi.");
}
