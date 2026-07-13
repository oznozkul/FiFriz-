import { ArrowRight, Check } from "lucide-react";
import type { LandingFeature } from "./landingData";

type FeatureCardProps = {
  feature: LandingFeature;
  index: number;
  isVisible: boolean;
};

export function FeatureCard({ feature, index, isVisible }: FeatureCardProps) {
  const Icon = feature.icon;
  const staggerClass = `landing-reveal-stagger-${Math.min(index + 1, 8)}`;

  return (
    <article
      className={`group flex h-full flex-col rounded-2xl border border-line bg-white p-6 shadow-panel transition duration-300 hover:-translate-y-1 hover:border-violet-brand/40 hover:shadow-lift landing-reveal ${staggerClass} ${
        isVisible ? "is-visible" : ""
      }`}
    >
      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-violet-soft text-violet-brand transition duration-300 group-hover:scale-105 group-hover:bg-violet-soft/90">
        <Icon className="h-5 w-5" />
      </div>

      <h3 className="mt-5 text-base font-bold text-ink">{feature.title}</h3>
      <p className="mt-2 flex-1 text-sm leading-relaxed text-slate-600">{feature.description}</p>

      <div className="mt-5 border-t border-line pt-4">
        <p className="text-[10px] font-bold uppercase tracking-wide text-slate-400">Öne Çıkan</p>
        <ul className="mt-2 space-y-1.5">
          {feature.highlights.map((highlight) => (
            <li key={highlight} className="flex items-center gap-2 text-xs font-semibold text-slate-600">
              <Check className="h-3.5 w-3.5 shrink-0 text-violet-brand" aria-hidden="true" />
              {highlight}
            </li>
          ))}
        </ul>
      </div>

      <div className="mt-4 flex items-center gap-1 text-xs font-semibold text-violet-brand opacity-0 transition duration-300 group-hover:translate-x-0.5 group-hover:opacity-100">
        <span>Detayları gör</span>
        <ArrowRight className="h-3.5 w-3.5" />
      </div>
    </article>
  );
}
