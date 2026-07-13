import { Loader2, Sparkles } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { deleteAiChatSession, getAiChatSession, listAiChatSessions, sendAiChatMessage } from "../../services/aiChatService";
import {
  AI_ASSISTANT_SUGGESTIONS,
  DOCUMENT_AI_SUGGESTIONS,
  mapSessionMessages,
  type AiChatFilterState,
  type AiChatMessage,
  type AiChatSessionSummary,
} from "../../types/aiChat";
import { AiChatFilters } from "./AiChatFilters";
import { AiChatHistorySidebar } from "./AiChatHistorySidebar";
import { AiChatInput } from "./AiChatInput";
import { AiMessageBubble } from "./AiMessageBubble";
import { AiSourceList } from "./AiSourceList";

type AiChatPanelProps = {
  documentId?: string;
  showFilters?: boolean;
  compact?: boolean;
  showSessionHistory?: boolean;
  sessionId?: string | null;
  onSessionIdChange?: (sessionId: string | null) => void;
  suggestions?: readonly string[];
  headerTitle?: string;
  headerSubtitle?: string;
  className?: string;
  onSummarizeDocument?: () => void;
  isSummarizingDocument?: boolean;
};

const defaultFilters: AiChatFilterState = {
  documentId: "",
  folderId: "",
  tagIds: [],
};

function createMessageId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

