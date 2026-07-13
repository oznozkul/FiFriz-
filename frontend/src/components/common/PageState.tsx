import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { EmptyState } from "./EmptyState";
import { ErrorState } from "./ErrorState";
import { PageLoader } from "./PageLoader";

type PageStateProps = {
  isLoading: boolean;
  error?: string | null;
  isEmpty?: boolean;
  emptyTitle: string;
  emptyDescription?: string;
  emptyAction?: ReactNode;
  emptyIcon?: LucideIcon;
  loadingLabel?: string;
  loadingVariant?: "spinner" | "table" | "cards";
  errorTitle?: string;
  onRetry?: () => void;
  variant?: "page" | "inline";
  className?: string;
  children: ReactNode;
};

export function PageState({
  isLoading,
  error,
  isEmpty = false,
  emptyTitle,
  emptyDescription,
  emptyAction,
  emptyIcon,
  loadingLabel = "Yükleniyor...",
  loadingVariant = "spinner",
  errorTitle = "Veriler yüklenemedi",
  onRetry,
  variant = "page",
  className = "",
  children,
}: PageStateProps) {
  const stateVariant = variant === "inline" ? "inline" : "panel";

  if (isLoading) {
    if (variant === "inline") {
      return <PageLoader label={loadingLabel} variant={loadingVariant} className={`border-0 bg-transparent shadow-none ${className}`} />;
    }
    return <PageLoader label={loadingLabel} variant={loadingVariant} className={className} />;
  }

  if (error) {
    return (
      <ErrorState
        title={errorTitle}
        description={error}
        onRetry={onRetry}
        variant={stateVariant}
        className={className}
      />
    );
  }

  if (isEmpty) {
    return (
      <EmptyState
        title={emptyTitle}
        description={emptyDescription}
        action={emptyAction}
        icon={emptyIcon}
        variant={stateVariant}
        className={className}
      />
    );
  }

  return <>{children}</>;
}
