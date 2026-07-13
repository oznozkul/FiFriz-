import type { WorkspaceInsight } from "../types/workspaceInsights";
import { apiRequest } from "./apiClient";

export function listWorkspaceInsights() {
  return apiRequest<WorkspaceInsight[]>("/api/v1/ai/workspace-insights", undefined, "AI içgörüleri yüklenemedi.");
}
