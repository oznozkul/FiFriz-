import { useMemo, useState } from "react";
import { AppModal, Button, PrimaryButton } from "../common";
import type { Folder } from "../../types/folder";
import { flattenFolders } from "./folderUtils";

type FolderMoveModalProps = {
  folder: Folder;
  folders: Folder[];
  isSaving: boolean;
  onClose: () => void;
  onMove: (parentId: string | null) => void;
};

export function FolderMoveModal({ folder, folders, isSaving, onClose, onMove }: FolderMoveModalProps) {
  const [parentId, setParentId] = useState<string>("");
  const options = useMemo(() => {
    const flat = flattenFolders(folders);
    return flat.filter((item) => item.id !== folder.id);
  }, [folder.id, folders]);

  return (
    <AppModal title={`"${folder.name}" taşı`} onClose={onClose}>
      <div className="space-y-4">
        <label className="block text-sm font-semibold text-slate-600">
          Hedef konum
          <select
            className="mt-2 w-full rounded-lg border border-line bg-white px-3 py-2 text-sm font-medium text-ink"
            value={parentId}
            onChange={(event) => setParentId(event.target.value)}
          >
            <option value="">Ana klasör (kök)</option>
            {options.map((item) => (
              <option key={item.id} value={item.id}>
                {`${"— ".repeat(item.depth)}${item.name}`}
              </option>
            ))}
          </select>
        </label>
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={onClose}>İptal</Button>
          <PrimaryButton disabled={isSaving} onClick={() => onMove(parentId || null)}>
            Taşı
          </PrimaryButton>
        </div>
      </div>
    </AppModal>
  );
}
