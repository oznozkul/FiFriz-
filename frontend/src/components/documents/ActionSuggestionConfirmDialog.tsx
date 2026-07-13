import { confirmationMessageFromPayload, toolNameFromPayload, type ActionSuggestion } from "../../types/actionSuggestion";
import { AppModal, Button } from "../common";

type ActionSuggestionConfirmDialogProps = {
  suggestion: ActionSuggestion | null;
  isSubmitting?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
};

export function ActionSuggestionConfirmDialog({ suggestion, isSubmitting = false, onConfirm, onCancel }: ActionSuggestionConfirmDialogProps) {
  if (!suggestion) return null;

  const hasTool = Boolean(toolNameFromPayload(suggestion.actionPayload));
  const message = confirmationMessageFromPayload(suggestion.actionPayload, suggestion.title);

  return (
    <AppModal
      title="Öneriyi onayla"
      onClose={onCancel}
      showCloseButton={false}
      footer={
        <>
          <Button variant="secondary" disabled={isSubmitting} onClick={onCancel}>
            Vazgeç
          </Button>
          <Button disabled={isSubmitting} onClick={onConfirm}>
            {isSubmitting ? "Uygulanıyor..." : "Onayla"}
          </Button>
        </>
      }
    >
      <p className="text-sm leading-6 text-slate-600">
        {hasTool ? "Bu işlem şu aksiyonu gerçekleştirecek:" : "Bu öneriyi kabul etmek istediğinize emin misiniz?"}
      </p>
      <p className="mt-2 rounded-lg bg-violet-soft/60 px-3 py-2 text-sm font-semibold text-ink">{message}</p>
      {suggestion.description ? <p className="mt-2 text-xs text-slate-500">{suggestion.description}</p> : null}
    </AppModal>
  );
}
