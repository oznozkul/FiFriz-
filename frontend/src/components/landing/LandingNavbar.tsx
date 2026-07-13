import { Globe, Menu, X } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { BrandMark } from "../common";
import { landingNavLinks } from "./landingData";

export function LandingNavbar() {
  const [mobileOpen, setMobileOpen] = useState(false);

  useEffect(() => {
    if (!mobileOpen) return;
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previous;
    };
  }, [mobileOpen]);

  const closeMobile = () => setMobileOpen(false);

  return (
    <header className="sticky top-0 z-50 border-b border-line/80 bg-white/90 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
        <Link to="/" className="flex shrink-0 items-center gap-2.5" onClick={closeMobile}>
          <BrandMark />
          <span className="text-lg font-bold tracking-tight text-ink">FiFriz</span>
        </Link>

        <nav className="hidden items-center gap-6 lg:flex" aria-label="Ana menü">
          {landingNavLinks.map((item) => (
            <a
              key={item.href}
              href={item.href}
              className="text-sm font-medium text-slate-600 transition hover:text-violet-brand"
            >
              {item.label}
            </a>
          ))}
        </nav>

        <div className="hidden items-center gap-3 lg:flex">
          <button
            type="button"
            className="inline-flex items-center gap-1.5 rounded-lg px-2 py-1.5 text-sm font-semibold text-slate-600 transition hover:bg-slate-50"
            aria-label="Dil seçimi"
          >
            <Globe className="h-4 w-4" />
            TR
          </button>
          <Link
            to="/login"
            className="inline-flex h-10 items-center rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-700 transition hover:border-violet-ring hover:text-violet-brand"
          >
            Giriş Yap
          </Link>
          <Link
            to="/register"
            className="inline-flex h-10 items-center rounded-lg bg-violet-brand px-4 text-sm font-semibold text-white shadow-lg shadow-violet-200/80 transition hover:bg-violet-hover active:scale-[0.98]"
          >
            FiFriz&apos;e Başla
          </Link>
        </div>

        <button
          type="button"
          className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-line text-slate-600 lg:hidden"
          aria-label={mobileOpen ? "Menüyü kapat" : "Menüyü aç"}
          onClick={() => setMobileOpen((open) => !open)}
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {mobileOpen ? (
        <div className="border-t border-line bg-white px-4 py-4 lg:hidden">
          <nav className="flex flex-col gap-1" aria-label="Mobil menü">
            {landingNavLinks.map((item) => (
              <a
                key={item.href}
                href={item.href}
                className="rounded-lg px-3 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-violet-soft/50 hover:text-violet-brand"
                onClick={closeMobile}
              >
                {item.label}
              </a>
            ))}
          </nav>
          <div className="mt-4 flex flex-col gap-2 border-t border-line pt-4">
            <button
              type="button"
              className="inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-semibold text-slate-600"
            >
              <Globe className="h-4 w-4" />
              Türkçe (TR)
            </button>
            <Link
              to="/login"
              className="inline-flex h-10 items-center justify-center rounded-lg border border-line text-sm font-semibold text-slate-700"
              onClick={closeMobile}
            >
              Giriş Yap
            </Link>
            <Link
              to="/register"
              className="inline-flex h-10 items-center justify-center rounded-lg bg-violet-brand text-sm font-semibold text-white"
              onClick={closeMobile}
            >
              FiFriz&apos;e Başla
            </Link>
          </div>
        </div>
      ) : null}
    </header>
  );
}
