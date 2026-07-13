import { Link } from "react-router-dom";
import type { DashboardTopFolder } from "../../services/dashboardService";
import { DashboardPanelCard } from "./DashboardPanelCard";

type TopFoldersPanelProps = {
  folders: DashboardTopFolder[];
  title: string;
  subtitle: string;
  emptyTitle: string;
  emptyDescription: string;
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
};

export function TopFoldersPanel({
  folders,
  title,
  subtitle,
  emptyTitle,
  emptyDescription,
  isLoading = false,
  error = null,
  onRetry,
}: TopFoldersPanelProps) {
  return (
    <DashboardPanelCard
      title={title}
      subtitle={subtitle}
      viewAllHref="/folders"
      isLoading={isLoading}
      error={error}
      onRetry={onRetry}
      isEmpty={!isLoading && !error && folders.length === 0}
      emptyTitle={emptyTitle}
      emptyDescription={emptyDescription}
    >
      <div className="divide-y divide-line">
        {folders.map((folder) => (
          <Link
            key={folder.folderId}
            to={`/folders?folderId=${folder.folderId}`}
            className="flex items-center justify-between gap-3 px-5 py-4 transition hover:bg-violet-soft/35"
          >
            <span className="truncate font-semibold text-ink">{folder.folderName}</span>
            <span className="shrink-0 rounded-full bg-violet-soft px-3 py-1 text-sm font-bold text-violet-brand">
              {folder.documentCount}
            </span>
          </Link>
        ))}
      </div>
    </DashboardPanelCard>
  );
}
