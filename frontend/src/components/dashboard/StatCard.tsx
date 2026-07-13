import type { AppIcon } from "../../types";

type StatCardProps = {
  label: string;
  value: string;
  change?: string;
  icon: AppIcon;
};

export function StatCard({ label, value, change, icon: Icon }: StatCardProps) {
  return (
    <section className="panel rounded-2xl p-5 transition hover:-translate-y-0.5 hover:border-violet-ring hover:shadow-lift">
      <div className="mb-4 flex items-center justify-between">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
          <Icon className="h-5 w-5" />
        </div>
        {change ? <span className="text-xs font-semibold text-emerald-500">{change}</span> : <span />}
      </div>
      <p className="text-sm font-medium text-slate-500">{label}</p>
      <p className="mt-1 text-3xl font-bold text-ink">{value}</p>
    </section>
  );
}
