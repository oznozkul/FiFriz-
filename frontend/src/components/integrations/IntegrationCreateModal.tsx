import { Plug, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import type { IntegrationProviderDescriptor } from "../../services/integrationService";
import { PrimaryButton, Spinner } from "../common";
import { defaultSyncFromDate } from "./imapFormShared";

type IntegrationCreateModalProps = {
  isOpen: boolean;
  provider: IntegrationProviderDescriptor | null;
  onClose: () => void;
  onSubmit: (input: {
    provider: string;
    displayName: string;
    emailAddress?: string;
    settings?: Record<string, unknown>;
    credentials?: { password: string };
  }) => Promise<void>;
};

export function IntegrationCreateModal({ isOpen, provider, onClose, onSubmit }: IntegrationCreateModalProps) {
  const [displayName, setDisplayName] = useState("");
  const [emailAddress, setEmailAddress] = useState("");
  const [password, setPassword] = useState("");
  const [syncFromDate, setSyncFromDate] = useState(defaultSyncFromDate());
  const [isSubmitting, setIsSubmitting] = useState(false);

  const isPasswordProvider = provider?.capabilities.supportsPassword ?? false;

  useEffect(() => {
    if (!isOpen) return;
    setDisplayName(provider?.displayName ?? "");
    setEmailAddress("");
    setPassword("");
    setSyncFromDate(defaultSyncFromDate());
  }, [isOpen, provider]);

  const canSubmit = useMemo(() => {
    if (isPasswordProvider) return emailAddress.trim() && password.trim() && syncFromDate.trim();
    if (!displayName.trim()) return false;
    return true;
  }, [displayName, isPasswordProvider, emailAddress, password, syncFromDate]);

  if (!isOpen || !provider) return null;

  const handleSubmit = async () => {
    if (!canSubmit) return;
    setIsSubmitting(true);
    try {
      const email = emailAddress.trim();
      await onSubmit({
        provider: provider.name,
        displayName: isPasswordProvider ? email : displayName.trim(),
        emailAddress: isPasswordProvider ? email : emailAddress.trim() || undefined,
        settings: isPasswordProvider ? { syncFromDate } : undefined,
        credentials: isPasswordProvider ? { password } : undefined,
      });
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[95] flex items-center justify-center bg-slate-900/40 p-4">
      <div className="flex max-h-[90vh] w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-line bg-white shadow-panel">
        <div className="flex shrink-0 items-center justify-between border-b border-line px-5 py-4">
          <h3 className="flex items-center gap-2 text-lg font-bold text-ink">
            <Plug className="h-5 w-5 text-violet-brand" />
            {provider.displayName} Bağla
          </h3>
          <button type="button" className="rounded-lg p-2 text-slate-400 hover:bg-slate-100 hover:text-slate-600" onClick={onClose}>
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
          {isPasswordProvider ? (
            <>
              <p className="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600">
                E-posta adresiniz ve şifreniz yeterli. Sunucu ayarları otomatik algılanır.
              </p>
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">E-posta adresi</span>
                <input
                  type="email"
                  className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand"
                  value={emailAddress}
                  onChange={(e) => setEmailAddress(e.target.value)}
                  autoComplete="email"
                />
              </label>
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">Şifre</span>
                <input
                  type="password"
                  className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                />
              </label>
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">Hangi tarihten itibaren</span>
                <input
                  type="date"
                  className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand"
                  value={syncFromDate}
                  max={new Date().toISOString().slice(0, 10)}
                  onChange={(e) => setSyncFromDate(e.target.value)}
                />
                <p className="text-xs text-slate-500">Bu tarihten sonraki e-postalar içe aktarılır.</p>
              </label>
            </>
          ) : (
            <>
              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">Görünen ad</span>
                <input className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand" value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
              </label>

              <label className="block space-y-2 text-sm">
                <span className="font-semibold text-slate-600">E-posta adresi</span>
                <input type="email" className="h-10 w-full rounded-lg border border-line px-3 outline-none focus:border-violet-brand" value={emailAddress} onChange={(e) => setEmailAddress(e.target.value)} />
              </label>
            </>
          )}
        </div>

        <div className="flex shrink-0 justify-end gap-2 border-t border-line px-5 py-4">
          <button type="button" className="h-10 rounded-lg border border-line px-4 text-sm font-semibold text-slate-600" onClick={onClose}>
            İptal
          </button>
          <PrimaryButton className="h-10" disabled={isSubmitting || !canSubmit} onClick={() => void handleSubmit()}>
            {isSubmitting ? <Spinner size="sm" className="text-white" /> : "Kaydet"}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
