import { PageTitle } from "../components/common";
import { useState } from "react";
import { AgentChatPanel, AiChatPanel, AiModeToggle, type AiChatMode } from "../components/ai";

export function AiAssistantPage() {
  const [mode, setMode] = useState<AiChatMode>("rag");
  const [ragSessionId, setRagSessionId] = useState<string | null>(null);
  const [agentSessionId, setAgentSessionId] = useState<string | null>(null);

  return (
    <div className="mx-auto flex min-h-[calc(100vh-8rem)] max-w-7xl flex-col">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <PageTitle
          title="AI Asistan"
          subtitle={
            mode === "rag"
              ? "Belgelerinizle ilgili sorular sorun, geçmiş sohbetlerinize dönün ve kaynaklı yanıtları birlikte görün."
              : "Workspace agent ile konuşun; aksiyon önerileri yalnızca onayınızla uygulanır."
          }
        />
        <AiModeToggle mode={mode} onChange={setMode} />
      </div>

      <section className="panel flex min-h-[620px] flex-1 flex-col overflow-hidden">
        {mode === "rag" ? (
          <AiChatPanel
            showFilters
            showSessionHistory
            sessionId={ragSessionId}
            onSessionIdChange={setRagSessionId}
            headerTitle="Belge sohbeti"
            headerSubtitle="Yanıtlar yalnızca erişebildiğiniz belgelerdeki içeriklere dayanır."
          />
        ) : (
          <AgentChatPanel
            showSessionHistory
            sessionId={agentSessionId}
            onSessionIdChange={setAgentSessionId}
            headerTitle="Workspace Agent"
            headerSubtitle="Etiket, klasör, hatırlatıcı ve daha fazlası için öneri alın."
          />
        )}
      </section>
    </div>
  );
}
