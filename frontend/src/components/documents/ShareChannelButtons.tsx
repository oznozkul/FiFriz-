import { Copy, Mail, Send, Share2 } from "lucide-react";
import { useMemo } from "react";
import type { DocumentShareLink } from "../../services/shareService";
import { logShareChannel } from "../../services/shareService";
import { toast } from "../../services/toast";
import {
  buildEmailShareUrl,
  buildTelegramShareUrl,
  buildWhatsAppShareUrl,
  canUseNativeShare,
  isMobileBrowser,
  openShareWindow,
  shareNative,
  type ShareChannel,
} from "../../utils/shareChannels";

type ShareChannelButtonsProps = {
  documentName: string;
  shareLink: DocumentShareLink;
  compact?: boolean;
};

export function ShareChannelButtons({ documentName, shareLink, compact = false }: ShareChannelButtonsProps) {
  const showNative = useMemo(() => canUseNativeShare(), []);
  const showDesktopChannels = !isMobileBrowser() || !showNative;

  const content = {
    documentName,
    shareUrl: shareLink.shareUrl,
    expiresAt: shareLink.expiresAt,
  };

  const logChannel = (channel: ShareChannel) => {
    void logShareChannel(shareLink.id, channel);
  };

  const handleWhatsApp = () => {
    openShareWindow(buildWhatsAppShareUrl(content));
    logChannel("whatsapp");
  };

  const handleTelegram = () => {
    openShareWindow(buildTelegramShareUrl(content));
    logChannel("telegram");
  };

  const handleEmail = () => {
    window.location.href = buildEmailShareUrl(content);
    logChannel("email");
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shareLink.shareUrl);
      toast.success("Link kopyalandı.");
      logChannel("copy");
    } catch {
      toast.error("Link kopyalanamadı.");
    }
  };

  const handleNative = async () => {
    try {
      await shareNative(content);
      logChannel("native");
    } catch (error) {
      if (error instanceof Error && error.name === "AbortError") return;
      toast.error("Paylaşım tamamlanamadı.");
    }
  };

  const buttonClass = compact
    ? "inline-flex h-9 items-center gap-1.5 rounded-lg border border-line bg-white px-3 text-xs font-semibold text-slate-600 transition hover:border-violet-brand hover:text-violet-brand"
    : "inline-flex h-10 items-center justify-center gap-2 rounded-lg border border-line bg-white px-4 text-sm font-semibold text-slate-600 transition hover:border-violet-brand hover:bg-violet-soft/30 hover:text-violet-brand";

  return (
    <div className="space-y-3">
      <h4 className="text-sm font-bold text-ink">Paylaş</h4>
      <div className={`flex flex-wrap gap-2 ${compact ? "" : "sm:grid sm:grid-cols-2"}`}>
        {showNative ? (
          <button className={`${buttonClass} border-violet-ring bg-violet-soft/40 text-violet-brand sm:col-span-2`} type="button" onClick={() => void handleNative()}>
            <Share2 className="h-4 w-4" />
            Paylaş
          </button>
        ) : null}

        {showDesktopChannels ? (
          <>
            <button className={`${buttonClass} ${!compact ? "text-[#25D366] hover:border-[#25D366] hover:text-[#25D366]" : ""}`} type="button" onClick={handleWhatsApp}>
              <WhatsAppIcon />
              WhatsApp
            </button>
            <button className={`${buttonClass} ${!compact ? "text-[#229ED9] hover:border-[#229ED9] hover:text-[#229ED9]" : ""}`} type="button" onClick={handleTelegram}>
              <Send className="h-4 w-4" />
              Telegram
            </button>
            <button className={buttonClass} type="button" onClick={handleEmail}>
              <Mail className="h-4 w-4" />
              Email
            </button>
          </>
        ) : null}

        <button className={`${buttonClass} ${!compact && !showDesktopChannels ? "sm:col-span-2" : ""}`} type="button" onClick={() => void handleCopy()}>
          <Copy className="h-4 w-4" />
          Linki Kopyala
        </button>
      </div>
    </div>
  );
}

function WhatsAppIcon() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.435 9.884-9.881 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z" />
    </svg>
  );
}
