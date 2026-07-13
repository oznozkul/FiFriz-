import { FileSearch, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";

type EmptyStateProps = {
  title: string;
  description?: string;
  action?: ReactNode;
  icon?: LucideIcon;
  variant?: "panel" | "inline";
  className?: string;
};

export function EmptyState({
  title,
  description,
  action,
  icon: Icon = FileSearch,
  variant = "panel",
  className = "",
}: EmptyStateProps) {
  const content = (
    <>
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
        <Icon className="h-6 w-6" aria-hidden="true" />
      </div>
      <h2 className="text-base font-bold text-ink">{title}</h2>
      {description ? <p className="mt-2 max-w-md text-sm text-slate-500">{description}</p> : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </>
  );

  if (variant === "inline") {
    return <div className={`flex flex-col items-center px-4 py-8 text-center ${className}`}>{content}</div>;
  }

  return (
    <div className={`panel flex flex-col items-center justify-center rounded-2xl px-6 py-12 text-center ${className}`}>
      {content}
    </div>
  );
}
