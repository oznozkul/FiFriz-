import { ArrowRight } from "lucide-react";
import { FeatureCard } from "./FeatureCard";
import { landingFeatures } from "./landingData";
import { useRevealOnScroll } from "./useRevealOnScroll";

export function FeatureSection() {
  const { ref: sectionRef, isVisible } = useRevealOnScroll<HTMLElement>();

  return (
    <section
      id="ozellikler"
      ref={sectionRef}
      className="relative scroll-mt-20 overflow-hidden bg-gradient-to-b from-slate-50/90 via-violet-soft/20 to-slate-50/80 py-20 sm:py-24"
    >
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(115,103,240,0.06),transparent_45%)]" />

      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div
          className={`mx-auto max-w-3xl text-center landing-reveal landing-reveal-stagger-1 ${isVisible ? "is-visible" : ""}`}
        >
          <h2 className="text-3xl font-extrabold tracking-tight text-ink sm:text-4xl">Neden FiFriz?</h2>
          <p className="mt-4 text-base leading-relaxed text-slate-600 sm:text-lg sm:leading-8">
            Yapay zeka destekli belge yönetimini, kurumsal güvenlik ve modern kullanıcı deneyimi ile bir araya getiriyoruz.
          </p>
        </div>

        <div className="mt-14 grid grid-cols-1 gap-5 sm:grid-cols-2 sm:gap-6 lg:grid-cols-3 lg:gap-7">
          {landingFeatures.map((feature, index) => (
            <FeatureCard key={feature.title} feature={feature} index={index} isVisible={isVisible} />
          ))}
        </div>

        <div
          className={`mt-16 rounded-2xl border border-line bg-white/80 px-6 py-8 text-center shadow-panel backdrop-blur-sm sm:px-10 sm:py-10 landing-reveal landing-reveal-stagger-8 ${
            isVisible ? "is-visible" : ""
          }`}
        >
          <h3 className="text-lg font-bold text-ink sm:text-xl">Daha fazlasını keşfetmeye hazır mısınız?</h3>
          <p className="mx-auto mt-2 max-w-xl text-sm text-slate-600">
            FiFriz&apos;in tüm yeteneklerini keşfedin ve ekibiniz için en uygun akışı oluşturun.
          </p>
          <a
            href="/features"
            className="mt-6 inline-flex h-11 items-center justify-center gap-2 rounded-xl border border-line bg-white px-5 text-sm font-bold text-slate-700 shadow-sm transition hover:-translate-y-0.5 hover:border-violet-ring hover:bg-violet-soft/40 hover:text-violet-brand active:scale-[0.98]"
          >
            Tüm Özellikleri İncele
            <ArrowRight className="h-4 w-4" />
          </a>
        </div>
      </div>
    </section>
  );
}
