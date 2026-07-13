import { apiRequest } from "./apiClient";

export type WorkflowStep = {
  id: string;
  stepOrder: number;
  approverUserId: string;
  approverFullName: string;
  status: "pending" | "approved" | "rejected" | "skipped";
  note: string | null;
  decidedAt: string | null;
  isCurrent: boolean;
  createdAt: string;
};

export type Workflow = {
  id: string;
  documentId: string;
  documentName: string;
  createdByUserId: string;
  createdByFullName: string;
  title: string;
  description: string | null;
  status: "pending" | "approved" | "rejected" | "cancelled";
  currentStepOrder: number;
  steps: WorkflowStep[];
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
};

export type PendingApproval = {
  workflowId: string;
  stepId: string;
  documentId: string;
  documentName: string;
  workflowTitle: string;
  stepOrder: number;
  createdByUserId: string;
  createdByFullName: string;
  createdAt: string;
};

export type CreateWorkflowStepInput = {
  approverUserId: string;
  stepOrder: number;
};

export function listDocumentWorkflows(documentId: string) {
  return apiRequest<Workflow[]>(`/api/v1/documents/${documentId}/workflows`, undefined, "Onay süreçleri yüklenemedi.");
}

export function createDocumentWorkflow(documentId: string, input: { title: string; description?: string; steps: CreateWorkflowStepInput[] }) {
  return apiRequest<Workflow>(`/api/v1/documents/${documentId}/workflows`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  }, "Onay süreci oluşturulamadı.");
}

export function getWorkflow(id: string) {
  return apiRequest<Workflow>(`/api/v1/workflows/${id}`, undefined, "Onay süreci yüklenemedi.");
}

export function listPendingApprovals() {
  return apiRequest<PendingApproval[]>("/api/v1/workflows/pending-for-me", undefined, "Bekleyen onaylar yüklenemedi.").then(
    (data) => data ?? [],
  );
}

export function approveWorkflow(id: string, note?: string) {
  return apiRequest<Workflow>(`/api/v1/workflows/${id}/approve`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ note: note ?? null }),
  }, "Onay işlemi başarısız.");
}

export function rejectWorkflow(id: string, note?: string) {
  return apiRequest<Workflow>(`/api/v1/workflows/${id}/reject`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ note: note ?? null }),
  }, "Red işlemi başarısız.");
}

export function cancelWorkflow(id: string) {
  return apiRequest<Workflow>(`/api/v1/workflows/${id}/cancel`, { method: "POST" }, "İptal işlemi başarısız.");
}

export function workflowStatusLabel(status: Workflow["status"]) {
  const labels: Record<Workflow["status"], string> = {
    pending: "Bekliyor",
    approved: "Onaylandı",
    rejected: "Reddedildi",
    cancelled: "İptal",
  };
  return labels[status];
}

export function stepStatusLabel(status: WorkflowStep["status"]) {
  const labels: Record<WorkflowStep["status"], string> = {
    pending: "Bekliyor",
    approved: "Onaylandı",
    rejected: "Reddedildi",
    skipped: "Atlandı",
  };
  return labels[status];
}
