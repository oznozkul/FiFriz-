import { ArrowLeft, RefreshCw } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { DocumentTable } from "../components/documents";
import { EmptyState, IconButton, LoadingState, PageTitle } from "../components/common";
import { useDocumentFavorite } from "../hooks/useDocumentFavorite";
import { listSmartCollectionDocuments, listSmartCollections, type SmartCollection } from "../services/smartCollectionService";
import { toast } from "../services/toast";
import type { ApiDocument } from "../types/document";

export function SmartCollectionDetailPage() {
  const { key = "" } = useParams();
  const navigate = useNavigate();
  const [collection, setCollection] = useState<SmartCollection | null>(null);
  const [documents, setDocuments] = useState<ApiDocument[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const { favoriteLoadingId, toggleFavorite } = useDocumentFavorite({
    onUpdated: (updatedDocument) => {
      setDocuments((current) => current.map((item) => (item.id === updatedDocument.id ? updatedDocument : item)));
    },
  });

  const fetchData = async () => {
    if (!key) return;
    setIsLoading(true);
    setLoadError(null);
    try {
      const [collections, items] = await Promise.all([listSmartCollections(), listSmartCollectionDocuments(key)]);
      setCollection(collections.find((item) => item.key === key) ?? null);
      setDocuments(items);
      setSelectedIds([]);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Koleksiyon yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchData();
  }, [key]);

  const title = collection?.title ?? "Akıllı Koleksiyon";
  const description = collection?.description ?? "Filtrelenmiş belge görünümü.";

  const emptyDescription = useMemo(() => "Bu koleksiyonda henüz belge yok.", []);

  if (isLoading) {
    return <LoadingState variant="panel" label="Koleksiyon yükleniyor..." />;
  }

  if (loadError) {
    return (
      <div className="space-y-4">
        <Link to="/collections" className="inline-flex items-center gap-2 text-sm font-semibold text-violet-brand hover:underline">
          <ArrowLeft className="h-4 w-4" />
          Tüm koleksiyonlar
        </Link>
        <EmptyState title="Koleksiyon yüklenemedi" description={loadError} />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-2">
          <Link to="/collections" className="inline-flex items-center gap-2 text-sm font-semibold text-violet-brand hover:underline">
            <ArrowLeft className="h-4 w-4" />
            Akıllı Koleksiyonlar
          </Link>
          <PageTitle title={title} subtitle={description} />
        </div>
        <IconButton aria-label="Yenile" onClick={() => void fetchData()}>
          <RefreshCw className="h-4 w-4" />
        </IconButton>
      </div>

      {documents.length === 0 ? (
        <EmptyState title="Henüz belge yok" description={emptyDescription} />
      ) : (
        <DocumentTable
          documents={documents}
          selectedIds={selectedIds}
          favoriteLoadingId={favoriteLoadingId}
          onEdit={(document) => navigate(`/documents/${document.id}`)}
          onDelete={() => toast.info("Koleksiyondan silme yapılamaz. Belgeyi Belgelerim sayfasından yönetin.")}
          onToggleFavorite={(document) => void toggleFavorite(document)}
          onSelect={(documentId, selected) =>
            setSelectedIds((current) => (selected ? [...new Set([...current, documentId])] : current.filter((id) => id !== documentId)))
          }
          onSelectAll={(selected) =>
            setSelectedIds(selected ? documents.map((document) => document.id) : [])
          }
        />
      )}
    </div>
  );
}
