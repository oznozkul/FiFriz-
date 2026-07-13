import { ChevronRight } from "lucide-react";
import { Link } from "react-router-dom";

export type BreadcrumbItem = {
  label: string;
  href?: string;
};

type BreadcrumbProps = {
  items: BreadcrumbItem[];
  maxVisible?: number;
  className?: string;
};

export function Breadcrumb({ items, maxVisible = 4, className = "" }: BreadcrumbProps) {
  if (items.length === 0) return null;

  const shouldCollapse = items.length > maxVisible;
  const visibleItems = shouldCollapse
    ? [items[0], { label: "...", href: undefined }, ...items.slice(-(maxVisible - 2))]
    : items;

  return (
    <nav aria-label="Breadcrumb" className={`min-w-0 ${className}`}>
      <ol className="flex min-w-0 items-center gap-1 text-sm">
        {visibleItems.map((item, index) => {
          const isLast = index === visibleItems.length - 1;
          const isEllipsis = item.label === "...";

          return (
            <li key={`${item.label}-${index}`} className="flex min-w-0 items-center gap-1">
              {index > 0 ? <ChevronRight className="h-3.5 w-3.5 shrink-0 text-slate-300" aria-hidden="true" /> : null}
              {isEllipsis ? (
                <span className="px-1 font-semibold text-slate-400">...</span>
              ) : item.href && !isLast ? (
                <Link to={item.href} className="truncate font-semibold text-slate-500 transition hover:text-violet-brand">
                  {item.label}
                </Link>
              ) : (
                <span className={`truncate font-semibold ${isLast ? "text-ink" : "text-slate-500"}`} aria-current={isLast ? "page" : undefined}>
                  {item.label}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
