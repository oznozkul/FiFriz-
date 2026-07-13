import { Cloud, HardDrive, History, KeyRound, Mail, Plug, Radio, Webhook } from "lucide-react";
import { useCallback, useEffect, useMemo, useState, type ComponentType } from "react";
import { useSearchParams } from "react-router-dom";
import { IntegrationCreateModal } from "../components/integrations/IntegrationCreateModal";
import { IntegrationDetailModal } from "../components/integrations/IntegrationDetailModal";
import { IntegrationStatusBadge } from "../components/integrations/IntegrationStatusBadge";
import { EmptyState, ErrorState, IconButton, PageLoader, PageTitle, PrimaryButton, useConfirm } from "../components/common";
import {
  buildIntegrationMenuItems,
  ContextMenuActions,
  getIntegrationMenuContext,
  useContextMenuActionHandler,
} from "../contextMenu";
import { useContextMenu } from "../hooks/useContextMenu";
import { useAuth } from "../services/AuthContext";
import {
  connectIntegrationProvider,
  createIntegration,
  deleteIntegration,
  listIntegrationProviders,
  listIntegrations,
  queueIntegrationSync,
  type IntegrationAccount,
  type IntegrationProviderDescriptor,
  type ProviderCapabilities,
} from "../services/integrationService";
import { toast } from "../services/toast";

function providerIcon(capabilities: ProviderCapabilities): ComponentType<{ className?: string }> {
  if (capabilities.supportsFolders) return Cloud;
  if (capabilities.supportsAttachments) return Mail;
  return HardDrive;
}

function CapabilityBadge({ icon: Icon, label }: { icon: ComponentType<{ className?: string }>; label: string }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-600">
      <Icon className="h-3 w-3" />
      {label}
    </span>
  );
}

function CapabilityBadges({ capabilities }: { capabilities: ProviderCapabilities }) {
  return (
    <div className="mt-3 flex flex-wrap gap-1.5">
      {capabilities.supportsOAuth ? <CapabilityBadge icon={KeyRound} label="OAuth" /> : null}
      {capabilities.supportsWebhook ? <CapabilityBadge icon={Webhook} label="Webhook" /> : null}
      {capabilities.supportsHistoryImport ? <CapabilityBadge icon={History} label="Geçmiş" /> : null}
      {capabilities.supportsRealtime ? <CapabilityBadge icon={Radio} label="Gerçek zamanlı" /> : null}
    </div>
  );
}

