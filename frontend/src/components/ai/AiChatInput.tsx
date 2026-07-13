import { Send } from "lucide-react";
import { PrimaryButton } from "../common";

type AiChatInputProps = {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  disabled?: boolean;
  placeholder?: string;
  compact?: boolean;
};

export function AiChatInput({ value, onChange, onSubmit, disabled, placeholder = "Bir şey sorun...", compact }: AiChatInputProps) {
  const canSubmit = value.trim().length > 0 && !disabled;

  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      if (canSubmit) {
        onSubmit();
      }
    }
  };

  return (
    <div className={`flex gap-3 ${compact ? "items-end" : "items-end"}`}>
      <textarea
        className="min-h-[48px] min-w-0 flex-1 resize-none rounded-lg border border-line bg-white px-4 py-3 text-sm outline-none transition placeholder:text-slate-400 focus:border-violet-brand focus:ring-4 focus:ring-violet-ring/70 disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-400"
        rows={compact ? 2 : 3}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        aria-label="Soru yazın"
      />
      <PrimaryButton className={`shrink-0 ${compact ? "h-11 w-11 px-0" : "h-12 w-12 px-0"}`} disabled={!canSubmit} onClick={onSubmit} aria-label="Gönder">
        <Send className="h-5 w-5" />
      </PrimaryButton>
    </div>
  );
}
