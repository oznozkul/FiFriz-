import { UploadCloud, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { IconButton, ModalFooter, ModalShell, PrimaryButton, Spinner, FileTypeIcon, extensionFromFileName } from "../common";
import { apiBaseUrl, clearAccessToken, getAccessToken } from "../../services/apiClient";
import { useFolderLabels } from "../../hooks/useFolderLabels";
import { listFolders } from "../../services/folderService";
import { listTags } from "../../services/tagService";
import { toast } from "../../services/toast";
import type { Folder } from "../../types/folder";
import type { Tag } from "../../types/tag";

const EMPTY_TAG_NAMES: string[] = [];

type UploadDocumentModalProps = {
  isOpen: boolean;
  onClose: () => void;
  initialTagNames?: string[];
};

export function UploadDocumentModal({ isOpen, onClose, initialTagNames = EMPTY_TAG_NAMES }: UploadDocumentModalProps) {
  const labels = useFolderLabels();
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [folders, setFolders] = useState<Folder[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [selectedFolderId, setSelectedFolderId] = useState("");
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const [isFoldersLoading, setIsFoldersLoading] = useState(false);
  const [isTagsLoading, setIsTagsLoading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);
  const initialTagKey = useMemo(() => initialTagNames.join("\0"), [initialTagNames]);
  const selectLoadFailed = labels.selectLoadFailed;

  useEffect(() => {
    if (!isOpen) {
      setSelectedTagIds((current) => (current.length === 0 ? current : []));
      return;
    }

    let cancelled = false;
    setIsFoldersLoading(true);
    setIsTagsLoading(true);
    Promise.all([listFolders(), listTags()])
      .then(([folderItems, tagItems]) => {
        if (!cancelled) {
          setFolders(folderItems);
          setTags(tagItems);
          if (initialTagNames.length > 0) {
            const normalized = initialTagNames.map((name) => name.trim().toLowerCase());
            const matchedIds = tagItems
              .filter((tag) => normalized.includes(tag.name.trim().toLowerCase()))
              .map((tag) => tag.id);
            setSelectedTagIds(matchedIds);
          }
        }
      })
      .catch(() => {
        if (!cancelled) {
          setFolders([]);
          setTags([]);
          toast.error(selectLoadFailed);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setIsFoldersLoading(false);
          setIsTagsLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
    // initialTagKey stabilizes array identity; initialTagNames used only when isOpen opens.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- intentional: avoid refetch loop on new [] each render
  }, [isOpen, initialTagKey, selectLoadFailed]);

  if (!isOpen) return null;

  const handleFiles = (files: FileList | null) => {
    setSelectedFiles((current) => [...current, ...Array.from(files ?? [])]);
    setUploadProgress("");
  };

  const handleClose = () => {
    if (isUploading) return;
    setSelectedFiles([]);
    setSelectedFolderId("");
    setSelectedTagIds([]);
    setUploadProgress("");
    onClose();
  };

  const removeFile = (fileIndex: number) => {
    setSelectedFiles((current) => current.filter((_, index) => index !== fileIndex));
  };

  const handleUpload = async () => {
    if (selectedFiles.length === 0 || isUploading) return;

    setIsUploading(true);
    setUploadProgress("");

    const failedFiles: string[] = [];
    const loadingToastId = toast.loading("Belgeler yükleniyor...");
    try {
      for (const [index, file] of selectedFiles.entries()) {
        setUploadProgress(`${index + 1}/${selectedFiles.length} yükleniyor: ${file.name}`);

        const formData = new FormData();
        formData.append("file", file);
        if (selectedFolderId) {
          formData.append("folderId", selectedFolderId);
        }
        if (selectedTagIds.length > 0) {
          formData.append("tagIds", JSON.stringify(selectedTagIds));
        }
        const token = getAccessToken();

        const response = await fetch(`${apiBaseUrl}/api/v1/documents/upload`, {
          method: "POST",
          headers: token ? { Authorization: `Bearer ${token}` } : undefined,
          body: formData,
        });

        const payload = await response.json().catch(() => null);
        if (!response.ok || payload?.success === false) {
          if (response.status === 401 || response.status === 403) {
            clearAccessToken();
            window.dispatchEvent(new CustomEvent("auth:logout"));
          }
          failedFiles.push(`${file.name}: ${payload?.error?.message ?? "Yüklenemedi"}`);
          continue;
        }

        const warnings = payload?.data?.warnings as string[] | undefined;
        const duplicateOf = payload?.data?.duplicateOf as { originalFileName?: string } | undefined;
        if (warnings?.length) {
          toast.info(warnings.join(" "));
        }
        if (duplicateOf?.originalFileName) {
          toast.info(`Benzer checksum: ${duplicateOf.originalFileName}`);
        }
      }

      if (failedFiles.length > 0) {
        toast.dismiss(loadingToastId);
        toast.error(failedFiles.join(" | "));
        if (failedFiles.length < selectedFiles.length) {
          window.dispatchEvent(new CustomEvent("documents:changed"));
        }
        return;
      }

      toast.dismiss(loadingToastId);
      toast.success(selectedFiles.length > 1 ? `${selectedFiles.length} belge yüklendi.` : "Belge yüklendi.");
      setSelectedFiles([]);
      setSelectedFolderId("");
      setUploadProgress("");
      window.dispatchEvent(new CustomEvent("documents:changed"));
      onClose();
    } catch (error) {
      toast.dismiss(loadingToastId);
      toast.error(error instanceof Error ? error.message : "Belgeler yüklenemedi.");
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <ModalShell onClose={handleClose} size="xl" zIndexClass="z-[70]" ariaLabel="Belge yükle">
      <div className="flex items-center justify-between border-b border-line px-4 py-4 sm:px-5">
        <div className="min-w-0 pr-3">
          <h2 className="text-lg font-bold text-ink">Belge Yükle</h2>
          <p className="mt-1 text-sm text-slate-500">{labels.uploadSubtitle}</p>
        </div>
        <IconButton className="h-8 w-8 shrink-0" aria-label="Kapat" disabled={isUploading} onClick={handleClose}>
          <X className="h-4 w-4" />
        </IconButton>
      </div>

      <div className="min-h-0 flex-1 space-y-5 overflow-y-auto p-4 sm:p-5">
          <label
            className="flex min-h-44 cursor-pointer flex-col items-center justify-center rounded-lg border border-dashed border-violet-brand bg-violet-soft/50 px-6 py-8 text-center transition hover:bg-violet-soft"
            onDragOver={(event) => event.preventDefault()}
            onDrop={(event) => {
              event.preventDefault();
              if (isUploading) return;
              handleFiles(event.dataTransfer.files);
            }}
          >
            <input
              ref={fileInputRef}
              className="hidden"
              type="file"
              accept=".pdf,.png,.jpg,.jpeg,.docx,.xlsx"
              multiple
              disabled={isUploading}
              onChange={(event) => handleFiles(event.target.files)}
            />
            <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-white text-violet-brand shadow-sm">
              <UploadCloud className="h-6 w-6" />
            </div>
            <p className="font-semibold text-ink">Dosyaları buraya sürükleyip bırakın</p>
            <p className="mt-2 text-sm text-slate-500">Desteklenen formatlar: PDF, PNG, JPG, JPEG, DOCX, XLSX</p>
            <button
              className="mt-4 rounded-lg border border-line bg-white px-4 py-2 text-sm font-semibold text-violet-brand transition hover:border-violet-brand disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              disabled={isUploading}
              onClick={() => fileInputRef.current?.click()}
            >
              Dosya Seç
            </button>
          </label>

          {selectedFiles.length > 0 && (
            <div className="rounded-lg border border-line bg-slate-50">
              <div className="flex items-center justify-between border-b border-line px-3 py-2 text-sm">
                <span className="font-semibold text-ink">{selectedFiles.length} dosya seçildi</span>
                <button className="font-semibold text-violet-brand hover:text-violet-hover disabled:cursor-not-allowed disabled:opacity-60" type="button" disabled={isUploading} onClick={() => setSelectedFiles([])}>
                  Temizle
                </button>
              </div>
              <div className="max-h-44 divide-y divide-line overflow-y-auto">
                {selectedFiles.map((file, index) => (
                  <div key={`${file.name}-${file.lastModified}-${index}`} className="flex items-center gap-3 p-3">
                    <div className="flex h-9 w-9 items-center justify-center">
                      <FileTypeIcon extension={extensionFromFileName(file.name)} size="lg" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-semibold text-ink">{file.name}</p>
                      <p className="text-xs text-slate-500">{formatFileSize(file.size)}</p>
                    </div>
                    <button className="rounded px-2 py-1 text-xs font-semibold text-slate-500 transition hover:bg-white hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-60" type="button" disabled={isUploading} onClick={() => removeFile(index)}>
                      Kaldır
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {uploadProgress && (
            <div className="flex items-center gap-2 rounded-lg border border-violet-ring bg-violet-soft/70 px-3 py-2 text-sm font-semibold text-violet-brand">
              <Spinner size="sm" />
              {uploadProgress}
            </div>
          )}

          <div className="grid gap-4 sm:grid-cols-2">
            <label className="space-y-2 sm:col-span-2">
              <span className="text-sm font-semibold text-slate-600">Etiketler (opsiyonel)</span>
              {isTagsLoading ? (
                <Spinner size="sm" label="Etiketler yükleniyor..." />
              ) : tags.length === 0 ? (
                <p className="text-xs text-slate-500">Henüz etiket yok.</p>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {tags.map((tag) => {
                    const isSelected = selectedTagIds.includes(tag.id);
                    const color = tag.color || "#7c3aed";
                    return (
                      <button
                        key={tag.id}
                        className={`rounded-full border px-3 py-1.5 text-xs font-semibold transition ${isSelected ? "border-transparent text-white" : "border-line bg-white text-slate-600 hover:border-violet-brand hover:text-violet-brand"}`}
                        style={isSelected ? { backgroundColor: color } : undefined}
                        type="button"
                        disabled={isUploading}
                        onClick={() =>
                          setSelectedTagIds((current) => (current.includes(tag.id) ? current.filter((id) => id !== tag.id) : [...current, tag.id]))
                        }
                      >
                        {tag.name}
                      </button>
                    );
                  })}
                </div>
              )}
            </label>
            <label className="space-y-2 sm:col-span-2">
              <span className="text-sm font-semibold text-slate-600">{labels.selectLabel} seç</span>
              <select className="h-11 w-full rounded-lg border border-line bg-white px-3 text-sm outline-none focus:border-violet-brand" value={selectedFolderId} disabled={isUploading} onChange={(event) => setSelectedFolderId(event.target.value)}>
                <option value="">{labels.noneLabel}</option>
                {flattenFolders(folders).map((folder) => (
                  <option key={folder.id} value={folder.id}>
                    {folder.label}
                  </option>
                ))}
              </select>
              {isFoldersLoading && <Spinner size="sm" label={labels.loadingLabel} />}
            </label>
          </div>
        </div>

      <ModalFooter>
          <button
            className="h-10 w-full rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-600 transition hover:border-violet-ring hover:bg-violet-soft/40 hover:text-violet-brand disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
            type="button"
            disabled={isUploading}
            onClick={handleClose}
          >
            Vazgeç
          </button>
          <PrimaryButton
            disabled={selectedFiles.length === 0 || isUploading}
            className={`w-full sm:w-auto ${selectedFiles.length === 0 ? "cursor-not-allowed opacity-60 hover:bg-violet-brand" : ""}`}
            onClick={handleUpload}
          >
            {isUploading ? <Spinner size="sm" label="Yükleniyor..." className="text-white" /> : selectedFiles.length > 1 ? `${selectedFiles.length} Dosyayı Yükle` : "Yükle"}
          </PrimaryButton>
      </ModalFooter>
    </ModalShell>
  );
}

function formatFileSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function flattenFolders(folders: Folder[], depth = 0): Array<{ id: string; label: string }> {
  return folders.flatMap((folder) => [
    { id: folder.id, label: `${"  ".repeat(depth)}${depth > 0 ? "- " : ""}${folder.name}` },
    ...flattenFolders(folder.children ?? [], depth + 1),
  ]);
}
