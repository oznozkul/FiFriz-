import { useCallback, useEffect, useState } from "react";
import {
  DashboardQuickActions,
  DashboardStatCards,
  DashboardWelcomeHeader,
  FavoriteDocumentsPanel,
  OnboardingChecklist,
  PendingApprovalsPanel,
  RecentActivitiesPanel,
  RecentDocumentsPanel,
  RecentSharesPanel,
  TopFoldersPanel,
  UpcomingImportantDatesPanel,
} from "../components/dashboard";
import { UploadDocumentModal } from "../components/documents";
import { ErrorState, PageTitle, PageState } from "../components/common";
import { listFavoriteDocuments } from "../data/documentsApi";
import { useDocumentFavorite } from "../hooks/useDocumentFavorite";
import { useFolderLabels } from "../hooks/useFolderLabels";
import { useImportantDateLabels } from "../hooks/useImportantDateLabels";
import { useOnboardingProgress } from "../hooks/useOnboardingProgress";
import { useAuth } from "../services/AuthContext";
import { getDashboardSummary, type DashboardRecentDocument, type DashboardSummary } from "../services/dashboardService";
import { listPendingApprovals, type PendingApproval } from "../services/workflowService";
import type { ApiDocument } from "../types/document";

export function DashboardPage() {
  const { user, activeTenant } = useAuth();
  const folderLabels = useFolderLabels();
  const dateLabels = useImportantDateLabels();
  const isPersonal = activeTenant?.type === "personal";

  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [summaryLoading, setSummaryLoading] = useState(true);
  const [summaryError, setSummaryError] = useState<string | null>(null);

  const [favorites, setFavorites] = useState<ApiDocument[]>([]);
  const [favoritesLoading, setFavoritesLoading] = useState(false);
  const [favoritesError, setFavoritesError] = useState<string | null>(null);

  const [pendingApprovals, setPendingApprovals] = useState<PendingApproval[]>([]);
  const [pendingLoading, setPendingLoading] = useState(false);
  const [pendingError, setPendingError] = useState<string | null>(null);

  const [uploadOpen, setUploadOpen] = useState(false);
  const [uploadTagNames, setUploadTagNames] = useState<string[]>([]);

  const onboarding = useOnboardingProgress({
    userId: user?.id,
    tenantId: activeTenant?.id,
    isPersonal,
    folderLabels,
    summary,
    isSummaryLoading: summaryLoading,
  });

  const { favoriteLoadingId, toggleFavorite } = useDocumentFavorite({
    onUpdated: (updatedDocument) => {
      setSummary((current) => {
        if (!current) return current;
        const previous = current.recentDocuments.find((item) => item.id === updatedDocument.id);
        const delta = updatedDocument.isFavorite && !previous?.isFavorite ? 1 : !updatedDocument.isFavorite && previous?.isFavorite ? -1 : 0;
        return {
          ...current,
          recentDocuments: current.recentDocuments.map((item) =>
            item.id === updatedDocument.id ? { ...item, isFavorite: updatedDocument.isFavorite } : item,
          ),
          favoriteDocuments: Math.max(0, current.favoriteDocuments + delta),
        };
      });
      setFavorites((current) => {
        if (!updatedDocument.isFavorite) {
          return current.filter((item) => item.id !== updatedDocument.id);
        }
        const exists = current.some((item) => item.id === updatedDocument.id);
        if (exists) {
          return current.map((item) => (item.id === updatedDocument.id ? updatedDocument : item));
        }
        return [updatedDocument, ...current].slice(0, 5);
      });
    },
  });

  const loadSummary = useCallback(async () => {
    setSummaryLoading(true);
    setSummaryError(null);
    try {
      setSummary(await getDashboardSummary());
    } catch (error) {
      setSummaryError(error instanceof Error ? error.message : "Dashboard verileri yüklenemedi.");
      setSummary(null);
    } finally {
      setSummaryLoading(false);
    }
  }, []);

  const loadFavorites = useCallback(async () => {
    if (!isPersonal) {
      setFavorites([]);
      return;
    }
    setFavoritesLoading(true);
    setFavoritesError(null);
    try {
      const items = await listFavoriteDocuments();
      setFavorites(items.slice(0, 5));
    } catch (error) {
      setFavoritesError(error instanceof Error ? error.message : "Favoriler yüklenemedi.");
      setFavorites([]);
    } finally {
      setFavoritesLoading(false);
    }
  }, [isPersonal]);

  const loadPendingApprovals = useCallback(async () => {
    if (isPersonal) {
      setPendingApprovals([]);
      return;
    }
    setPendingLoading(true);
    setPendingError(null);
    try {
      setPendingApprovals(await listPendingApprovals());
    } catch (error) {
      setPendingError(error instanceof Error ? error.message : "Bekleyen onaylar yüklenemedi.");
      setPendingApprovals([]);
    } finally {
      setPendingLoading(false);
    }
  }, [isPersonal]);

  useEffect(() => {
    void loadSummary();
    void loadFavorites();
    void loadPendingApprovals();
  }, [loadSummary, loadFavorites, loadPendingApprovals]);

  useEffect(() => {
    const refresh = () => {
      void loadSummary();
      void loadFavorites();
      void loadPendingApprovals();
    };
    window.addEventListener("documents:changed", refresh);
    return () => window.removeEventListener("documents:changed", refresh);
  }, [loadSummary, loadFavorites, loadPendingApprovals]);

  const openUpload = (tagNames: string[] = []) => {
    setUploadTagNames(tagNames);
    setUploadOpen(true);
  };

  const handleToggleRecentFavorite = (document: DashboardRecentDocument) => {
    void toggleFavorite({ id: document.id, isFavorite: document.isFavorite });
  };

  const handleToggleFavoriteDocument = (document: ApiDocument) => {
    void toggleFavorite({ id: document.id, isFavorite: document.isFavorite });
  };

  const refreshAll = () => {
    void loadSummary();
    void loadFavorites();
    void loadPendingApprovals();
  };

  if (summaryError && !summary && !summaryLoading) {
    return (
      <div className="space-y-6">
        <PageTitle title="Dashboard" subtitle="Çalışma alanı özetiniz" />
        <ErrorState title="Dashboard yüklenemedi" description={summaryError} onRetry={refreshAll} />
      </div>
    );
  }

  const topFoldersTitle = isPersonal ? "Kategorilere göre belgeler" : "En aktif klasörler";
  const topFoldersSubtitle = isPersonal
    ? "En çok belge içeren kategorileriniz"
    : "En çok belge barındıran klasörler";
  const topFoldersEmptyTitle = isPersonal ? "Henüz kategori yok" : "Henüz klasör yok";
  const topFoldersEmptyDescription = isPersonal
    ? "Kategori oluşturup belgelerinizi grupladıkça burada görünecek."
    : "Klasör oluşturup belgelerinizi organize ettikçe burada görünecek.";

  return (
    <div className="space-y-6">
      {isPersonal && user && activeTenant ? (
        <DashboardWelcomeHeader fullName={user.fullName} tenantName={activeTenant.name} isPersonal />
      ) : (
        <PageTitle title="Çalışma alanı" subtitle="Onaylar, aktiviteler ve belge hareketlerinizin özeti." />
      )}

      {onboarding.isVisible ? (
        <OnboardingChecklist
          steps={onboarding.steps}
          completedCount={onboarding.completedCount}
          totalCount={onboarding.totalCount}
          isLoading={onboarding.isLoading}
          isPersonal={isPersonal}
          onDismiss={onboarding.dismiss}
          onUpload={() => openUpload()}
        />
      ) : null}

      <DashboardQuickActions
        isPersonal={isPersonal}
        folderLabels={folderLabels}
        onUpload={() => openUpload()}
        onUploadWithTags={(tagNames) => openUpload(tagNames)}
      />

      <DashboardStatCards
        isPersonal={isPersonal}
        summary={summary}
        pendingApprovalsCount={pendingApprovals.length}
        isLoading={summaryLoading}
        upcomingLabel={dateLabels.dashboardUpcoming}
      />

      {isPersonal ? (
        <>
          <UpcomingImportantDatesPanel
            items={summary?.upcomingImportantDateItems ?? []}
            upcomingCount={summary?.upcomingImportantDates ?? 0}
            overdueCount={summary?.overdueImportantDates ?? 0}
            assignedCount={summary?.assignedImportantDates ?? 0}
            isLoading={summaryLoading}
            error={summaryError}
            onRetry={loadSummary}
          />
          <div className="grid gap-5 xl:grid-cols-2">
            <RecentDocumentsPanel
              documents={summary?.recentDocuments ?? []}
              favoriteLoadingId={favoriteLoadingId}
              onToggleFavorite={handleToggleRecentFavorite}
              isLoading={summaryLoading}
              error={summaryError}
              onRetry={loadSummary}
              onUpload={() => openUpload()}
            />
            <FavoriteDocumentsPanel
              documents={favorites}
              favoriteLoadingId={favoriteLoadingId}
              onToggleFavorite={handleToggleFavoriteDocument}
              isLoading={favoritesLoading}
              error={favoritesError}
              onRetry={loadFavorites}
              onUpload={() => openUpload()}
            />
          </div>
          <TopFoldersPanel
            folders={summary?.topFolders ?? []}
            title={topFoldersTitle}
            subtitle={topFoldersSubtitle}
            emptyTitle={topFoldersEmptyTitle}
            emptyDescription={topFoldersEmptyDescription}
            isLoading={summaryLoading}
            error={summaryError}
            onRetry={loadSummary}
          />
        </>
      ) : (
        <>
          <div className="grid gap-5 xl:grid-cols-2">
            <PendingApprovalsPanel items={pendingApprovals} isLoading={pendingLoading} error={pendingError} onRetry={loadPendingApprovals} />
            <UpcomingImportantDatesPanel
              items={summary?.upcomingImportantDateItems ?? []}
              upcomingCount={summary?.upcomingImportantDates ?? 0}
              overdueCount={summary?.overdueImportantDates ?? 0}
              assignedCount={summary?.assignedImportantDates ?? 0}
              isLoading={summaryLoading}
              error={summaryError}
              onRetry={loadSummary}
            />
          </div>
          <div className="grid gap-5 xl:grid-cols-2">
            <RecentActivitiesPanel
              activities={summary?.recentActivities ?? []}
              isLoading={summaryLoading}
              error={summaryError}
              onRetry={loadSummary}
            />
            <RecentSharesPanel
              activities={summary?.recentActivities ?? []}
              isLoading={summaryLoading}
              error={summaryError}
              onRetry={loadSummary}
            />
          </div>
          <div className="grid gap-5 xl:grid-cols-2">
            <RecentDocumentsPanel
              documents={summary?.recentDocuments ?? []}
              favoriteLoadingId={favoriteLoadingId}
              onToggleFavorite={handleToggleRecentFavorite}
              isLoading={summaryLoading}
              error={summaryError}
              onRetry={loadSummary}
              onUpload={() => openUpload()}
            />
            <TopFoldersPanel
              folders={summary?.topFolders ?? []}
              title={topFoldersTitle}
              subtitle={topFoldersSubtitle}
              emptyTitle={topFoldersEmptyTitle}
              emptyDescription={topFoldersEmptyDescription}
              isLoading={summaryLoading}
              error={summaryError}
              onRetry={loadSummary}
            />
          </div>
        </>
      )}

      <UploadDocumentModal
        isOpen={uploadOpen}
        initialTagNames={uploadTagNames}
        onClose={() => {
          setUploadOpen(false);
          setUploadTagNames([]);
        }}
      />
    </div>
  );
}
