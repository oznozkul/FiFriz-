import { History, RotateCcw, Upload, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import type { ApiDocument } from "../../types/document";
import { EmptyState, LoadingState, PrimaryButton, Spinner, useConfirm } from "../common";
import {
  buildVersionMenuItems,
  ContextMenuActions,
  getVersionMenuContext,
  useContextMenuActionHandler,
} from "../../contextMenu";
import { useContextMenu } from "../../hooks/useContextMenu";
import { useAuth } from "../../services/AuthContext";
import {
  formatVersionFileSize,
  listDocumentVersions,
  restoreDocumentVersion,
  uploadDocumentVersion,
  type DocumentVersion,
} from "../../services/versionService";
import { toast } from "../../services/toast";

type DocumentVersionsPanelProps = {
  document: ApiDocument;
  onVersionsChanged: () => Promise<void>;
};

export function DocumentVersionsPanel({ document, onVersionsChanged }: DocumentVersionsPanelProps) {
  const { activeTenant } = useAuth();
  const role = activeTenant?.role ?? "member";
  const confirm = useConfirm();
  const { menu, bindContextMenu } = useContextMenu();
  const [versions, setVersions] = useState<DocumentVersion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [changeNote, setChangeNote] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const [restoringVersionId, setRestoringVersionId] = useState<string | null>(null);

  const fetchVersions = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setVersions(await listDocumentVersions(document.id));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Versiyonlar yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchVersions();
  }, [document.id]);

  const handleUpload = async () => {
    if (!uploadFile) {
      toast.error("Lütfen bir dosya seçin.");
      return;
    }
    setIsUploading(true);
    try {
      await uploadDocumentVersion(document.id, uploadFile, changeNote);
      setIsUploadModalOpen(false);
      setUploadFile(null);
      setChangeNote("");
      await fetchVersions();
      await onVersionsChanged();
      toast.success("Yeni versiyon yüklendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Yeni versiyon yüklenemedi.");
    } finally {
      setIsUploading(false);
    }
  };

  const handleRestore = async (version: DocumentVersion) => {
    if (version.isCurrent) return;
    const shouldRestore = await confirm({
      title: "Versiyonu geri yükle",
      message: `Versiyon ${version.versionNumber} aktif versiyon yapılsın mı?`,
      confirmLabel: "Geri yükle",
      variant: "primary",
    });
    if (!shouldRestore) return;

    setRestoringVersionId(version.id);
    try {
      await restoreDocumentVersion(document.id, version.id);
      await fetchVersions();
      await onVersionsChanged();
      toast.success(`Versiyon ${version.versionNumber} geri yüklendi.`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Versiyon geri yüklenemedi.");
    } finally {
      setRestoringVersionId(null);
    }
  };

  useContextMenuActionHandler(
    useCallback(
      (detail) => {
        if (detail.context.entityType !== "version") return;
        const version = detail.context.entity as DocumentVersion;

        switch (detail.action) {
          case ContextMenuActions.version.restore:
            void handleRestore(version);
            return;
          case ContextMenuActions.version.download:
          case ContextMenuActions.version.preview:
            toast.info("Bu özellik yakında eklenecek.");
            return;
        }
      },
      [document.id],
    ),
  );

  return (
    <section className="panel overflow-hidden">
      <div className="flex items-center justify-between border-b border-line px-5 py-4">
        <div>
          <h2 className="flex items-center gap-2 text-base font-bold text-ink">
            <History className="h-4 w-4 text-violet-brand" />
            Versiyonlar
          </h2>
          <p className="mt-1 text-sm text-slate-500">Belge geçmişi ve geri yükleme</p>
        </div>
        <PrimaryButton className="h-9 px-3 text-sm" onClick={() => setIsUploadModalOpen(true)}>
          <Upload className="h-4 w-4" />
          Yeni Versiyon
        </PrimaryButton>
      </div>

      {isLoading ? (
        <div className="p-5">
          <LoadingState label="Versiyonlar yükleniyor..." />
        </div>
      ) : loadError ? (
        <div className="p-5">
          <EmptyState title="Versiyonlar yüklenemedi" description={loadError} action={<PrimaryButton onClick={() => void fetchVersions()}>Tekrar Dene</PrimaryButton>} />
        </div>
      ) : versions.length === 0 ? (
        <div className="p-5">
          <EmptyState title="Henüz versiyon yok" description="İlk versiyon yükleme sırasında otomatik oluşturulur." />
        </div>
      ) : (
        <div className="divide-y divide-line">
          {versions.map((version) => {
            const menuProps = bindContextMenu({
              items: buildVersionMenuItems({ version, role }),
              context: getVersionMenuContext(version, document.id),
            });

            return (
            <div
              key={version.id}
              className="space-y-3 px-5 py-4"
              onContextMenu={menuProps.onContextMenu}
              onTouchStart={menuProps.onTouchStart}
              onTouchMove={menuProps.onTouchMove}
              onTouchEnd={menuProps.onTouchEnd}
              onTouchCancel={menuProps.onTouchCancel}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-bold text-ink">v{version.versionNumber}</span>
                    {version.isCurrent ? (
                      <span className="rounded-full bg-violet-soft px-2 py-0.5 text-xs font-semibold text-violet-brand">Aktif</span>
                    ) : null}
                  </div>
                  <p className="mt-1 truncate text-sm font-semibold text-slate-700">{version.originalFileName}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    {version.uploadedByFullName || "Kullanıcı"} · {formatVersionFileSize(version.fileSize)} · {formatDateTime(version.createdAt)}
                  </p>
                  {version.changeNote ? <p className="mt-2 text-sm text-slate-600">{version.changeNote}</p> : null}
                </div>
                {!version.isCurrent ? (
                  <button
                    type="button"
                    className="inline-flex h-9 shrink-0 items-center gap-1 rounded-lg border border-line px-3 text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand disabled:opacity-60"
                    disabled={restoringVersionId === version.id}
                    onClick={() => void handleRestore(version)}
                  >
                    {restoringVersionId === version.id ? <Spinner size="sm" /> : <RotateCcw className="h-4 w-4" />}
                    Geri Yükle
                  </button>
                ) : null}
              </div>
            </div>
            );
          })}
        </div>
      )}

      {menu}

      {isUploadModalOpen ? (
        <div className="fixed inset-0 z-[95] flex items-center justify-center bg-slate-900/40 p-4">
          <div className="w-full max-w-lg rounded-xl border border-line bg-white shadow-panel">
            <div className="flex items-center justify-between border-b border-line px-5 py-4">
              <h3 className="text-lg font-bold text-ink">Yeni Versiyon Yükle</h3>
              <button type="button" className="rounded-lg p-2 text-slate-500 hover:bg-slate-100" onClick={() => setIsUploadModalOpen(false)}>
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="space-y-4 p-5">
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">Dosya</span>
                <input
                  className="block w-full text-sm text-slate-600 file:mr-3 file:rounded-lg file:border-0 file:bg-violet-soft file:px-3 file:py-2 file:text-sm file:font-semibold file:text-violet-brand"
                  type="file"
                  accept=".pdf,.png,.jpg,.jpeg,.docx,.xlsx"
                  onChange={(event) => setUploadFile(event.target.files?.[0] ?? null)}
                />
              </label>
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">Değişiklik notu (opsiyonel)</span>
                <textarea
                  className="min-h-24 w-full rounded-lg border border-line px-3 py-2 outline-none focus:border-violet-brand"
                  placeholder="Bu versiyonda ne değişti?"
                  value={changeNote}
                  onChange={(event) => setChangeNote(event.target.value)}
                />
              </label>
              <div className="flex justify-end gap-2">
                <button type="button" className="h-10 rounded-lg border border-line px-4 text-sm font-semibold text-slate-600" onClick={() => setIsUploadModalOpen(false)}>
                  İptal
                </button>
                <PrimaryButton disabled={isUploading || !uploadFile} onClick={() => void handleUpload()}>
                  {isUploading ? <Spinner size="sm" label="Yükleniyor..." className="text-white" /> : "Yükle"}
                </PrimaryButton>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
