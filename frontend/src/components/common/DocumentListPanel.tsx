import type { ReactNode } from "react";

type DocumentListPanelProps = {
  children: ReactNode;
  minWidth?: string;
  footer?: ReactNode;
};

export function DocumentListPanel({ children, minWidth = "860px", footer }: DocumentListPanelProps) {
  return (
    <section className="panel overflow-hidden rounded-2xl">
      <div className="overflow-x-auto">
        <div style={{ minWidth }}>{children}</div>
      </div>
      {footer ? <div className="border-t border-line bg-slate-50/40">{footer}</div> : null}
    </section>
  );
}

type DocumentListHeaderProps = {
  columns: string[];
  gridClassName: string;
};

export function DocumentListHeader({ columns, gridClassName }: DocumentListHeaderProps) {
  return (
    <div className={`${gridClassName} border-b border-line px-5 py-3 text-xs font-semibold uppercase tracking-wide text-slate-400`}>
      {columns.map((column) => (
        <span key={column}>{column}</span>
      ))}
    </div>
  );
}
