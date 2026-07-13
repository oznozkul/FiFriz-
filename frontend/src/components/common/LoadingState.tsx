import { Spinner } from "./Spinner";

type LoadingStateProps = {
  label?: string;
  variant?: "inline" | "panel" | "fullscreen";
  className?: string;
  size?: "sm" | "md" | "lg";
};

export function LoadingState({ label = "Yükleniyor...", variant = "panel", className = "", size = "md" }: LoadingStateProps) {
  if (variant === "inline") {
    return <Spinner size={size} label={label} className={className} />;
  }

  if (variant === "fullscreen") {
    return (
      <div className={`flex min-h-screen flex-col items-center justify-center gap-3 bg-slate-50 ${className}`}>
        <Spinner size="lg" />
        <p className="text-sm font-semibold text-slate-500">{label}</p>
      </div>
    );
  }

  return (
    <section className={`panel flex items-center justify-center gap-3 p-8 ${className}`}>
      <Spinner size={size} label={label} />
    </section>
  );
}
