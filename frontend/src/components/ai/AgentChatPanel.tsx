import { Bot, Loader2, ShieldCheck } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { listAiChatSessions } from "../../services/aiChatService";
import { sendAgentChatMessage } from "../../services/agentChatService";
import type { AiChatSessionSummary } from "../../types/aiChat";
import {
  AGENT_ASSISTANT_SUGGESTIONS,
  DOCUMENT_AGENT_SUGGESTIONS,
  type AgentChatMessage,
  type AgentSuggestedAction,
} from "../../types/agentChat";
import { AiChatHistorySidebar } from "./AiChatHistorySidebar";
import { AiChatInput } from "./AiChatInput";
import { AiMessageBubble } from "./AiMessageBubble";
import { AiSourceList } from "./AiSourceList";
import { AgentSuggestedActionCard } from "./AgentSuggestedActionCard";

type AgentChatPanelProps = {
  documentId?: string;
  folderId?: string;
  compact?: boolean;
  showSessionHistory?: boolean;
  sessionId?: string | null;
  onSessionIdChange?: (sessionId: string | null) => void;
  suggestions?: readonly string[];
  headerTitle?: string;
  headerSubtitle?: string;
  className?: string;
};

function createMessageId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

function agentSessionScope(documentId?: string) {
  if (documentId) {
    return { scopeType: "agent", scopeId: documentId };
  }
  return { scopeType: "workspace", scopeId: undefined as string | undefined };
}

