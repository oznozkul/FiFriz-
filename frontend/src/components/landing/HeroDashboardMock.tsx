import type { ReactNode } from "react";
import {
  Bot,
  CalendarClock,
  CheckCircle2,
  ClipboardCheck,
  FileText,
  FolderOpen,
  LayoutDashboard,
  Mail,
  Search,
  Sparkles,
  Upload,
} from "lucide-react";

type FloatingCardProps = {
  icon: ReactNode;
  title: string;
  className: string;
};

function FloatingCard({ icon, title, className }: FloatingCardProps) {
  return (
    <div
      className={`absolute z-20 flex max-w-[168px] items-center gap-2 rounded-xl border border-violet-ring/80 bg-white/95 px-3 py-2 shadow-lift backdrop-blur-sm ${className}`}
    >
      <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">{icon}</span>
      <p className="text-[11px] font-semibold leading-snug text-ink">{title}</p>
    </div>
  );
}

const documents = [
  { name: "Kira_Sozlesmesi.pdf", status: "Hazır", statusTone: "emerald" as const, active: true },
  { name: "Fatura_Ocak.pdf", status: "İşleniyor", statusTone: "amber" as const, active: false },
  { name: "Garanti_Belgesi.jpg", status: "Yüklendi", statusTone: "sky" as const, active: false },
  { name: "Sozlesme_v2.docx", status: "Onay", statusTone: "violet" as const, active: false },
];

const statusClasses = {
  emerald: "bg-emerald-100 text-emerald-700",
  amber: "bg-amber-100 text-amber-700",
  sky: "bg-sky-100 text-sky-700",
  violet: "bg-violet-soft text-violet-brand",
};

