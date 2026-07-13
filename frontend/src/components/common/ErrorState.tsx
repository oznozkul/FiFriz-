import { AlertCircle, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { RetryButton } from "./RetryButton";

type ErrorStateProps = {
  title?: string;
  description?: string;
  onRetry?: () => void;
  retryLabel?: string;
  action?: ReactNode;
  icon?: LucideIcon;
  variant?: "panel" | "inline";
  className?: string;
};

export function ErrorState({
  title = "Bir şeyler ters gitti",
  description,
  onRetry,
  retryLabel,
  action,
  icon: Icon = AlertCircle,
  variant = "panel",
  className = "",
}: ErrorStateProps) {
  const content = (
    <>
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-rose-100 text-rose-600">
        <Icon className="h-6 w-6" aria-hidden="true" />
      </div>
      <h2 className="text-base font-bold text-ink">{title}</h2>
      {description ? <p className="mt-2 max-w-md text-sm text-slate-500">{description}</p> : null}
      {(onRetry || action) && (
        <div className="mt-5 flex flex-wrap items-center justify-center gap-2">
          {onRetry ? <RetryButton onClick={onRetry} label={retryLabel} /> : null}
          {action}
        </div>
      )}
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
