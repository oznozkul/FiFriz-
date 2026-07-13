import type { ReactNode } from "react";

export function SettingsSection({ title, description, children }: { title: string; description?: string; children: ReactNode }) {
  return (
    <section className="w-full rounded-2xl border border-line bg-white shadow-panel">
      <div className="border-b border-line px-5 py-4 sm:px-6 sm:py-5">
        <h1 className="text-xl font-bold text-ink">{title}</h1>
        {description ? <p className="mt-1 text-sm text-slate-500">{description}</p> : null}
      </div>
      <div className="px-5 py-5 sm:px-6 sm:py-6">{children}</div>
    </section>
  );
}

export function SettingsField({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return (
    <label className="block space-y-2">
      <span className="text-sm font-bold text-slate-600">{label}</span>
      {children}
      {hint ? <p className="text-xs text-slate-500">{hint}</p> : null}
    </label>
  );
}

export function settingsInputClassName(readonly = false) {
  return `h-11 w-full rounded-lg border border-line px-3 text-sm outline-none transition focus:border-violet-brand focus:ring-4 focus:ring-violet-ring/60 ${
    readonly ? "bg-slate-50 text-slate-500" : "bg-white"
  }`;
}

export function InfoItem({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded-lg border border-line bg-slate-50/60 px-4 py-3">
      <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</dt>
      <dd className={`mt-1 text-sm font-semibold text-ink ${mono ? "font-mono text-xs" : ""}`}>{value}</dd>
    </div>
  );
}
