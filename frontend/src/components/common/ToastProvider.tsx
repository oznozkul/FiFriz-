import { AlertCircle, CheckCircle2, Info, Loader2, X } from "lucide-react";
import { useEffect, useState } from "react";
import { dismiss, subscribe, type ToastItem } from "../../services/toast";

const iconMap = {
  success: CheckCircle2,
  error: AlertCircle,
  info: Info,
  loading: Loader2,
} as const;

const styleMap = {
  success: "border-emerald-200/80 bg-emerald-50 text-emerald-800",
  error: "border-red-200/80 bg-red-50 text-red-800",
  info: "border-sky-200/80 bg-sky-50 text-sky-800",
  loading: "border-violet-ring bg-violet-soft/40 text-slate-700",
} as const;

const iconColorMap = {
  success: "text-emerald-600",
  error: "text-red-600",
  info: "text-sky-700",
  loading: "text-violet-brand",
} as const;

export function ToastProvider() {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  useEffect(() => subscribe(setToasts), []);

  if (toasts.length === 0) return null;

  return (
    <div
      className="pointer-events-none fixed bottom-4 right-4 z-[200] flex w-full max-w-sm flex-col gap-2.5 px-4 sm:px-0"
      aria-live="polite"
      aria-relevant="additions"
    >
      {toasts.map((item) => {
        const Icon = iconMap[item.type];
        return (
          <div
            key={item.id}
            className={`pointer-events-auto toast-enter flex items-start gap-3 rounded-xl border px-4 py-3 shadow-lift ${styleMap[item.type]}`}
            role={item.type === "error" ? "alert" : "status"}
          >
            <Icon
              className={`mt-0.5 h-5 w-5 shrink-0 ${item.type === "loading" ? "animate-spin" : ""} ${iconColorMap[item.type]}`}
            />
            <p className="min-w-0 flex-1 text-sm font-semibold leading-5">{item.message}</p>
            {item.type !== "loading" ? (
              <button
                className="rounded-md p-1 text-current/60 transition hover:bg-black/5 hover:text-current focus:outline-none focus:ring-4 focus:ring-violet-ring/50"
                type="button"
                aria-label="Bildirimi kapat"
                onClick={() => dismiss(item.id)}
              >
                <X className="h-4 w-4" />
              </button>
            ) : null}
          </div>
        );
      })}
    </div>
  );
}
