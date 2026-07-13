import { Search } from "lucide-react";
import type { InputHTMLAttributes } from "react";

type SearchInputProps = InputHTMLAttributes<HTMLInputElement> & {
  wrapperClassName?: string;
  shortcut?: string;
};

export function SearchInput({ wrapperClassName = "", shortcut, className = "", ...props }: SearchInputProps) {
  return (
    <div className={`flex h-10 items-center gap-3 rounded-lg border border-line bg-white px-3 text-sm transition focus-within:border-violet-brand focus-within:ring-4 focus-within:ring-violet-ring/70 ${wrapperClassName}`}>
      <Search className="h-4 w-4 shrink-0 text-slate-400" />
      <input className={`min-w-0 flex-1 border-0 bg-transparent text-sm font-medium text-ink outline-none placeholder:font-normal placeholder:text-slate-400 ${className}`} {...props} />
      {shortcut && <kbd className="rounded bg-slate-100 px-2 py-1 text-[11px] font-semibold text-slate-400">{shortcut}</kbd>}
    </div>
  );
}
