import { X } from "lucide-react";
import type { ReactNode } from "react";
import { IconButton } from "./IconButton";
import { ModalFooter, modalBodyClass, modalHeaderClass, ModalShell } from "./ModalShell";

type AppModalProps = {
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
  size?: "sm" | "md" | "lg" | "xl" | "2xl";
  showCloseButton?: boolean;
};

export function AppModal({
  title,
  description,
  children,
  footer,
  onClose,
  size = "md",
  showCloseButton = true,
}: AppModalProps) {
  return (
    <ModalShell onClose={onClose} size={size} ariaLabel={title}>
      <div className={modalHeaderClass}>
        <div className="min-w-0 pr-2">
          <h2 id="app-modal-title" className="text-lg font-bold text-ink">
            {title}
          </h2>
          {description ? <p className="mt-1 text-sm leading-relaxed text-slate-500">{description}</p> : null}
        </div>
        {showCloseButton ? (
          <IconButton aria-label="Kapat" className="h-8 w-8 shrink-0" onClick={onClose}>
            <X className="h-4 w-4" />
          </IconButton>
        ) : null}
      </div>
      <div className={modalBodyClass}>{children}</div>
      {footer ? <ModalFooter>{footer}</ModalFooter> : null}
    </ModalShell>
  );
}
