import { useState } from "react";
import type { ApiDocument, DocumentStatus } from "../../types/document";
import { updateDocument } from "../../data/documentsApi";
import { toast } from "../../services/toast";
import { PrimaryButton } from "../common";

const statusOptions: Array<{ label: string; value: DocumentStatus }> = [
  { label: "Yüklendi", value: "uploaded" },
  { label: "İşleniyor", value: "processing" },
  { label: "İşlendi", value: "processed" },
  { label: "Başarısız", value: "failed" },
];

type DocumentRenameModalProps = {
  document: ApiDocument;
  onClose: () => void;
  onSaved: (document: ApiDocument) => void;
};

export function DocumentRenameModal({ document, onClose, onSaved }: DocumentRenameModalProps) {
  const [name, setName] = useState(document.original_file_name);
  const [status, setStatus] = useState<DocumentStatus>(document.status);
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = async () => {
    if (!name.trim()) return;
    setIsSaving(true);
    try {
      const updated = await updateDocument(document.id, {
        original_file_name: name.trim(),
        status,
      });
      onSaved(updated);
      toast.success("Belge güncellendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Belge güncellenemedi.");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[80] flex items-center justify-center bg-slate-900/35 p-4">
      <div className="w-full max-w-md rounded-xl border border-line bg-white shadow-panel">
        <div className="border-b border-line px-5 py-4">
          <h2 className="text-lg font-bold text-ink">Belgeyi Düzenle</h2>
        </div>
        <div className="space-y-4 p-5">
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">Dosya adı</span>
            <input className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">Durum</span>
            <select className="h-11 w-full rounded-lg border border-line bg-white px-3 text-sm outline-none focus:border-violet-brand" value={status} onChange={(event) => setStatus(event.target.value as DocumentStatus)}>
              {statusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="flex justify-end gap-3 border-t border-line px-5 py-4">
          <button className="h-10 rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-600 disabled:opacity-60" type="button" disabled={isSaving} onClick={onClose}>
            Vazgeç
          </button>
          <PrimaryButton disabled={isSaving || !name.trim()} onClick={() => void handleSave()}>
            {isSaving ? "Kaydediliyor..." : "Kaydet"}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