export function AgentChatPanel({
  documentId,
  folderId,
  compact = false,
  showSessionHistory = false,
  sessionId = null,
  onSessionIdChange,
  suggestions = documentId ? DOCUMENT_AGENT_SUGGESTIONS : AGENT_ASSISTANT_SUGGESTIONS,
  headerTitle = "Workspace Agent",
  headerSubtitle = "Belgeleriniz üzerinde aksiyon önerebilir; hiçbir işlem onayınız olmadan çalışmaz.",
  className = "",
}: AgentChatPanelProps) {
  const [messages, setMessages] = useState<AgentChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [sessions, setSessions] = useState<AiChatSessionSummary[]>([]);
  const [isSessionsLoading, setIsSessionsLoading] = useState(false);
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
    if (onSessionIdChange) return;
    setInternalSessionId(sessionId ?? null);
  }, [onSessionIdChange, sessionId]);

  const refreshSessions = useCallback(async () => {
    if (!showSessionHistory) return;
    const scope = agentSessionScope(documentId);
    setIsSessionsLoading(true);
    try {
      const items = await listAiChatSessions(scope.scopeType, scope.scopeId);
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

  const resolveAction = useCallback((messageId: string, suggestionId: string) => {
    setMessages((current) =>
      current.map((message) => {
        if (message.id !== messageId || !message.suggestedActions) return message;
        return {
          ...message,
          suggestedActions: message.suggestedActions.filter((action) => action.suggestionId !== suggestionId),
        };
      }),
    );
  }, []);

  const submitMessage = useCallback(
    async (rawMessage: string) => {
      const message = rawMessage.trim();
      if (!message || isLoading) return;

      const userMessage: AgentChatMessage = {
        id: createMessageId(),
        role: "user",
        text: message,
      };

      setMessages((current) => [...current, userMessage]);
      setDraft("");
      setIsLoading(true);

      try {
        const response = await sendAgentChatMessage({
          message,
          sessionId: activeSessionId,
          documentId: documentId ?? null,
          folderId: folderId ?? null,
        });

        updateSessionId(response.sessionId);

        const assistantMessage: AgentChatMessage = {
          id: createMessageId(),
          role: "assistant",
          text: response.answer,
          sources: response.sources ?? [],
          suggestedActions: response.suggestedActions ?? [],
          question: message,
        };

        setMessages((current) => [...current, assistantMessage]);
        void refreshSessions();
      } catch (error) {
        const assistantMessage: AgentChatMessage = {
          id: createMessageId(),
          role: "assistant",
          text: error instanceof Error ? error.message : "Agent şu anda cevap veremiyor. Lütfen tekrar deneyin.",
          isError: true,
          question: message,
        };
        setMessages((current) => [...current, assistantMessage]);
      } finally {
        setIsLoading(false);
      }
    },
    [activeSessionId, documentId, folderId, isLoading, refreshSessions, updateSessionId],
  );

  const handleNewChat = () => {
    updateSessionId(null);
    setMessages([]);
    setDraft("");
  };

  const showEmptyState = messages.length === 0 && !isLoading;

  const chatContent = (
    <section className={`flex min-h-0 flex-1 flex-col overflow-hidden ${className}`}>
      {!compact ? (
        <div className="border-b border-line bg-white px-5 py-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
              <Bot className="h-5 w-5" />
            </div>
            <div className="min-w-0 flex-1">
              <h2 className="font-bold text-ink">{headerTitle}</h2>
              <p className="text-sm text-slate-500">{headerSubtitle}</p>
            </div>
          </div>
          <div className="mt-3 flex items-start gap-2 rounded-lg border border-violet-200 bg-violet-50/60 px-3 py-2 text-xs text-violet-800">
            <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0" />
            <span>AI bu işlemi senin onayın olmadan yapmaz. Önerilen aksiyonlar yalnızca onayladığınızda çalışır.</span>
          </div>
        </div>
      ) : (
        <div className="border-b border-line bg-violet-50/50 px-3 py-2 text-[11px] text-violet-800">
          <span className="inline-flex items-center gap-1 font-semibold">
            <ShieldCheck className="h-3.5 w-3.5" />
            AI bu işlemi senin onayın olmadan yapmaz.
          </span>
        </div>
      )}

      <div className={`flex-1 space-y-5 overflow-auto ${compact ? "bg-white p-3" : "bg-slate-50/60 p-5"}`}>
        {showEmptyState ? (
          <div className="mx-auto max-w-[760px] rounded-lg border border-dashed border-violet-200 bg-violet-50/50 p-5">
            <p className="text-sm font-semibold text-violet-brand">Agent önerileri</p>
            <p className="mt-1 text-sm text-slate-500">
              {documentId
                ? "Bu belge için etiket, klasör, hatırlatıcı gibi işlemler önerebilirim."
                : "Çalışma alanınız hakkında soru sorun veya aksiyon isteyin."}
            </p>
            <div className="mt-4 flex flex-wrap gap-2">
              {suggestions.map((suggestion) => (
                <button
                  key={suggestion}
                  type="button"
                  className="rounded-full border border-violet-200 bg-white px-3 py-1.5 text-left text-xs font-semibold text-slate-600 transition hover:border-violet-brand hover:bg-violet-soft hover:text-violet-brand disabled:opacity-60"
                  disabled={isLoading}
                  onClick={() => void submitMessage(suggestion)}
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
                <p className="mb-2 text-xs font-semibold text-violet-brand">Kaynaklar</p>
                <AiSourceList sources={message.sources} question={message.question} />
              </div>
            ) : null}
            {message.role === "assistant" && message.suggestedActions && message.suggestedActions.length > 0 ? (
              <div className="mt-3 space-y-2">
                <p className="text-xs font-semibold text-violet-brand">Önerilen aksiyonlar</p>
                {message.suggestedActions.map((action: AgentSuggestedAction) => (
                  <AgentSuggestedActionCard
                    key={action.suggestionId}
                    action={action}
                    onResolved={(suggestionId) => resolveAction(message.id, suggestionId)}
                  />
                ))}
              </div>
            ) : null}
          </div>
        ))}

        {isLoading ? (
          <div className="flex justify-start">
            <div className="inline-flex items-center gap-2 rounded-lg border border-line bg-white px-4 py-3 text-sm text-slate-500 shadow-sm">
              <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
              Agent yanıt hazırlıyor...
            </div>
          </div>
        ) : null}
      </div>

      <div className={`border-t border-line bg-white ${compact ? "p-3" : "p-4"}`}>
        <AiChatInput
          value={draft}
          onChange={setDraft}
          onSubmit={() => void submitMessage(draft)}
          disabled={isLoading}
          compact={compact}
          placeholder={documentId ? "Bu belge için bir işlem isteyin..." : "Workspace agent'a yazın..."}
        />
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
          isDeletingId={null}
          onNewChat={handleNewChat}
          onSelect={(id) => updateSessionId(id)}
          onDelete={() => undefined}
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
        isDeletingId={null}
        onNewChat={handleNewChat}
        onSelect={(id) => updateSessionId(id)}
        onDelete={() => undefined}
      />
      {chatContent}
    </div>
  );
}
