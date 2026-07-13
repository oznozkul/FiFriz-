import { ChevronLeft, ChevronRight } from "lucide-react";

const DEFAULT_PAGE_SIZE_OPTIONS = [10, 20, 50, 100];

type TablePaginationProps = {
  page: number;
  pageSize: number;
  totalCount: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  pageSizeOptions?: number[];
  itemLabel?: string;
};

export function TablePagination({
  page,
  pageSize,
  totalCount,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = DEFAULT_PAGE_SIZE_OPTIONS,
  itemLabel = "kayıt",
}: TablePaginationProps) {
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
  const safePage = Math.min(page, totalPages);
  const rangeStart = totalCount === 0 ? 0 : (safePage - 1) * pageSize + 1;
  const rangeEnd = Math.min(safePage * pageSize, totalCount);

  return (
    <div className="flex flex-col gap-3 border-t border-line bg-slate-50/40 px-5 py-4 text-sm text-slate-500 sm:flex-row sm:items-center sm:justify-between">
      <span>
        {totalCount === 0 ? (
          <>Toplam 0 {itemLabel}</>
        ) : (
          <>
            {rangeStart}–{rangeEnd} / {totalCount} {itemLabel} · Sayfa {safePage}/{totalPages}
          </>
        )}
      </span>

      <div className="flex flex-wrap items-center gap-3">
        <label className="flex items-center gap-2 text-sm font-semibold text-slate-600">
          <span className="whitespace-nowrap">Sayfa başına</span>
          <select
            className="h-9 rounded-lg border border-line bg-white px-2 text-sm font-semibold text-slate-600 outline-none transition focus:border-violet-brand focus:ring-4 focus:ring-violet-ring/70"
            value={pageSize}
            onChange={(event) => onPageSizeChange(Number(event.target.value))}
          >
            {pageSizeOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>

        <div className="flex items-center gap-2">
          <button
            className="inline-flex h-9 items-center gap-1 rounded-lg border border-line bg-white px-3 font-semibold text-slate-600 transition hover:border-violet-ring hover:bg-violet-soft/40 active:scale-[0.98] focus:outline-none focus:ring-4 focus:ring-violet-ring/70 disabled:cursor-not-allowed disabled:opacity-50"
            type="button"
            disabled={safePage <= 1}
            onClick={() => onPageChange(safePage - 1)}
          >
            <ChevronLeft className="h-4 w-4" />
            Önceki
          </button>
          <button
            className="inline-flex h-9 items-center gap-1 rounded-lg border border-line bg-white px-3 font-semibold text-slate-600 transition hover:border-violet-ring hover:bg-violet-soft/40 active:scale-[0.98] focus:outline-none focus:ring-4 focus:ring-violet-ring/70 disabled:cursor-not-allowed disabled:opacity-50"
            type="button"
            disabled={safePage >= totalPages}
            onClick={() => onPageChange(safePage + 1)}
          >
            Sonraki
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