export function AiChatPanel({
  documentId,
  showFilters = false,
  compact = false,
  showSessionHistory = false,
  sessionId = null,
  onSessionIdChange,
  suggestions = documentId ? DOCUMENT_AI_SUGGESTIONS : AI_ASSISTANT_SUGGESTIONS,
  headerTitle = "Belge sohbeti",
  headerSubtitle = "Belgelerinizdeki içeriklere dayalı yanıtlar alın.",
  className = "",
  onSummarizeDocument,
  isSummarizingDocument = false,
}: AiChatPanelProps) {
  const [messages, setMessages] = useState<AiChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSessionLoading, setIsSessionLoading] = useState(false);
  const [filters, setFilters] = useState<AiChatFilterState>(defaultFilters);
  const [sessions, setSessions] = useState<AiChatSessionSummary[]>([]);
  const [isSessionsLoading, setIsSessionsLoading] = useState(false);
  const [deletingSessionId, setDeletingSessionId] = useState<string | null>(null);
  const [internalSessionId, setInternalSessionId] = useState<string | null>(sessionId ?? null);

  const activeSessionId = onSessionIdChange ? sessionId ?? null : internalSessionId;

  const updateSessionId = useCallback(
    (nextSessionId: string | null) => {
      if (onSessionIdChange) {
        onSessionIdChange(nextSessionId);
      } else {
        setInternalSessionId(nextSessionId);
      }
    },
    [onSessionIdChange],
  );

  useEffect(() => {
    if (onSessionIdChange) {
      return;
    }
    setInternalSessionId(sessionId ?? null);
  }, [onSessionIdChange, sessionId]);

  const refreshSessions = useCallback(async () => {
    if (!showSessionHistory) return;

    setIsSessionsLoading(true);
    try {
      const items = await listAiChatSessions(documentId ? "document" : undefined, documentId);
      setSessions(items);
    } catch {
      setSessions([]);
    } finally {
      setIsSessionsLoading(false);
    }
  }, [documentId, showSessionHistory]);

  useEffect(() => {
    void refreshSessions();
  }, [refreshSessions]);

  useEffect(() => {
    if (!activeSessionId) {
      setMessages([]);
      setDraft("");
      return;
    }

    let isMounted = true;
    const loadSession = async () => {
      setIsSessionLoading(true);
      try {
        const session = await getAiChatSession(activeSessionId);
        if (!isMounted) return;
        setMessages(mapSessionMessages(session.messages));
      } catch {
        if (isMounted) {
          setMessages([]);
        }
      } finally {
        if (isMounted) {
          setIsSessionLoading(false);
        }
      }
    };

    void loadSession();
    return () => {
      isMounted = false;
    };
  }, [activeSessionId]);

  const submitQuestion = useCallback(
    async (rawQuestion: string) => {
      const question = rawQuestion.trim();
      if (!question || isLoading) {
        return;
      }

      const userMessage: AiChatMessage = {
        id: createMessageId(),
        role: "user",
        text: question,
      };

      setMessages((current) => [...current, userMessage]);
      setDraft("");
      setIsLoading(true);

      try {
        const response = await sendAiChatMessage({
          question,
          sessionId: activeSessionId,
          documentId: documentId ?? (filters.documentId || null),
          folderId: filters.folderId || null,
          tagIds: filters.tagIds,
          collectionKey: null,
        });

        updateSessionId(response.sessionId);

        const assistantMessage: AiChatMessage = {
          id: createMessageId(),
          role: "assistant",
          text: response.answer,
          sources: response.sources ?? [],
          question,
        };

        setMessages((current) => [...current, assistantMessage]);
        void refreshSessions();
      } catch (error) {
        const assistantMessage: AiChatMessage = {
          id: createMessageId(),
          role: "assistant",
          text: error instanceof Error ? error.message : "AI şu anda cevap veremiyor. Lütfen tekrar deneyin.",
          isError: true,
          question,
        };
        setMessages((current) => [...current, assistantMessage]);
      } finally {
        setIsLoading(false);
      }
    },
    [activeSessionId, documentId, filters, isLoading, refreshSessions, updateSessionId],
  );

  const handleSubmit = () => {
    void submitQuestion(draft);
  };

  const handleNewChat = () => {
    updateSessionId(null);
    setMessages([]);
    setDraft("");
    setFilters(defaultFilters);
  };

  const handleDeleteSession = async (targetSessionId: string) => {
    setDeletingSessionId(targetSessionId);
    try {
      await deleteAiChatSession(targetSessionId);
      if (activeSessionId === targetSessionId) {
        handleNewChat();
      }
      await refreshSessions();
    } catch {
      // keep list unchanged on failure
    } finally {
      setDeletingSessionId(null);
    }
  };

  const showEmptyState = messages.length === 0 && !isLoading && !isSessionLoading;

  const chatContent = (
    <section className={`flex min-h-0 flex-1 flex-col overflow-hidden ${className}`}>
      {!compact ? (
        <div className="border-b border-line bg-white px-5 py-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
              <Sparkles className="h-5 w-5" />
            </div>
            <div>
              <h2 className="font-bold text-ink">{headerTitle}</h2>
              <p className="text-sm text-slate-500">{headerSubtitle}</p>
            </div>
          </div>
        </div>
      ) : null}

      {showFilters && !documentId ? <AiChatFilters value={filters} onChange={setFilters} disabled={isLoading || isSessionLoading} /> : null}

      <div className={`flex-1 space-y-5 overflow-auto ${compact ? "bg-white p-3" : "bg-slate-50/60 p-5"}`}>
        {isSessionLoading ? (
          <div className="flex justify-center py-8">
            <div className="inline-flex items-center gap-2 text-sm text-slate-500">
              <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
              Sohbet yükleniyor...
            </div>
          </div>
        ) : null}

        {showEmptyState ? (
          <div className="mx-auto max-w-[760px] rounded-lg border border-dashed border-violet-200 bg-violet-50/50 p-5">
            <p className="text-sm font-semibold text-violet-brand">Önerilen sorular</p>
            <p className="mt-1 text-sm text-slate-500">
              {documentId ? "Bu belge hakkında soru sorabilir veya aşağıdaki önerilerden birini seçebilirsiniz." : "Başlamak için bir soru yazın veya önerilerden birini seçin."}
            </p>
            {documentId && onSummarizeDocument ? (
              <button
                type="button"
                className="mt-4 inline-flex h-10 items-center gap-2 rounded-lg border border-violet-brand bg-white px-4 text-sm font-semibold text-violet-brand transition hover:bg-violet-soft disabled:opacity-60"
                disabled={isLoading || isSessionLoading || isSummarizingDocument}
                onClick={onSummarizeDocument}
              >
                {isSummarizingDocument ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                Bu belgeyi özetle
              </button>
            ) : null}
            <div className="mt-4 flex flex-wrap gap-2">
              {suggestions.map((suggestion) => (
                <button
                  key={suggestion}
                  type="button"
                  className="rounded-full border border-violet-200 bg-white px-3 py-1.5 text-left text-xs font-semibold text-slate-600 transition hover:border-violet-brand hover:bg-violet-soft hover:text-violet-brand disabled:opacity-60"
                  disabled={isLoading || isSessionLoading}
                  onClick={() => void submitQuestion(suggestion)}
                >
                  {suggestion}
                </button>
              ))}
            </div>
          </div>
        ) : null}

        {messages.map((message) => (
          <div key={message.id} className={message.role === "user" ? "" : "max-w-[760px]"}>
            <AiMessageBubble
              role={message.role}
              text={message.text}
              maxWidthClassName={message.role === "user" ? "max-w-[760px]" : "w-full max-w-none"}
              variant={message.isError ? "error" : "default"}
            />
            {message.role === "assistant" && message.sources && message.sources.length > 0 ? (
              <div className="mt-1">
                <p className="mb-2 text-xs font-semibold text-violet-brand">Kaynaklı yanıt</p>
                <AiSourceList sources={message.sources} question={message.question} />
              </div>
            ) : null}
          </div>
        ))}

        {isLoading ? (
          <div className="flex justify-start">
            <div className="inline-flex items-center gap-2 rounded-lg border border-line bg-white px-4 py-3 text-sm text-slate-500 shadow-sm">
              <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
              Yanıt hazırlanıyor...
            </div>
          </div>
        ) : null}
      </div>

      <div className={`border-t border-line bg-white ${compact ? "p-3" : "p-4"}`}>
        {compact && documentId && onSummarizeDocument ? (
          <button
            type="button"
            className="mb-3 inline-flex h-9 items-center gap-2 rounded-lg border border-violet-brand bg-violet-soft px-3 text-xs font-semibold text-violet-brand transition hover:bg-violet-brand hover:text-white disabled:opacity-60"
            disabled={isLoading || isSessionLoading || isSummarizingDocument}
            onClick={onSummarizeDocument}
          >
            {isSummarizingDocument ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Sparkles className="h-3.5 w-3.5" />}
            Bu belgeyi özetle
          </button>
        ) : null}
        <AiChatInput
          value={draft}
          onChange={setDraft}
          onSubmit={handleSubmit}
          disabled={isLoading || isSessionLoading}
          compact={compact}
          placeholder={documentId ? "Bu belge hakkında sorun..." : "Bir şey sorun..."}
        />
        {!compact ? (
          <p className="mt-2 text-xs text-slate-400">Enter ile gönderin, Shift+Enter ile yeni satır ekleyin.</p>
        ) : null}
      </div>
    </section>
  );

  if (!showSessionHistory) {
    return chatContent;
  }

  if (compact) {
    return (
      <div className="flex min-h-[480px] flex-col overflow-hidden rounded-lg border border-line">
        <AiChatHistorySidebar
          compact
          sessions={sessions}
          activeSessionId={activeSessionId}
          isLoading={isSessionsLoading}
          isDeletingId={deletingSessionId}
          onNewChat={handleNewChat}
          onSelect={(id) => updateSessionId(id)}
          onDelete={(id) => void handleDeleteSession(id)}
        />
        {chatContent}
      </div>
    );
  }

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden">
      <AiChatHistorySidebar
        sessions={sessions}
        activeSessionId={activeSessionId}
        isLoading={isSessionsLoading}
        isDeletingId={deletingSessionId}
        onNewChat={handleNewChat}
        onSelect={(id) => updateSessionId(id)}
        onDelete={(id) => void handleDeleteSession(id)}
      />
      {chatContent}
    </div>
  );
}