export function HeroDashboardMock() {
  return (
    <div className="relative mx-auto w-full max-w-[640px] lg:max-w-none lg:pl-4">
      <div className="pointer-events-none absolute -right-4 top-6 h-56 w-56 rounded-full bg-violet-brand/15 blur-3xl sm:-right-10 sm:h-72 sm:w-72" />
      <div className="pointer-events-none absolute -bottom-8 left-0 h-44 w-44 rounded-full bg-violet-soft blur-2xl" />

      <div
        className="pointer-events-none absolute inset-0 opacity-[0.35]"
        style={{
          backgroundImage: "radial-gradient(circle, rgba(115,103,240,0.18) 1px, transparent 1px)",
          backgroundSize: "22px 22px",
        }}
      />

      <FloatingCard
        icon={<Sparkles className="h-3.5 w-3.5" />}
        title="AI özet hazır"
        className="-left-2 top-8 hidden sm:flex lg:-left-10"
      />
      <FloatingCard
        icon={<CalendarClock className="h-3.5 w-3.5" />}
        title="3 önemli tarih bulundu"
        className="-right-1 top-20 hidden md:flex lg:-right-8"
      />
      <FloatingCard
        icon={<Mail className="h-3.5 w-3.5" />}
        title="Gmail'den 12 belge içe aktarıldı"
        className="-left-1 bottom-16 hidden sm:flex lg:-left-12"
      />
      <FloatingCard
        icon={<ClipboardCheck className="h-3.5 w-3.5" />}
        title="Workflow onay bekliyor"
        className="-right-2 bottom-8 flex sm:bottom-10 lg:-right-10"
      />

      <div className="relative overflow-hidden rounded-2xl border border-line/90 bg-white p-1.5 shadow-lift">
        <div className="flex h-[300px] overflow-hidden rounded-xl bg-slate-50/90 sm:h-[360px] lg:h-[420px]">
          <aside className="hidden w-[52px] shrink-0 flex-col items-center gap-2.5 border-r border-line bg-white py-3 sm:flex">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-violet-soft text-violet-brand shadow-sm">
              <LayoutDashboard className="h-4 w-4" />
            </div>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-100 text-slate-400">
              <FolderOpen className="h-3.5 w-3.5" />
            </div>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-100 text-slate-400">
              <FileText className="h-3.5 w-3.5" />
            </div>
            <div className="mt-auto flex h-8 w-8 items-center justify-center rounded-lg bg-slate-100 text-slate-400">
              <Bot className="h-3.5 w-3.5" />
            </div>
          </aside>

          <div className="min-w-0 flex-1 p-3 sm:p-4">
            <div className="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-[10px] font-semibold uppercase tracking-wide text-slate-400">Workspace</p>
                <p className="text-sm font-bold text-ink">Belgelerim</p>
              </div>
              <div className="flex h-8 w-full max-w-none items-center gap-2 rounded-lg border border-line bg-white px-2.5 shadow-sm sm:max-w-[200px]">
                <Search className="h-3.5 w-3.5 shrink-0 text-slate-400" />
                <span className="truncate text-xs text-slate-400">Belge, etiket veya içerik ara...</span>
              </div>
            </div>

            <div className="mb-3 rounded-lg border border-amber-200/80 bg-amber-50/80 px-2.5 py-2">
              <div className="flex items-center gap-2">
                <CalendarClock className="h-3.5 w-3.5 shrink-0 text-amber-600" />
                <p className="text-[10px] font-semibold text-amber-800 sm:text-[11px]">Önemli tarih: Kira ödemesi 3 gün içinde</p>
              </div>
            </div>

            <div className="mb-3 rounded-lg border border-line bg-white p-2.5 shadow-sm">
              <div className="mb-1.5 flex items-center justify-between">
                <div className="flex items-center gap-1.5">
                  <Upload className="h-3 w-3 text-violet-brand" />
                  <span className="text-[10px] font-semibold text-slate-600">Yükleme</span>
                </div>
                <span className="text-[10px] font-bold text-violet-brand">78%</span>
              </div>
              <div className="h-1.5 overflow-hidden rounded-full bg-slate-100">
                <div className="h-full w-[78%] rounded-full bg-gradient-to-r from-violet-brand to-violet-hover" />
              </div>
            </div>

            <div className="space-y-1.5">
              {documents.map((doc) => (
                <div
                  key={doc.name}
                  className={`flex items-center gap-2 rounded-lg border px-2.5 py-2 ${
                    doc.active ? "border-violet-ring bg-violet-soft/50" : "border-line bg-white"
                  }`}
                >
                  <FileText className={`h-3.5 w-3.5 shrink-0 ${doc.active ? "text-violet-brand" : "text-slate-400"}`} />
                  <span className="min-w-0 flex-1 truncate text-[11px] font-semibold text-ink">{doc.name}</span>
                  <span className={`shrink-0 rounded px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide ${statusClasses[doc.statusTone]}`}>
                    {doc.status}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <aside className="hidden w-[130px] shrink-0 border-l border-line bg-white p-3 lg:block">
            <div className="mb-2 flex items-center justify-between">
              <div className="flex items-center gap-1.5">
                <Bot className="h-4 w-4 text-violet-brand" />
                <span className="text-[11px] font-bold text-ink">AI Asistan</span>
              </div>
              <span className="rounded bg-emerald-100 px-1 py-0.5 text-[8px] font-bold text-emerald-700">CANLI</span>
            </div>

            <div className="space-y-2 rounded-xl border border-violet-ring/60 bg-violet-soft/40 p-2">
              <div className="flex items-start gap-1.5">
                <Sparkles className="mt-0.5 h-3 w-3 shrink-0 text-violet-brand" />
                <p className="text-[10px] leading-4 text-slate-600">Kira sözleşmesi özetlendi. Son ödeme: 15 Mart.</p>
              </div>
              <div className="rounded-md border border-line bg-white/90 p-1.5">
                <p className="text-[9px] font-semibold text-slate-500">Önerilen etiketler</p>
                <div className="mt-1 flex flex-wrap gap-1">
                  {["Kira", "Sözleşme"].map((tag) => (
                    <span key={tag} className="rounded bg-violet-soft px-1.5 py-0.5 text-[8px] font-bold text-violet-brand">
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            <div className="mt-2 flex items-center gap-1.5 rounded-lg border border-violet-ring bg-violet-soft/30 px-2 py-1.5">
              <CheckCircle2 className="h-3 w-3 text-violet-brand" />
              <span className="text-[9px] font-semibold text-slate-600">Onay bekliyor</span>
            </div>
          </aside>
        </div>
      </div>

      <div className="pointer-events-none absolute -bottom-2 left-1/2 hidden h-1 w-[88%] -translate-x-1/2 rounded-full bg-gradient-to-r from-transparent via-violet-brand/25 to-transparent sm:block" />
    </div>
  );
}
