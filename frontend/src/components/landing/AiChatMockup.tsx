import type { ComponentType } from "react";
import { Bot, Sparkles } from "lucide-react";
import { aiFloatingBadges, aiShowcaseConversation } from "./aiShowcaseData";

type AiChatMockupProps = {
  isVisible: boolean;
};

function FloatingBadge({
  label,
  icon: Icon,
  className,
  isVisible,
  delayClass,
}: {
  label: string;
  icon: ComponentType<{ className?: string }>;
  className: string;
  isVisible: boolean;
  delayClass: string;
}) {
  return (
    <div
      className={`absolute z-20 hidden max-w-[170px] items-center gap-2 rounded-xl border border-violet-ring/80 bg-white/95 px-3 py-2 shadow-lift backdrop-blur-sm sm:flex landing-reveal ${delayClass} ${
        isVisible ? "is-visible" : ""
      } ${className}`}
    >
      <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
        <Icon className="h-3.5 w-3.5" />
      </span>
      <p className="text-[11px] font-semibold leading-snug text-ink">{label}</p>
    </div>
  );
}

export function AiChatMockup({ isVisible }: AiChatMockupProps) {
  return (
    <div
      className={`relative mx-auto w-full max-w-xl lg:max-w-none landing-reveal landing-reveal-stagger-2 ${
        isVisible ? "is-visible" : ""
      }`}
    >
      <div className="pointer-events-none absolute -right-6 top-0 h-48 w-48 rounded-full bg-violet-brand/10 blur-3xl" />

      {aiFloatingBadges.map((badge, index) => (
        <FloatingBadge
          key={badge.label}
          label={badge.label}
          icon={badge.icon}
          className={badge.className}
          isVisible={isVisible}
          delayClass={`landing-reveal-stagger-${index + 3}`}
        />
      ))}

      <div className="relative overflow-hidden rounded-2xl border border-line bg-white shadow-lift">
        <div className="flex items-center justify-between border-b border-line bg-slate-50/80 px-4 py-3">
          <div className="flex items-center gap-2.5">
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-violet-soft text-violet-brand">
              <Bot className="h-4 w-4" />
            </span>
            <div>
              <p className="text-sm font-bold text-ink">FiFriz AI</p>
              <p className="text-[11px] font-semibold text-emerald-600">Çevrimiçi · Belgelerinize hazır</p>
            </div>
          </div>
          <span className="hidden items-center gap-1 rounded-full bg-violet-soft px-2.5 py-1 text-[10px] font-bold text-violet-brand sm:inline-flex">
            <Sparkles className="h-3 w-3" />
            Pro AI
          </span>
        </div>

        <div className="space-y-4 bg-gradient-to-b from-white to-slate-50/60 p-4 sm:p-5">
          {aiShowcaseConversation.map((message, index) => {
            const isUser = message.role === "user";
            const bubbleDelay = `landing-chat-delay-${index + 1}`;

            return (
              <div
                key={`${message.role}-${index}`}
                className={`flex ${isUser ? "justify-end" : "justify-start"} landing-chat-bubble ${bubbleDelay} ${
                  isVisible ? "is-visible" : ""
                }`}
              >
                {!isUser ? (
                  <span className="mr-2 mt-1 flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
                    <Sparkles className="h-3.5 w-3.5" />
                  </span>
                ) : null}
                <div
                  className={`max-w-[88%] rounded-2xl px-4 py-3 text-sm leading-relaxed shadow-sm sm:max-w-[78%] ${
                    isUser
                      ? "rounded-br-md bg-ink text-white"
                      : "rounded-bl-md border border-violet-ring/60 bg-violet-soft/50 text-slate-700"
                  }`}
                >
                  {message.content.split("\n").map((line, lineIndex) => (
                    <p key={lineIndex} className={lineIndex > 0 ? "mt-2" : undefined}>
                      {line}
                    </p>
                  ))}
                </div>
              </div>
            );
          })}
        </div>

        <div className="border-t border-line bg-white px-4 py-3">
          <div className="flex items-center gap-2 rounded-xl border border-line bg-slate-50 px-3 py-2.5">
            <span className="flex-1 text-sm text-slate-400">Belgeniz hakkında bir soru sorun...</span>
            <span className="rounded-lg bg-violet-brand px-3 py-1.5 text-xs font-bold text-white">Gönder</span>
          </div>
        </div>
      </div>
    </div>
  );
}
