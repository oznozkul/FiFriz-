import { ArrowRight, RefreshCw, Share2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { DocumentListPanel, FileTypeIcon, IconButton, PageState, PageTitle, PrimaryButton } from "../components/common";
import { internalSharePermissionLabels, listSharedWithMe, type SharedWithMeItem } from "../services/internalShareService";
import { toast } from "../services/toast";

const SHARED_GRID = "grid grid-cols-[36px_minmax(220px,1fr)_180px_140px_150px_140px]";

export function SharedWithMePage() {
  const navigate = useNavigate();
  const [items, setItems] = useState<SharedWithMeItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const fetchItems = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setItems(await listSharedWithMe());
    } catch (error) {
      const message = error instanceof Error ? error.message : "Paylaşılan belgeler yüklenemedi.";
      setLoadError(message);
      setItems([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchItems();
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title="Benimle Paylaşılanlar" subtitle="Tenant üyelerinin sizinle paylaştığı belgeler." />
        <IconButton aria-label="Yenile" onClick={() => void fetchItems()}>
          <RefreshCw className="h-4 w-4" />
        </IconButton>
      </div>

      <PageState
        isLoading={isLoading}
        error={loadError}
        isEmpty={items.length === 0}
        loadingLabel="Paylaşılan belgeler yükleniyor..."
        loadingVariant="table"
        errorTitle="Paylaşılan belgeler yüklenemedi"
        onRetry={() => void fetchItems()}
        emptyTitle="Henüz paylaşılan belge yok"
        emptyDescription="Bir kullanıcı sizinle belge paylaştığında burada görünecek."
        emptyIcon={Share2}
        emptyAction={<PrimaryButton onClick={() => navigate("/documents")}>Belgelerime Git</PrimaryButton>}
      >
        <DocumentListPanel minWidth="900px">
          <div className={`${SHARED_GRID} border-b border-line px-5 py-3 text-xs font-semibold text-slate-400`}>
            <span />
            <span>Belge Adı</span>
            <span>Paylaşan</span>
            <span>Yetki</span>
            <span>Tarih</span>
            <span />
          </div>
          {items.map((item) => (
            <div key={item.shareId} className={`${SHARED_GRID} items-center border-b border-line px-5 py-4 last:border-b-0`}>
              <FileTypeIcon extension={item.documentName.split(".").pop() ?? ""} className="h-5 w-5" />
              <p className="truncate text-sm font-semibold text-ink">{item.documentName}</p>
              <p className="truncate text-sm text-slate-600">{item.sharedByFullName}</p>
              <p className="text-sm font-semibold text-violet-brand">{internalSharePermissionLabels[item.permission]}</p>
              <p className="text-sm text-slate-500">{formatDate(item.createdAt)}</p>
              <PrimaryButton className="h-9 justify-self-end px-3 text-xs" onClick={() => navigate(`/documents/${item.documentId}`)}>
                Belgeye Git
                <ArrowRight className="h-4 w-4" />
              </PrimaryButton>
            </div>
          ))}
        </DocumentListPanel>
      </PageState>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}
