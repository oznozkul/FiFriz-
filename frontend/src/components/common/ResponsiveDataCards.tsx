import type { ReactNode } from "react";

type ResponsiveDataCardsProps = {
  children: ReactNode;
  className?: string;
};

export function ResponsiveDataCards({ children, className = "" }: ResponsiveDataCardsProps) {
  return <div className={`divide-y divide-line md:hidden ${className}`}>{children}</div>;
}

type ResponsiveDataCardProps = {
  children: ReactNode;
  onClick?: () => void;
  className?: string;
};

export function ResponsiveDataCard({ children, onClick, className = "" }: ResponsiveDataCardProps) {
  const interactive = Boolean(onClick);

  if (interactive) {
    return (
      <button
        type="button"
        className={`w-full bg-white px-4 py-4 text-left transition hover:bg-violet-soft/30 focus:outline-none focus:ring-4 focus:ring-inset focus:ring-violet-ring/70 ${className}`}
        onClick={onClick}
      >
        {children}
      </button>
    );
  }

  return <div className={`bg-white px-4 py-4 ${className}`}>{children}</div>;
}

type ResponsiveDataCardRowProps = {
  label: string;
  children: ReactNode;
  className?: string;
};

export function ResponsiveDataCardRow({ label, children, className = "" }: ResponsiveDataCardRowProps) {
  return (
    <div className={`flex items-start justify-between gap-3 text-sm ${className}`}>
      <span className="shrink-0 text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</span>
      <span className="min-w-0 text-right font-medium text-slate-700">{children}</span>
    </div>
  );
}
