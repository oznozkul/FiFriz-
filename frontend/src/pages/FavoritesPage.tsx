import { ArrowRight, RefreshCw, Star } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { FavoriteStarButton } from "../components/documents/FavoriteStarButton";
import { formatDateTime } from "../components/documents/documentPresentation";
import { DocumentListPanel, FileTypeIcon, IconButton, PageState, PageTitle, PrimaryButton } from "../components/common";
import { listFavoriteDocuments, unfavoriteDocument } from "../data/documentsApi";
import { toast } from "../services/toast";
import type { ApiDocument } from "../types/document";

const FAVORITES_GRID = "grid grid-cols-[36px_minmax(220px,1fr)_150px_100px_150px_170px_140px]";

export function FavoritesPage() {
  const navigate = useNavigate();
  const [documents, setDocuments] = useState<ApiDocument[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [favoriteLoadingId, setFavoriteLoadingId] = useState<string | null>(null);

  const fetchFavorites = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setDocuments(await listFavoriteDocuments());
    } catch (fetchError) {
      const message = fetchError instanceof Error ? fetchError.message : "Favori belgeler yüklenemedi.";
      setLoadError(message);
      setDocuments([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchFavorites();

    const refresh = () => void fetchFavorites();
    window.addEventListener("favorites:changed", refresh);
    return () => window.removeEventListener("favorites:changed", refresh);
  }, []);

  const handleToggleFavorite = async (document: ApiDocument) => {
    setFavoriteLoadingId(document.id);
    try {
      await unfavoriteDocument(document.id);
      setDocuments((current) => current.filter((item) => item.id !== document.id));
      window.dispatchEvent(new CustomEvent("favorites:changed"));
      toast.success("Favorilerden çıkarıldı.");
    } catch (favoriteError) {
      toast.error(favoriteError instanceof Error ? favoriteError.message : "Favoriden çıkarılamadı.");
    } finally {
      setFavoriteLoadingId(null);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title="Favoriler" subtitle="Önemli belgelerine hızlı erişim için favori listesi." />
        <IconButton aria-label="Yenile" onClick={() => void fetchFavorites()}>
          <RefreshCw className="h-4 w-4" />
        </IconButton>
      </div>

      <PageState
        isLoading={isLoading}
        error={loadError}
        isEmpty={documents.length === 0}
        loadingLabel="Favori belgeler yükleniyor..."
        loadingVariant="table"
        errorTitle="Favoriler yüklenemedi"
        onRetry={() => void fetchFavorites()}
        emptyTitle="Henüz favori belgen yok"
        emptyDescription="Önemli belgelerini yıldızlayarak buraya ekleyebilirsin."
        emptyIcon={Star}
        emptyAction={<PrimaryButton onClick={() => navigate("/documents")}>Belgelerime Git</PrimaryButton>}
      >
        <DocumentListPanel minWidth="960px" footer={`Toplam ${documents.length} favori belge`}>
          <div className={`${FAVORITES_GRID} border-b border-line px-5 py-3 text-xs font-semibold text-slate-400`}>
            <span />
            <span>Belge Adı</span>
            <span>Klasör</span>
            <span>Tip</span>
            <span>Yüklenme Tarihi</span>
            <span>Favoriye Eklenme</span>
            <span />
          </div>
          <div className="divide-y divide-line">
            {documents.map((document) => (
              <div key={document.id} className={`${FAVORITES_GRID} items-center px-5 py-4`}>
                <FavoriteStarButton
                  isFavorite={document.isFavorite}
                  isLoading={favoriteLoadingId === document.id}
                  onToggle={() => void handleToggleFavorite(document)}
                />
                <div className="flex min-w-0 items-center gap-3">
                  <FileTypeIcon extension={document.file_extension} />
                  <span className="truncate font-semibold text-ink">{document.original_file_name}</span>
                </div>
                <span className="truncate text-slate-500">{document.folder_name || "Klasörsüz"}</span>
                <span className="uppercase text-slate-500">{document.file_extension}</span>
                <span className="text-slate-500">{formatDateTime(document.created_at)}</span>
                <span className="text-slate-500">{document.favoritedAt ? formatDateTime(document.favoritedAt) : "-"}</span>
                <div className="flex justify-end">
                  <PrimaryButton className="h-9 px-3 text-xs" onClick={() => navigate(`/documents/${document.id}`)}>
                    Belgeye Git
                    <ArrowRight className="h-4 w-4" />
                  </PrimaryButton>
                </div>
              </div>
            ))}
          </div>
        </DocumentListPanel>
      </PageState>
    </div>
  );
}
