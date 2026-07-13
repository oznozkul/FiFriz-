import { useCallback, useEffect, useMemo, useState } from "react";
import {
  getOrganizationOnboardingSteps,
  getPersonalOnboardingSteps,
  type OnboardingStepDef,
  type OnboardingStepId,
} from "../components/dashboard/onboarding/onboardingConfig";
import {
  dismissOnboarding,
  isOnboardingDismissed,
  isStepManuallyCompleted,
  markOnboardingStep,
} from "../components/dashboard/onboarding/onboardingStorage";
import type { FolderLabels } from "./useFolderLabels";
import { getRecentSearches } from "../services/commandPaletteStorage";
import type { DashboardRecentActivity, DashboardSummary } from "../services/dashboardService";
import { listIntegrations } from "../services/integrationService";
import { listTenantMembers } from "../services/tenantService";

export type OnboardingStepProgress = OnboardingStepDef & {
  completed: boolean;
};

type UseOnboardingProgressOptions = {
  userId?: string;
  tenantId?: string;
  isPersonal: boolean;
  folderLabels: FolderLabels;
  summary: DashboardSummary | null;
  isSummaryLoading: boolean;
};

function hasActivity(activities: DashboardRecentActivity[] | undefined, ...actions: string[]) {
  return (activities ?? []).some((item) => actions.includes(item.action));
}

function hasCategorySelection(summary: DashboardSummary) {
  if (summary.totalFolders > 0) return true;
  if (summary.topFolders.some((folder) => folder.documentCount > 0)) return true;
  return summary.recentDocuments.some((document) => Boolean(document.folderName));
}

function hasImportantDate(summary: DashboardSummary) {
  const total =
    summary.upcomingImportantDates + summary.overdueImportantDates + summary.assignedImportantDates;
  if (total > 0) return true;
  return hasActivity(summary.recentActivities, "important_date_created", "ai_important_date_accepted");
}

function isStepCompleted(
  stepId: OnboardingStepId,
  isPersonal: boolean,
  summary: DashboardSummary | null,
  userId: string,
  tenantId: string,
  memberCount: number,
  hasConnectedIntegration: boolean,
): boolean {
  if (isStepManuallyCompleted(userId, tenantId, stepId)) return true;
  if (!summary) return stepId === "workspace-created" && !isPersonal;

  switch (stepId) {
    case "workspace-created":
      return !isPersonal;
    case "upload-document":
      return summary.totalDocuments > 0;
    case "select-category":
    case "create-folder":
      return hasCategorySelection(summary);
    case "add-important-date":
      return hasImportantDate(summary);
    case "search":
      return getRecentSearches().length > 0;
    case "ai-summary":
      return hasActivity(summary.recentActivities, "document_summary_generated");
    case "invite-member":
      return memberCount > 1;
    case "start-workflow":
      return hasActivity(summary.recentActivities, "workflow_created");
    case "connect-integration":
      return (
        hasConnectedIntegration ||
        hasActivity(
          summary.recentActivities,
          "integration_created",
          "integration_gmail_connected",
          "integration_imap_connected",
        )
      );
    default:
      return false;
  }
}

export function useOnboardingProgress({
  userId,
  tenantId,
  isPersonal,
  folderLabels,
  summary,
  isSummaryLoading,
}: UseOnboardingProgressOptions) {
  const [dismissed, setDismissed] = useState(false);
  const [memberCount, setMemberCount] = useState(1);
  const [hasConnectedIntegration, setHasConnectedIntegration] = useState(false);
  const [extrasLoading, setExtrasLoading] = useState(!isPersonal);
  const [manualRevision, setManualRevision] = useState(0);

  const stepDefs = useMemo(
    () => (isPersonal ? getPersonalOnboardingSteps(folderLabels) : getOrganizationOnboardingSteps(folderLabels)),
    [folderLabels, isPersonal],
  );

  useEffect(() => {
    if (!userId || !tenantId) return;
    setDismissed(isOnboardingDismissed(userId, tenantId));
  }, [userId, tenantId]);

  useEffect(() => {
    if (!userId || !tenantId) return;
    const handler = (event: Event) => {
      const stepId = (event as CustomEvent<{ stepId: OnboardingStepId }>).detail?.stepId;
      if (!stepId) return;
      markOnboardingStep(userId, tenantId, stepId);
      setManualRevision((current) => current + 1);
    };
    window.addEventListener("fifriz:onboarding-step", handler);
    return () => window.removeEventListener("fifriz:onboarding-step", handler);
  }, [userId, tenantId]);

  useEffect(() => {
    if (!userId || !tenantId || isPersonal) {
      setMemberCount(1);
      setHasConnectedIntegration(false);
      setExtrasLoading(false);
      return;
    }

    let cancelled = false;
    setExtrasLoading(true);

    Promise.all([listTenantMembers(), listIntegrations()])
      .then(([members, integrations]) => {
        if (cancelled) return;
        setMemberCount(members.length);
        setHasConnectedIntegration(
          integrations.some((item) => item.isActive && item.status !== "disconnected"),
        );
      })
      .catch(() => {
        if (!cancelled) {
          setMemberCount(1);
          setHasConnectedIntegration(false);
        }
      })
      .finally(() => {
        if (!cancelled) setExtrasLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [isPersonal, userId, tenantId, summary?.totalDocuments, summary?.totalFolders]);

  const steps = useMemo<OnboardingStepProgress[]>(() => {
    if (!userId || !tenantId) return [];
    return stepDefs.map((step) => ({
      ...step,
      completed: isStepCompleted(
        step.id,
        isPersonal,
        summary,
        userId,
        tenantId,
        memberCount,
        hasConnectedIntegration,
      ),
    }));
  }, [stepDefs, userId, tenantId, isPersonal, summary, memberCount, hasConnectedIntegration, manualRevision]);

  const completedCount = steps.filter((step) => step.completed).length;
  const totalCount = steps.length;
  const allCompleted = totalCount > 0 && completedCount === totalCount;
  const isLoading = isSummaryLoading || extrasLoading;
  const isVisible = Boolean(userId && tenantId && !dismissed && !allCompleted);

  const dismiss = useCallback(() => {
    if (!userId || !tenantId) return;
    dismissOnboarding(userId, tenantId);
    setDismissed(true);
  }, [userId, tenantId]);

  return {
    steps,
    completedCount,
    totalCount,
    allCompleted,
    isLoading,
    isVisible,
    dismiss,
  };
}
