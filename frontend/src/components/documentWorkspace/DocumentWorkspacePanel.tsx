import { Link } from "react-router-dom";
import { useState } from "react";
import {
  Activity,
  Bot,
  CalendarClock,
  ClipboardCheck,
  FileText,
  FileType,
  FolderOpen,
  GitBranch,
  HeartPulse,
  Layers,
  Link2,
  ListChecks,
  MessageSquare,
  PanelRightClose,
  ScanSearch,
  Share2,
  Sparkles,
  Tags,
  WandSparkles,
} from "lucide-react";
import { AiChatPanel, AgentChatPanel, AiModeToggle, type AiChatMode } from "../ai";
import { IconButton } from "../common";
import { useFolderLabels } from "../../hooks/useFolderLabels";
import { DocumentHealthPanel } from "../documents/DocumentHealthPanel";
import { DocumentActionSuggestionsPanel } from "../documents/DocumentActionSuggestionsPanel";
import { DocumentAiMetadataPanel } from "../documents/DocumentAiMetadataPanel";
import { DocumentClassificationPanel } from "../documents/DocumentClassificationPanel";
import { DocumentEntitiesPanel } from "../documents/DocumentEntitiesPanel";
import { DocumentCommentsPanel } from "../documents/DocumentCommentsPanel";
import { DocumentAiProgressPanel } from "../documents/DocumentAiProgressPanel";
import { DocumentSuggestedTagsPanel } from "../documents/DocumentSuggestedTagsPanel";
import { DocumentSummaryPanel } from "../documents/DocumentSummaryPanel";
import { DocumentImportantDatesPanel } from "../importantDates";
import { DocumentSourceBadge } from "../documents/DocumentSourceBadge";
import { DocumentVersionsPanel } from "../documents/DocumentVersionsPanel";
import { DocumentWorkflowPanel } from "../documents/DocumentWorkflowPanel";
import { DocumentTagsPanel } from "../tags";
import type { ApiDocument } from "../../types/document";
import { DocumentWorkspaceActivitySection } from "./DocumentWorkspaceActivitySection";
import { DocumentWorkspaceRelatedSection } from "./DocumentWorkspaceRelatedSection";
import { DocumentWorkspaceSharesSection } from "./DocumentWorkspaceSharesSection";
import { WorkspacePanelNav } from "./WorkspacePanelNav";
import { WorkspaceAccordionSection } from "./WorkspaceAccordionSection";
import { StatusBadge } from "../common";
import { formatDateTime, formatFileSize, sourceLabel, type WorkspacePanelSection } from "./documentWorkspaceUtils";
import { generateDocumentSummary } from "../../services/documentSummaryService";
import { notifyOnboardingStep } from "../dashboard/onboarding/onboardingStorage";

type DocumentWorkspacePanelProps = {
  document: ApiDocument;
  openSections: Record<WorkspacePanelSection, boolean>;
  focusToken: number;
  focusSection?: WorkspacePanelSection;
  onToggleSection: (section: WorkspacePanelSection) => void;
  onTagsUpdated: (tags: NonNullable<ApiDocument["tags"]>) => void;
  onVersionsChanged: () => Promise<void>;
  onCreateShare: () => void;
  onFocusSection?: (section: WorkspacePanelSection) => void;
  onClose?: () => void;
  onCollapsePanel?: () => void;
  activeSection?: WorkspacePanelSection;
  isDrawer?: boolean;
};

