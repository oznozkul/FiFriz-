import { useEffect, useState } from "react";
import { PageState } from "../../components/common";
import { useAuth } from "../../services/AuthContext";
import { getProfileTenant } from "../../services/profileService";
import type { TenantProfile } from "../../types/auth";
import { roleLabels, tenantTypeLabels } from "./constants";
import { InfoItem, SettingsSection } from "./shared";

export function WorkspaceSettingsPage() {
  const { activeTenant } = useAuth();
  const [tenantInfo, setTenantInfo] = useState<TenantProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const loadTenant = () => {
    setIsLoading(true);
    setLoadError(null);
    getProfileTenant()
      .then(setTenantInfo)
      .catch((error) => {
        setTenantInfo(null);
        setLoadError(error instanceof Error ? error.message : "Workspace bilgisi yüklenemedi.");
      })
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    loadTenant();
  }, []);

  return (
    <SettingsSection title="Workspace" description="Aktif çalışma alanınız ve workspace bilgileri.">
      <PageState
        isLoading={isLoading}
        error={loadError}
        isEmpty={!tenantInfo}
        variant="inline"
        loadingLabel="Workspace bilgisi yükleniyor..."
        errorTitle="Workspace bilgisi yüklenemedi"
        onRetry={loadTenant}
        emptyTitle="Workspace bilgisi bulunamadı"
        emptyDescription="Çalışma alanı bilgileri şu an görüntülenemiyor."
      >
        <dl className="grid gap-4 sm:grid-cols-2">
          <InfoItem label="Workspace adı" value={tenantInfo!.tenantName} />
          <InfoItem label="Workspace tipi" value={tenantTypeLabels[tenantInfo!.tenantType] ?? tenantInfo!.tenantType} />
          <InfoItem label="Rolünüz" value={roleLabels[tenantInfo!.userRole] ?? tenantInfo!.userRole} />
          <InfoItem label="Üye sayısı" value={String(tenantInfo!.memberCount)} />
          <InfoItem label="Oluşturulma" value={new Date(tenantInfo!.createdAt).toLocaleDateString("tr-TR")} />
          {activeTenant ? <InfoItem label="Workspace ID" value={activeTenant.id} mono /> : null}
        </dl>
      </PageState>
    </SettingsSection>
  );
}
