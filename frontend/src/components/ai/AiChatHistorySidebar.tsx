import { MessageSquarePlus, Trash2 } from "lucide-react";
import type { AiChatSessionSummary } from "../../types/aiChat";

type AiChatHistorySidebarProps = {
  sessions: AiChatSessionSummary[];
  activeSessionId: string | null;
  isLoading: boolean;
  isDeletingId?: string | null;
  onNewChat: () => void;
  onSelect: (sessionId: string) => void;
  onDelete: (sessionId: string) => void;
  compact?: boolean;
};

function formatSessionTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function AiChatHistorySidebar({
  sessions,
  activeSessionId,
  isLoading,
  isDeletingId,
  onNewChat,
  onSelect,
  onDelete,
  compact = false,
}: AiChatHistorySidebarProps) {
  return (
    <aside className={`flex shrink-0 flex-col border-line bg-white ${compact ? "w-full border-b" : "w-72 border-r"}`}>
      <div className={`border-b border-line ${compact ? "px-3 py-2" : "p-4"}`}>
        <button
          type="button"
          className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-violet-brand bg-violet-soft px-3 text-sm font-semibold text-violet-brand transition hover:bg-violet-brand hover:text-white focus:outline-none focus:ring-4 focus:ring-violet-ring/70"
          onClick={onNewChat}
        >
          <MessageSquarePlus className="h-4 w-4" />
          Yeni Sohbet
        </button>
      </div>

      <div className={`min-h-0 flex-1 overflow-y-auto ${compact ? "max-h-40 p-2" : "p-3"}`}>
        {isLoading ? (
          <p className="px-2 py-3 text-sm text-slate-500">Sohbetler yükleniyor...</p>
        ) : sessions.length === 0 ? (
          <p className="px-2 py-3 text-sm text-slate-500">Henüz AI sohbetin yok.</p>
        ) : (
          <ul className="space-y-1">
            {sessions.map((session) => {
              const isActive = session.id === activeSessionId;
              const isDeleting = isDeletingId === session.id;

              return (
                <li key={session.id}>
                  <div
                    className={`group flex items-stretch gap-1 rounded-lg border transition ${
                      isActive ? "border-violet-brand bg-violet-soft/70" : "border-transparent hover:border-line hover:bg-slate-50"
                    }`}
                  >
                    <button
                      type="button"
                      className="min-w-0 flex-1 px-3 py-2.5 text-left"
                      onClick={() => onSelect(session.id)}
                    >
                      <p className="truncate text-sm font-semibold text-ink">{session.title}</p>
                      <p className="mt-0.5 text-xs text-slate-400">{formatSessionTime(session.updatedAt)}</p>
                    </button>
                    <button
                      type="button"
                      className="mr-1 inline-flex w-9 shrink-0 items-center justify-center rounded-lg text-slate-400 opacity-0 transition hover:bg-rose-50 hover:text-rose-600 group-hover:opacity-100 focus:opacity-100 disabled:opacity-50"
                      aria-label="Sohbeti sil"
                      disabled={isDeleting}
                      onClick={() => onDelete(session.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </aside>
  );
}
