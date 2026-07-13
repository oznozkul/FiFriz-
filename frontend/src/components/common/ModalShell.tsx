import type { ReactNode } from "react";

export const modalOverlayClass =
  "fixed inset-0 z-[90] flex items-end justify-center bg-slate-900/40 backdrop-blur-[1px] sm:items-center sm:p-4";

export const modalPanelBaseClass =
  "flex max-h-[100dvh] w-full flex-col overflow-hidden rounded-t-2xl border border-line bg-white shadow-lift sm:max-h-[min(90dvh,900px)] sm:rounded-2xl sm:border";

export const modalHeaderClass =
  "flex shrink-0 items-start justify-between gap-3 border-b border-line px-4 py-4 sm:px-5";

export const modalBodyClass = "min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5";

export const modalFooterClass =
  "flex shrink-0 flex-col-reverse gap-2 border-t border-line bg-slate-50/70 px-4 py-4 sm:flex-row sm:items-center sm:justify-end sm:gap-3 sm:px-5";

export const modalSizeClass = {
  sm: "sm:max-w-sm",
  md: "sm:max-w-md",
  lg: "sm:max-w-lg",
  xl: "sm:max-w-xl",
  "2xl": "sm:max-w-2xl",
  "3xl": "sm:max-w-3xl",
} as const;

type ModalShellProps = {
  children: ReactNode;
  onClose: () => void;
  size?: keyof typeof modalSizeClass;
  zIndexClass?: string;
  className?: string;
  ariaLabel?: string;
};

export function ModalShell({
  children,
  onClose,
  size = "md",
  zIndexClass = "z-[90]",
  className = "",
  ariaLabel,
}: ModalShellProps) {
  return (
    <div className={`${modalOverlayClass} ${zIndexClass}`} onClick={onClose}>
      <div
        className={`${modalPanelBaseClass} ${modalSizeClass[size]} ${className}`}
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        onClick={(event) => event.stopPropagation()}
      >
        {children}
      </div>
    </div>
  );
}

type ModalFooterProps = {
  children: ReactNode;
  className?: string;
};

export function ModalFooter({ children, className = "" }: ModalFooterProps) {
  return <div className={`${modalFooterClass} ${className}`}>{children}</div>;
}
