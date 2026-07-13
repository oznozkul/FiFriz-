import { Pencil, Plug, RefreshCw, Trash2, X } from "lucide-react";
import { useEffect, useMemo, useState, type ReactNode } from "react";
import type { IntegrationAccount, IntegrationProviderDescriptor, IntegrationSyncLog } from "../../services/integrationService";
import {
  deleteIntegration,
  getIntegration,
  listIntegrationSyncLogs,
  queueIntegrationSync,
  resolveProviderDisplayName,
  syncLogStatusLabel,
  updateIntegration,
} from "../../services/integrationService";
import { toast } from "../../services/toast";
import { EmptyState, LoadingState, PrimaryButton, Spinner, useConfirm } from "../common";
import { GmailSettingsFields } from "./GmailSettingsFields";
import { gmailFormStateFromSettings, gmailSettingsFromFormState, type GmailFormState } from "./gmailFormShared";
import { GoogleDriveSettingsFields } from "./GoogleDriveSettingsFields";
import { googleDriveFormStateFromSettings, googleDriveSettingsFromFormState, type GoogleDriveFormState } from "./googleDriveFormShared";
import { defaultSyncFromDate, maxEmailsPerSyncFromSettings, syncFromDateFromSettings } from "./imapFormShared";
import { OutlookSettingsFields } from "./OutlookSettingsFields";
import { outlookFormStateFromSettings, outlookSettingsFromFormState, type OutlookFormState } from "./outlookFormShared";
import { IntegrationStatusBadge } from "./IntegrationStatusBadge";

type IntegrationDetailModalProps = {
  integration: IntegrationAccount | null;
  providers?: IntegrationProviderDescriptor[];
  isOpen: boolean;
  onClose: () => void;
  onChanged: () => void;
};

