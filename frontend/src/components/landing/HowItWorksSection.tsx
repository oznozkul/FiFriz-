import { ArrowRight } from "lucide-react";
import { landingSteps } from "./landingData";

export function HowItWorksSection() {
  return (
    <section id="nasil-calisir" className="scroll-mt-20 border-y border-line bg-slate-50/70 py-16 sm:py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-2xl font-bold tracking-tight text-ink sm:text-3xl">Nasıl Çalışır?</h2>
          <p className="mt-3 text-sm leading-relaxed text-slate-600 sm:text-base">
            Dört basit adımda belgelerinizi yükleyin, yapay zeka işlesin ve ekibinizle hızlanın.
          </p>
        </div>

        <div id="cozumler" className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {landingSteps.map((step, index) => {
            const Icon = step.icon;
            return (
              <div key={step.title} className="relative">
                <article className="flex h-full flex-col items-center rounded-2xl border border-line bg-white p-6 text-center shadow-panel">
                  <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-violet-soft text-violet-brand">
                    <Icon className="h-6 w-6" />
                  </div>
                  <h3 className="mt-4 text-base font-bold text-ink">{step.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-slate-500">{step.description}</p>
                </article>
                {index < landingSteps.length - 1 ? (
                  <ArrowRight className="absolute -right-3 top-1/2 hidden h-5 w-5 -translate-y-1/2 text-violet-brand/40 lg:block" />
                ) : null}
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
