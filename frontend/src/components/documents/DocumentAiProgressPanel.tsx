import { CheckCircle2, CircleDashed, Clock3, Loader2, XCircle } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { getDocumentAiProgress } from "../../services/documentAiProgressService";
import { AI_PIPELINE_STEP_LABELS, type DocumentAiStepProgress } from "../../types/documentAiProgress";

type DocumentAiProgressPanelProps = {
  documentId: string;
  refreshToken?: number;
};

function stepLabel(stepName: string) {
  return AI_PIPELINE_STEP_LABELS[stepName] ?? stepName;
}

function stepStatusLabel(status: DocumentAiStepProgress["status"]) {
  const labels: Record<DocumentAiStepProgress["status"], string> = {
    completed: "Tamamlandı",
    running: "Çalışıyor",
    retrying: "Yeniden deneniyor",
    failed: "Başarısız",
    skipped: "Atlandı",
    pending: "Bekliyor",
  };
  return labels[status] ?? status;
}

function StepIcon({ status }: { status: DocumentAiStepProgress["status"] }) {
  switch (status) {
    case "completed":
      return <CheckCircle2 className="h-4 w-4 text-emerald-600" />;
    case "running":
    case "retrying":
      return <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />;
    case "failed":
      return <XCircle className="h-4 w-4 text-rose-500" />;
    case "skipped":
      return <CircleDashed className="h-4 w-4 text-slate-400" />;
    default:
      return <Clock3 className="h-4 w-4 text-slate-300" />;
  }
}

export function DocumentAiProgressPanel({ documentId, refreshToken = 0 }: DocumentAiProgressPanelProps) {
  const [steps, setSteps] = useState<DocumentAiStepProgress[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadProgress = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await getDocumentAiProgress(documentId);
      setSteps(result);
    } catch (loadError) {
      setSteps([]);
      setError(loadError instanceof Error ? loadError.message : "AI ilerlemesi yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void loadProgress();
  }, [documentId, loadProgress, refreshToken]);

  const hasActiveStep = steps.some((step) => step.status === "running" || step.status === "retrying" || step.status === "pending");

  useEffect(() => {
    if (!hasActiveStep) {
      return;
    }
    const timer = window.setInterval(() => {
      void loadProgress();
    }, 4000);
    return () => window.clearInterval(timer);
  }, [hasActiveStep, loadProgress]);

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-4 text-sm text-slate-500">
        <Loader2 className="h-4 w-4 animate-spin text-violet-brand" />
        AI ilerlemesi yükleniyor...
      </div>
    );
  }

  if (error) {
    return <p className="text-sm text-rose-600">{error}</p>;
  }

  if (steps.length === 0) {
    return <p className="text-sm text-slate-500">Bu belge için henüz AI işlemi başlamadı.</p>;
  }

  return (
    <div className="space-y-2">
      <p className="text-xs font-bold uppercase tracking-wide text-slate-400">AI İşlem Durumu</p>
      <ul className="space-y-2">
        {steps.map((step) => (
          <li key={step.stepName} className="flex items-center gap-3 rounded-lg border border-line bg-slate-50/70 px-3 py-2">
            <StepIcon status={step.status} />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold text-ink">{stepLabel(step.stepName)}</p>
              {step.error ? <p className="truncate text-xs text-rose-600">{step.error}</p> : null}
            </div>
            <span className="text-xs font-semibold text-slate-500">{stepStatusLabel(step.status)}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