export function IntegrationDetailModal({ integration, providers, isOpen, onClose, onChanged }: IntegrationDetailModalProps) {
  const confirm = useConfirm();
  const [syncLogs, setSyncLogs] = useState<IntegrationSyncLog[]>([]);
  const [detail, setDetail] = useState(integration);
  const [isLoadingLogs, setIsLoadingLogs] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [displayName, setDisplayName] = useState("");
  const [isActive, setIsActive] = useState(true);
  const [gmailForm, setGmailForm] = useState<GmailFormState>(gmailFormStateFromSettings(null));
  const [googleDriveForm, setGoogleDriveForm] = useState<GoogleDriveFormState>(googleDriveFormStateFromSettings(null));
  const [outlookForm, setOutlookForm] = useState<OutlookFormState>(outlookFormStateFromSettings(null));
  const [password, setPassword] = useState("");
  const [syncFromDate, setSyncFromDate] = useState(defaultSyncFromDate());
  const [maxEmailsPerSync, setMaxEmailsPerSync] = useState("500");

  useEffect(() => {
    if (!isOpen || !integration) return;
    setDetail(integration);
    setIsEditing(false);
    setIsLoadingLogs(true);
    Promise.all([getIntegration(integration.id), listIntegrationSyncLogs(integration.id)])
      .then(([detailResponse, logs]) => {
        setDetail(detailResponse);
        setSyncLogs(logs);
      })
      .catch((error) => toast.error(error instanceof Error ? error.message : "Entegrasyon detayı yüklenemedi."))
      .finally(() => setIsLoadingLogs(false));
  }, [isOpen, integration]);

  useEffect(() => {
    if (!detail) return;
    setDisplayName(detail.displayName);
    setIsActive(detail.isActive);
    setGmailForm(gmailFormStateFromSettings(detail.settings));
    setGoogleDriveForm(googleDriveFormStateFromSettings(detail.settings));
    setOutlookForm(outlookFormStateFromSettings(detail.settings));
    setSyncFromDate(syncFromDateFromSettings(detail.settings));
    setMaxEmailsPerSync(maxEmailsPerSyncFromSettings(detail.settings));
    setPassword("");
  }, [detail, isEditing]);

  const isIMAP = detail?.provider === "imap";
  const isGmail = detail?.provider === "gmail";
  const isGoogleDrive = detail?.provider === "google_drive";
  const isOneDrive = detail?.provider === "onedrive";
  const isCloudDrive = isGoogleDrive || isOneDrive;
  const isOutlook = detail?.provider === "outlook";

  const canSave = useMemo(() => {
    if (isIMAP) return syncFromDate.trim().length > 0;
    if (!displayName.trim()) return false;
    return true;
  }, [displayName, isIMAP, syncFromDate]);

  if (!isOpen || !integration || !detail) return null;

  const latestLog = syncLogs[0];
  const totalImportedFromLogs = syncLogs.reduce((sum, log) => sum + log.importedItems, 0);

  const startEditing = () => {
    setDisplayName(detail.displayName);
    setIsActive(detail.isActive);
    setGmailForm(gmailFormStateFromSettings(detail.settings));
    setGoogleDriveForm(googleDriveFormStateFromSettings(detail.settings));
    setOutlookForm(outlookFormStateFromSettings(detail.settings));
    setSyncFromDate(syncFromDateFromSettings(detail.settings));
    setMaxEmailsPerSync(maxEmailsPerSyncFromSettings(detail.settings));
    setPassword("");
    setIsEditing(true);
  };

  const cancelEditing = () => {
    setDisplayName(detail.displayName);
    setIsActive(detail.isActive);
    setGmailForm(gmailFormStateFromSettings(detail.settings));
    setGoogleDriveForm(googleDriveFormStateFromSettings(detail.settings));
    setOutlookForm(outlookFormStateFromSettings(detail.settings));
    setSyncFromDate(syncFromDateFromSettings(detail.settings));
    setMaxEmailsPerSync(maxEmailsPerSyncFromSettings(detail.settings));
    setPassword("");
    setIsEditing(false);
  };

  const handleSave = async () => {
    if (!canSave) return;
    setIsSaving(true);
    try {
      const updated = await updateIntegration(detail.id, {
        displayName: isIMAP ? undefined : displayName.trim(),
        isActive,
        settings: isIMAP
          ? { syncFromDate, maxEmailsPerSync: Number(maxEmailsPerSync) || 500 }
            : isGmail
            ? gmailSettingsFromFormState(gmailForm)
            : isCloudDrive
              ? googleDriveSettingsFromFormState(googleDriveForm)
            : isOutlook
              ? outlookSettingsFromFormState(outlookForm)
              : undefined,
        credentials: password.trim() ? { password } : undefined,
      });
      setDetail(updated);
      setIsEditing(false);
      setPassword("");
      onChanged();
      toast.success("Ayarlar kaydedildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Ayarlar kaydedilemedi.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleSync = async () => {
    setIsSyncing(true);
    try {
      await queueIntegrationSync(detail.id);
      toast.success("Senkronizasyon kuyruğa alındı.");
      onChanged();
      const [detailResponse, logs] = await Promise.all([getIntegration(detail.id), listIntegrationSyncLogs(detail.id)]);
      setDetail(detailResponse);
      setSyncLogs(logs);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Senkronizasyon başlatılamadı.");
    } finally {
      setIsSyncing(false);
    }
  };

  const handleDelete = async () => {
    const shouldDelete = await confirm({
      title: "Entegrasyonu sil",
      message: "Bu entegrasyonu kaldırmak istediğinize emin misiniz?",
      confirmLabel: "Sil",
      variant: "danger",
    });
    if (!shouldDelete) return;
    setIsDeleting(true);
    try {
      await deleteIntegration(detail.id);
      toast.success("Entegrasyon kaldırıldı.");
      onChanged();
      onClose();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Entegrasyon silinemedi.");
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[95] flex items-center justify-center bg-slate-900/40 p-4">
      <div className="flex max-h-[90vh] w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-line bg-white shadow-panel">
        <div className="flex shrink-0 items-center justify-between border-b border-line px-5 py-4">
          <div>
            <h3 className="flex items-center gap-2 text-lg font-bold text-ink">
              <Plug className="h-5 w-5 text-violet-brand" />
              {isEditing ? "Entegrasyonu Düzenle" : integration.displayName}
            </h3>
            <p className="mt-1 text-sm text-slate-500">{resolveProviderDisplayName(detail.provider, providers)}</p>
          </div>
          <button type="button" className="rounded-lg p-2 text-slate-400 hover:bg-slate-100 hover:text-slate-600" onClick={onClose}>
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-4">
          {isEditing ? (
            <div className="space-y-4">
              {!isIMAP ? (
                <label className="block space-y-2 text-sm">
                  <span className="font-semibold text-slate-600">Görünen ad</span>
                  <input
                    className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                  />
                </label>
              ) : null}

              <label className="flex items-center gap-2 text-sm font-semibold text-slate-600">
                <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
                Entegrasyon aktif
              </label>

              {isGmail ? (
                <GmailSettingsFields state={gmailForm} onChange={(patch) => setGmailForm((current) => ({ ...current, ...patch }))} />
              ) : null}
              {isCloudDrive ? (
                <GoogleDriveSettingsFields
                  state={googleDriveForm}
                  onChange={(patch) => setGoogleDriveForm((current) => ({ ...current, ...patch }))}
                  providerLabel={isOneDrive ? "OneDrive" : "Google Drive"}
                  folderPlaceholder={isOneDrive ? "Boş bırakılırsa kök klasör" : undefined}
                  folderHelpText={
                    isOneDrive
                      ? "Belirli bir klasörden başlamak için OneDrive klasör ID'sini girin."
                      : undefined
                  }
                />
              ) : null}
              {isOutlook ? (
                <OutlookSettingsFields state={outlookForm} onChange={(patch) => setOutlookForm((current) => ({ ...current, ...patch }))} />
              ) : null}
              {isIMAP ? (
                <label className="block space-y-2 text-sm">
                  <span className="font-semibold text-slate-600">Hangi tarihten itibaren</span>
                  <input
                    type="date"
                    className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand"
                    value={syncFromDate}
                    max={new Date().toISOString().slice(0, 10)}
                    onChange={(e) => setSyncFromDate(e.target.value)}
                  />
                </label>
              ) : null}
              {isIMAP ? (
                <label className="block space-y-2 text-sm">
                  <span className="font-semibold text-slate-600">Senkron basina max e-posta</span>
                  <input
                    type="number"
                    min={1}
                    max={2000}
                    className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand"
                    value={maxEmailsPerSync}
                    onChange={(e) => setMaxEmailsPerSync(e.target.value)}
                  />
                  <p className="text-xs text-slate-500">Tarih aralığındaki tüm mailler için 2000&apos;ye kadar artırabilirsiniz. Daha fazlası için tekrar senkronize edin.</p>
                </label>
              ) : null}
              {isIMAP ? (
                <label className="block space-y-2 text-sm">
                  <span className="font-semibold text-slate-600">Şifre (değiştirmek için doldurun)</span>
                  <input
                    type="password"
                    className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="new-password"
                  />
                </label>
              ) : null}
            </div>
          ) : (
            <>
              <dl className="grid gap-3 text-sm sm:grid-cols-2">
                <DetailItem label="Durum" value={<IntegrationStatusBadge status={detail.status} />} />
                <DetailItem label="E-posta" value={detail.emailAddress || "-"} />
                <DetailItem label="Son senkron" value={detail.lastSyncAt ? formatDateTime(detail.lastSyncAt) : "-"} />
                <DetailItem label="İçe aktarılan" value={String(detail.importedItemCount ?? 0)} />
                <DetailItem label="Aktif" value={detail.isActive ? "Evet" : "Hayır"} />
                {isIMAP ? (
                  <>
                    <DetailItem label="IMAP host" value={String(detail.settings?.host ?? "-")} />
                    <DetailItem label="Port" value={String(detail.settings?.port ?? "-")} />
                    <DetailItem label="Kullanıcı adı" value={String(detail.settings?.username ?? "-")} />
                    <DetailItem
                      label="Bağlantı"
                      value={
                        detail.settings?.useTLS === false
                          ? "Şifresiz"
                          : Number(detail.settings?.port) === 143
                            ? "STARTTLS"
                            : "SSL / TLS"
                      }
                    />
                    <DetailItem label="Başlangıç tarihi" value={String(detail.settings?.syncFromDate ?? "-")} />
                  </>
                ) : null}
                {isOutlook ? (
                  <>
                    <DetailItem label="Klasör" value={String(detail.settings?.folder ?? "-")} />
                    <DetailItem label="Başlangıç tarihi" value={String(detail.settings?.syncFromDate ?? "-")} />
                  </>
                ) : null}
                {isGmail ? (
                  <>
                    <DetailItem label="Gmail query" value={String(detail.settings?.query ?? "-")} />
                    <DetailItem label="Senkron başlangıcı" value={String(detail.settings?.syncFromDate ?? "-")} />
                  </>
                ) : null}
                {isCloudDrive ? (
                  <>
                    <DetailItem label="Klasör ID" value={detail.settings?.folderId == null || detail.settings?.folderId === "" ? "Kök klasör" : String(detail.settings.folderId)} />
                    <DetailItem label="Senkron başlangıcı" value={String(detail.settings?.syncFromDate ?? "-")} />
                    <DetailItem label="Alt klasörler" value={detail.settings?.includeSubfolders === false ? "Hayır" : "Evet"} />
                    <DetailItem label="Max dosya / senkron" value={String(detail.settings?.maxFilesPerSync ?? "-")} />
                  </>
                ) : null}
              </dl>

              {detail.lastError ? (
                <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{detail.lastError}</div>
              ) : null}

              {latestLog ? (
                <div className="rounded-lg border border-line bg-slate-50 px-4 py-3 text-sm text-slate-700">
                  Son sync: {syncLogStatusLabel(latestLog.status)} · İçe aktarılan {latestLog.importedItems} · Atlanan {latestLog.skippedItems} · Başarısız {latestLog.failedItems}
                  {totalImportedFromLogs > 0 ? <p className="mt-1 text-xs text-slate-500">Toplam log içe aktarımı: {totalImportedFromLogs}</p> : null}
                </div>
              ) : null}
            </>
          )}

          <div className="flex flex-wrap gap-2">
            {isEditing ? (
              <>
                <PrimaryButton className="h-10" disabled={isSaving || !canSave} onClick={() => void handleSave()}>
                  {isSaving ? <Spinner size="sm" className="text-white" /> : "Kaydet"}
                </PrimaryButton>
                <button type="button" className="h-10 rounded-lg border border-line px-4 text-sm font-semibold text-slate-600" onClick={cancelEditing}>
                  İptal
                </button>
              </>
            ) : (
              <>
                <button
                  type="button"
                  className="inline-flex h-10 items-center gap-2 rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-700 hover:bg-slate-50"
                  onClick={startEditing}
                >
                  <Pencil className="h-4 w-4" />
                  Ayarları Düzenle
                </button>
                <PrimaryButton className="h-10" disabled={isSyncing || !detail.isActive} onClick={() => void handleSync()}>
                  {isSyncing ? <Spinner size="sm" className="text-white" /> : <RefreshCw className="h-4 w-4" />}
                  Şimdi Senkronize Et
                </PrimaryButton>
                <button type="button" className="inline-flex h-10 items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 text-sm font-semibold text-red-700 disabled:opacity-60" disabled={isDeleting} onClick={() => void handleDelete()}>
                  <Trash2 className="h-4 w-4" />
                  Kaldır
                </button>
              </>
            )}
          </div>
          {!isEditing ? (
            <p className="text-xs text-slate-500">
              {isCloudDrive ? "Dosyalar bağlantı sonrası ve periyodik olarak otomatik senkronize edilir." : "E-postalar bağlantı sonrası ve periyodik olarak otomatik senkronize edilir."}
            </p>
          ) : null}

          {!isEditing ? (
            <section>
              <h4 className="text-sm font-bold text-ink">Senkronizasyon kayıtları</h4>
              {isLoadingLogs ? (
                <div className="mt-3"><LoadingState label="Kayıtlar yükleniyor..." /></div>
              ) : syncLogs.length === 0 ? (
                <div className="mt-3"><EmptyState title="Kayıt yok" description="İlk otomatik senkronizasyon yakında başlayacak." /></div>
              ) : (
                <div className="mt-3 divide-y divide-line overflow-hidden rounded-lg border border-line">
                  {syncLogs.map((log) => (
                    <div key={log.id} className={`space-y-1 px-4 py-3 text-sm ${log.status === "failed" ? "bg-red-50/60" : ""}`}>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`font-semibold ${log.status === "failed" ? "text-red-700" : "text-ink"}`}>{syncLogStatusLabel(log.status)}</span>
                        <span className="text-xs text-slate-500">{log.syncType}</span>
                      </div>
                      <p className="text-slate-600">
                        Toplam: {log.totalItems} · İçe aktarılan: {log.importedItems} · Atlanan: {log.skippedItems} · Başarısız: {log.failedItems}
                      </p>
                      <p className="text-xs text-slate-400">{formatDateTime(log.createdAt)}</p>
                      {log.errorMessage ? <p className="rounded-md border border-red-200 bg-white px-3 py-2 text-xs text-red-700">{log.errorMessage}</p> : null}
                    </div>
                  ))}
                </div>
              )}
            </section>
          ) : null}
        </div>
      </div>
    </div>
  );
}

function DetailItem({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</dt>
      <dd className="mt-1 font-medium text-slate-700">{value}</dd>
    </div>
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("tr-TR", { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}