export function IntegrationsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const { activeTenant } = useAuth();
  const role = activeTenant?.role ?? "member";
  const { menu, bindContextMenu } = useContextMenu();
  const confirm = useConfirm();
  const [integrations, setIntegrations] = useState<IntegrationAccount[]>([]);
  const [providers, setProviders] = useState<IntegrationProviderDescriptor[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [createProvider, setCreateProvider] = useState<IntegrationProviderDescriptor | null>(null);
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationAccount | null>(null);

  const integrationsByProvider = useMemo(() => {
    const map = new Map<string, IntegrationAccount[]>();
    for (const item of integrations) {
      const current = map.get(item.provider) ?? [];
      current.push(item);
      map.set(item.provider, current);
    }
    return map;
  }, [integrations]);

  const fetchIntegrations = async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      const [items, providerList] = await Promise.all([listIntegrations(), listIntegrationProviders()]);
      setIntegrations(items);
      setProviders(providerList);
      setSelectedIntegration((current) => (current ? items.find((item) => item.id === current.id) ?? null : null));
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Entegrasyonlar yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchIntegrations();
  }, []);

  useEffect(() => {
    if (providers.length === 0) return;

    let changed = false;
    const next = new URLSearchParams(searchParams);

    for (const provider of providers) {
      if (!provider.capabilities.supportsOAuth) continue;
      const status = searchParams.get(provider.name);
      if (!status) continue;

      if (status === "connected") {
        toast.success(`${provider.displayName} hesabı başarıyla bağlandı.`);
        void fetchIntegrations();
      } else if (status === "error") {
        toast.error(`${provider.displayName} bağlantısı tamamlanamadı.`);
      }

      next.delete(provider.name);
      changed = true;
    }

    if (changed) {
      setSearchParams(next, { replace: true });
    }
  }, [providers, searchParams, setSearchParams]);

  const handleConnect = async (descriptor: IntegrationProviderDescriptor) => {
    if (!descriptor.available) {
      toast.error(`${descriptor.displayName} henüz yapılandırılmamış.`);
      return;
    }

    if (descriptor.capabilities.supportsOAuth) {
      try {
        const response = await connectIntegrationProvider(descriptor.name);
        window.location.href = response.authUrl;
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "Bağlantı başlatılamadı.");
      }
      return;
    }

    if (descriptor.capabilities.supportsPassword) {
      setCreateProvider(descriptor);
      return;
    }

    toast.error("Bu entegrasyon için bağlantı akışı henüz desteklenmiyor.");
  };

  const handleCreate = async (input: {
    provider: string;
    displayName: string;
    emailAddress?: string;
    settings?: Record<string, unknown>;
    credentials?: { password: string };
  }) => {
    try {
      await createIntegration(input);
      await fetchIntegrations();
      toast.success("Entegrasyon oluşturuldu.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Entegrasyon oluşturulamadı.");
      throw error;
    }
  };

  const handleContextMenuAction = useCallback(
    async (detail: { action: string; context: { entityType: string; entity: unknown } }) => {
      if (detail.context.entityType !== "integration") return;
      const integration = detail.context.entity as IntegrationAccount;

      switch (detail.action) {
        case ContextMenuActions.integration.sync:
          try {
            await queueIntegrationSync(integration.id);
            toast.success("Senkronizasyon kuyruğa alındı.");
          } catch (error) {
            toast.error(error instanceof Error ? error.message : "Senkronizasyon başlatılamadı.");
          }
          return;
        case ContextMenuActions.integration.settings:
          setSelectedIntegration(integration);
          return;
        case ContextMenuActions.integration.disable:
          toast.info("Devre dışı bırakma yakında eklenecek.");
          return;
        case ContextMenuActions.integration.delete: {
          const shouldDelete = await confirm({
            title: "Entegrasyonu sil",
            message: `${integration.displayName} entegrasyonu silinsin mi?`,
            confirmLabel: "Sil",
            variant: "danger",
          });
          if (!shouldDelete) return;
          try {
            await deleteIntegration(integration.id);
            await fetchIntegrations();
            toast.success("Entegrasyon silindi.");
          } catch (error) {
            toast.error(error instanceof Error ? error.message : "Entegrasyon silinemedi.");
          }
          return;
        }
      }
    },
    [],
  );

  useContextMenuActionHandler(handleContextMenuAction);

  return (
    <div className="space-y-6">
      <PageTitle
        title="Entegrasyonlar"
        subtitle="Mail ve bulut hesaplarınızı bağlayarak belgelerinizi otomatik içe aktarın."
      />

      {isLoading ? (
        <PageLoader label="Entegrasyonlar yükleniyor..." />
      ) : loadError ? (
        <ErrorState title="Entegrasyonlar yüklenemedi" description={loadError} onRetry={() => void fetchIntegrations()} />
      ) : (
        <>
          {integrations.length === 0 ? (
            <EmptyState title="Henüz entegrasyon yok" description="Aşağıdaki kartlardan birini seçerek ilk entegrasyonunuzu oluşturun." />
          ) : null}

          <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {providers.map((descriptor) => {
              const Icon = providerIcon(descriptor.capabilities);
              const linked = integrationsByProvider.get(descriptor.name) ?? [];
              const primary = linked[0];
              const unavailable = !descriptor.available;
              const integrationMenuProps = primary
                ? bindContextMenu({
                    items: buildIntegrationMenuItems({ integration: primary, role }),
                    context: getIntegrationMenuContext(primary),
                  })
                : null;

              return (
                <article
                  key={descriptor.name}
                  className={`panel flex flex-col p-5 ${unavailable ? "opacity-80" : ""}`}
                  onContextMenu={integrationMenuProps?.onContextMenu}
                  onTouchStart={integrationMenuProps?.onTouchStart}
                  onTouchMove={integrationMenuProps?.onTouchMove}
                  onTouchEnd={integrationMenuProps?.onTouchEnd}
                  onTouchCancel={integrationMenuProps?.onTouchCancel}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
                      <Icon className="h-5 w-5" />
                    </div>
                    {primary ? (
                      <IntegrationStatusBadge status={primary.status} />
                    ) : unavailable ? (
                      <span className="rounded-full border border-slate-200 bg-slate-100 px-2.5 py-0.5 text-xs font-semibold text-slate-600">Yapılandırılmadı</span>
                    ) : (
                      <span className="rounded-full border border-slate-200 bg-slate-100 px-2.5 py-0.5 text-xs font-semibold text-slate-600">Bağlı değil</span>
                    )}
                  </div>

                  <h3 className="mt-4 text-base font-bold text-ink">{descriptor.displayName}</h3>
                  <p className="mt-2 flex-1 text-sm text-slate-600">{descriptor.description}</p>
                  <CapabilityBadges capabilities={descriptor.capabilities} />

                  <div className="mt-4 flex flex-wrap gap-2">
                    {unavailable ? (
                      <button type="button" className="h-10 rounded-lg border border-line px-4 text-sm font-semibold text-slate-400" disabled>
                        Yapılandırılmadı
                      </button>
                    ) : primary ? (
                      <PrimaryButton className="h-10" onClick={() => setSelectedIntegration(primary)}>
                        <Plug className="h-4 w-4" />
                        Yönet
                      </PrimaryButton>
                    ) : (
                      <PrimaryButton className="h-10" onClick={() => void handleConnect(descriptor)}>
                        Bağla
                      </PrimaryButton>
                    )}
                  </div>

                  {linked.length > 1 ? (
                    <p className="mt-3 text-xs text-slate-500">{linked.length} bağlı hesap</p>
                  ) : null}
                </article>
              );
            })}
          </section>
        </>
      )}

      <IntegrationCreateModal
        isOpen={!!createProvider}
        provider={createProvider}
        onClose={() => setCreateProvider(null)}
        onSubmit={handleCreate}
      />

      <IntegrationDetailModal
        integration={selectedIntegration}
        providers={providers}
        isOpen={!!selectedIntegration}
        onClose={() => setSelectedIntegration(null)}
        onChanged={() => void fetchIntegrations()}
      />

      {menu}
    </div>
  );
}
