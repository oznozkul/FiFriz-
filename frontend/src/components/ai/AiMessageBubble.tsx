type AiMessageBubbleProps = {
  role: "user" | "assistant";
  text: string;
  source?: string;
  maxWidthClassName?: string;
  variant?: "default" | "error";
};

export function AiMessageBubble({ role, text, source, maxWidthClassName = "max-w-[760px]", variant = "default" }: AiMessageBubbleProps) {
  const isUser = role === "user";

  const assistantClassName =
    variant === "error"
      ? "border border-rose-200 bg-rose-50 text-rose-800 shadow-sm"
      : "border border-line bg-white text-slate-700 shadow-sm";

  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div className={`${maxWidthClassName} rounded-lg px-4 py-3 ${isUser ? "bg-violet-soft text-violet-950" : assistantClassName}`}>
        <p className="font-medium leading-6">{text}</p>
        {source && <p className="mt-3 text-xs font-semibold text-slate-500">{source}</p>}
      </div>
    </div>
  );
}
