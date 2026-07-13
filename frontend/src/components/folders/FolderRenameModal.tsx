import { useState } from "react";
import { Button, PrimaryButton } from "../common";
import { useFolderLabels } from "../../hooks/useFolderLabels";
import type { Folder } from "../../types/folder";

type FolderRenameModalProps = {
  folder: Folder;
  isSaving: boolean;
  onClose: () => void;
  onSave: (input: { name: string; description: string }) => void;
};

export function FolderRenameModal({ folder, isSaving, onClose, onSave }: FolderRenameModalProps) {
  const labels = useFolderLabels();
  const [name, setName] = useState(folder.name);
  const [description, setDescription] = useState(folder.description ?? "");

  return (
    <div className="fixed inset-0 z-[80] flex items-center justify-center bg-slate-900/35 p-4">
      <div className="w-full max-w-md rounded-2xl border border-line bg-white shadow-panel">
        <div className="border-b border-line px-5 py-4">
          <h2 className="text-lg font-bold text-ink">{labels.renameTitle}</h2>
        </div>
        <div className="space-y-4 p-5">
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">{labels.nameLabel}</span>
            <input className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label className="space-y-2">
            <span className="text-sm font-semibold text-slate-600">Açıklama</span>
            <textarea className="min-h-24 w-full rounded-lg border border-line px-3 py-2 text-sm outline-none focus:border-violet-brand" value={description} onChange={(event) => setDescription(event.target.value)} />
          </label>
        </div>
        <div className="flex justify-end gap-3 border-t border-line px-5 py-4">
          <Button variant="secondary" disabled={isSaving} onClick={onClose}>
            Vazgeç
          </Button>
          <PrimaryButton disabled={isSaving || name.trim() === ""} onClick={() => onSave({ name: name.trim(), description: description.trim() })}>
            {isSaving ? "Kaydediliyor..." : "Kaydet"}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
