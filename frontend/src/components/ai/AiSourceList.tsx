import type { AiChatSource } from "../../types/aiChat";
import { AiSourceCard } from "./AiSourceCard";

type AiSourceListProps = {
  sources: AiChatSource[];
  question?: string;
};

export function AiSourceList({ sources, question }: AiSourceListProps) {
  if (sources.length === 0) {
    return null;
  }

  return (
    <div className="mt-3 space-y-2">
      <p className="text-xs font-bold uppercase tracking-wide text-slate-400">Kaynaklar</p>
      <div className="space-y-2">
        {sources.map((source) => (
          <AiSourceCard key={source.chunkId} source={source} question={question} />
        ))}
      </div>
    </div>
  );
}
