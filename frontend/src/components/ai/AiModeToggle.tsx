export type AiChatMode = "rag" | "agent";

type AiModeToggleProps = {
  mode: AiChatMode;
  onChange: (mode: AiChatMode) => void;
  disabled?: boolean;
  compact?: boolean;
};

export function AiModeToggle({ mode, onChange, disabled = false, compact = false }: AiModeToggleProps) {
  return (
    <div
      className={`inline-flex rounded-lg border border-line bg-slate-50 p-0.5 ${compact ? "text-xs" : "text-sm"}`}
      role="group"
      aria-label="AI modu"
    >
      <button
        type="button"
        className={`rounded-md px-3 py-1.5 font-semibold transition disabled:opacity-60 ${
          mode === "rag" ? "bg-white text-violet-brand shadow-sm" : "text-slate-500 hover:text-ink"
        }`}
        disabled={disabled}
        onClick={() => onChange("rag")}
      >
        {compact ? "RAG" : "RAG Chat"}
      </button>
      <button
        type="button"
        className={`rounded-md px-3 py-1.5 font-semibold transition disabled:opacity-60 ${
          mode === "agent" ? "bg-white text-violet-brand shadow-sm" : "text-slate-500 hover:text-ink"
        }`}
        disabled={disabled}
        onClick={() => onChange("agent")}
      >
        {compact ? "Agent" : "Agent Mode"}
      </button>
    </div>
  );
}
