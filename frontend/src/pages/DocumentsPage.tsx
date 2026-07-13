import { Grid3X3, List, RefreshCw, RotateCcw, Search, SlidersHorizontal, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { DocumentTable } from "../components/documents";
import { IconButton, PageTitle, PrimaryButton, SearchInput, Spinner, useConfirm } from "../components/common";
import { useContextMenuActionHandler } from "../contextMenu";
import { deleteDocument, listDocuments, listTrashDocuments, permanentlyDeleteDocument, restoreDocument, updateDocument } from "../data/documentsApi";
import { useDebounce } from "../hooks/useDebounce";
import { useDocumentFavorite } from "../hooks/useDocumentFavorite";
import { useDocumentMenuActions } from "../hooks/useDocumentMenuActions";
import { hybridSearchDocuments } from "../services/hybridSearchService";
import { searchDocuments } from "../services/documentService";
import { SearchModeToggle } from "../components/search/SearchModeToggle";
import { getSearchMode } from "../utils/searchMode";
import type { SearchMode } from "../types/hybridSearch";
import { listFolders } from "../services/folderService";
import { listTags } from "../services/tagService";
import { toast } from "../services/toast";
import type { ApiDocument, DocumentStatus } from "../types/document";
import type { Folder } from "../types/folder";
import type { Tag } from "../types/tag";
import { useNavigate } from "react-router-dom";

const statusOptions: Array<{ label: string; value: "all" | DocumentStatus }> = [
  { label: "Tümü", value: "all" },
  { label: "Yükleniyor", value: "uploading" },
  { label: "İşleniyor", value: "processing" },
  { label: "Hazır", value: "ready" },
  { label: "Başarısız", value: "failed" },
  { label: "Kilitli", value: "locked" },
];

const fileTypeOptions = ["all", "pdf", "png", "jpg", "jpeg", "docx", "xlsx"];

const PAGE_SIZE_STORAGE_KEY = "documentsPageSize";
const DEFAULT_PAGE_SIZE = 20;

function readStoredPageSize() {
  if (typeof window === "undefined") return DEFAULT_PAGE_SIZE;
  const stored = Number(window.localStorage.getItem(PAGE_SIZE_STORAGE_KEY));
  return [10, 20, 50, 100].includes(stored) ? stored : DEFAULT_PAGE_SIZE;
}

type DocumentsPageProps = {
  mode?: "documents" | "trash";
};

export function DocumentsPage({ mode = "documents" }: DocumentsPageProps) {
  const isTrash = mode === "trash";
  const navigate = useNavigate();
  const [documents, setDocuments] = useState<ApiDocument[]>([]);
  const [folders, setFolders] = useState<Folder[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isSearchLoading, setIsSearchLoading] = useState(false);
  const [query, setQuery] = useState("");
  const [searchMode, setSearchMode] = useState<SearchMode>(() => getSearchMode());
  const [contentSearchIds, setContentSearchIds] = useState<string[]>([]);
  const [status, setStatus] = useState<"all" | DocumentStatus>("all");
  const [fileType, setFileType] = useState("all");
  const [folderId, setFolderId] = useState("");
  const [tagId, setTagId] = useState("");
  const [editingDocument, setEditingDocument] = useState<ApiDocument | null>(null);
  const [editName, setEditName] = useState("");
  const [editStatus, setEditStatus] = useState<DocumentStatus>("ready");
  const [isSaving, setIsSaving] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(readStoredPageSize);
  const debouncedQuery = useDebounce(query, 400);
  const confirm = useConfirm();
  const { favoriteLoadingId, toggleFavorite } = useDocumentFavorite({
    onUpdated: (updatedDocument) => {
      setDocuments((current) => current.map((item) => (item.id === updatedDocument.id ? updatedDocument : item)));
    },
  });

  const fetchDocuments = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setDocuments(isTrash ? await listTrashDocuments() : await listDocuments(folderId || undefined, tagId || undefined));
      setSelectedIds([]);
    } catch (fetchError) {
      const message = fetchError instanceof Error ? fetchError.message : "Belgeler yüklenemedi.";
      setLoadError(message);
      setDocuments([]);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchDocuments();

    const refresh = () => void fetchDocuments();
    window.addEventListener("documents:changed", refresh);
    return () => window.removeEventListener("documents:changed", refresh);
  }, [folderId, tagId, isTrash]);

  useEffect(() => {
    if (isTrash) return;
    const hasProcessing = documents.some(
      (document) =>
        document.status === "processing" ||
        document.status === "uploading" ||
        document.status === "uploaded",
    );
    if (!hasProcessing) return;

    const intervalId = window.setInterval(() => {
      void fetchDocuments();
    }, 5000);

    return () => window.clearInterval(intervalId);
  }, [documents, isTrash]);

  useEffect(() => {
    if (isTrash) return;
    let cancelled = false;
    listFolders()
      .then((items) => {
        if (!cancelled) setFolders(items);
      })
      .catch(() => {
        if (!cancelled) setFolders([]);
      });
    listTags()
      .then((items) => {
        if (!cancelled) setTags(items);
      })
      .catch(() => {
        if (!cancelled) setTags([]);
      });
    return () => {
      cancelled = true;
    };
  }, [isTrash]);

  useEffect(() => {
    const trimmedQuery = debouncedQuery.trim();
    if (isTrash || trimmedQuery.length < 2) {
      setContentSearchIds([]);
      setIsSearchLoading(false);
      return;
    }

    let cancelled = false;
    setIsSearchLoading(true);

    const searchPromise =
      searchMode === "smart" ? hybridSearchDocuments(trimmedQuery) : searchDocuments(trimmedQuery);

    searchPromise
      .then((results) => {
        if (!cancelled) setContentSearchIds(results.map((result) => result.documentId));
      })
      .catch((searchFailure) => {
        if (!cancelled) {
          setContentSearchIds([]);
          toast.error(searchFailure instanceof Error ? searchFailure.message : "Arama tamamlanamadı.");
        }
      })
      .finally(() => {
        if (!cancelled) setIsSearchLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [debouncedQuery, isTrash, searchMode]);

  const filteredDocuments = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase("tr-TR");
    return documents.filter((document) => {
      const matchesQuery = isTrash || normalizedQuery === "" || (normalizedQuery.length < 2 ? false : contentSearchIds.includes(document.id));

      const matchesStatus = status === "all" || document.status === status;
      const matchesFileType = fileType === "all" || document.file_extension === fileType;

      return matchesQuery && matchesStatus && matchesFileType;
    });
  }, [contentSearchIds, documents, fileType, isTrash, query, status]);

  useEffect(() => {
    setPage(1);
  }, [status, fileType, folderId, tagId, debouncedQuery, isTrash]);

  const paginatedDocuments = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredDocuments.slice(start, start + pageSize);
  }, [filteredDocuments, page, pageSize]);

  useEffect(() => {
    const totalPages = Math.max(1, Math.ceil(filteredDocuments.length / pageSize));
    if (page > totalPages) {
      setPage(totalPages);
    }
  }, [filteredDocuments.length, page, pageSize]);

  const handlePageSizeChange = (nextPageSize: number) => {
    setPageSize(nextPageSize);
    setPage(1);
    window.localStorage.setItem(PAGE_SIZE_STORAGE_KEY, String(nextPageSize));
  };

  const submitSearch = () => {
    const trimmedQuery = query.trim();
    if (trimmedQuery.length < 2) return;
    navigate(`/search?q=${encodeURIComponent(trimmedQuery)}`);
  };

  const openEditModal = (document: ApiDocument) => {
    setEditingDocument(document);
    setEditName(document.original_file_name);
    setEditStatus(document.status);
  };

  const { handleContextMenuAction, modals: documentActionModals } = useDocumentMenuActions({
    isTrash,
    navigationQuery: !isTrash ? query : "",
    onToggleFavorite: toggleFavorite,
    onRename: openEditModal,
    onDocumentUpdated: (updated) => {
      setDocuments((current) => current.map((item) => (item.id === updated.id ? updated : item)));
    },
    onDocumentRemoved: (removed) => {
      setDocuments((current) => current.filter((item) => item.id !== removed.id));
      setSelectedIds((current) => current.filter((id) => id !== removed.id));
    },
  });

  useContextMenuActionHandler(handleContextMenuAction);

  const closeEditModal = () => {
    if (isSaving) return;
    setEditingDocument(null);
    setEditName("");
    setEditStatus("ready");
  };

  const handleSaveEdit = async () => {
    if (!editingDocument || editName.trim() === "") return;

    setIsSaving(true);
    try {
      const updatedDocument = await updateDocument(editingDocument.id, {
        original_file_name: editName.trim(),
        status: editStatus,
      });
      setDocuments((current) => current.map((document) => (document.id === updatedDocument.id ? updatedDocument : document)));
      closeEditModal();
      toast.success("Belge güncellendi.");
    } catch (saveError) {
      toast.error(saveError instanceof Error ? saveError.message : "Belge güncellenemedi.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (document: ApiDocument) => {
    const shouldDelete = await confirm({
      title: isTrash ? "Kalıcı sil" : "Çöp kutusuna taşı",
      message: isTrash
        ? `${document.original_file_name} kalıcı olarak silinsin mi? Bu işlem geri alınamaz.`
        : `${document.original_file_name} çöp kutusuna taşınsın mı?`,
      confirmLabel: isTrash ? "Kalıcı sil" : "Taşı",
      variant: "danger",
    });
    if (!shouldDelete) return;

    try {
      if (isTrash) {
        await permanentlyDeleteDocument(document.id);
        toast.success("Belge kalıcı olarak silindi.");
      } else {
        await deleteDocument(document.id);
        toast.success("Belge çöp kutusuna taşındı.");
      }
      setDocuments((current) => current.filter((item) => item.id !== document.id));
      setSelectedIds((current) => current.filter((id) => id !== document.id));
    } catch (deleteError) {
      toast.error(deleteError instanceof Error ? deleteError.message : "Belge silinemedi.");
    }
  };

  const handleRestore = async (document: ApiDocument) => {
    try {
      await restoreDocument(document.id);
      setDocuments((current) => current.filter((item) => item.id !== document.id));
      setSelectedIds((current) => current.filter((id) => id !== document.id));
      toast.success("Belge geri yüklendi.");
    } catch (restoreError) {
      toast.error(restoreError instanceof Error ? restoreError.message : "Belge geri yüklenemedi.");
    }
  };

  const handleSelect = (documentId: string, selected: boolean) => {
    setSelectedIds((current) => (selected ? [...new Set([...current, documentId])] : current.filter((id) => id !== documentId)));
  };

  const handleSelectAll = (selected: boolean) => {
    const pageIds = paginatedDocuments.map((document) => document.id);
    setSelectedIds((current) => (selected ? [...new Set([...current, ...pageIds])] : current.filter((id) => !pageIds.includes(id))));
  };

  const handleBulkDelete = async () => {
    if (selectedIds.length === 0) return;
    const shouldDelete = await confirm({
      title: isTrash ? "Kalıcı sil" : "Çöp kutusuna taşı",
      message: isTrash
        ? `${selectedIds.length} belge kalıcı olarak silinsin mi? Bu işlem geri alınamaz.`
        : `${selectedIds.length} belge çöp kutusuna taşınsın mı?`,
      confirmLabel: isTrash ? "Kalıcı sil" : "Taşı",
      variant: "danger",
    });
    if (!shouldDelete) return;

    try {
      await Promise.all(selectedIds.map((id) => (isTrash ? permanentlyDeleteDocument(id) : deleteDocument(id))));
      setDocuments((current) => current.filter((document) => !selectedIds.includes(document.id)));
      setSelectedIds([]);
      toast.success(isTrash ? "Seçili belgeler kalıcı olarak silindi." : "Seçili belgeler çöp kutusuna taşındı.");
    } catch (bulkError) {
      toast.error(bulkError instanceof Error ? bulkError.message : "Toplu silme işlemi tamamlanamadı.");
    }
  };

  const handleBulkRestore = async () => {
    if (selectedIds.length === 0) return;

    try {
      await Promise.all(selectedIds.map((id) => restoreDocument(id)));
      setDocuments((current) => current.filter((document) => !selectedIds.includes(document.id)));
      setSelectedIds([]);
      toast.success("Seçili belgeler geri yüklendi.");
    } catch (bulkError) {
      toast.error(bulkError instanceof Error ? bulkError.message : "Toplu geri yükleme işlemi tamamlanamadı.");
    }
  };

  return (
    <div className="space-y-5">
      <PageTitle title={isTrash ? "Çöp Kutusu" : "Belgelerim"} />

      <section className="panel p-4">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
          <div className="min-w-0 flex-1">
            <SearchInput
              value={query}
              wrapperClassName="w-full"
              placeholder={isTrash ? "Çöp kutusunda ara..." : "Döküman içeriklerinde ara..."}
              shortcut="Enter"
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  submitSearch();
                }
              }}
            />
            {!isTrash && query.trim().length > 0 && query.trim().length < 2 && <p className="mt-2 text-xs font-semibold text-slate-500">En az 2 karakter yazın.</p>}
            {!isTrash && isSearchLoading && (
              <div className="mt-2">
                <Spinner size="sm" label="İçeriklerde aranıyor..." />
              </div>
            )}
          </div>

          {!isTrash ? <SearchModeToggle onChange={setSearchMode} /> : null}

          <select className="h-10 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-600 outline-none focus:border-violet-brand" value={status} onChange={(event) => setStatus(event.target.value as "all" | DocumentStatus)}>
            {statusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>

          <select className="h-10 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-600 outline-none focus:border-violet-brand" value={fileType} onChange={(event) => setFileType(event.target.value)}>
            {fileTypeOptions.map((option) => (
              <option key={option} value={option}>
                {option === "all" ? "Tüm tipler" : option.toUpperCase()}
              </option>
            ))}
          </select>

          {!isTrash && (
            <select className="h-10 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-600 outline-none focus:border-violet-brand" value={tagId} onChange={(event) => setTagId(event.target.value)}>
              <option value="">Tüm etiketler</option>
              {tags.map((tag) => (
                <option key={tag.id} value={tag.id}>
                  {tag.name}
                </option>
              ))}
            </select>
          )}

          {!isTrash && (
            <select className="h-10 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-600 outline-none focus:border-violet-brand" value={folderId} onChange={(event) => setFolderId(event.target.value)}>
              <option value="">Tüm klasörler</option>
              <option value="root">Klasörsüz</option>
              {flattenFolders(folders).map((folder) => (
                <option key={folder.id} value={folder.id}>
                  {folder.label}
                </option>
              ))}
            </select>
          )}

          <div className="flex items-center gap-2 lg:ml-auto">
            <IconButton aria-label="Yenile" onClick={() => void fetchDocuments()}>
              <RefreshCw className="h-4 w-4" />
            </IconButton>
            <IconButton aria-label="Filtrele">
              <SlidersHorizontal className="h-4 w-4" />
            </IconButton>
            <IconButton className="text-violet-brand" aria-label="Liste görünümü">
              <List className="h-4 w-4" />
            </IconButton>
            <IconButton aria-label="Grid">
              <Grid3X3 className="h-4 w-4" />
            </IconButton>
          </div>
        </div>
      </section>

      {selectedIds.length > 0 && (
        <section className="flex flex-col gap-3 rounded-lg border border-violet-ring bg-violet-soft/70 px-4 py-3 text-sm sm:flex-row sm:items-center sm:justify-between">
          <span className="font-semibold text-violet-brand">{selectedIds.length} belge seçildi</span>
          <div className="flex flex-wrap gap-2">
            {isTrash && (
              <PrimaryButton onClick={() => void handleBulkRestore()}>
                <RotateCcw className="h-4 w-4" />
                Geri Yükle
              </PrimaryButton>
            )}
            <button className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-red-600 px-4 text-sm font-semibold text-white shadow-lg shadow-red-100 transition hover:bg-red-700" type="button" onClick={() => void handleBulkDelete()}>
              <Trash2 className="h-4 w-4" />
              {isTrash ? "Kalıcı Sil" : "Çöp Kutusuna Taşı"}
            </button>
          </div>
        </section>
      )}

      <DocumentTable
        documents={paginatedDocuments}
        isLoading={isLoading}
        isTrash={isTrash}
        error={loadError}
        onRetry={() => void fetchDocuments()}
        emptyTitle={isTrash ? "Çöp kutusu boş" : "Henüz belge yok"}
        emptyDescription={
          isTrash
            ? "Çöp kutusuna taşınan belgeler burada görünür. Kalıcı silme işlemi geri alınamaz."
            : "Belge yüklediğinizde veya filtreleri değiştirdiğinizde sonuçlar burada listelenir."
        }
        selectedIds={selectedIds}
        page={page}
        pageSize={pageSize}
        totalCount={filteredDocuments.length}
        onPageChange={setPage}
        onPageSizeChange={handlePageSizeChange}
        onSelect={handleSelect}
        onSelectAll={handleSelectAll}
        onEdit={openEditModal}
        onDelete={(document) => void handleDelete(document)}
        onRestore={(document) => void handleRestore(document)}
        onToggleFavorite={(document) => void toggleFavorite(document)}
        favoriteLoadingId={favoriteLoadingId}
        navigationQuery={!isTrash ? query : ""}
      />

      {!isTrash && editingDocument && (
        <div className="fixed inset-0 z-[80] flex items-center justify-center bg-slate-900/35 p-4">
          <div className="w-full max-w-md rounded-xl border border-line bg-white shadow-panel">
            <div className="border-b border-line px-5 py-4">
              <h2 className="text-lg font-bold text-ink">Belgeyi Düzenle</h2>
            </div>
            <div className="space-y-4 p-5">
              <label className="space-y-2">
                <span className="text-sm font-semibold text-slate-600">Dosya adı</span>
                <input className="h-11 w-full rounded-lg border border-line px-3 text-sm outline-none focus:border-violet-brand" value={editName} onChange={(event) => setEditName(event.target.value)} />
              </label>
              <label className="space-y-2">
                <span className="text-sm font-semibold text-slate-600">Durum</span>
                <select className="h-11 w-full rounded-lg border border-line bg-white px-3 text-sm outline-none focus:border-violet-brand" value={editStatus} onChange={(event) => setEditStatus(event.target.value as DocumentStatus)}>
                  {statusOptions
                    .filter((option) => option.value !== "all")
                    .map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                </select>
              </label>
            </div>
            <div className="flex justify-end gap-3 border-t border-line px-5 py-4">
              <button className="h-10 rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand disabled:opacity-60" type="button" disabled={isSaving} onClick={closeEditModal}>
                Vazgeç
              </button>
              <PrimaryButton disabled={isSaving || editName.trim() === ""} onClick={handleSaveEdit}>
                {isSaving ? "Kaydediliyor..." : "Kaydet"}
              </PrimaryButton>
            </div>
          </div>
        </div>
      )}

      {documentActionModals}
    </div>
  );
}

function flattenFolders(folders: Folder[], depth = 0): Array<{ id: string; label: string }> {
  return folders.flatMap((folder) => [
    { id: folder.id, label: `${"  ".repeat(depth)}${depth > 0 ? "- " : ""}${folder.name}` },
    ...flattenFolders(folder.children ?? [], depth + 1),
  ]);
}
