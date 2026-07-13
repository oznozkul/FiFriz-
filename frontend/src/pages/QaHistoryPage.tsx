import { MessageSquareText, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { EmptyState, IconButton, LoadingState, PageTitle, PrimaryButton } from "../components/common";
import { listAiChatSessions } from "../services/aiChatService";
import { toast } from "../services/toast";
import type { AiChatSessionSummary } from "../types/aiChat";

export function QaHistoryPage() {
  const navigate = useNavigate();
  const [sessions, setSessions] = useState<AiChatSessionSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [query, setQuery] = useState("");

  const fetchSessions = useCallback(async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setSessions(await listAiChatSessions());
    } catch (error) {
      const message = error instanceof Error ? error.message : "Sohbet geçmişi yüklenemedi.";
      setLoadError(message);
      setSessions([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchSessions();
  }, [fetchSessions]);

  const filteredSessions = useMemo(() => {
    const trimmed = query.trim().toLowerCase();
    if (!trimmed) return sessions;
    return sessions.filter((session) => session.title.toLowerCase().includes(trimmed));
  }, [query, sessions]);

  return (
    <div className="mx-auto max-w-4xl space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title="Sohbet Geçmişi" subtitle="AI asistan ile yaptığınız sohbetlere dönün ve devam edin." />
        <IconButton aria-label="Yenile" onClick={() => void fetchSessions()}>
          <RefreshCw className="h-4 w-4" />
        </IconButton>
      </div>

      <input
        className="h-10 w-full max-w-md rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand"
        placeholder="Sohbetlerde ara..."
        value={query}
        onChange={(event) => setQuery(event.target.value)}
      />

      {isLoading ? (
        <LoadingState label="Sohbet geçmişi yükleniyor..." />
      ) : loadError ? (
        <EmptyState
          title="Sohbet geçmişi yüklenemedi"
          description={loadError}
          action={
            <PrimaryButton onClick={() => void fetchSessions()}>Tekrar dene</PrimaryButton>
          }
        />
      ) : filteredSessions.length === 0 ? (
        <EmptyState
          title={query.trim() ? "Sonuç bulunamadı" : "Henüz sohbet geçmişi yok"}
          description={
            query.trim()
              ? "Farklı bir arama terimi deneyin."
              : "AI asistan ile soru sorduğunuzda sohbetleriniz burada listelenir."
          }
          action={
            !query.trim() ? (
              <PrimaryButton onClick={() => navigate("/ai")}>AI Asistana Git</PrimaryButton>
            ) : undefined
          }
        />
      ) : (
        <section className="space-y-3">
          {filteredSessions.map((session) => (
            <QaHistoryRow key={session.id} session={session} onOpen={() => navigate(`/ai?session=${session.id}`)} />
          ))}
        </section>
      )}
    </div>
  );
}

type QaHistoryRowProps = {
  session: AiChatSessionSummary;
  onOpen: () => void;
};

function QaHistoryRow({ session, onOpen }: QaHistoryRowProps) {
  return (
    <button
      className="panel flex w-full flex-col gap-3 p-4 text-left transition hover:border-violet-ring hover:bg-violet-soft/40 focus:outline-none focus:ring-4 focus:ring-violet-ring/70 sm:flex-row sm:items-center"
      type="button"
      onClick={onOpen}
    >
      <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
        <MessageSquareText className="h-4 w-4" />
      </span>
      <span className="min-w-0 flex-1 font-semibold text-ink sm:truncate">{session.title}</span>
      <span className="shrink-0 text-xs font-medium text-slate-500">{formatDateTime(session.updatedAt)}</span>
    </button>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
