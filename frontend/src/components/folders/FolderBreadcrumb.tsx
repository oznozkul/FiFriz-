import { ChevronRight, Folder as FolderIcon, Home } from "lucide-react";
import { useFolderLabels } from "../../hooks/useFolderLabels";
import type { Folder } from "../../types/folder";

type FolderBreadcrumbProps = {
  path: Folder[];
  showRoot?: boolean;
  onSelect: (folder: Folder | null) => void;
};

export function FolderBreadcrumb({ path, showRoot = false, onSelect }: FolderBreadcrumbProps) {
  const labels = useFolderLabels();

  if (!showRoot && path.length === 0) {
    return (
      <div className="flex items-center gap-2 text-sm font-semibold text-slate-500">
        <FolderIcon className="h-4 w-4" />
        {labels.breadcrumbEmptyLabel}
      </div>
    );
  }

  return (
    <nav
      className="flex min-w-0 flex-wrap items-center gap-1 text-sm font-semibold text-slate-500"
      aria-label={labels.breadcrumbAriaLabel}
    >
      {showRoot && (
        <>
          <button
            className={`inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 transition focus:outline-none focus:ring-4 focus:ring-violet-ring/70 ${
              path.length === 0 ? "bg-violet-soft text-violet-brand" : "hover:bg-violet-soft/50 hover:text-violet-brand"
            }`}
            type="button"
            onClick={() => onSelect(null)}
          >
            <Home className="h-4 w-4" />
            {labels.rootLabel}
          </button>
          {path.length > 0 && <ChevronRight className="h-4 w-4 shrink-0 text-slate-300" />}
        </>
      )}
      {path.map((folder, index) => {
        const isLast = index === path.length - 1;
        return (
          <span key={folder.id} className="flex min-w-0 items-center gap-1">
            {index > 0 && <ChevronRight className="h-4 w-4 shrink-0 text-slate-300" />}
            <button
              className={`max-w-[200px] truncate rounded-lg px-2.5 py-1.5 transition focus:outline-none focus:ring-4 focus:ring-violet-ring/70 ${
                isLast ? "bg-violet-soft font-bold text-violet-brand" : "hover:bg-violet-soft/50 hover:text-violet-brand"
              }`}
              type="button"
              onClick={() => onSelect(folder)}
            >
              {folder.name}
            </button>
          </span>
        );
      })}
    </nav>
  );
}
