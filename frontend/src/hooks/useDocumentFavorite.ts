import { useCallback, useState } from "react";
import { favoriteDocument, unfavoriteDocument } from "../data/documentsApi";
import { toast } from "../services/toast";
import type { ApiDocument } from "../types/document";

type UseDocumentFavoriteOptions = {
  onUpdated?: (document: ApiDocument) => void;
  onError?: (message: string) => void;
  silent?: boolean;
};

export function useDocumentFavorite(options: UseDocumentFavoriteOptions = {}) {
  const { onUpdated, onError, silent = false } = options;
  const [favoriteLoadingId, setFavoriteLoadingId] = useState<string | null>(null);

  const toggleFavorite = useCallback(
    async (document: Pick<ApiDocument, "id" | "isFavorite">) => {
      setFavoriteLoadingId(document.id);

      try {
        const updatedDocument = document.isFavorite ? await unfavoriteDocument(document.id) : await favoriteDocument(document.id);
        onUpdated?.(updatedDocument);
        if (!silent) {
          toast.success(updatedDocument.isFavorite ? "Favorilere eklendi." : "Favorilerden çıkarıldı.");
        }
        window.dispatchEvent(new CustomEvent("favorites:changed"));
        return updatedDocument;
      } catch (error) {
        const message = error instanceof Error ? error.message : "Favori durumu güncellenemedi.";
        onError?.(message);
        toast.error(message);
        throw error;
      } finally {
        setFavoriteLoadingId(null);
      }
    },
    [onError, onUpdated, silent],
  );

  return {
    favoriteLoadingId,
    toggleFavorite,
    isFavoriteLoading: (documentId: string) => favoriteLoadingId === documentId,
  };
}
