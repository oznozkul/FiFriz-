import { PanelRightOpen } from "lucide-react";
import { createRef, useCallback, useEffect, useMemo, useState } from "react";
import { useLocation, useParams, useSearchParams } from "react-router-dom";
import { findFolderPath } from "../folders/folderUtils";
import { useBreadcrumbContext } from "../../context/BreadcrumbContext";
import { ShareLinkModal } from "../documents/ShareLinkModal";
import { EmptyState, LoadingState } from "../common";
import {
  favoriteDocument,
  getDocument,
  getDocumentText,
  getDocumentWorkspace,
  reprocessDocumentText,
  unfavoriteDocument,
} from "../../data/documentsApi";
import { apiBaseUrl, clearAccessToken, getAccessToken } from "../../services/apiClient";
import { listFolders } from "../../services/folderService";
import { toast } from "../../services/toast";
import type { ApiDocument } from "../../types/document";
import { DocumentWorkspacePanel } from "./DocumentWorkspacePanel";
import { DocumentWorkspaceToolbar } from "./DocumentWorkspaceToolbar";
import {
  buildHighlightParts,
  ExtractedTextPanel,
  FilePreview,
  SearchMatchesPanel,
} from "./DocumentFilePreview";
import type { WorkspacePanelSection } from "./documentWorkspaceUtils";

function resolveWorkspacePanel(value: string | null): WorkspacePanelSection | null {
  if (!value) return null;
  const aliases: Record<string, WorkspacePanelSection> = {
    general: "general",
    health: "health",
    tags: "tags",
    tag: "tags",
    "suggested-tags": "suggestedTags",
    "ai-tags": "suggestedTags",
    folder: "folder",
    shares: "shares",
    share: "shares",
    workflow: "workflow",
    comments: "comments",
    versions: "versions",
    dates: "dates",
    "important-date": "dates",
    "important-dates": "dates",
    activity: "activity",
    related: "related",
    "ai-metadata": "aiMetadata",
    metadata: "aiMetadata",
    classification: "classification",
    "document-type": "classification",
    entities: "entities",
    "ai-entities": "entities",
    summary: "summary",
    "ai-progress": "aiProgress",
    ai: "ai",
    "action-suggestions": "actionSuggestions",
    "ai-suggestions": "actionSuggestions",
  };
  return aliases[value] ?? null;
}

const defaultOpenSections: Record<WorkspacePanelSection, boolean> = {
  general: true,
  health: false,
  tags: false,
  suggestedTags: false,
  folder: false,
  shares: false,
  workflow: false,
  comments: false,
  versions: false,
  dates: false,
  activity: false,
  related: false,
  aiMetadata: false,
  classification: false,
  entities: false,
  summary: false,
  aiProgress: false,
  ai: false,
  actionSuggestions: false,
};

