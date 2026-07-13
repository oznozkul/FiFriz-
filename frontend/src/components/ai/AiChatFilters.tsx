import { useEffect, useMemo, useState } from "react";
import { listDocuments } from "../../data/documentsApi";
import { listFolders } from "../../services/folderService";
import { listTags } from "../../services/tagService";
import type { AiChatFilterState } from "../../types/aiChat";
import type { ApiDocument } from "../../types/document";
import type { Folder } from "../../types/folder";
import type { Tag } from "../../types/tag";

type AiChatFiltersProps = {
  value: AiChatFilterState;
  onChange: (value: AiChatFilterState) => void;
  disabled?: boolean;
};

function flattenFolders(folders: Folder[], depth = 0): Array<{ id: string; label: string }> {
  const items: Array<{ id: string; label: string }> = [];
  for (const folder of folders) {
    items.push({ id: folder.id, label: `${"— ".repeat(depth)}${folder.name}` });
    if (folder.children?.length) {
      items.push(...flattenFolders(folder.children, depth + 1));
    }
  }
  return items;
}

const emptyFilters: AiChatFilterState = {
  documentId: "",
  folderId: "",
  tagIds: [],
};

export function AiChatFilters({ value, onChange, disabled }: AiChatFiltersProps) {
  const [documents, setDocuments] = useState<ApiDocument[]>([]);
  const [folders, setFolders] = useState<Folder[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;

    const loadFilters = async () => {
      setIsLoading(true);
      try {
        const [documentItems, folderItems, tagItems] = await Promise.all([listDocuments(), listFolders(), listTags()]);
        if (!isMounted) return;
        setDocuments(documentItems);
        setFolders(folderItems);
        setTags(tagItems);
      } catch {
        if (isMounted) {
          setDocuments([]);
          setFolders([]);
          setTags([]);
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    void loadFilters();

    return () => {
      isMounted = false;
    };
  }, []);

  const folderOptions = useMemo(() => flattenFolders(folders), [folders]);

  const selectClassName =
    "h-10 w-full rounded-lg border border-line bg-white px-3 text-sm font-medium text-slate-700 outline-none transition focus:border-violet-brand focus:ring-4 focus:ring-violet-ring/70 disabled:cursor-not-allowed disabled:bg-slate-50";

  return (
    <div className="grid gap-3 border-b border-line bg-white px-4 py-3 sm:grid-cols-3">
      <label className="block space-y-1.5">
        <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Belge</span>
        <select
          className={selectClassName}
          value={value.documentId}
          disabled={disabled || isLoading}
          onChange={(event) => onChange({ ...value, documentId: event.target.value })}
        >
          <option value="">Tüm belgeler</option>
          {documents.map((document) => (
            <option key={document.id} value={document.id}>
              {document.original_file_name}
            </option>
          ))}
        </select>
      </label>

      <label className="block space-y-1.5">
        <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Klasör</span>
        <select
          className={selectClassName}
          value={value.folderId}
          disabled={disabled || isLoading}
          onChange={(event) => onChange({ ...value, folderId: event.target.value })}
        >
          <option value="">Tüm klasörler</option>
          {folderOptions.map((folder) => (
            <option key={folder.id} value={folder.id}>
              {folder.label}
            </option>
          ))}
        </select>
      </label>

      <label className="block space-y-1.5">
        <span className="text-xs font-semibold uppercase tracking-wide text-slate-400">Etiket</span>
        <select
          className={selectClassName}
          value={value.tagIds[0] ?? ""}
          disabled={disabled || isLoading}
          onChange={(event) => {
            const tagId = event.target.value;
            onChange({ ...value, tagIds: tagId ? [tagId] : [] });
          }}
        >
          <option value="">Tüm etiketler</option>
          {tags.map((tag) => (
            <option key={tag.id} value={tag.id}>
              {tag.name}
            </option>
          ))}
        </select>
      </label>

      {(value.documentId || value.folderId || value.tagIds.length > 0) && (
        <div className="sm:col-span-3">
          <button
            type="button"
            className="text-xs font-semibold text-violet-brand transition hover:text-violet-hover disabled:opacity-60"
            disabled={disabled}
            onClick={() => onChange(emptyFilters)}
          >
            Filtreleri temizle
          </button>
        </div>
      )}
    </div>
  );
}
