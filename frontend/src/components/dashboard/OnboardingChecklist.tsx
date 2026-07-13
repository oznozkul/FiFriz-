import { Check, ChevronRight, Sparkles, X } from "lucide-react";
import { Link } from "react-router-dom";
import type { OnboardingStepProgress } from "../../hooks/useOnboardingProgress";
import { IconButton } from "../common";

type OnboardingChecklistProps = {
  steps: OnboardingStepProgress[];
  completedCount: number;
  totalCount: number;
  isLoading?: boolean;
  isPersonal: boolean;
  onDismiss: () => void;
  onUpload?: () => void;
};

export function OnboardingChecklist({
  steps,
  completedCount,
  totalCount,
  isLoading = false,
  isPersonal,
  onDismiss,
  onUpload,
}: OnboardingChecklistProps) {
  const progressPercent = totalCount > 0 ? Math.round((completedCount / totalCount) * 100) : 0;

  return (
    <section className="panel overflow-hidden rounded-2xl border border-violet-ring/80 bg-gradient-to-br from-white via-white to-violet-soft/30 shadow-panel">
      <div className="flex items-start justify-between gap-4 border-b border-line px-5 py-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-violet-soft text-violet-brand">
              <Sparkles className="h-4 w-4" />
            </span>
            <div>
              <h2 className="text-sm font-bold text-ink">Başlangıç rehberi</h2>
              <p className="mt-0.5 text-xs text-slate-500">
                {isPersonal
                  ? "Kişisel çalışma alanınızda ilk adımlar"
                  : "Kurumsal çalışma alanınızı kurmaya başlayın"}
              </p>
            </div>
          </div>
        </div>
        <IconButton className="h-8 w-8 shrink-0" aria-label="Rehberi kapat" onClick={onDismiss}>
          <X className="h-4 w-4" />
        </IconButton>
      </div>

      <div className="space-y-4 px-5 py-4">
        <div>
          <div className="flex items-center justify-between text-xs font-semibold text-slate-500">
            <span>{completedCount}/{totalCount} adım tamamlandı</span>
            <span className="text-violet-brand">{progressPercent}%</span>
          </div>
          <div className="mt-2 h-2 overflow-hidden rounded-full bg-slate-100">
            <div
              className="h-full rounded-full bg-violet-brand transition-all duration-300"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>

        <ol className="space-y-2">
          {steps.map((step, index) => (
            <OnboardingStepRow
              key={step.id}
              step={step}
              index={index + 1}
              isLoading={isLoading}
              onUpload={onUpload}
            />
          ))}
        </ol>
      </div>
    </section>
  );
}

type OnboardingStepRowProps = {
  step: OnboardingStepProgress;
  index: number;
  isLoading: boolean;
  onUpload?: () => void;
};

function OnboardingStepRow({ step, index, isLoading, onUpload }: OnboardingStepRowProps) {
  const content = (
    <>
      <span
        className={`mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-xs font-bold ${
          step.completed
            ? "border-emerald-200 bg-emerald-50 text-emerald-600"
            : "border-line bg-white text-slate-400"
        }`}
      >
        {step.completed ? <Check className="h-3.5 w-3.5" /> : index}
      </span>
      <span className="min-w-0 flex-1">
        <span className={`block text-sm font-semibold ${step.completed ? "text-slate-500 line-through" : "text-ink"}`}>
          {step.title}
        </span>
        <span className="mt-0.5 block text-xs leading-5 text-slate-500">{step.description}</span>
      </span>
      {!step.completed && (step.href || step.action) ? (
        <ChevronRight className="mt-1 h-4 w-4 shrink-0 text-slate-300" />
      ) : null}
    </>
  );

  const className = `flex w-full items-start gap-3 rounded-xl border px-3 py-3 text-left transition ${
    step.completed
      ? "border-line/80 bg-slate-50/60"
      : "border-line bg-white hover:border-violet-ring hover:bg-violet-soft/20"
  } ${isLoading ? "opacity-70" : ""}`;

  if (step.completed || isLoading) {
    return <li className={className}>{content}</li>;
  }

  if (step.action === "upload" && onUpload) {
    return (
      <li>
        <button type="button" className={className} onClick={onUpload}>
          {content}
        </button>
      </li>
    );
  }

  if (step.href) {
    return (
      <li>
        <Link to={step.href} className={className}>
          {content}
        </Link>
      </li>
    );
  }

  return <li className={className}>{content}</li>;
}
