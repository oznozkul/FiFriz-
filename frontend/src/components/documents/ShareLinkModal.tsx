import { BarChart3, Link2, Shield, SlidersHorizontal, Trash2, UserPlus, Users, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { IconButton, PrimaryButton, Spinner, useConfirm } from "../common";
import {
  createShareLink,
  deleteShareLink,
  expirationPresetLabels,
  getShareLinkAnalytics,
  listShareLinks,
  shareStatusLabels,
  type CreateShareLinkInput,
  type DocumentShareLink,
  type ExpirationPreset,
  type ShareLinkAnalytics,
  type SharePermission,
} from "../../services/shareService";
import {
  createInternalShare,
  deleteInternalShare,
  internalSharePermissionDescriptions,
  internalSharePermissionLabels,
  listInternalShares,
  updateInternalShare,
  type DocumentInternalShare,
  type InternalSharePermission,
} from "../../services/internalShareService";
import { listTenantMembers, type TenantMember } from "../../services/tenantService";
import { useAuth } from "../../services/AuthContext";
import { toast } from "../../services/toast";
import { ShareChannelButtons } from "./ShareChannelButtons";

type ShareLinkModalProps = {
  documentId: string;
  documentName: string;
  isOpen: boolean;
  onClose: () => void;
};

type ShareTab = "public" | "people";
type PublicConfigTab = "general" | "security" | "permissions" | "analytics";

export function ShareLinkModal({ documentId, documentName, isOpen, onClose }: ShareLinkModalProps) {
  const { user } = useAuth();
  const confirm = useConfirm();
  const [activeTab, setActiveTab] = useState<ShareTab>("public");
  const [configTab, setConfigTab] = useState<PublicConfigTab>("general");

  const [links, setLinks] = useState<DocumentShareLink[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [createdLink, setCreatedLink] = useState<DocumentShareLink | null>(null);
  const [selectedLink, setSelectedLink] = useState<DocumentShareLink | null>(null);
  const [analytics, setAnalytics] = useState<ShareLinkAnalytics | null>(null);
  const [isAnalyticsLoading, setIsAnalyticsLoading] = useState(false);

  const [permission, setPermission] = useState<SharePermission>("view");
  const [expirationPreset, setExpirationPreset] = useState<ExpirationPreset>("never");
  const [allowDownload, setAllowDownload] = useState(true);
  const [allowPrint, setAllowPrint] = useState(false);
  const [allowCopyText, setAllowCopyText] = useState(false);
  const [expiresAt, setExpiresAt] = useState("");
  const [maxViewCount, setMaxViewCount] = useState("");
  const [password, setPassword] = useState("");
  const [allowedDomains, setAllowedDomains] = useState("");
  const [watermarkEnabled, setWatermarkEnabled] = useState(false);
  const [invitedEmail, setInvitedEmail] = useState("");

  const [members, setMembers] = useState<TenantMember[]>([]);
  const [internalShares, setInternalShares] = useState<DocumentInternalShare[]>([]);
  const [isPeopleLoading, setIsPeopleLoading] = useState(false);
  const [isPeopleSaving, setIsPeopleSaving] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState("");
  const [internalPermission, setInternalPermission] = useState<InternalSharePermission>("view");

  const availableMembers = useMemo(
    () => members.filter((member) => member.userId !== user?.id),
    [members, user?.id],
  );

  const fetchLinks = async () => {
    setIsLoading(true);
    try {
      setLinks(await listShareLinks(documentId));
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Paylaşım linkleri yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  const fetchPeopleData = async () => {
    setIsPeopleLoading(true);
    try {
      const [tenantMembers, shares] = await Promise.all([listTenantMembers(), listInternalShares(documentId)]);
      setMembers(tenantMembers);
      setInternalShares(shares);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Kişi paylaşımları yüklenemedi.");
    } finally {
      setIsPeopleLoading(false);
    }
  };

  const loadAnalytics = async (link: DocumentShareLink) => {
    setIsAnalyticsLoading(true);
    try {
      setAnalytics(await getShareLinkAnalytics(link.id));
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Analitik yüklenemedi.");
      setAnalytics(null);
    } finally {
      setIsAnalyticsLoading(false);
    }
  };

  useEffect(() => {
    if (!isOpen) return;
    setActiveTab("public");
    setConfigTab("general");
    setCreatedLink(null);
    setSelectedLink(null);
    setAnalytics(null);
    setPermission("view");
    setExpirationPreset("never");
    setAllowDownload(true);
    setAllowPrint(false);
    setAllowCopyText(false);
    setExpiresAt("");
    setMaxViewCount("");
    setPassword("");
    setAllowedDomains("");
    setWatermarkEnabled(false);
    setInvitedEmail("");
    setSelectedUserId("");
    setInternalPermission("view");
    void fetchLinks();
    void fetchPeopleData();
  }, [isOpen, documentId]);

  useEffect(() => {
    if (configTab === "analytics" && selectedLink) {
      void loadAnalytics(selectedLink);
    }
  }, [configTab, selectedLink?.id]);

  if (!isOpen) return null;

  const handleCreate = async () => {
    setIsSaving(true);
    try {
      const domains = allowedDomains
        .split(/[,\n]/)
        .map((item) => item.trim())
        .filter(Boolean);

      const input: CreateShareLinkInput = {
        shareType: invitedEmail.trim() ? "email_invite" : "public_link",
        permission,
        expirationPreset,
        allowDownload: permission === "download" ? true : allowDownload,
        allowPrint,
        allowCopyText,
        watermarkEnabled,
      };
      if (password.trim()) input.password = password.trim();
      if (expirationPreset === "custom" && expiresAt) input.expiresAt = new Date(`${expiresAt}T23:59:59`).toISOString();
      if (maxViewCount.trim()) input.maxViewCount = Number(maxViewCount);
      if (domains.length > 0) input.allowedDomains = domains;
      if (invitedEmail.trim()) input.invitedEmail = invitedEmail.trim();

      const link = await createShareLink(documentId, input);
      setCreatedLink(link);
      setPassword("");
      await fetchLinks();
      toast.success("Paylaşım linki oluşturuldu.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Paylaşım linki oluşturulamadı.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (link: DocumentShareLink) => {
    const shouldDelete = await confirm({
      title: "Paylaşımı kaldır",
      message: "Bu paylaşım linki pasifleştirilsin mi?",
      confirmLabel: "Kaldır",
      variant: "danger",
    });
    if (!shouldDelete) return;
    try {
      await deleteShareLink(documentId, link.id);
      setLinks((current) => current.filter((item) => item.id !== link.id));
      if (createdLink?.id === link.id) setCreatedLink(null);
      if (selectedLink?.id === link.id) {
        setSelectedLink(null);
        setAnalytics(null);
        setConfigTab("general");
      }
      toast.success("Paylaşım linki silindi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Paylaşım linki silinemedi.");
    }
  };

  const handleInternalShare = async () => {
    if (!selectedUserId) {
      toast.error("Lütfen bir kullanıcı seçin.");
      return;
    }
    setIsPeopleSaving(true);
    try {
      await createInternalShare(documentId, {
        sharedWithUserId: selectedUserId,
        permission: internalPermission,
      });
      setSelectedUserId("");
      setInternalPermission("view");
      await fetchPeopleData();
      toast.success("Belge kullanıcıyla paylaşıldı.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Belge paylaşılamadı.");
    } finally {
      setIsPeopleSaving(false);
    }
  };

  const handleUpdateInternalShare = async (share: DocumentInternalShare, nextPermission: InternalSharePermission) => {
    try {
      const updated = await updateInternalShare(documentId, share.id, nextPermission);
      setInternalShares((current) => current.map((item) => (item.id === share.id ? updated : item)));
      toast.success("Paylaşım yetkisi güncellendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Paylaşım yetkisi güncellenemedi.");
    }
  };

  const handleDeleteInternalShare = async (share: DocumentInternalShare) => {
    const shouldDelete = await confirm({
      title: "Paylaşımı kaldır",
      message: `${share.sharedWithFullName} için paylaşım kaldırılsın mı?`,
      confirmLabel: "Kaldır",
      variant: "danger",
    });
    if (!shouldDelete) return;
    try {
      await deleteInternalShare(documentId, share.id);
      setInternalShares((current) => current.filter((item) => item.id !== share.id));
      toast.success("Paylaşım kaldırıldi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Paylaşım kaldırılamadı.");
    }
  };

  const configTabs: { id: PublicConfigTab; label: string; icon: typeof Link2 }[] = [
    { id: "general", label: "Genel", icon: Link2 },
    { id: "security", label: "Güvenlik", icon: Shield },
    { id: "permissions", label: "İzinler", icon: SlidersHorizontal },
    { id: "analytics", label: "Analitik", icon: BarChart3 },
  ];

  return (
    <div className="fixed inset-0 z-[90] flex items-center justify-center bg-slate-900/40 p-4">
      <div className="w-full max-w-3xl rounded-xl border border-line bg-white shadow-panel">
        <div className="flex items-center justify-between border-b border-line px-5 py-4">
          <div>
            <h2 className="text-lg font-bold text-ink">Belgeyi Paylaş</h2>
            <p className="mt-1 truncate text-sm text-slate-500">{documentName}</p>
          </div>
          <IconButton className="h-8 w-8" aria-label="Kapat" onClick={onClose}>
            <X className="h-4 w-4" />
          </IconButton>
        </div>

        <div className="border-b border-line px-5">
          <div className="flex gap-1">
            <button
              type="button"
              className={`border-b-2 px-4 py-3 text-sm font-semibold transition ${
                activeTab === "public"
                  ? "border-violet-brand text-violet-brand"
                  : "border-transparent text-slate-500 hover:text-slate-700"
              }`}
              onClick={() => setActiveTab("public")}
            >
              <span className="inline-flex items-center gap-2">
                <Link2 className="h-4 w-4" />
                Herkese Açık Link
              </span>
            </button>
            <button
              type="button"
              className={`border-b-2 px-4 py-3 text-sm font-semibold transition ${
                activeTab === "people"
                  ? "border-violet-brand text-violet-brand"
                  : "border-transparent text-slate-500 hover:text-slate-700"
              }`}
              onClick={() => setActiveTab("people")}
            >
              <span className="inline-flex items-center gap-2">
                <Users className="h-4 w-4" />
                Kişiler
              </span>
            </button>
          </div>
        </div>

        <div className="max-h-[70vh] space-y-5 overflow-y-auto p-5">
          {activeTab === "public" ? (
            <>
              <div className="flex flex-wrap gap-1 rounded-lg border border-line bg-slate-50 p-1">
                {configTabs.map((tab) => {
                  const Icon = tab.icon;
                  const disabled = tab.id === "analytics" && !selectedLink;
                  return (
                    <button
                      key={tab.id}
                      type="button"
                      disabled={disabled}
                      className={`inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm font-semibold transition ${
                        configTab === tab.id
                          ? "bg-white text-violet-brand shadow-sm"
                          : disabled
                            ? "cursor-not-allowed text-slate-300"
                            : "text-slate-500 hover:text-slate-700"
                      }`}
                      onClick={() => setConfigTab(tab.id)}
                    >
                      <Icon className="h-4 w-4" />
                      {tab.label}
                    </button>
                  );
                })}
              </div>

              {configTab === "general" ? (
                <section className="space-y-4 rounded-lg border border-line bg-slate-50/70 p-4">
                  <h3 className="text-sm font-bold text-ink">Yeni link oluştur</h3>
                  <div className="grid gap-4 sm:grid-cols-2">
                    <label className="space-y-2 text-sm sm:col-span-2">
                      <span className="font-semibold text-slate-600">Yetki</span>
                      <select
                        className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-700 outline-none focus:border-violet-brand"
                        value={permission}
                        onChange={(event) => {
                          const next = event.target.value as SharePermission;
                          setPermission(next);
                          if (next === "download") setAllowDownload(true);
                        }}
                      >
                        <option value="view">Sadece görüntüle</option>
                        <option value="download">İndirebilir</option>
                      </select>
                    </label>
                    <label className="space-y-2 text-sm">
                      <span className="font-semibold text-slate-600">Süre</span>
                      <select
                        className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-700 outline-none focus:border-violet-brand"
                        value={expirationPreset}
                        onChange={(event) => setExpirationPreset(event.target.value as ExpirationPreset)}
                      >
                        {(Object.keys(expirationPresetLabels) as ExpirationPreset[]).map((key) => (
                          <option key={key} value={key}>
                            {expirationPresetLabels[key]}
                          </option>
                        ))}
                      </select>
                    </label>
                    {expirationPreset === "custom" ? (
                      <label className="space-y-2 text-sm">
                        <span className="font-semibold text-slate-600">Özel tarih</span>
                        <input
                          className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
                          type="date"
                          value={expiresAt}
                          onChange={(event) => setExpiresAt(event.target.value)}
                        />
                      </label>
                    ) : (
                      <div />
                    )}
                    <label className="space-y-2 text-sm sm:col-span-2">
                      <span className="font-semibold text-slate-600">E-posta daveti (opsiyonel)</span>
                      <input
                        className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
                        type="email"
                        placeholder="ornek@company.com"
                        value={invitedEmail}
                        onChange={(event) => setInvitedEmail(event.target.value)}
                      />
                    </label>
                  </div>
                  <PrimaryButton disabled={isSaving} onClick={() => void handleCreate()}>
                    {isSaving ? <Spinner size="sm" label="Oluşturuluyor..." className="text-white" /> : "Link Oluştur"}
                  </PrimaryButton>
                  {createdLink ? (
                    <div className="space-y-4 rounded-lg border border-violet-ring bg-violet-soft/50 p-4">
                      <div>
                        <p className="text-xs font-bold uppercase tracking-wide text-violet-brand">Oluşturulan link</p>
                        <p className="mt-2 break-all text-sm font-semibold text-ink">{createdLink.shareUrl}</p>
                      </div>
                      <ShareChannelButtons documentName={documentName} shareLink={createdLink} />
                    </div>
                  ) : null}
                </section>
              ) : null}

              {configTab === "security" ? (
                <section className="space-y-4 rounded-lg border border-line bg-slate-50/70 p-4">
                  <h3 className="text-sm font-bold text-ink">Güvenlik ayarları</h3>
                  <div className="grid gap-4">
                    <label className="space-y-2 text-sm">
                      <span className="font-semibold text-slate-600">Şifre (opsiyonel)</span>
                      <input
                        className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
                        type="password"
                        placeholder="Link şifresi"
                        value={password}
                        onChange={(event) => setPassword(event.target.value)}
                      />
                    </label>
                    <label className="space-y-2 text-sm">
                      <span className="font-semibold text-slate-600">Maks. erişim sayısı</span>
                      <input
                        className="h-10 w-full rounded-lg border border-line bg-white px-3 outline-none focus:border-violet-brand"
                        type="number"
                        min={1}
                        placeholder="Sınırsız"
                        value={maxViewCount}
                        onChange={(event) => setMaxViewCount(event.target.value)}
                      />
                    </label>
                    <label className="space-y-2 text-sm">
                      <span className="font-semibold text-slate-600">İzinli e-posta domainleri</span>
                      <textarea
                        className="min-h-24 w-full rounded-lg border border-line bg-white px-3 py-2 text-sm outline-none focus:border-violet-brand"
                        placeholder="company.com, company.org"
                        value={allowedDomains}
                        onChange={(event) => setAllowedDomains(event.target.value)}
                      />
                      <p className="text-xs text-slate-500">Virgül veya satır ile ayırın. Boş bırakılırsa domain kısıtı uygulanmaz.</p>
                    </label>
                    <label className="flex items-center gap-3 text-sm">
                      <input
                        className="h-4 w-4 accent-violet-brand"
                        type="checkbox"
                        checked={watermarkEnabled}
                        onChange={(event) => setWatermarkEnabled(event.target.checked)}
                      />
                      <span className="font-semibold text-slate-600">Önizlemede filigran (altyapı hazır)</span>
                    </label>
                  </div>
                  <PrimaryButton disabled={isSaving} onClick={() => void handleCreate()}>
                    {isSaving ? <Spinner size="sm" label="Oluşturuluyor..." className="text-white" /> : "Link Oluştur"}
                  </PrimaryButton>
                </section>
              ) : null}

              {configTab === "permissions" ? (
                <section className="space-y-4 rounded-lg border border-line bg-slate-50/70 p-4">
                  <h3 className="text-sm font-bold text-ink">İzin politikası</h3>
                  <div className="space-y-3">
                    <label className="flex items-center gap-3 text-sm">
                      <input
                        className="h-4 w-4 accent-violet-brand"
                        type="checkbox"
                        checked={allowDownload}
                        disabled={permission === "download"}
                        onChange={(event) => setAllowDownload(event.target.checked)}
                      />
                      <span className="font-semibold text-slate-600">İndirmeye izin ver</span>
                    </label>
                    <label className="flex items-center gap-3 text-sm">
                      <input
                        className="h-4 w-4 accent-violet-brand"
                        type="checkbox"
                        checked={allowPrint}
                        onChange={(event) => setAllowPrint(event.target.checked)}
                      />
                      <span className="font-semibold text-slate-600">Yazdırmaya izin ver</span>
                    </label>
                    <label className="flex items-center gap-3 text-sm">
                      <input
                        className="h-4 w-4 accent-violet-brand"
                        type="checkbox"
                        checked={allowCopyText}
                        onChange={(event) => setAllowCopyText(event.target.checked)}
                      />
                      <span className="font-semibold text-slate-600">Metin kopyalamaya izin ver</span>
                    </label>
                  </div>
                  <PrimaryButton disabled={isSaving} onClick={() => void handleCreate()}>
                    {isSaving ? <Spinner size="sm" label="Oluşturuluyor..." className="text-white" /> : "Link Oluştur"}
                  </PrimaryButton>
                </section>
              ) : null}

              {configTab === "analytics" && selectedLink ? (
                <section className="space-y-4 rounded-lg border border-line bg-slate-50/70 p-4">
                  <h3 className="text-sm font-bold text-ink">Paylaşım analitiği</h3>
                  {isAnalyticsLoading ? (
                    <Spinner size="sm" label="Analitik yükleniyor..." />
                  ) : analytics ? (
                    <div className="grid gap-3 sm:grid-cols-2">
                      <AnalyticsCard label="Durum" value={shareStatusLabels[analytics.status] ?? analytics.status} />
                      <AnalyticsCard label="Açılma" value={String(analytics.openCount)} />
                      <AnalyticsCard label="İndirme" value={String(analytics.downloadCount)} />
                      <AnalyticsCard label="Oluşturan" value={analytics.createdByName || analytics.createdByUserId} />
                      <AnalyticsCard label="Son açılma" value={formatDateTime(analytics.lastOpenedAt)} />
                      <AnalyticsCard label="Son indirme" value={formatDateTime(analytics.lastDownloadedAt)} />
                      <AnalyticsCard label="Son IP" value={analytics.lastIpAddress ?? "—"} />
                      <AnalyticsCard label="Son tarayıcı" value={truncate(analytics.lastUserAgent, 48)} />
                    </div>
                  ) : (
                    <p className="text-sm text-slate-500">Analitik verisi bulunamadı.</p>
                  )}
                </section>
              ) : null}

              <section className="space-y-3">
                <h3 className="text-sm font-bold text-ink">Aktif linkler</h3>
                {isLoading ? (
                  <Spinner size="sm" label="Linkler yükleniyor..." />
                ) : links.length === 0 ? (
                  <p className="text-sm text-slate-500">Henüz aktif paylaşım linki yok.</p>
                ) : (
                  <div className="space-y-3">
                    {links.map((link) => (
                      <div
                        key={link.id}
                        className={`space-y-4 rounded-lg border p-4 transition ${
                          selectedLink?.id === link.id ? "border-violet-brand bg-violet-soft/30" : "border-line"
                        }`}
                      >
                        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                          <button
                            type="button"
                            className="min-w-0 flex-1 text-left"
                            onClick={() => {
                              setSelectedLink(link);
                              setConfigTab("analytics");
                            }}
                          >
                            <div className="flex flex-wrap items-center gap-2 text-sm font-semibold text-ink">
                              <Link2 className="h-4 w-4 text-violet-brand" />
                              {link.permission === "download" ? "İndirilebilir" : "Görüntüleme"}
                              <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-600">
                                {shareStatusLabels[link.status] ?? link.status}
                              </span>
                            </div>
                            <p className="mt-2 break-all text-xs text-slate-500">{link.shareUrl}</p>
                            <p className="mt-2 text-xs text-slate-400">
                              {link.viewCount} açılma · {link.downloadCount} indirme
                              {link.maxViewCount ? ` / ${link.maxViewCount} limit` : ""}
                              {link.expiresAt ? ` · Son: ${formatDate(link.expiresAt)}` : ""}
                              {link.hasPassword ? " · Şifreli" : ""}
                              {link.allowedDomains?.length ? ` · ${link.allowedDomains.length} domain` : ""}
                            </p>
                          </button>
                          <button
                            className="inline-flex h-9 items-center gap-1 self-start rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-600 transition hover:bg-red-50"
                            type="button"
                            onClick={() => void handleDelete(link)}
                          >
                            <Trash2 className="h-4 w-4" />
                            Sil
                          </button>
                        </div>
                        <ShareChannelButtons documentName={documentName} shareLink={link} compact />
                      </div>
                    ))}
                  </div>
                )}
              </section>
            </>
          ) : (
            <>
              <section className="space-y-4 rounded-lg border border-line bg-slate-50/70 p-4">
                <h3 className="text-sm font-bold text-ink">Kullanıcıyla paylaş</h3>
                {isPeopleLoading ? (
                  <Spinner size="sm" label="Üyeler yükleniyor..." />
                ) : availableMembers.length === 0 ? (
                  <p className="text-sm text-slate-500">Paylaşılabilecek başka tenant üyesi bulunamadı.</p>
                ) : (
                  <div className="grid gap-4">
                    <label className="space-y-2 text-sm">
                      <span className="font-semibold text-slate-600">Kullanıcı</span>
                      <select
                        className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-700 outline-none focus:border-violet-brand"
                        value={selectedUserId}
                        onChange={(event) => setSelectedUserId(event.target.value)}
                      >
                        <option value="">Kullanıcı seçin</option>
                        {availableMembers.map((member) => (
                          <option key={member.userId} value={member.userId}>
                            {member.fullName} ({member.email})
                          </option>
                        ))}
                      </select>
                    </label>
                    <label className="space-y-2 text-sm">
                      <span className="font-semibold text-slate-600">Yetki</span>
                      <select
                        className="h-10 w-full rounded-lg border border-line bg-white px-3 font-semibold text-slate-700 outline-none focus:border-violet-brand"
                        value={internalPermission}
                        onChange={(event) => setInternalPermission(event.target.value as InternalSharePermission)}
                      >
                        {(Object.keys(internalSharePermissionLabels) as InternalSharePermission[]).map((key) => (
                          <option key={key} value={key}>
                            {internalSharePermissionLabels[key]}
                          </option>
                        ))}
                      </select>
                      <p className="text-xs text-slate-500">{internalSharePermissionDescriptions[internalPermission]}</p>
                    </label>
                    <PrimaryButton disabled={isPeopleSaving || !selectedUserId} onClick={() => void handleInternalShare()}>
                      {isPeopleSaving ? (
                        <Spinner size="sm" label="Paylaşılıyor..." className="text-white" />
                      ) : (
                        <span className="inline-flex items-center gap-2">
                          <UserPlus className="h-4 w-4" />
                          Paylaş
                        </span>
                      )}
                    </PrimaryButton>
                  </div>
                )}
              </section>
              <section className="space-y-3">
                <h3 className="text-sm font-bold text-ink">Mevcut paylaşımlar</h3>
                {isPeopleLoading ? (
                  <Spinner size="sm" label="Paylaşımlar yükleniyor..." />
                ) : internalShares.length === 0 ? (
                  <p className="text-sm text-slate-500">Bu belge henüz kimseyle paylaşılmamış.</p>
                ) : (
                  <div className="space-y-3">
                    {internalShares.map((share) => (
                      <div key={share.id} className="rounded-lg border border-line p-4">
                        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                          <div className="min-w-0 flex-1">
                            <p className="text-sm font-semibold text-ink">{share.sharedWithFullName}</p>
                            <p className="text-xs text-slate-500">{share.sharedWithEmail}</p>
                            <p className="mt-2 text-xs text-slate-400">
                              Paylaşan: {share.sharedByFullName} · {formatDate(share.createdAt)}
                            </p>
                          </div>
                          <div className="flex flex-col gap-2 sm:items-end">
                            <select
                              className="h-9 rounded-lg border border-line bg-white px-3 text-sm font-semibold text-slate-700 outline-none focus:border-violet-brand"
                              value={share.permission}
                              onChange={(event) =>
                                void handleUpdateInternalShare(share, event.target.value as InternalSharePermission)
                              }
                            >
                              {(Object.keys(internalSharePermissionLabels) as InternalSharePermission[]).map((key) => (
                                <option key={key} value={key}>
                                  {internalSharePermissionLabels[key]}
                                </option>
                              ))}
                            </select>
                            <button
                              className="inline-flex h-9 items-center gap-1 rounded-lg border border-red-200 px-3 text-sm font-semibold text-red-600 transition hover:bg-red-50"
                              type="button"
                              onClick={() => void handleDeleteInternalShare(share)}
                            >
                              <Trash2 className="h-4 w-4" />
                              Kaldır
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </section>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function AnalyticsCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-line bg-white px-3 py-2">
      <p className="text-xs font-semibold uppercase tracking-wide text-slate-400">{label}</p>
      <p className="mt-1 text-sm font-semibold text-ink">{value}</p>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

function formatDateTime(value: string | null) {
  if (!value) return "—";
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function truncate(value: string | null, max: number) {
  if (!value) return "—";
  return value.length > max ? `${value.slice(0, max)}…` : value;
}
