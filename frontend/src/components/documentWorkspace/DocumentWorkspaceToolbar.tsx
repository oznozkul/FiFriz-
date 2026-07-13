import {
  ChevronRight,
  Download,
  ExternalLink,
  Maximize2,
  Minimize2,
  PanelRightClose,
  PanelRightOpen,
  Share2,
} from "lucide-react";
import { Link } from "react-router-dom";
import { Breadcrumb, Button, IconButton } from "../common";
import { FavoriteStarButton } from "../documents/FavoriteStarButton";
import { FileTypeIcon } from "../common";
import { StatusBadge } from "../common";
import { useBreadcrumbs } from "../../hooks/useBreadcrumbs";
import type { ApiDocument } from "../../types/document";
import { WorkspaceToolbarMoreMenu } from "./WorkspaceToolbarMoreMenu";
import type { WorkspacePanelSection } from "./documentWorkspaceUtils";

type DocumentWorkspaceToolbarProps = {
  document: ApiDocument;
  fileUrl: string;
  isFileLoading: boolean;
  isFavoriteLoading: boolean;
  isPanelOpen: boolean;
  isFullscreen: boolean;
  onToggleFavorite: () => void;
  onShare: () => void;
  onTogglePanel: () => void;
  onToggleFullscreen: () => void;
  onOpenMobilePanel: () => void;
  onFocusSection: (section: WorkspacePanelSection) => void;
};

export function DocumentWorkspaceToolbar({
  document,
  fileUrl,
  isFileLoading,
  isFavoriteLoading,
  isPanelOpen,
  isFullscreen,
  onToggleFavorite,
  onShare,
  onTogglePanel,
  onToggleFullscreen,
  onOpenMobilePanel,
  onFocusSection,
}: DocumentWorkspaceToolbarProps) {
  const breadcrumbs = useBreadcrumbs();
  const fileReady = Boolean(fileUrl) && !isFileLoading;

  return (
    <header className="shrink-0 border-b border-line bg-white">
      <div className="flex min-h-[3.75rem] flex-wrap items-center justify-between gap-3 px-4 py-3 lg:px-5">
        <div className="flex min-w-0 flex-1 items-start gap-3">
          <FileTypeIcon extension={document.file_extension} size="lg" />
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="truncate text-base font-bold text-ink lg:text-lg">{document.original_file_name}</h1>
              <StatusBadge status={document.lifecycleStatus ?? document.status} />
            </div>
            <div className="mt-1 hidden sm:block">
              <Breadcrumb items={breadcrumbs} maxVisible={4} />
            </div>
            <div className="mt-1 flex items-center gap-1 text-xs font-semibold text-slate-500 sm:hidden">
              <Link className="hover:text-violet-brand" to="/documents">
                Belgeler
              </Link>
              <ChevronRight className="h-3 w-3" />
              <span className="truncate text-ink">{document.original_file_name}</span>
            </div>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-1.5 sm:gap-2">
          <ToolbarAction
            label="Aç"
            icon={<ExternalLink className="h-4 w-4" />}
            href={fileReady ? fileUrl : undefined}
            disabled={!fileReady}
            target="_blank"
          />
          <ToolbarAction
            label="İndir"
            icon={<Download className="h-4 w-4" />}
            href={fileReady ? fileUrl : undefined}
            download={document.original_file_name}
            disabled={!fileReady}
          />
          <Button variant="secondary" className="h-9 px-3 text-xs sm:text-sm" disabled={isFileLoading} onClick={onShare}>
            <Share2 className="h-4 w-4" />
            <span className="hidden sm:inline">Paylaş</span>
          </Button>
          <FavoriteStarButton isFavorite={document.isFavorite ?? false} isLoading={isFavoriteLoading} onToggle={onToggleFavorite} />
          <IconButton
            className="hidden lg:inline-flex"
            aria-label={isPanelOpen ? "Paneli gizle" : "Paneli göster"}
            onClick={onTogglePanel}
          >
            {isPanelOpen ? <PanelRightClose className="h-4 w-4" /> : <PanelRightOpen className="h-4 w-4" />}
          </IconButton>
          <IconButton aria-label={isFullscreen ? "Tam ekrandan çık (Esc)" : "Tam ekran (F)"} onClick={onToggleFullscreen}>
            {isFullscreen ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
          </IconButton>
          <WorkspaceToolbarMoreMenu
            isFullscreen={isFullscreen}
            isPanelOpen={isPanelOpen}
            onToggleFullscreen={onToggleFullscreen}
            onTogglePanel={onTogglePanel}
            onOpenMobilePanel={onOpenMobilePanel}
            onFocusSection={onFocusSection}
          />
        </div>
      </div>
    </header>
  );
}

function ToolbarAction({
  label,
  icon,
  href,
  download,
  disabled,
  target,
}: {
  label: string;
  icon: React.ReactNode;
  href?: string;
  download?: string;
  disabled?: boolean;
  target?: string;
}) {
  if (!href || disabled) {
    return (
      <Button variant="secondary" className="h-9 px-3 text-xs sm:text-sm" disabled>
        {icon}
        <span className="hidden sm:inline">{label}</span>
      </Button>
    );
  }

  return (
    <a
      className="inline-flex h-9 items-center gap-1.5 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-slate-600 transition hover:border-violet-ring hover:bg-violet-soft/40 hover:text-ink focus:outline-none focus:ring-4 focus:ring-violet-ring/70 sm:text-sm"
      href={href}
      download={download}
      target={target}
      rel={target === "_blank" ? "noreferrer" : undefined}
    >
      {icon}
      <span className="hidden sm:inline">{label}</span>
    </a>
  );
}
