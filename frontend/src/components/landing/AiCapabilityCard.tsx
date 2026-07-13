import { Check } from "lucide-react";
import type { AiCapability } from "./aiShowcaseData";

type AiCapabilityCardProps = {
  capability: AiCapability;
  index: number;
  isVisible: boolean;
};

export function AiCapabilityCard({ capability, index, isVisible }: AiCapabilityCardProps) {
  const Icon = capability.icon;
  const staggerClass = `landing-reveal-stagger-${Math.min(index + 3, 8)}`;

  return (
    <article
      className={`group flex h-full flex-col rounded-2xl border border-line bg-white/90 p-4 shadow-panel backdrop-blur-sm transition duration-300 hover:-translate-y-0.5 hover:border-violet-brand/35 hover:shadow-lift landing-reveal ${staggerClass} ${
        isVisible ? "is-visible" : ""
      }`}
    >
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-violet-soft text-violet-brand transition duration-300 group-hover:scale-105">
          <Icon className="h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <Check className="h-3.5 w-3.5 shrink-0 text-violet-brand" aria-hidden="true" />
            <h3 className="text-sm font-bold text-ink">{capability.title}</h3>
          </div>
          <p className="mt-1.5 text-sm leading-relaxed text-slate-600">{capability.description}</p>
        </div>
      </div>
    </article>
  );
}
