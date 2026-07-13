import { Loader2 } from "lucide-react";

type SpinnerSize = "sm" | "md" | "lg";

const sizeClasses: Record<SpinnerSize, string> = {
  sm: "h-4 w-4",
  md: "h-5 w-5",
  lg: "h-8 w-8",
};

type SpinnerProps = {
  size?: SpinnerSize;
  className?: string;
  label?: string;
};

export function Spinner({ size = "md", className = "", label }: SpinnerProps) {
  return (
    <span className={`inline-flex items-center gap-2 ${className}`} role="status" aria-live="polite" aria-busy="true">
      <Loader2 className={`animate-spin text-violet-brand ${sizeClasses[size]}`} aria-hidden="true" />
      {label ? <span className="text-sm font-semibold text-slate-500">{label}</span> : null}
      {!label ? <span className="sr-only">Yükleniyor</span> : null}
    </span>
  );
}
