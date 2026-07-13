import { useEffect, useState } from "react";
import { LoadingState } from "../common";
import { formatFileSize } from "../documentWorkspace/documentWorkspaceUtils";
import { getFolderMetadata } from "../../services/folderService";
import type { FolderMetadata } from "../../types/folder";

type FolderMetadataPanelProps = {
  folderId: string;
};

export function FolderMetadataPanel({ folderId }: FolderMetadataPanelProps) {
  const [metadata, setMetadata] = useState<FolderMetadata | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    setIsLoading(true);
    setError(null);
    void getFolderMetadata(folderId)
      .then((data) => {
        if (active) setMetadata(data);
      })
      .catch((fetchError) => {
        if (active) setError(fetchError instanceof Error ? fetchError.message : "Metadata yüklenemedi");
      })
      .finally(() => {
        if (active) setIsLoading(false);
      });
    return () => {
      active = false;
    };
  }, [folderId]);

  if (isLoading) {
    return <LoadingState variant="inline" label="Klasör bilgileri yükleniyor..." />;
  }
  if (error || !metadata) {
    return <p className="text-sm text-rose-600">{error ?? "Metadata bulunamadı"}</p>;
  }

  return (
    <dl className="grid gap-3 text-sm sm:grid-cols-2">
      <MetaItem label="Belge sayısı" value={String(metadata.documentCount)} />
      <MetaItem label="Alt klasör" value={String(metadata.subFolderCount)} />
      <MetaItem label="Toplam boyut" value={formatFileSize(metadata.totalSize)} />
      <MetaItem label="Depolama" value={formatFileSize(metadata.storageUsage)} />
      <MetaItem label="AI indeks" value={metadata.aiIndexed ? "Evet" : "Hayır"} />
      <MetaItem
        label="Son aktivite"
        value={metadata.lastActivityAt ? new Date(metadata.lastActivityAt).toLocaleString("tr-TR") : "-"}
      />
    </dl>
  );
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-line bg-slate-50 px-3 py-2">
      <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</dt>
      <dd className="mt-1 font-semibold text-ink">{value}</dd>
    </div>
  );
}
