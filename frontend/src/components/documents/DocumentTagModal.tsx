import type { ApiDocument } from "../../types/document";
import { DocumentTagsPanel } from "../tags";

type DocumentTagModalProps = {
  document: ApiDocument;
  onClose: () => void;
  onUpdated: (document: ApiDocument) => void;
};

export function DocumentTagModal({ document, onClose, onUpdated }: DocumentTagModalProps) {
  return (
    <div className="fixed inset-0 z-[80] flex items-center justify-center bg-slate-900/35 p-4">
      <div className="w-full max-w-lg rounded-xl border border-line bg-white shadow-panel">
        <div className="flex items-center justify-between border-b border-line px-5 py-4">
          <div>
            <h2 className="text-lg font-bold text-ink">Etiketle</h2>
            <p className="mt-1 truncate text-sm text-slate-500">{document.original_file_name}</p>
          </div>
          <button type="button" className="text-sm font-semibold text-violet-brand" onClick={onClose}>
            Kapat
          </button>
        </div>
        <div className="max-h-[70vh] overflow-y-auto p-2">
          <DocumentTagsPanel
            documentId={document.id}
            selectedTags={document.tags ?? []}
            onUpdated={(tags) => onUpdated({ ...document, tags })}
          />
        </div>
      </div>
    </div>
  );
}