export function DocumentWorkspacePanel({
  document,
  openSections,
  focusToken,
  focusSection,
  onToggleSection,
  onTagsUpdated,
  onVersionsChanged,
  onCreateShare,
  onFocusSection,
  onClose,
  onCollapsePanel,
  activeSection,
  isDrawer = false,
}: DocumentWorkspacePanelProps) {
  const folderLabels = useFolderLabels();
  const sectionFocus = (section: WorkspacePanelSection) => (focusSection === section ? focusToken : undefined);
  const [summaryRefreshToken, setSummaryRefreshToken] = useState(0);
  const [aiProgressRefreshToken, setAiProgressRefreshToken] = useState(0);
  const [suggestedTagsRefreshToken, setSuggestedTagsRefreshToken] = useState(0);
  const [aiMetadataRefreshToken, setAiMetadataRefreshToken] = useState(0);
  const [classificationRefreshToken, setClassificationRefreshToken] = useState(0);
  const [entitiesRefreshToken, setEntitiesRefreshToken] = useState(0);
  const [relatedRefreshToken, setRelatedRefreshToken] = useState(0);
  const [actionSuggestionsRefreshToken, setActionSuggestionsRefreshToken] = useState(0);
  const [isSummarizingDocument, setIsSummarizingDocument] = useState(false);
  const [aiMode, setAiMode] = useState<AiChatMode>("rag");

  const handleSummarizeDocument = async () => {
    onFocusSection?.("summary");
    setIsSummarizingDocument(true);
    try {
      await generateDocumentSummary(document.id);
      notifyOnboardingStep("ai-summary");
      setSummaryRefreshToken((value) => value + 1);
      setAiProgressRefreshToken((value) => value + 1);
      setSuggestedTagsRefreshToken((value) => value + 1);
      setAiMetadataRefreshToken((value) => value + 1);
      setClassificationRefreshToken((value) => value + 1);
      setEntitiesRefreshToken((value) => value + 1);
      setRelatedRefreshToken((value) => value + 1);
      setActionSuggestionsRefreshToken((value) => value + 1);
    } catch {
      setSummaryRefreshToken((value) => value + 1);
      setAiProgressRefreshToken((value) => value + 1);
      setSuggestedTagsRefreshToken((value) => value + 1);
      setAiMetadataRefreshToken((value) => value + 1);
      setClassificationRefreshToken((value) => value + 1);
      setEntitiesRefreshToken((value) => value + 1);
      setRelatedRefreshToken((value) => value + 1);
      setActionSuggestionsRefreshToken((value) => value + 1);
    } finally {
      setIsSummarizingDocument(false);
    }
  };

  return (
    <aside className={`flex h-full min-h-0 w-full flex-col bg-white ${isDrawer ? "" : "border-l border-line"}`}>
      <div className="flex items-center justify-between gap-2 border-b border-line px-4 py-3">
        <div>
          <h2 className="text-sm font-bold text-ink">Belge paneli</h2>
          <p className="text-xs text-slate-500">Metadata, paylaşım ve AI araçları</p>
        </div>
        <div className="flex items-center gap-1">
          {!isDrawer && onCollapsePanel ? (
            <IconButton aria-label="Paneli gizle" className="hidden lg:inline-flex" onClick={onCollapsePanel}>
              <PanelRightClose className="h-4 w-4" />
            </IconButton>
          ) : null}
          {isDrawer ? (
            <button type="button" className="text-sm font-semibold text-violet-brand" onClick={onClose}>
              Kapat
            </button>
          ) : null}
        </div>
      </div>

      <WorkspacePanelNav
        activeSection={activeSection}
        folderLabel={folderLabels.singular}
        onSelect={(section) => onFocusSection?.(section)}
      />

      <div className="min-h-0 flex-1 overflow-y-auto bg-slate-50/60">
        <WorkspaceAccordionSection
          id="workspace-general"
          title="Genel Bilgiler"
          icon={FileText}
          isOpen={openSections.general}
          onToggle={() => onToggleSection("general")}
          focusToken={sectionFocus("general")}
        >
          <dl className="space-y-2 text-sm">
            <InfoRow label="Belge adı" value={document.original_file_name} />
            <InfoRow label="Orijinal boyut" value={formatFileSize(document.file_size)} />
            <InfoRow label="Sıkıştırılmış boyut" value={document.compressedSize ? formatFileSize(document.compressedSize) : "-"} />
            <InfoRow label="MIME türü" value={document.mime_type || "-"} />
            <InfoRow label="Kodlama" value={document.contentEncoding || "-"} />
            <InfoRow label="Dosya uzantısı" value={document.file_extension.toUpperCase()} />
            <InfoRow label="Checksum (SHA-256)" value={document.sha256Checksum || "-"} mono />
            <InfoRow label="Depolama sağlayıcı" value={document.storageProvider || "-"} />
            <InfoRow label="Depolama durumu" value={document.storageStatus || "available"} />
            <InfoRow label="Upload tarihi" value={formatDateTime(document.created_at)} />
            <InfoRow label="Son güncelleme" value={formatDateTime(document.updated_at)} />
            <InfoRow label="Kaynak" value={sourceLabel(document.sourceType)} />
            <div className="rounded-xl border border-line bg-white px-3 py-2.5">
              <dt className="text-xs font-semibold text-slate-500">Yaşam döngüsü</dt>
              <dd className="mt-1.5">
                <StatusBadge status={document.lifecycleStatus ?? document.status} />
              </dd>
            </div>
            {document.isLocked ? (
              <div className="rounded-lg border border-violet-200 bg-violet-soft/50 px-3 py-2 text-xs font-semibold text-violet-brand">
                Bu belge kilitli. Yeniden adlandırma, silme ve versiyon yükleme engellenir.
              </div>
            ) : null}
          </dl>
          {document.sourceType && document.sourceType !== "manual_upload" ? (
            <div className="mt-3">
              <DocumentSourceBadge document={document} />
            </div>
          ) : null}
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-tags"
          title="Etiketler"
          icon={Tags}
          isOpen={openSections.tags}
          onToggle={() => onToggleSection("tags")}
          focusToken={sectionFocus("tags")}
        >
          <div className="[&_section.panel]:border-0 [&_section.panel]:shadow-none [&_section.panel>div:first-child]:hidden">
            <DocumentTagsPanel documentId={document.id} selectedTags={document.tags ?? []} onUpdated={onTagsUpdated} />
          </div>
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-folder"
          title={folderLabels.singular}
          icon={FolderOpen}
          isOpen={openSections.folder}
          onToggle={() => onToggleSection("folder")}
          focusToken={sectionFocus("folder")}
        >
          {document.folder_id ? (
            <Link className="text-sm font-semibold text-violet-brand hover:underline" to={`/folders?folderId=${document.folder_id}`}>
              {document.folder_name || "Klasöre git"}
            </Link>
          ) : (
            <p className="text-sm text-slate-500">Bu belge bir {folderLabels.singular.toLowerCase()}e atanmamış.</p>
          )}
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-shares"
          title="Paylaşımlar"
          icon={Share2}
          isOpen={openSections.shares}
          onToggle={() => onToggleSection("shares")}
          focusToken={sectionFocus("shares")}
        >
          <DocumentWorkspaceSharesSection documentId={document.id} onCreateShare={onCreateShare} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-comments"
          title="Yorumlar"
          icon={MessageSquare}
          isOpen={openSections.comments}
          onToggle={() => onToggleSection("comments")}
          focusToken={sectionFocus("comments")}
        >
          <div className="[&_section.panel]:border-0 [&_section.panel]:shadow-none [&_section.panel>div:first-child]:hidden">
            <DocumentCommentsPanel document={document} />
          </div>
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-versions"
          title="Versiyonlar"
          icon={GitBranch}
          isOpen={openSections.versions}
          onToggle={() => onToggleSection("versions")}
          focusToken={sectionFocus("versions")}
        >
          <div className="[&_section.panel]:border-0 [&_section.panel]:shadow-none [&_section.panel>div:first-child]:hidden">
            <DocumentVersionsPanel document={document} onVersionsChanged={onVersionsChanged} />
          </div>
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-dates"
          title="Önemli Tarihler"
          icon={CalendarClock}
          isOpen={openSections.dates}
          onToggle={() => onToggleSection("dates")}
          focusToken={sectionFocus("dates")}
        >
          <div className="[&_section.panel]:border-0 [&_section.panel]:shadow-none [&_section.panel>div:first-child]:hidden">
            <DocumentImportantDatesPanel document={document} refreshToken={aiProgressRefreshToken + aiMetadataRefreshToken} />
          </div>
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-workflow"
          title="Onay Süreci"
          icon={ClipboardCheck}
          isOpen={openSections.workflow}
          onToggle={() => onToggleSection("workflow")}
          focusToken={sectionFocus("workflow")}
        >
          <div className="[&_section.panel]:border-0 [&_section.panel]:shadow-none [&_section.panel>div:first-child]:hidden">
            <DocumentWorkflowPanel document={document} />
          </div>
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-activity"
          title="Aktivite"
          icon={Activity}
          isOpen={openSections.activity}
          onToggle={() => onToggleSection("activity")}
          focusToken={sectionFocus("activity")}
        >
          <DocumentWorkspaceActivitySection documentId={document.id} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-related"
          title="İlgili Belgeler"
          icon={Link2}
          isOpen={openSections.related}
          onToggle={() => onToggleSection("related")}
          focusToken={sectionFocus("related")}
        >
          <DocumentWorkspaceRelatedSection document={document} refreshToken={relatedRefreshToken + aiProgressRefreshToken} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-summary"
          title="AI Özeti"
          icon={Sparkles}
          isOpen={openSections.summary}
          onToggle={() => onToggleSection("summary")}
          focusToken={sectionFocus("summary")}
        >
          <DocumentSummaryPanel documentId={document.id} refreshToken={summaryRefreshToken} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-ai-metadata"
          title="AI Metadata"
          icon={ScanSearch}
          isOpen={openSections.aiMetadata}
          onToggle={() => onToggleSection("aiMetadata")}
          focusToken={sectionFocus("aiMetadata")}
        >
          <DocumentAiMetadataPanel documentId={document.id} refreshToken={aiMetadataRefreshToken + aiProgressRefreshToken} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-action-suggestions"
          title="AI Önerileri"
          icon={ListChecks}
          isOpen={openSections.actionSuggestions}
          onToggle={() => onToggleSection("actionSuggestions")}
          focusToken={sectionFocus("actionSuggestions")}
        >
          <DocumentActionSuggestionsPanel
            documentId={document.id}
            refreshToken={actionSuggestionsRefreshToken + aiProgressRefreshToken + classificationRefreshToken + aiMetadataRefreshToken}
            onFocusSection={onFocusSection}
            onActionExecuted={() => {
              setSummaryRefreshToken((value) => value + 1);
              setAiProgressRefreshToken((value) => value + 1);
              setSuggestedTagsRefreshToken((value) => value + 1);
              setAiMetadataRefreshToken((value) => value + 1);
              setClassificationRefreshToken((value) => value + 1);
              setEntitiesRefreshToken((value) => value + 1);
              setRelatedRefreshToken((value) => value + 1);
              setActionSuggestionsRefreshToken((value) => value + 1);
            }}
          />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-health"
          title="Belge Sağlığı"
          icon={HeartPulse}
          isOpen={openSections.health}
          onToggle={() => onToggleSection("health")}
          focusToken={sectionFocus("health")}
        >
          <DocumentHealthPanel documentId={document.id} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-suggested-tags"
          title="AI Etiket Önerileri"
          icon={WandSparkles}
          isOpen={openSections.suggestedTags}
          onToggle={() => onToggleSection("suggestedTags")}
          focusToken={sectionFocus("suggestedTags")}
        >
          <DocumentSuggestedTagsPanel
            documentId={document.id}
            refreshToken={suggestedTagsRefreshToken + aiProgressRefreshToken}
            onTagsUpdated={onTagsUpdated}
          />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-classification"
          title="Belge Türü"
          icon={FileType}
          isOpen={openSections.classification}
          onToggle={() => onToggleSection("classification")}
          focusToken={sectionFocus("classification")}
        >
          <DocumentClassificationPanel
            documentId={document.id}
            refreshToken={classificationRefreshToken + aiProgressRefreshToken}
          />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-entities"
          title="AI Varlıklar"
          icon={Layers}
          isOpen={openSections.entities}
          onToggle={() => onToggleSection("entities")}
          focusToken={sectionFocus("entities")}
        >
          <DocumentEntitiesPanel documentId={document.id} refreshToken={entitiesRefreshToken + aiProgressRefreshToken} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-ai-progress"
          title="AI İlerlemesi"
          icon={ListChecks}
          isOpen={openSections.aiProgress}
          onToggle={() => onToggleSection("aiProgress")}
          focusToken={sectionFocus("aiProgress")}
        >
          <DocumentAiProgressPanel documentId={document.id} refreshToken={aiProgressRefreshToken} />
        </WorkspaceAccordionSection>

        <WorkspaceAccordionSection
          id="workspace-ai"
          title="AI Asistan"
          icon={Bot}
          isOpen={openSections.ai}
          onToggle={() => onToggleSection("ai")}
          focusToken={sectionFocus("ai")}
        >
          <div className="flex min-h-[480px] flex-col">
            <div className="border-b border-line px-3 py-2">
              <AiModeToggle mode={aiMode} onChange={setAiMode} compact />
            </div>
            {aiMode === "rag" ? (
              <AiChatPanel
                documentId={document.id}
                compact
                showSessionHistory
                className="min-h-[480px]"
                onSummarizeDocument={() => void handleSummarizeDocument()}
                isSummarizingDocument={isSummarizingDocument}
              />
            ) : (
              <AgentChatPanel documentId={document.id} compact showSessionHistory className="min-h-[480px]" />
            )}
          </div>
        </WorkspaceAccordionSection>
      </div>
    </aside>
  );
}

function InfoRow({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="rounded-xl border border-line bg-white px-3 py-2.5">
      <dt className="text-xs font-semibold text-slate-500">{label}</dt>
      <dd className={`mt-1 break-words text-sm font-semibold text-ink ${mono ? "font-mono text-xs" : ""}`}>{value}</dd>
    </div>
  );
}
