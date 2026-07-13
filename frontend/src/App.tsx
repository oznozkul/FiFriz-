import { Navigate, Outlet, Route, Routes, useLocation } from "react-router-dom";
import { LoadingState } from "./components/common";
import { AppLayout, SettingsLayout } from "./layout";
import {
  AiAssistantPage,
  ActivityPage,
  AppearanceSettingsPage,
  BillingSettingsPage,
  DashboardPage,
  DocumentDetailPage,
  DocumentsPage,
  FavoritesPage,
  FoldersPage,
  ForgotPasswordPage,
  ImportantDatesPage,
  IntegrationsPage,
  LoginPage,
  MembersSettingsPage,
  NotificationSettingsPage,
  NotificationsPage,
  PendingApprovalsPage,
  PlanSettingsPage,
  ProfilePage,
  ProfileSettingsPage,
  PublicSharePage,
  QaHistoryPage,
  RecentPage,
  RegisterPage,
  ResetPasswordPage,
  RolesPermissionsSettingsPage,
  SearchResultsPage,
  SecuritySettingsPage,
  SharedWithMePage,
  SmartCollectionDetailPage,
  SmartCollectionsPage,
  SummariesPage,
  StorageSettingsPage,
  TagsPage,
  UsageSettingsPage,
  WorkspaceSettingsPage,
  LandingPage,
} from "./pages";
import { useAuth } from "./services/AuthContext";

function LandingRoute() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingState variant="fullscreen" label="Yükleniyor..." />;
  }

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return <LandingPage />;
}

function RequireAuthLayout() {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return <LoadingState variant="fullscreen" label="Oturum yükleniyor..." />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <Outlet />;
}

function RequireAuditAccessLayout() {
  const { activeTenant } = useAuth();
  if (activeTenant?.role === "member") {
    return <Navigate to="/dashboard" replace />;
  }
  return <Outlet />;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<LandingRoute />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} />
      <Route path="/share/:token" element={<PublicSharePage />} />

      <Route element={<RequireAuthLayout />}>
        <Route element={<AppLayout />}>
          <Route path="dashboard" element={<DashboardPage />} />
          <Route path="documents" element={<DocumentsPage />} />
          <Route path="collections" element={<SmartCollectionsPage />} />
          <Route path="collections/:key" element={<SmartCollectionDetailPage />} />
          <Route path="favorites" element={<FavoritesPage />} />
          <Route path="important-dates" element={<ImportantDatesPage />} />
          <Route path="shared-with-me" element={<SharedWithMePage />} />
          <Route path="notifications" element={<NotificationsPage />} />
          <Route path="approvals" element={<PendingApprovalsPage />} />
          <Route path="integrations" element={<IntegrationsPage />} />
          <Route element={<RequireAuditAccessLayout />}>
            <Route path="activity" element={<ActivityPage />} />
          </Route>
          <Route path="folders" element={<FoldersPage />} />
          <Route path="tags" element={<TagsPage />} />
          <Route path="search" element={<SearchResultsPage />} />
          <Route path="recent" element={<RecentPage />} />
          <Route path="trash" element={<DocumentsPage mode="trash" />} />
          <Route path="documents/:id" element={<DocumentDetailPage />} />
          <Route path="ai-assistant" element={<AiAssistantPage />} />
          <Route path="qa-history" element={<QaHistoryPage />} />
          <Route path="summaries" element={<SummariesPage />} />
          <Route path="settings" element={<SettingsLayout />}>
            <Route index element={<Navigate to="/settings/profile" replace />} />
            <Route path="profile" element={<ProfileSettingsPage />} />
            <Route path="security" element={<SecuritySettingsPage />} />
            <Route path="workspace" element={<WorkspaceSettingsPage />} />
            <Route path="members" element={<MembersSettingsPage />} />
            <Route path="notifications" element={<NotificationSettingsPage />} />
            <Route path="roles" element={<RolesPermissionsSettingsPage />} />
            <Route path="storage" element={<StorageSettingsPage />} />
            <Route path="appearance" element={<AppearanceSettingsPage />} />
            <Route path="plan" element={<PlanSettingsPage />} />
            <Route path="usage" element={<UsageSettingsPage />} />
            <Route path="billing" element={<BillingSettingsPage />} />
            <Route path="integrations" element={<Navigate to="/integrations" replace />} />
          </Route>
          <Route path="profile" element={<ProfilePage />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Route>
      </Route>
    </Routes>
  );
}

function App() {
  return <AppRoutes />;
}

export default App;
