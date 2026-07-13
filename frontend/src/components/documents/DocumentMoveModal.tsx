import { useEffect, useState } from "react";
import type { ApiDocument } from "../../types/document";
import type { Folder } from "../../types/folder";
import { updateDocument } from "../../data/documentsApi";
import { listFolders } from "../../services/folderService";
import { toast } from "../../services/toast";
import { LoadingState, PrimaryButton } from "../common";

type DocumentMoveModalProps = {
  document: ApiDocument;
  onClose: () => void;
  onMoved: (document: ApiDocument) => void;
};

export function DocumentMoveModal({ document, onClose, onMoved }: DocumentMoveModalProps) {
  const [folders, setFolders] = useState<Folder[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedFolderId, setSelectedFolderId] = useState(document.folder_id ?? "root");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    listFolders()
      .then((items) => {
        if (!cancelled) setFolders(items);
      })
      .catch(() => {
        if (!cancelled) setFolders([]);
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const updated = await updateDocument(document.id, {
        original_file_name: document.original_file_name,
        folder_id: selectedFolderId === "root" ? "root" : selectedFolderId,
      });
      onMoved(updated);
      toast.success("Belge taşındı.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Belge taşınamadı.");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[80] flex items-center justify-center bg-slate-900/35 p-4">
      <div className="w-full max-w-md rounded-xl border border-line bg-white shadow-panel">
        <div className="border-b border-line px-5 py-4">
          <h2 className="text-lg font-bold text-ink">Belgeyi Taşı</h2>
          <p className="mt-1 truncate text-sm text-slate-500">{document.original_file_name}</p>
        </div>
        <div className="p-5">
          {isLoading ? (
            <LoadingState variant="inline" label="Klasörler yükleniyor..." />
          ) : (
            <label className="block space-y-2 text-sm">
              <span className="font-semibold text-slate-600">Hedef klasör</span>
              <select
                className="h-11 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
                value={selectedFolderId}
                onChange={(event) => setSelectedFolderId(event.target.value)}
              >
                <option value="root">Klasörsüz (kök)</option>
                {flattenFolders(folders).map((folder) => (
                  <option key={folder.id} value={folder.id}>
                    {folder.label}
                  </option>
                ))}
              </select>
            </label>
          )}
        </div>
        <div className="flex justify-end gap-3 border-t border-line px-5 py-4">
          <button className="h-10 rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-600" type="button" disabled={isSaving} onClick={onClose}>
            Vazgeç
          </button>
          <PrimaryButton disabled={isSaving || isLoading} onClick={() => void handleSave()}>
            {isSaving ? "Taşınıyor..." : "Taşı"}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}

function flattenFolders(folders: Folder[], depth = 0): Array<{ id: string; label: string }> {
  return folders.flatMap((folder) => [
    { id: folder.id, label: `${"  ".repeat(depth)}${depth > 0 ? "- " : ""}${folder.name}` },
    ...flattenFolders(folder.children ?? [], depth + 1),
  ]);
}
