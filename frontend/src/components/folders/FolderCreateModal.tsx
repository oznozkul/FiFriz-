import { useState } from "react";
import { Button, ModalFooter, ModalShell, PrimaryButton } from "../common";
import { useFolderLabels } from "../../hooks/useFolderLabels";
import type { Folder } from "../../types/folder";

type FolderCreateModalProps = {
  parent: Folder | null;
  isSaving: boolean;
  onClose: () => void;
  onCreate: (input: { name: string; description: string; parentId?: string }) => void;
};

export function FolderCreateModal({ parent, isSaving, onClose, onCreate }: FolderCreateModalProps) {
  const labels = useFolderLabels();
  const [name, setName] = useState("");
  const [showDescription, setShowDescription] = useState(false);
  const [description, setDescription] = useState("");

  const handleSubmit = () => {
    if (!name.trim()) return;
    onCreate({ name: name.trim(), description: description.trim(), parentId: parent?.id });
  };

  return (
    <ModalShell onClose={onClose} size="md" zIndexClass="z-[80]" ariaLabel={labels.createTitle}>
      <div className="border-b border-line px-4 py-4 sm:px-5">
        <h2 id="folder-create-title" className="text-lg font-bold text-ink">
          {labels.createTitle}
        </h2>
        <p className="mt-1 text-sm text-slate-500">
          {parent ? labels.createParentHint(parent.name) : labels.createRootHint}
        </p>
      </div>
      <div className="space-y-4 p-4 sm:p-5">
          <label className="block space-y-2">
            <span className="text-sm font-semibold text-slate-600">{labels.nameLabel}</span>
            <input
              autoFocus
              className="h-10 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand focus:ring-4 focus:ring-violet-ring/70"
              placeholder={labels.createPlaceholder}
              value={name}
              onChange={(event) => setName(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter" && name.trim()) handleSubmit();
              }}
            />
          </label>
          {showDescription ? (
            <label className="block space-y-2">
              <span className="text-sm font-semibold text-slate-600">Açıklama</span>
              <textarea
                className="min-h-20 w-full rounded-lg border border-line px-3 py-2 text-sm outline-none focus:border-violet-brand focus:ring-4 focus:ring-violet-ring/70"
                placeholder="İsteğe bağlı"
                value={description}
                onChange={(event) => setDescription(event.target.value)}
              />
            </label>
          ) : (
            <button
              type="button"
              className="text-sm font-semibold text-violet-brand transition hover:text-violet-hover"
              onClick={() => setShowDescription(true)}
            >
              + Açıklama ekle
            </button>
          )}
        </div>
      <ModalFooter>
          <Button variant="secondary" className="w-full sm:w-auto" disabled={isSaving} onClick={onClose}>
            Vazgeç
          </Button>
          <PrimaryButton className="w-full sm:w-auto" disabled={isSaving || name.trim() === ""} onClick={handleSubmit}>
            {isSaving ? "Oluşturuluyor..." : "Oluştur"}
          </PrimaryButton>
      </ModalFooter>
    </ModalShell>
  );
}
