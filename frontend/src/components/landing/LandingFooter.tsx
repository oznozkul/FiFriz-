import { Globe, Linkedin, Twitter } from "lucide-react";
import { Link } from "react-router-dom";
import { BrandMark } from "../common";
import {
  footerCompanyLinks,
  footerLegalLinks,
  footerProductLinks,
  footerResourceLinks,
} from "./landingData";

function FooterColumn({ title, links }: { title: string; links: ReadonlyArray<{ label: string; href: string }> }) {
  return (
    <div>
      <h3 className="text-sm font-bold text-white">{title}</h3>
      <ul className="mt-4 space-y-2.5">
        {links.map((link) => (
          <li key={link.label}>
            <a href={link.href} className="text-sm text-slate-400 transition hover:text-white">
              {link.label}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function LandingFooter() {
  return (
    <footer id="iletisim" className="scroll-mt-20 bg-slate-900 text-slate-300">
      <div className="mx-auto max-w-7xl px-4 py-14 sm:px-6 lg:px-8">
        <div className="grid gap-10 lg:grid-cols-[1.4fr_repeat(4,1fr)]">
          <div id="hakkimizda">
            <Link to="/" className="flex items-center gap-2.5">
              <BrandMark />
              <span className="text-lg font-bold text-white">FiFriz</span>
            </Link>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-slate-400">
              Yapay zeka destekli doküman yönetim platformu. Belgelerinizi güvenle yönetin, ekibinizle hızlanın.
            </p>
            <div className="mt-5 flex items-center gap-3">
              <a
                href="https://linkedin.com"
                target="_blank"
                rel="noreferrer"
                className="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-700 text-slate-400 transition hover:border-slate-500 hover:text-white"
                aria-label="LinkedIn"
              >
                <Linkedin className="h-4 w-4" />
              </a>
              <a
                href="https://twitter.com"
                target="_blank"
                rel="noreferrer"
                className="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-700 text-slate-400 transition hover:border-slate-500 hover:text-white"
                aria-label="Twitter"
              >
                <Twitter className="h-4 w-4" />
              </a>
              <a
                href="https://github.com"
                target="_blank"
                rel="noreferrer"
                className="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-700 text-slate-400 transition hover:border-slate-500 hover:text-white"
                aria-label="Web"
              >
                <Globe className="h-4 w-4" />
              </a>
            </div>
          </div>

          <FooterColumn title="Ürün" links={footerProductLinks} />
          <FooterColumn title="Kaynaklar" links={footerResourceLinks} />
          <FooterColumn title="Şirket" links={footerCompanyLinks} />
          <FooterColumn title="Yasal" links={footerLegalLinks} />
        </div>

        <div className="mt-12 flex flex-col items-center justify-between gap-3 border-t border-slate-800 pt-6 text-sm text-slate-500 sm:flex-row">
          <p>© {new Date().getFullYear()} FiFriz. Tüm hakları saklıdır.</p>
          <p>Türkiye&apos;de ❤️ ile üretildi</p>
        </div>
      </div>
    </footer>
  );
}
