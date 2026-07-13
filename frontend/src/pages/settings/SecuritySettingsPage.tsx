import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { PrimaryButton, Spinner } from "../../components/common";
import { changePassword } from "../../services/authService";
import { toast } from "../../services/toast";
import { minPasswordLength } from "./constants";
import { SettingsField, SettingsSection, settingsInputClassName } from "./shared";

export function SecuritySettingsPage() {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isChanging, setIsChanging] = useState(false);

  const passwordsMatch = newPassword === confirmPassword;
  const canSubmit = currentPassword.length > 0 && newPassword.length >= minPasswordLength && passwordsMatch;

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!canSubmit) {
      toast.error("Şifreler eşleşmeli ve en az 8 karakter olmalıdır.");
      return;
    }

    setIsChanging(true);
    try {
      await changePassword({ currentPassword, newPassword });
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      toast.success("Şifreniz değiştirildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Şifre değiştirilemedi.");
    } finally {
      setIsChanging(false);
    }
  };

  return (
    <div className="space-y-5">
      <SettingsSection title="Güvenlik" description="Hesap şifrenizi yönetin ve güvenlik tercihlerinizi kontrol edin.">
        <form className="mx-auto max-w-xl space-y-4" onSubmit={handleSubmit}>
          <SettingsField label="Mevcut şifre">
            <input
              className={settingsInputClassName()}
              type="password"
              autoComplete="current-password"
              value={currentPassword}
              onChange={(event) => setCurrentPassword(event.target.value)}
            />
          </SettingsField>
          <SettingsField label="Yeni şifre">
            <input
              className={settingsInputClassName()}
              type="password"
              autoComplete="new-password"
              placeholder="En az 8 karakter"
              value={newPassword}
              onChange={(event) => setNewPassword(event.target.value)}
            />
          </SettingsField>
          <SettingsField label="Yeni şifre tekrar">
            <input
              className={settingsInputClassName()}
              type="password"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(event) => setConfirmPassword(event.target.value)}
            />
          </SettingsField>
          {confirmPassword && !passwordsMatch ? <p className="text-sm text-rose-600">Şifreler eşleşmiyor.</p> : null}
          <div className="pt-2">
            <PrimaryButton type="submit" disabled={isChanging || !canSubmit}>
              {isChanging ? <Spinner size="sm" label="Güncelleniyor..." className="text-white" /> : "Şifreyi güncelle"}
            </PrimaryButton>
          </div>
        </form>
      </SettingsSection>

      <SettingsSection title="Şifrenizi mi unuttunuz?" description="Oturum açmadan şifre sıfırlama bağlantısı almak için e-posta adresinizi kullanabilirsiniz.">
        <Link className="inline-flex text-sm font-bold text-violet-brand hover:text-violet-hover" to="/forgot-password">
          Şifre sıfırlama sayfasına git
        </Link>
      </SettingsSection>
    </div>
  );
}
