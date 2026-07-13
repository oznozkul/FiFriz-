export type ShareChannel = "whatsapp" | "telegram" | "email" | "copy" | "native";

type ShareContent = {
  documentName: string;
  shareUrl: string;
  expiresAt?: string | null;
};

export function isMobileBrowser() {
  if (typeof window === "undefined") return false;
  return window.matchMedia("(max-width: 768px)").matches || /Android|iPhone|iPad|iPod/i.test(navigator.userAgent);
}

export function canUseNativeShare() {
  return typeof navigator !== "undefined" && typeof navigator.share === "function" && isMobileBrowser();
}

export function buildWhatsAppShareUrl({ documentName, shareUrl }: ShareContent) {
  const message = `Seninle bir belge paylaşıldı: ${documentName}\nLink: ${shareUrl}`;
  return `https://wa.me/?text=${encodeURIComponent(message)}`;
}

export function buildTelegramShareUrl({ documentName, shareUrl }: ShareContent) {
  const text = `Seninle bir belge paylaşıldı: ${documentName}`;
  return `https://t.me/share/url?url=${encodeURIComponent(shareUrl)}&text=${encodeURIComponent(text)}`;
}

export function buildEmailShareUrl({ documentName, shareUrl, expiresAt }: ShareContent) {
  const subject = `Belge paylaşımı: ${documentName}`;
  const expiryNote = expiresAt
    ? `\n\nBu link ${formatExpiryDate(expiresAt)} tarihine kadar geçerlidir.`
    : "\n\nBu linkin süresi varsa belirtilen tarihe kadar geçerlidir.";

  const body = `Seninle bir belge paylaşıldı.

Belge: ${documentName}
Link: ${shareUrl}${expiryNote}`;

  return `mailto:?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
}

export function openShareWindow(url: string) {
  window.open(url, "_blank", "noopener,noreferrer");
}

export async function shareNative({ documentName, shareUrl }: ShareContent) {
  await navigator.share({
    title: `Belge paylaşımı: ${documentName}`,
    text: `Seninle bir belge paylaşıldı: ${documentName}`,
    url: shareUrl,
  });
}

function formatExpiryDate(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(new Date(value));
}
