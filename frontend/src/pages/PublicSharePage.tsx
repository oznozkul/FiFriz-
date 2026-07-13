import { Download, Lock } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { EmptyState, FileTypeIcon, LoadingState, PrimaryButton, Spinner } from "../components/common";
import {
  formatShareFileSize,
  getPublicShare,
  getPublicShareDownloadUrl,
  verifySharePassword,
  type PublicShareInfo,
} from "../services/shareService";
import { toast } from "../services/toast";

export function PublicSharePage() {
  const { token = "" } = useParams();
  const [info, setInfo] = useState<PublicShareInfo | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(() => sessionStorage.getItem(shareSessionKey(token)));
  const [password, setPassword] = useState("");
  const [visitorEmail, setVisitorEmail] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isVerifying, setIsVerifying] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const previewUrl = useMemo(() => {
    if (!info?.canAccess || !token) return "";
    return getPublicShareDownloadUrl(token, accessToken ?? undefined, false, visitorEmail || undefined);
  }, [accessToken, info?.canAccess, token, visitorEmail]);

  const loadShare = async (nextAccessToken?: string | null) => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      const data = await getPublicShare(token, nextAccessToken ?? accessToken ?? undefined, visitorEmail || undefined);
      setInfo(data);
      if (data.expired) {
        setErrorMessage("Bu paylaşım linkinin süresi dolmuş.");
      } else if (data.maxViewsReached) {
        setErrorMessage("Bu paylaşım linkinin görüntüleme limiti doldu.");
      }
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Paylaşım linki açılamadı.");
      setInfo(null);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!token) return;
    void loadShare();
  }, [token]);

  const handleVerifyPassword = async () => {
    if (!password.trim()) return;
    setIsVerifying(true);
    try {
      const result = await verifySharePassword(token, password, visitorEmail || undefined);
      setAccessToken(result.accessToken);
      sessionStorage.setItem(shareSessionKey(token), result.accessToken);
      setPassword("");
      await loadShare(result.accessToken);
      toast.success("Şifre doğrulandı.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Şifre hatali.");
    } finally {
      setIsVerifying(false);
    }
  };

  const handleDownload = () => {
    if (!token) return;
    window.open(getPublicShareDownloadUrl(token, accessToken ?? undefined, true, visitorEmail || undefined), "_blank", "noopener,noreferrer");
  };

  if (isLoading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
        <LoadingState label="Paylaşım linki yükleniyor..." />
      </main>
    );
  }

  if (errorMessage && !info) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
        <div className="w-full max-w-lg">
          <EmptyState title="Paylaşım linki kullanılamıyor" description={errorMessage} />
        </div>
      </main>
    );
  }

  if (!info) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
        <EmptyState title="Paylaşım linki bulunamadı" description="Link geçersiz veya kaldırılmış olabilir." />
      </main>
    );
  }

  const isPreviewable = ["pdf", "png", "jpg", "jpeg", "gif", "webp"].includes(info.fileExtension.toLowerCase());

  return (
    <main className="min-h-screen bg-slate-50 px-4 py-10">
      <div className="mx-auto max-w-4xl space-y-5">
        <section className="panel p-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex min-w-0 items-center gap-4">
              <FileTypeIcon extension={info.fileExtension} size="xl" />
              <div className="min-w-0">
                <h1 className="truncate text-xl font-bold text-ink">{info.documentName}</h1>
                <p className="mt-1 text-sm text-slate-500">
                  {info.fileExtension.toUpperCase()} · {formatShareFileSize(info.fileSize)}
                </p>
              </div>
            </div>
            {info.canAccess && info.allowDownload ? (
              <PrimaryButton onClick={handleDownload}>
                <Download className="h-4 w-4" />
                İndir
              </PrimaryButton>
            ) : null}
          </div>

          {errorMessage ? (
            <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-semibold text-amber-900">{errorMessage}</div>
          ) : null}

          {info.domainRestricted && !info.canAccess ? (
            <div className="mt-6 rounded-lg border border-line bg-slate-50 p-4">
              <p className="mb-3 text-sm font-semibold text-slate-600">Bu paylaşım yalnızca belirli e-posta domainlerine açıktır.</p>
              <input
                className="h-11 w-full rounded-lg border border-line bg-white px-3 text-sm outline-none focus:border-violet-brand"
                type="email"
                placeholder="E-posta adresiniz"
                value={visitorEmail}
                onChange={(event) => setVisitorEmail(event.target.value)}
              />
              <PrimaryButton className="mt-3 h-11" disabled={!visitorEmail.trim()} onClick={() => void loadShare()}>
                Erişimi Doğrula
              </PrimaryButton>
            </div>
          ) : null}

          {info.passwordRequired && !info.canAccess && (!info.domainRestricted || visitorEmail.trim()) ? (
            <div className="mt-6 rounded-lg border border-line bg-slate-50 p-4">
              <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-600">
                <Lock className="h-4 w-4" />
                Bu link şifre ile korunuyor
              </div>
              <div className="flex flex-col gap-3 sm:flex-row">
                <input
                  className="h-11 min-w-0 flex-1 rounded-lg border border-line bg-white px-3 text-sm outline-none focus:border-violet-brand"
                  type="password"
                  placeholder="Şifreyi girin"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") void handleVerifyPassword();
                  }}
                />
                <PrimaryButton className="h-11" disabled={isVerifying || !password.trim()} onClick={() => void handleVerifyPassword()}>
                  {isVerifying ? <Spinner size="sm" label="Doğrulanıyor..." className="text-white" /> : "Devam Et"}
                </PrimaryButton>
              </div>
            </div>
          ) : null}
        </section>

        {info.canAccess && previewUrl ? (
          <section className="panel overflow-hidden">
            <div className="border-b border-line px-5 py-4">
              <h2 className="font-bold text-ink">Onizleme</h2>
            </div>
            <div className="bg-slate-100 p-4">
              {isPreviewable ? (
                info.fileExtension.toLowerCase() === "pdf" ? (
                  <iframe className="h-[70vh] w-full rounded-lg border border-line bg-white" src={previewUrl} title={info.documentName} />
                ) : (
                  <img className="mx-auto max-h-[70vh] max-w-full rounded-lg border border-line bg-white object-contain" src={previewUrl} alt={info.documentName} />
                )
              ) : (
                <div className="rounded-lg border border-line bg-white p-8 text-center text-sm text-slate-500">
                  Bu dosya türü tarayıcıda önizlenemiyor. {info.allowDownload ? "İndir butonunu kullanabilirsiniz." : "Dosyaya erişim sağlandı."}
                </div>
              )}
            </div>
          </section>
        ) : null}
      </div>
    </main>
  );
}

function shareSessionKey(token: string) {
  return `fifriz_share_access_${token}`;
}