export function DocumentWorkspace() {
  const { id } = useParams();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const searchQuery = searchParams.get("q") ?? "";
  const breadcrumbContext = useBreadcrumbContext();
  const setBreadcrumbItems = breadcrumbContext?.setItems;
  const resetBreadcrumbItems = breadcrumbContext?.resetItems;
  const panelQuery = searchParams.get("panel");

  const [document, setDocument] = useState<ApiDocument | null>(null);
  const [documentText, setDocumentText] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isTextLoading, setIsTextLoading] = useState(false);
  const [fileObjectUrl, setFileObjectUrl] = useState("");
  const [isFileLoading, setIsFileLoading] = useState(false);
  const [fileError, setFileError] = useState<string | null>(null);
  const [isReprocessing, setIsReprocessing] = useState(false);
  const [isFavoriteLoading, setIsFavoriteLoading] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);
  const [textError, setTextError] = useState<string | null>(null);
  const [activeMatchIndex, setActiveMatchIndex] = useState(0);
  const [isShareModalOpen, setIsShareModalOpen] = useState(false);
  const [isPanelOpen, setIsPanelOpen] = useState(true);
  const [isMobileDrawerOpen, setIsMobileDrawerOpen] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [openSections, setOpenSections] = useState(defaultOpenSections);
  const [focusSection, setFocusSection] = useState<WorkspacePanelSection | undefined>();
  const [focusToken, setFocusToken] = useState(0);
  const [fileReloadToken, setFileReloadToken] = useState(0);

  useEffect(() => {
    const panelFromQuery = resolveWorkspacePanel(panelQuery);
    const panelFromHash = resolveWorkspacePanel(location.hash.replace("#", ""));
    const panel = panelFromQuery ?? panelFromHash;
    if (!panel) return;

    setOpenSections((current) => ({ ...current, [panel]: true }));
    setFocusSection(panel);
    setFocusToken((current) => current + 1);
    setIsPanelOpen(true);
  }, [id, location.hash, panelQuery]);

  useEffect(() => {
    if (!id) return;

    const fetchDocument = async () => {
      setIsLoading(true);
      setLoadFailed(false);
      setDocumentText("");
      try {
        setDocument(await getDocumentWorkspace(id));
      } catch {
        try {
          setDocument(await getDocument(id));
        } catch (fetchError) {
          setLoadFailed(true);
          toast.error(fetchError instanceof Error ? fetchError.message : "Belge yüklenemedi.");
        }
      } finally {
        setIsLoading(false);
      }
    };

    void fetchDocument();
  }, [id]);

  useEffect(() => {
    if (!document?.id || !setBreadcrumbItems) return;

    let cancelled = false;
    const documentId = document.id;
    const folderId = document.folder_id;
    const fileName = document.original_file_name;

    void listFolders()
      .then((folders) => {
        if (cancelled) return;
        const items = [{ label: "Belgeler", href: "/documents" }];
        if (folderId) {
          for (const folder of findFolderPath(folders, folderId)) {
            items.push({ label: folder.name, href: "/folders" });
          }
        }
        items.push({ label: fileName, href: `/documents/${documentId}` });
        setBreadcrumbItems(items);
      })
      .catch(() => {
        if (!cancelled) {
          setBreadcrumbItems([
            { label: "Belgeler", href: "/documents" },
            { label: fileName, href: `/documents/${documentId}` },
          ]);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [document?.folder_id, document?.id, document?.original_file_name, setBreadcrumbItems]);

  useEffect(() => {
    return () => {
      resetBreadcrumbItems?.();
    };
  }, [resetBreadcrumbItems]);

  useEffect(() => {
    if (!document?.fileUrl) return;

    const controller = new AbortController();
    let objectUrl = "";

    const loadFile = async () => {
      setIsFileLoading(true);
      setFileError(null);
      setFileObjectUrl("");

      try {
        const token = getAccessToken();
        const response = await fetch(`${apiBaseUrl}${document.fileUrl}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : undefined,
          signal: controller.signal,
        });

        if (response.status === 401 || response.status === 403) {
          clearAccessToken();
          window.dispatchEvent(new CustomEvent("auth:logout"));
        }
        if (!response.ok) {
          throw new Error("Dosya yüklenemedi.");
        }

        const blob = await response.blob();
        if (controller.signal.aborted) return;

        objectUrl = URL.createObjectURL(blob);
        setFileObjectUrl(objectUrl);
      } catch (loadError) {
        if (controller.signal.aborted) return;
        setFileError(loadError instanceof Error ? loadError.message : "Dosya yüklenemedi.");
      } finally {
        if (!controller.signal.aborted) setIsFileLoading(false);
      }
    };

    void loadFile();

    return () => {
      controller.abort();
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [document?.fileUrl, document?.id, fileReloadToken]);

  useEffect(() => {
    const shouldLoadText =
      searchQuery.trim().length >= 2 || ["docx", "xlsx", "txt", "png", "jpg", "jpeg"].includes(document?.file_extension?.toLowerCase() ?? "");
    if (!id || !document?.hasExtractedText || !shouldLoadText) return;

    let cancelled = false;
    setIsTextLoading(true);
    setTextError(null);

    getDocumentText(id)
      .then((textResponse) => {
        if (!cancelled) setDocumentText(textResponse.full_text);
      })
      .catch(() => {
        if (!cancelled) {
          setDocumentText("");
          setTextError("Çıkarılmış metin yüklenemedi.");
        }
      })
      .finally(() => {
        if (!cancelled) setIsTextLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [document?.file_extension, document?.hasExtractedText, id, searchQuery]);

  const highlight = useMemo(() => buildHighlightParts(documentText, searchQuery), [documentText, searchQuery]);
  const matchRefs = useMemo(() => highlight.matches.map(() => createRef<HTMLElement>()), [highlight.matches.length]);

  useEffect(() => {
    setActiveMatchIndex(0);
  }, [searchQuery, documentText]);

  useEffect(() => {
    if (highlight.matches.length === 0) return;
    matchRefs[activeMatchIndex]?.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
  }, [activeMatchIndex, highlight.matches.length, matchRefs]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      const isTyping = target?.tagName === "INPUT" || target?.tagName === "TEXTAREA" || target?.isContentEditable;

      if (event.key === "Escape" && isFullscreen) {
        event.preventDefault();
        setIsFullscreen(false);
        return;
      }

      if (!isTyping && (event.key === "f" || event.key === "F")) {
        event.preventDefault();
        setIsFullscreen((current) => !current);
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [isFullscreen]);

  const focusPanelSection = useCallback((section: WorkspacePanelSection) => {
    setOpenSections((current) => ({ ...current, [section]: true }));
    setFocusSection(section);
    setFocusToken((current) => current + 1);
    setIsPanelOpen(true);
    setIsMobileDrawerOpen(true);
  }, []);

  const reloadFilePreview = useCallback(() => {
    setFileReloadToken((current) => current + 1);
  }, []);

  const toggleSection = useCallback((section: WorkspacePanelSection) => {
    setOpenSections((current) => ({ ...current, [section]: !current[section] }));
  }, []);

  const handleReprocessText = async () => {
    if (!id) return;

    setIsReprocessing(true);
    setTextError(null);
    try {
      const updatedDocument = await reprocessDocumentText(id);
      setDocument(updatedDocument);
      if (updatedDocument.hasExtractedText) {
        const textResponse = await getDocumentText(id);
        setDocumentText(textResponse.full_text);
      } else {
        setDocumentText("");
      }
    } catch (reprocessError) {
      const message = reprocessError instanceof Error ? reprocessError.message : "Metin yeniden işlenemedi.";
      setTextError(message);
      toast.error(message);
    } finally {
      setIsReprocessing(false);
    }
  };

  const handleVersionsChanged = async () => {
    if (!id) return;

    setDocumentText("");
    setTextError(null);
    const updatedDocument = await getDocument(id);
    setDocument(updatedDocument);
    if (updatedDocument.hasExtractedText) {
      try {
        const textResponse = await getDocumentText(id);
        setDocumentText(textResponse.full_text);
      } catch {
        setTextError("Çıkarılmış metin yüklenemedi.");
      }
    }
  };

  const handleToggleFavorite = async () => {
    if (!document) return;

    const nextFavorite = !document.isFavorite;
    const previousDocument = document;

    setIsFavoriteLoading(true);
    setDocument({
      ...document,
      isFavorite: nextFavorite,
      favoritedAt: nextFavorite ? new Date().toISOString() : null,
    });

    try {
      const updatedDocument = nextFavorite ? await favoriteDocument(document.id) : await unfavoriteDocument(document.id);
      setDocument(updatedDocument);
      window.dispatchEvent(new CustomEvent("favorites:changed"));
      toast.success(updatedDocument.isFavorite ? "Favorilere eklendi." : "Favorilerden çıkarıldı.");
    } catch (favoriteError) {
      setDocument(previousDocument);
      toast.error(favoriteError instanceof Error ? favoriteError.message : "Favori durumu güncellenemedi.");
    } finally {
      setIsFavoriteLoading(false);
    }
  };

  if (isLoading) {
    return <LoadingState label="Belge yükleniyor..." />;
  }

  if (loadFailed || !document) {
    return <EmptyState title="Belge bulunamadı" description="Belge silinmiş olabilir veya erişim yetkiniz olmayabilir." />;
  }

  const fileUrl = fileObjectUrl;
  const extension = document.file_extension.toLowerCase();
  const isOfficeDocument = extension === "docx" || extension === "xlsx";
  const isTextDocument = extension === "txt";
  const isImageDocument = ["png", "jpg", "jpeg"].includes(extension);
  const supportsExtractedTextPanel = isOfficeDocument || isImageDocument || isTextDocument;

  const goToPreviousMatch = () => {
    setActiveMatchIndex((current) => (current === 0 ? highlight.matches.length - 1 : current - 1));
  };

  const goToNextMatch = () => {
    setActiveMatchIndex((current) => (current + 1) % highlight.matches.length);
  };

  return (
    <div
      className={`flex min-h-[calc(100dvh-10rem)] flex-col overflow-hidden bg-white lg:min-h-[720px] ${isFullscreen ? "fixed inset-0 z-50" : "panel rounded-2xl"}`}
    >
      <DocumentWorkspaceToolbar
        document={document}
        fileUrl={fileUrl}
        isFileLoading={isFileLoading}
        isFavoriteLoading={isFavoriteLoading}
        isPanelOpen={isPanelOpen}
        isFullscreen={isFullscreen}
        onToggleFavorite={() => void handleToggleFavorite()}
        onShare={() => setIsShareModalOpen(true)}
        onTogglePanel={() => setIsPanelOpen((current) => !current)}
        onToggleFullscreen={() => setIsFullscreen((current) => !current)}
        onOpenMobilePanel={() => setIsMobileDrawerOpen(true)}
        onFocusSection={focusPanelSection}
      />

      <div className="relative flex min-h-0 flex-1">
        <section className="flex min-w-0 flex-1 flex-col overflow-hidden bg-slate-100/60">
          {searchQuery.trim().length >= 2 ? (
            <div className="border-b border-line bg-white p-3">
              <SearchMatchesPanel
                query={searchQuery}
                isLoading={isTextLoading}
                hasExtractedText={document.hasExtractedText}
                highlight={highlight}
                matchRefs={matchRefs}
                activeMatchIndex={activeMatchIndex}
                onPrevious={goToPreviousMatch}
                onNext={goToNextMatch}
                onSelect={setActiveMatchIndex}
              />
            </div>
          ) : null}

          <article className="min-h-0 flex-1 overflow-auto p-4 sm:p-5 md:p-6">
            <FilePreview
              document={document}
              fileUrl={fileUrl}
              isFileLoading={isFileLoading}
              fileError={fileError}
              onRetry={reloadFilePreview}
              extractedText={documentText}
              isTextLoading={isTextLoading}
              searchQuery={searchQuery}
              activeMatchIndex={activeMatchIndex}
              activeSnippet={highlight.snippets[activeMatchIndex]}
            />
          </article>

          {supportsExtractedTextPanel && searchQuery.trim().length < 2 ? (
            <div className="border-t border-line bg-white p-3 lg:hidden">
              <ExtractedTextPanel
                fileTypeLabel={extension.toUpperCase()}
                text={documentText}
                isLoading={isTextLoading}
                isReprocessing={isReprocessing}
                hasExtractedText={document.hasExtractedText}
                error={textError}
                onReprocess={() => void handleReprocessText()}
              />
            </div>
          ) : null}
        </section>

        {isPanelOpen ? (
          <div className="hidden h-full min-h-0 w-[380px] shrink-0 self-stretch lg:flex">
            <DocumentWorkspacePanel
              document={document}
              openSections={openSections}
              focusToken={focusToken}
              focusSection={focusSection}
              activeSection={focusSection}
              onToggleSection={toggleSection}
              onTagsUpdated={(tags) => setDocument((current) => (current ? { ...current, tags } : current))}
              onVersionsChanged={handleVersionsChanged}
              onCreateShare={() => setIsShareModalOpen(true)}
              onFocusSection={focusPanelSection}
              onCollapsePanel={() => setIsPanelOpen(false)}
            />
          </div>
        ) : null}

        {!isPanelOpen ? (
          <button
            type="button"
            className="absolute right-3 top-3 hidden h-10 w-10 items-center justify-center rounded-lg border border-line bg-white text-slate-600 shadow-panel lg:inline-flex"
            aria-label="Bilgi panelini aç"
            onClick={() => setIsPanelOpen(true)}
          >
            <PanelRightOpen className="h-4 w-4" />
          </button>
        ) : null}

        <button
          type="button"
          className="absolute bottom-4 right-4 inline-flex h-11 items-center gap-2 rounded-full border border-line bg-white px-4 text-sm font-semibold text-slate-700 shadow-panel lg:hidden"
          onClick={() => setIsMobileDrawerOpen(true)}
        >
          <PanelRightOpen className="h-4 w-4" />
          Bilgiler
        </button>

        {isMobileDrawerOpen ? (
          <div className="fixed inset-0 z-40 lg:hidden">
            <button type="button" className="absolute inset-0 bg-black/40" aria-label="Paneli kapat" onClick={() => setIsMobileDrawerOpen(false)} />
            <div className="absolute inset-y-0 right-0 flex w-full shadow-panel sm:max-w-[360px]">
              <DocumentWorkspacePanel
                document={document}
                isDrawer
                openSections={openSections}
                focusToken={focusToken}
                focusSection={focusSection}
                activeSection={focusSection}
                onToggleSection={toggleSection}
                onTagsUpdated={(tags) => setDocument((current) => (current ? { ...current, tags } : current))}
                onVersionsChanged={handleVersionsChanged}
                onCreateShare={() => setIsShareModalOpen(true)}
                onFocusSection={focusPanelSection}
                onClose={() => setIsMobileDrawerOpen(false)}
              />
            </div>
          </div>
        ) : null}
      </div>

      <ShareLinkModal documentId={document.id} documentName={document.original_file_name} isOpen={isShareModalOpen} onClose={() => setIsShareModalOpen(false)} />
    </div>
  );
}
