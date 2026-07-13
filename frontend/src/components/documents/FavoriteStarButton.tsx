import { Star } from "lucide-react";
import { Spinner } from "../common/Spinner";

type FavoriteStarButtonProps = {
  isFavorite: boolean;
  isLoading?: boolean;
  onToggle: () => void;
  className?: string;
  label?: string;
};

export function FavoriteStarButton({ isFavorite, isLoading = false, onToggle, className = "", label }: FavoriteStarButtonProps) {
  const ariaLabel = label ?? (isFavorite ? "Favoriden çıkar" : "Favoriye ekle");

  return (
    <button
      type="button"
      className={`inline-flex h-8 w-8 items-center justify-center rounded-lg transition hover:bg-amber-50 focus:outline-none focus:ring-4 focus:ring-amber-200/70 disabled:cursor-not-allowed disabled:opacity-60 ${className}`}
      aria-label={ariaLabel}
      aria-pressed={isFavorite}
      title={ariaLabel}
      disabled={isLoading}
      onClick={(event) => {
        event.stopPropagation();
        onToggle();
      }}
    >
      {isLoading ? (
        <Spinner size="sm" className="text-amber-400" />
      ) : (
        <Star className={`h-4 w-4 ${isFavorite ? "fill-amber-400 text-amber-400" : "text-slate-300 hover:text-amber-300"}`} />
      )}
    </button>
  );
}
