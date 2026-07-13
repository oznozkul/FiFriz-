export type ActionSuggestionType =
  | "add_tag"
  | "move_folder"
  | "create_reminder"
  | "start_approval"
  | "link_related"
  | "regenerate_metadata"
  | "generate_summary"
  | "favorite_document";

export type ActionSuggestionStatus = "pending" | "accepted" | "rejected";

export type ActionSuggestion = {
  id: string;
  documentId?: string | null;
  suggestionType: ActionSuggestionType;
  title: string;
  description?: string | null;
  actionPayload: Record<string, unknown>;
  status: ActionSuggestionStatus;
  createdAt: string;
  updatedAt: string;
};

export type ActionSuggestionExecution = {
  toolName: string;
  success: boolean;
  output?: Record<string, unknown>;
  error?: string;
};

export type ActionSuggestionAcceptResponse = {
  suggestion: ActionSuggestion;
  executed: boolean;
  execution?: ActionSuggestionExecution | null;
};

export const ACTION_SUGGESTION_TYPE_LABELS: Record<ActionSuggestionType, string> = {
  add_tag: "Etiket ekle",
  move_folder: "Klasöre taşı",
  create_reminder: "Hatırlatıcı",
  start_approval: "Onay süreci",
  link_related: "İlişkilendir",
  regenerate_metadata: "Metadata analizi",
  generate_summary: "Özet oluştur",
  favorite_document: "Favorilere ekle",
};

export type WorkspacePanelFromPayload = "tags" | "folder" | "dates" | "workflow" | "related" | "aiMetadata" | "summary";

export function panelFromActionPayload(payload: Record<string, unknown>): WorkspacePanelFromPayload | null {
  const panel = payload.panel;
  if (typeof panel !== "string") return null;
  const allowed: WorkspacePanelFromPayload[] = ["tags", "folder", "dates", "workflow", "related", "aiMetadata", "summary"];
  return allowed.includes(panel as WorkspacePanelFromPayload) ? (panel as WorkspacePanelFromPayload) : null;
}

export function toolNameFromPayload(payload: Record<string, unknown>): string | null {
  const toolName = payload.toolName;
  return typeof toolName === "string" && toolName.trim() ? toolName.trim() : null;
}

export function confirmationMessageFromPayload(payload: Record<string, unknown>, fallbackTitle: string): string {
  const confirmMessage = payload.confirmMessage;
  if (typeof confirmMessage === "string" && confirmMessage.trim()) {
    return confirmMessage;
  }
  const toolName = toolNameFromPayload(payload);
  if (!toolName) {
    return `Bu işlem şu aksiyonu gerçekleştirecek: ${fallbackTitle}`;
  }
  return TOOL_CONFIRMATION_MESSAGES[toolName] ?? `Bu işlem şu aksiyonu gerçekleştirecek: ${fallbackTitle}`;
}

const TOOL_CONFIRMATION_MESSAGES: Record<string, string> = {
  add_tag_to_document: "Bu belgeye önerilen etiket eklenecek.",
  move_document_to_folder: "Bu belge önerilen klasöre taşınacak.",
  create_important_date: "Bu belge için önemli tarih oluşturulacak.",
  create_workflow: "Bu belge için onay süreci başlatılacak.",
  create_document_relationship: "Bu belge önerilen belgeyle ilişkilendirilecek.",
  generate_summary: "Bu belge için AI özeti oluşturulacak.",
  reprocess_document: "Bu belge yeniden işlenecek ve analiz süreci başlatılacak.",
  favorite_document: "Bu belge favorilere eklenecek.",
};
