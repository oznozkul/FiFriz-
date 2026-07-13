import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ArrowUpRight } from "lucide-react";
import { PageState, PrimaryButton } from "../../components/common";
import { integrationStatusLabel, listIntegrations, resolveProviderDisplayName } from "../../services/integrationService";
import type { IntegrationAccount } from "../../services/integrationService";
import { SettingsSection } from "./shared";

export function IntegrationSettingsPage() {
  const [integrations, setIntegrations] = useState<IntegrationAccount[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const loadIntegrations = () => {
    setIsLoading(true);
    setLoadError(null);
    listIntegrations()
      .then(setIntegrations)
      .catch((error) => {
        setIntegrations([]);
        setLoadError(error instanceof Error ? error.message : "Entegrasyonlar yüklenemedi.");
      })
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    loadIntegrations();
  }, []);

  const connected = integrations.filter((item) => item.isActive && item.status !== "disconnected");

  return (
    <div className="space-y-5">
      <SettingsSection title="Entegrasyonlar" description="Bağlı hesaplarınızın özeti ve yönetim bağlantıları.">
        <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-slate-500">
            {connected.length > 0 ? `${connected.length} aktif entegrasyon` : "Henüz bağlı entegrasyon yok."}
          </p>
          <Link to="/integrations">
            <PrimaryButton className="h-10">
              Entegrasyonları yönet
              <ArrowUpRight className="h-4 w-4" />
            </PrimaryButton>
          </Link>
        </div>

        <PageState
          isLoading={isLoading}
          error={loadError}
          isEmpty={integrations.length === 0}
          variant="inline"
          loadingLabel="Entegrasyonlar yükleniyor..."
          errorTitle="Entegrasyonlar yüklenemedi"
          onRetry={loadIntegrations}
          emptyTitle="Entegrasyon bulunamadı"
          emptyDescription="Gmail, Outlook, Google Drive ve diğer kaynakları bağlamak için entegrasyonlar sayfasını kullanın."
          emptyAction={
            <Link to="/integrations">
              <PrimaryButton>Entegrasyon ekle</PrimaryButton>
            </Link>
          }
        >
          <div className="space-y-3">
            {integrations.map((integration) => (
              <div key={integration.id} className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-line bg-white px-4 py-3">
                <div>
                  <p className="text-sm font-bold text-ink">{integration.displayName}</p>
                  <p className="text-xs text-slate-500">
                    {resolveProviderDisplayName(integration.provider)}
                    {integration.emailAddress ? ` · ${integration.emailAddress}` : ""}
                  </p>
                </div>
                <span
                  className={`rounded-md px-2 py-1 text-xs font-bold ${
                    integration.isActive ? "bg-emerald-50 text-emerald-700" : "bg-slate-100 text-slate-600"
                  }`}
                >
                  {integrationStatusLabel(integration.status)}
                </span>
              </div>
            ))}
          </div>
        </PageState>
      </SettingsSection>
    </div>
  );
}
