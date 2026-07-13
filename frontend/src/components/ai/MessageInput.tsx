import { Send } from "lucide-react";
import { PrimaryButton } from "../common";

type MessageInputProps = {
  compact?: boolean;
};

export function MessageInput({ compact }: MessageInputProps) {
  return (
    <div className={`flex gap-3 ${compact ? "" : "mt-6"}`}>
      <input className="h-12 min-w-0 flex-1 rounded-lg border border-line bg-white px-4 outline-none focus:border-violet-brand" placeholder="Bir şey sorun..." />
      <PrimaryButton className="h-12 w-12 px-0" aria-label="Gönder">
        <Send className="h-5 w-5" />
      </PrimaryButton>
    </div>
  );
}
