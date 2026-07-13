import { ArrowRight, Sparkles } from "lucide-react";
import { AiCapabilityCard } from "./AiCapabilityCard";
import { AiChatMockup } from "./AiChatMockup";
import { aiCapabilities } from "./aiShowcaseData";
import { useRevealOnScroll } from "./useRevealOnScroll";

export function AiShowcaseSection() {
  const { ref: sectionRef, isVisible } = useRevealOnScroll<HTMLElement>();

  return (
    <section
      id="ai-showcase"
      ref={sectionRef}
      className="relative scroll-mt-20 overflow-hidden bg-gradient-to-b from-white via-violet-soft/25 to-white py-20 sm:py-24"
    >
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_center,rgba(115,103,240,0.08),transparent_55%)]" />
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.2]"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(148,163,184,0.1) 1px, transparent 1px), linear-gradient(to bottom, rgba(148,163,184,0.1) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        }}
      />

      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div
          className={`mx-auto max-w-3xl text-center landing-reveal landing-reveal-stagger-1 ${
            isVisible ? "is-visible" : ""
          }`}
        >
          <span className="inline-flex items-center gap-1.5 rounded-full border border-violet-ring bg-violet-soft/70 px-3 py-1 text-[11px] font-bold uppercase tracking-wide text-violet-brand">
            <Sparkles className="h-3.5 w-3.5" />
            FiFriz AI
          </span>
          <h2 className="mt-4 text-3xl font-extrabold tracking-tight text-ink sm:text-4xl">Belgelerinizle Konuşun.</h2>
          <p className="mt-4 text-base leading-relaxed text-slate-600 sm:text-lg sm:leading-8">
            FiFriz AI belgelerinizi sadece saklamaz. Anlar, analiz eder, sınıflandırır ve sizin yerinize çalışır.
          </p>
        </div>

        <div className="mt-14 grid grid-cols-1 items-start gap-10 lg:grid-cols-2 lg:gap-12 xl:gap-16">
          <AiChatMockup isVisible={isVisible} />

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-2">
            {aiCapabilities.map((capability, index) => (
              <AiCapabilityCard key={capability.title} capability={capability} index={index} isVisible={isVisible} />
            ))}
          </div>
        </div>

        <div
          className={`mt-14 flex flex-col items-center justify-between gap-4 rounded-2xl border border-line bg-white/85 px-6 py-6 text-center shadow-panel backdrop-blur-sm sm:flex-row sm:px-8 sm:text-left landing-reveal landing-reveal-stagger-8 ${
            isVisible ? "is-visible" : ""
          }`}
        >
          <div>
            <p className="text-sm font-bold uppercase tracking-wide text-violet-brand">FiFriz AI&apos;ı keşfet</p>
            <h3 className="mt-1 text-lg font-bold text-ink">Yapay zeka ile belge yönetimini bir üst seviyeye taşıyın.</h3>
          </div>
          <a
            href="/ai"
            className="inline-flex h-11 shrink-0 items-center justify-center gap-2 rounded-xl bg-violet-brand px-5 text-sm font-bold text-white shadow-lg shadow-violet-200/80 transition hover:-translate-y-0.5 hover:bg-violet-hover hover:shadow-xl active:scale-[0.98]"
          >
            AI Özelliklerini İncele
            <ArrowRight className="h-4 w-4" />
          </a>
        </div>
      </div>
    </section>
  );
}
