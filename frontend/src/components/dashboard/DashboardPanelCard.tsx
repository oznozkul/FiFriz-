import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { EmptyState, ErrorState } from "../common";
import { DashboardListSkeleton } from "./DashboardListSkeleton";

type DashboardPanelCardProps = {
  title: string;
  subtitle?: string;
  viewAllHref?: string;
  viewAllLabel?: string;
  isLoading?: boolean;
  error?: string | null;
  onRetry?: () => void;
  isEmpty?: boolean;
  emptyTitle?: string;
  emptyDescription?: string;
  emptyAction?: ReactNode;
  children: ReactNode;
};

export function DashboardPanelCard({
  title,
  subtitle,
  viewAllHref,
  viewAllLabel = "Tümünü Gör",
  isLoading = false,
  error = null,
  onRetry,
  isEmpty = false,
  emptyTitle = "Veri yok",
  emptyDescription,
  emptyAction,
  children,
}: DashboardPanelCardProps) {
  return (
    <section className="panel flex h-full flex-col overflow-hidden rounded-2xl">
      <div className="flex items-start justify-between gap-3 border-b border-line px-5 py-4">
        <div className="min-w-0">
          <h2 className="text-lg font-bold text-ink">{title}</h2>
          {subtitle ? <p className="mt-1 text-sm leading-relaxed text-slate-500">{subtitle}</p> : null}
        </div>
        {viewAllHref && !isLoading && !error ? (
          <Link className="shrink-0 text-sm font-semibold text-violet-brand transition hover:text-violet-hover" to={viewAllHref}>
            {viewAllLabel} →
          </Link>
        ) : null}
      </div>

      <div className="min-h-0 flex-1">
        {isLoading ? (
          <DashboardListSkeleton />
        ) : error ? (
          <ErrorState title="Yüklenemedi" description={error} onRetry={onRetry} variant="inline" />
        ) : isEmpty ? (
          <EmptyState title={emptyTitle} description={emptyDescription} action={emptyAction} variant="inline" />
        ) : (
          children
        )}
      </div>
    </section>
  );
}
