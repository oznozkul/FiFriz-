import { getSearchMode, setSearchMode, SEARCH_MODE_LABELS } from "../../utils/searchMode";
import type { SearchMode } from "../../types/hybridSearch";
import { useEffect, useState } from "react";

type SearchModeToggleProps = {
  className?: string;
  onChange?: (mode: SearchMode) => void;
};

export function SearchModeToggle({ className = "", onChange }: SearchModeToggleProps) {
  const [mode, setMode] = useState<SearchMode>(() => getSearchMode());

  useEffect(() => {
    setMode(getSearchMode());
  }, []);

  const updateMode = (next: SearchMode) => {
    setSearchMode(next);
    setMode(next);
    onChange?.(next);
  };

  return (
    <div className={`inline-flex rounded-lg border border-line bg-white p-1 ${className}`}>
      {(["quick", "smart"] as SearchMode[]).map((option) => (
        <button
          key={option}
          type="button"
          className={`rounded-md px-3 py-1.5 text-xs font-semibold transition ${
            mode === option ? "bg-violet-brand text-white" : "text-slate-600 hover:text-violet-brand"
          }`}
          onClick={() => updateMode(option)}
        >
          {SEARCH_MODE_LABELS[option]}
        </button>
      ))}
    </div>
  );
}
