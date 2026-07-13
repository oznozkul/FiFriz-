import { ChevronDown, type LucideIcon } from "lucide-react";
import { useEffect, useId, useRef, type ReactNode } from "react";

type WorkspaceAccordionSectionProps = {
  id: string;
  title: string;
  icon?: LucideIcon;
  isOpen: boolean;
  onToggle: () => void;
  focusToken?: number;
  children: ReactNode;
};

export function WorkspaceAccordionSection({ id, title, icon: Icon, isOpen, onToggle, focusToken, children }: WorkspaceAccordionSectionProps) {
  const sectionRef = useRef<HTMLElement | null>(null);
  const contentId = useId();

  useEffect(() => {
    if (!focusToken || !isOpen) return;
    sectionRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  }, [focusToken, isOpen]);

  return (
    <section ref={sectionRef} id={id} className="border-b border-line bg-white last:border-b-0">
      <button
        type="button"
        className="flex w-full items-center justify-between gap-3 bg-white px-4 py-3 text-left transition hover:bg-slate-50/80"
        aria-expanded={isOpen}
        aria-controls={contentId}
        onClick={onToggle}
      >
        <div className="flex min-w-0 items-center gap-2.5">
          {Icon ? (
            <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
              <Icon className="h-4 w-4" aria-hidden="true" />
            </span>
          ) : null}
          <span className="text-sm font-bold text-ink">{title}</span>
        </div>
        <ChevronDown className={`h-4 w-4 shrink-0 text-slate-400 transition ${isOpen ? "rotate-180" : ""}`} />
      </button>
      {isOpen ? (
        <div id={contentId} className="border-t border-line bg-white px-4 py-4">
          {children}
        </div>
      ) : null}
    </section>
  );
}
