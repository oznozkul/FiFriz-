import { useCallback, useState, type ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { DocumentMoveModal } from "../components/documents/DocumentMoveModal";
import { DocumentRenameModal } from "../components/documents/DocumentRenameModal";
import { DocumentTagModal } from "../components/documents/DocumentTagModal";
import { ShareLinkModal } from "../components/documents/ShareLinkModal";
import { WorkflowCreateModal } from "../components/documents/WorkflowCreateModal";
import { ImportantDateFormModal } from "../components/importantDates/ImportantDateFormModal";
import { deleteDocument, restoreDocument, permanentlyDeleteDocument } from "../data/documentsApi";
import { ContextMenuActions } from "../contextMenu";
import { useConfirm } from "../components/common";
import { createDocumentImportantDate } from "../services/importantDateService";
import { createDocumentWorkflow } from "../services/workflowService";
import { notifyOnboardingStep } from "../components/dashboard/onboarding/onboardingStorage";
import { toast } from "../services/toast";
import type { ApiDocument } from "../types/document";
import { downloadDocumentFile, workspacePanelPath } from "../utils/documentActions";

type UseDocumentMenuActionsOptions = {
  isTrash?: boolean;
  navigationQuery?: string;
  onDocumentUpdated?: (document: ApiDocument) => void;
  onDocumentRemoved?: (document: ApiDocument) => void;
  onToggleFavorite?: (document: ApiDocument) => void | Promise<unknown>;
  onRename?: (document: ApiDocument) => void;
};

export function useDocumentMenuActions({
  isTrash = false,
  navigationQuery = "",
  onDocumentUpdated,
  onDocumentRemoved,
  onToggleFavorite,
  onRename,
}: UseDocumentMenuActionsOptions = {}) {
  const navigate = useNavigate();
  const confirm = useConfirm();
  const [shareDocument, setShareDocument] = useState<ApiDocument | null>(null);
  const [tagDocument, setTagDocument] = useState<ApiDocument | null>(null);
  const [moveDocument, setMoveDocument] = useState<ApiDocument | null>(null);
  const [renameDocument, setRenameDocument] = useState<ApiDocument | null>(null);
  const [importantDateDocument, setImportantDateDocument] = useState<ApiDocument | null>(null);
  const [workflowDocument, setWorkflowDocument] = useState<ApiDocument | null>(null);
  const [isImportantDateSaving, setIsImportantDateSaving] = useState(false);

  const openWorkspacePanel = useCallback(
    (documentId: string, panel: string) => {
      navigate(workspacePanelPath(documentId, panel, navigationQuery));
    },
    [navigate, navigationQuery],
  );

  const handleDelete = useCallback(
    async (document: ApiDocument) => {
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
        onDocumentRemoved?.(document);
        window.dispatchEvent(new CustomEvent("documents:changed"));
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "Belge silinemedi.");
      }
    },
    [isTrash, onDocumentRemoved, confirm],
  );

  const handleRestore = useCallback(
    async (document: ApiDocument) => {
      try {
        await restoreDocument(document.id);
        onDocumentRemoved?.(document);
        toast.success("Belge geri yüklendi.");
        window.dispatchEvent(new CustomEvent("documents:changed"));
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "Belge geri yüklenemedi.");
      }
    },
    [onDocumentRemoved],
  );

  const handleContextMenuAction = useCallback(
    (detail: { action: string; context: { entityType: string; entity: unknown } }) => {
      if (detail.context.entityType !== "document") return;
      const document = detail.context.entity as ApiDocument;

      switch (detail.action) {
        case ContextMenuActions.document.open:
          navigate(workspacePanelPath(document.id, "general", navigationQuery));
          return;
        case ContextMenuActions.document.toggleFavorite:
          void onToggleFavorite?.(document);
          return;
        case ContextMenuActions.document.rename:
          if (onRename) {
            onRename(document);
            return;
          }
          setRenameDocument(document);
          return;
        case ContextMenuActions.document.trash:
          void handleDelete(document);
          return;
        case ContextMenuActions.document.restore:
          void handleRestore(document);
          return;
        case ContextMenuActions.document.download:
          void downloadDocumentFile(document).catch((error) => {
            toast.error(error instanceof Error ? error.message : "Dosya indirilemedi.");
          });
          return;
        case ContextMenuActions.document.tag:
          setTagDocument(document);
          return;
        case ContextMenuActions.document.move:
          setMoveDocument(document);
          return;
        case ContextMenuActions.document.share:
          setShareDocument(document);
          return;
        case ContextMenuActions.document.versions:
          openWorkspacePanel(document.id, "versions");
          return;
        case ContextMenuActions.document.activity:
          openWorkspacePanel(document.id, "activity");
          return;
        case ContextMenuActions.document.addImportantDate:
          setImportantDateDocument(document);
          return;
        case ContextMenuActions.document.startWorkflow:
          setWorkflowDocument(document);
          return;
        case ContextMenuActions.ai.summarize:
        case ContextMenuActions.ai.ask:
        case ContextMenuActions.ai.keywords:
          openWorkspacePanel(document.id, "ai");
          toast.info("Yapay zeka özellikleri yakında eklenecek.");
          return;
      }
    },
    [handleDelete, handleRestore, navigate, navigationQuery, onRename, onToggleFavorite, openWorkspacePanel],
  );

  const modals: ReactNode = (
    <>
      {shareDocument ? (
        <ShareLinkModal
          documentId={shareDocument.id}
          documentName={shareDocument.original_file_name}
          isOpen
          onClose={() => setShareDocument(null)}
        />
      ) : null}
      {tagDocument ? (
        <DocumentTagModal
          document={tagDocument}
          onClose={() => setTagDocument(null)}
          onUpdated={(updated) => {
            onDocumentUpdated?.(updated);
            setTagDocument(null);
          }}
        />
      ) : null}
      {moveDocument ? (
        <DocumentMoveModal
          document={moveDocument}
          onClose={() => setMoveDocument(null)}
          onMoved={(updated) => {
            onDocumentUpdated?.(updated);
            setMoveDocument(null);
            window.dispatchEvent(new CustomEvent("documents:changed"));
          }}
        />
      ) : null}
      {renameDocument ? (
        <DocumentRenameModal
          document={renameDocument}
          onClose={() => setRenameDocument(null)}
          onSaved={(updated) => {
            onDocumentUpdated?.(updated);
            setRenameDocument(null);
          }}
        />
      ) : null}
      {importantDateDocument ? (
        <ImportantDateFormModal
          isSaving={isImportantDateSaving}
          onClose={() => setImportantDateDocument(null)}
          onSave={(input) => {
            setIsImportantDateSaving(true);
            void createDocumentImportantDate(importantDateDocument.id, input)
              .then(() => {
                toast.success("Önemli tarih eklendi.");
                setImportantDateDocument(null);
                openWorkspacePanel(importantDateDocument.id, "dates");
              })
              .catch((error) => {
                toast.error(error instanceof Error ? error.message : "Önemli tarih eklenemedi.");
              })
              .finally(() => setIsImportantDateSaving(false));
          }}
        />
      ) : null}
      {workflowDocument ? (
        <WorkflowCreateModal
          isOpen
          onClose={() => setWorkflowDocument(null)}
          onSubmit={async (input) => {
            await createDocumentWorkflow(workflowDocument.id, input);
            toast.success("Onay süreci başlatıldı.");
            notifyOnboardingStep("start-workflow");
            setWorkflowDocument(null);
            openWorkspacePanel(workflowDocument.id, "workflow");
          }}
        />
      ) : null}
    </>
  );

  return { handleContextMenuAction, modals };
}
