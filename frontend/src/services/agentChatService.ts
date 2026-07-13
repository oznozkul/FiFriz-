import type { AgentChatRequest, AgentChatResponse } from "../types/agentChat";
import { apiPost } from "./apiClient";

const agentChatFallbackMessage = "Agent sohbeti şu anda tamamlanamıyor. Lütfen tekrar deneyin.";

export function sendAgentChatMessage(request: AgentChatRequest) {
  return apiPost<AgentChatResponse>(
    "/api/v1/ai/agent-chat",
    {
      message: request.message,
      sessionId: request.sessionId ?? null,
      documentId: request.documentId ?? null,
      folderId: request.folderId ?? null,
    },
    agentChatFallbackMessage,
  );
}
