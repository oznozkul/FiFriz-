export type AiChatSource = {
  documentId: string;
  documentTitle: string;
  chunkId: string;
  chunkIndex: number;
  similarity: number;
  contentPreview: string;
};

export type AiChatResponse = {
  sessionId: string;
  answer: string;
  sources: AiChatSource[];
};

export type AiChatRequest = {
  question: string;
  sessionId?: string | null;
  documentId?: string | null;
  folderId?: string | null;
  tagIds?: string[];
  collectionKey?: string | null;
};

export type AiChatMessage = {
  id: string;
  role: "user" | "assistant";
  text: string;
  sources?: AiChatSource[];
  question?: string;
  isError?: boolean;
};

export type AiChatFilterState = {
  documentId: string;
  folderId: string;
  tagIds: string[];
};

export type AiChatSessionSummary = {
  id: string;
  title: string;
  scopeType?: string | null;
  scopeId?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type AiChatSessionMessage = {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  sources?: AiChatSource[];
  createdAt: string;
};

export type AiChatSessionDetail = AiChatSessionSummary & {
  messages: AiChatSessionMessage[];
};

export const AI_ASSISTANT_SUGGESTIONS = [
  "Belgelerimde bu ay yüklenenleri özetle",
  "Kira sözleşmemde önemli tarih var mı?",
  "Bu klasördeki belgelerde ödeme bilgisi var mı?",
  "Benimle paylaşılan belgelerde önemli noktaları bul",
] as const;

export const DOCUMENT_AI_SUGGESTIONS = [
  "Bu belgedeki ana konuları özetle",
  "Bu belgede ödeme bilgisi var mı?",
  "Bu belgede önemli tarihler neler?",
] as const;

export function mapSessionMessages(messages: AiChatSessionMessage[]): AiChatMessage[] {
  let lastUserQuestion = "";

  return messages
    .filter((message) => message.role === "user" || message.role === "assistant")
    .map((message) => {
      if (message.role === "user") {
        lastUserQuestion = message.content;
        return {
          id: message.id,
          role: "user" as const,
          text: message.content,
        };
      }

      return {
        id: message.id,
        role: "assistant" as const,
        text: message.content,
        sources: message.sources ?? [],
        question: lastUserQuestion,
      };
    });
}
