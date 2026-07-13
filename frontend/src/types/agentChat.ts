import type { AiChatSource } from "./aiChat";

export type AgentSuggestedAction = {
  suggestionId: string;
  title: string;
  description?: string | null;
  toolName: string;
  actionPayload: Record<string, unknown>;
  requiresConfirmation: boolean;
};

export type AgentChatResponse = {
  sessionId: string;
  answer: string;
  sources?: AiChatSource[];
  suggestedActions?: AgentSuggestedAction[];
};

export type AgentChatRequest = {
  message: string;
  sessionId?: string | null;
  documentId?: string | null;
  folderId?: string | null;
};

export type AgentChatMessage = {
  id: string;
  role: "user" | "assistant";
  text: string;
  sources?: AiChatSource[];
  suggestedActions?: AgentSuggestedAction[];
  question?: string;
  isError?: boolean;
};

export const AGENT_ASSISTANT_SUGGESTIONS = [
  "Onay bekleyen belgeleri listele",
  "Benimle paylaşılan önemli belgeleri göster",
  "Bu ay yüklenen belgeleri özetle",
  "Hukuki belgeler klasöründeki dosyaları listele",
] as const;

export const DOCUMENT_AGENT_SUGGESTIONS = [
  "Bu belgeye sözleşme etiketi ekle",
  "Bu belgeyi Hukuki Belgeler klasörüne taşı",
  "Bu belge için 30 gün önce hatırlatıcı oluştur",
  "Bu belgeyi özetle",
] as const;
