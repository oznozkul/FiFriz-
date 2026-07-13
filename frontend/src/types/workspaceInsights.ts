export type WorkspaceInsightRelatedDocument = {
  documentId: string;
  originalFileName: string;
  fileExtension?: string;
};

export type WorkspaceInsightType =
  | "upcoming_dates"
  | "pending_approvals"
  | "recent_important_documents"
  | "classified_documents"
  | "missing_metadata"
  | "failed_processing"
  | "related_suggestions"
  | "popular_tags"
  | "active_folders";

export type WorkspaceInsightPriority = "high" | "medium" | "low";

export type WorkspaceInsight = {
  title: string;
  description: string;
  type: WorkspaceInsightType;
  priority: WorkspaceInsightPriority;
  relatedDocuments: WorkspaceInsightRelatedDocument[];
  actionLabel: string;
  actionUrl: string;
};

export const WORKSPACE_INSIGHT_PRIORITY_LABELS: Record<WorkspaceInsightPriority, string> = {
  high: "Yüksek",
  medium: "Orta",
  low: "Düşük",
};
