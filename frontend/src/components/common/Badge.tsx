import type { ReactNode } from "react";

type BadgeProps = {
  children: ReactNode;
  tone?: "slate" | "violet" | "green" | "amber" | "rose" | "blue";
  className?: string;
};

export type { BadgeProps };

const toneClasses: Record<NonNullable<BadgeProps["tone"]>, string> = {
  slate: "bg-slate-100 text-slate-600 ring-slate-200/80",
  violet: "bg-violet-soft text-violet-brand ring-violet-ring/80",
  green: "bg-emerald-100 text-emerald-600 ring-emerald-200/80",
  amber: "bg-amber-100 text-amber-600 ring-amber-200/80",
  rose: "bg-rose-100 text-rose-600 ring-rose-200/80",
  blue: "bg-sky-100 text-sky-700 ring-sky-200/80",
};

export function Badge({ children, tone = "slate", className = "" }: BadgeProps) {
  return (
    <span
      className={`inline-flex w-fit items-center rounded-md px-2 py-1 text-xs font-semibold ring-1 ring-inset ${toneClasses[tone]} ${className}`}
    >
      {children}
    </span>
  );
}
