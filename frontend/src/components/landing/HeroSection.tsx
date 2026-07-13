import { ChevronRight, Play, Star } from "lucide-react";
import { Link } from "react-router-dom";
import { HeroDashboardMock } from "./HeroDashboardMock";

const teamAvatars = [
  { initials: "AK", className: "bg-gradient-to-br from-violet-brand to-violet-hover" },
  { initials: "SY", className: "bg-gradient-to-br from-sky-500 to-sky-600" },
  { initials: "MT", className: "bg-gradient-to-br from-emerald-500 to-emerald-600" },
  { initials: "EL", className: "bg-gradient-to-br from-amber-500 to-amber-600" },
];

export function HeroSection() {
  return (
    <section id="urun" className="relative overflow-hidden scroll-mt-20">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,rgba(115,103,240,0.14),transparent_50%)]" />
      <div className="pointer-events-none absolute -left-24 top-32 h-72 w-72 rounded-full bg-violet-brand/10 blur-3xl" />
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-full opacity-[0.22]"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(148,163,184,0.08) 1px, transparent 1px), linear-gradient(to bottom, rgba(148,163,184,0.08) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
          maskImage: "linear-gradient(to bottom, black 40%, transparent 100%)",
        }}
      />
      <div className="pointer-events-none absolute bottom-0 left-0 right-0 h-28 bg-gradient-to-t from-white to-transparent" />

      <div className="relative mx-auto grid max-w-7xl items-center gap-10 px-4 py-12 sm:px-6 sm:py-14 lg:min-h-[calc(100vh-4rem)] lg:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)] lg:gap-14 lg:px-8 lg:py-16">
        <div className="mx-auto w-full max-w-xl lg:mx-0">
          <span className="inline-flex items-center rounded-full border border-violet-ring bg-violet-soft/70 px-3.5 py-1 text-[11px] font-bold uppercase tracking-[0.12em] text-violet-brand">
            Yapay Zeka Destekli Doküman Yönetimi
          </span>

          <h1 className="mt-6 text-[2rem] font-extrabold leading-[1.12] tracking-tight text-ink sm:text-5xl sm:leading-[1.08] lg:text-[3.25rem]">
            Belgelerinizi{" "}
            <span className="bg-gradient-to-r from-violet-brand to-violet-hover bg-clip-text text-transparent">Akıllandırın</span>,{" "}
            <br className="hidden sm:block" />
            İşinizi{" "}
            <span className="bg-gradient-to-r from-violet-brand to-violet-hover bg-clip-text text-transparent">Hızlandırın</span>.
          </h1>

          <p className="mt-5 max-w-lg text-base leading-7 text-slate-600 sm:text-lg sm:leading-8">
            FiFriz, yapay zeka destekli doküman yönetim sistemi ile belgelerinizi güvenli, düzenli ve erişilebilir hale getirir.
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:items-center">
            <Link
              to="/register"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-xl bg-violet-brand px-6 text-sm font-bold text-white shadow-lg shadow-violet-200/90 transition hover:-translate-y-0.5 hover:bg-violet-hover hover:shadow-xl hover:shadow-violet-200/80 active:translate-y-0 active:scale-[0.98]"
            >
              FiFriz&apos;e Başla
              <ChevronRight className="h-4 w-4" />
            </Link>
            <a
              href="#nasil-calisir"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-xl border border-line bg-white px-6 text-sm font-bold text-slate-700 shadow-sm transition hover:-translate-y-0.5 hover:border-violet-ring hover:bg-violet-soft/40 hover:text-violet-brand active:translate-y-0 active:scale-[0.98]"
            >
              <span className="flex h-7 w-7 items-center justify-center rounded-full bg-violet-soft">
                <Play className="h-3.5 w-3.5 text-violet-brand" />
              </span>
              Canlı Demo İzle
            </a>
          </div>

          <p className="mt-3 text-xs font-semibold text-slate-500 sm:text-sm">
            Kredi kartı gerekmez · Dakikalar içinde başlayın
          </p>

          <div className="mt-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:gap-5">
            <div className="flex items-center gap-3">
              <div className="flex -space-x-2.5">
                {teamAvatars.map((avatar) => (
                  <span
                    key={avatar.initials}
                    className={`flex h-9 w-9 items-center justify-center rounded-full border-2 border-white text-[10px] font-bold text-white shadow-sm ${avatar.className}`}
                  >
                    {avatar.initials}
                  </span>
                ))}
              </div>
              <p className="text-sm leading-snug text-slate-600">
                <span className="font-bold text-ink">1.000+ ekip</span> FiFriz ile belgelerini yönetiyor
              </p>
            </div>

            <div className="inline-flex w-fit items-center gap-2 rounded-full border border-line bg-white px-3 py-1.5 shadow-sm">
              <div className="flex items-center gap-0.5 text-amber-500">
                {Array.from({ length: 5 }).map((_, index) => (
                  <Star key={index} className="h-3.5 w-3.5 fill-current" />
                ))}
              </div>
              <span className="text-xs font-bold text-ink">4.9</span>
              <span className="text-xs font-semibold text-slate-500">Kurumsal güven</span>
            </div>
          </div>
        </div>

        <div className="w-full min-w-0 lg:justify-self-end">
          <HeroDashboardMock />
        </div>
      </div>
    </section>
  );
}
