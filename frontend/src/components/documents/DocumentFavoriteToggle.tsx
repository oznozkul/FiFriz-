import type { ApiDocument } from "../../types/document";
import { useDocumentFavorite } from "../../hooks/useDocumentFavorite";
import { FavoriteStarButton } from "./FavoriteStarButton";

type DocumentFavoriteToggleProps = {
  document: Pick<ApiDocument, "id" | "isFavorite">;
  onUpdated?: (document: ApiDocument) => void;
  onError?: (message: string) => void;
  className?: string;
};

export function DocumentFavoriteToggle({ document, onUpdated, onError, className }: DocumentFavoriteToggleProps) {
  const { favoriteLoadingId, toggleFavorite } = useDocumentFavorite({ onUpdated, onError });

  return (
    <FavoriteStarButton
      className={className}
      isFavorite={document.isFavorite ?? false}
      isLoading={favoriteLoadingId === document.id}
      onToggle={() => void toggleFavorite(document as ApiDocument)}
    />
  );
}
