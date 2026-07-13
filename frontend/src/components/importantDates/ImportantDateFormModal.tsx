import { useEffect, useState } from "react";
import { EmptyState, LoadingState, PrimaryButton } from "../common";
import { useImportantDateLabels } from "../../hooks/useImportantDateLabels";
import { listTenantMembers, type TenantMember } from "../../services/tenantService";
import type { ImportantDate, ImportantDateInput, ImportantDateType } from "../../types/importantDate";
import { IMPORTANT_DATE_TYPE_OPTIONS } from "../../types/importantDate";

type ImportantDateFormModalProps = {
  initial?: ImportantDate | null;
  isSaving: boolean;
  onClose: () => void;
  onSave: (input: ImportantDateInput) => void;
};

export function ImportantDateFormModal({ initial, isSaving, onClose, onSave }: ImportantDateFormModalProps) {
  const labels = useImportantDateLabels();
  const [title, setTitle] = useState(initial?.title ?? "");
  const [date, setDate] = useState(initial?.date ?? "");
  const [type, setType] = useState<ImportantDateType>(initial?.type ?? "custom");
  const [reminderDaysBefore, setReminderDaysBefore] = useState(initial?.reminderDaysBefore?.toString() ?? "");
  const [note, setNote] = useState(initial?.note ?? "");
  const [assignedToUserId, setAssignedToUserId] = useState(initial?.assignedToUserId ?? "");
  const [members, setMembers] = useState<TenantMember[]>([]);

  useEffect(() => {
    listTenantMembers()
      .then(setMembers)
      .catch(() => setMembers([]));
  }, []);

  const handleSubmit = () => {
    const trimmedTitle = title.trim();
    if (!trimmedTitle || !date) return;
    onSave({
      title: trimmedTitle,
      date,
      type: type || null,
      note: note.trim() || null,
      assignedToUserId: assignedToUserId || null,
      reminderDaysBefore: reminderDaysBefore ? Number(reminderDaysBefore) : null,
    });
  };

  return (
    <div className="fixed inset-0 z-[80] flex items-center justify-center bg-slate-900/35 p-4">
      <div className="w-full max-w-lg rounded-xl border border-line bg-white shadow-panel">
        <div className="border-b border-line px-5 py-4">
          <h2 className="text-lg font-bold text-ink">{initial ? labels.editLabel : labels.addLabel}</h2>
        </div>
        <div className="space-y-4 p-5">
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">Başlık</span>
            <input className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" value={title} onChange={(e) => setTitle(e.target.value)} />
          </label>
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">Tarih</span>
            <input className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </label>
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">Tür</span>
            <select className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" value={type} onChange={(e) => setType(e.target.value as ImportantDateType)}>
              {IMPORTANT_DATE_TYPE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">{labels.reminderLabel} (gün önce)</span>
            <input className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" type="number" min="0" placeholder="Örn. 7" value={reminderDaysBefore} onChange={(e) => setReminderDaysBefore(e.target.value)} />
          </label>
          {members.length > 0 && (
            <label className="space-y-2">
              <span className="text-sm font-semibold text-slate-600">{labels.assigneeLabel}</span>
              <select className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" value={assignedToUserId} onChange={(e) => setAssignedToUserId(e.target.value)}>
                <option value="">Seçilmedi</option>
                {members.map((member) => (
                  <option key={member.userId} value={member.userId}>
                    {member.fullName}
                  </option>
                ))}
              </select>
            </label>
          )}
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">Not</span>
            <textarea className="min-h-24 w-full rounded-lg border border-line px-3 py-2 text-sm outline-none focus:border-violet-brand" value={note} onChange={(e) => setNote(e.target.value)} />
          </label>
        </div>
        <div className="flex justify-end gap-3 border-t border-line px-5 py-4">
          <button className="h-10 rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand" type="button" disabled={isSaving} onClick={onClose}>
            Vazgeç
          </button>
          <PrimaryButton disabled={isSaving || title.trim() === "" || !date} onClick={handleSubmit}>
            {isSaving ? "Kaydediliyor..." : "Kaydet"}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
