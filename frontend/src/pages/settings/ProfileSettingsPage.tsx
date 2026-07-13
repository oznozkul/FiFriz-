import { FormEvent, useEffect, useState } from "react";
import { PageLoader, PrimaryButton, Spinner } from "../../components/common";
import { updateProfile } from "../../services/authService";
import { useAuth } from "../../services/AuthContext";
import { toast } from "../../services/toast";
import { SettingsField, SettingsSection, settingsInputClassName } from "./shared";

export function ProfileSettingsPage() {
  const { user, refreshSession } = useAuth();
  const [fullName, setFullName] = useState(user?.fullName ?? "");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (user?.fullName) setFullName(user.fullName);
  }, [user?.fullName]);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!fullName.trim()) {
      toast.error("Ad soyad gereklidir.");
      return;
    }

    setIsSaving(true);
    try {
      await updateProfile({ fullName: fullName.trim() });
      await refreshSession();
      toast.success("Profil bilgileri kaydedildi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Profil güncellenemedi.");
    } finally {
      setIsSaving(false);
    }
  };

  if (!user) {
    return <PageLoader label="Profil yükleniyor..." className="border-0 shadow-none" />;
  }

  return (
    <SettingsSection title="Profil" description="Kişisel bilgilerinizi görüntüleyin ve güncelleyin.">
      <form className="mx-auto max-w-xl space-y-4" onSubmit={handleSubmit}>
        <SettingsField label="Ad Soyad">
          <input className={settingsInputClassName()} value={fullName} onChange={(event) => setFullName(event.target.value)} required />
        </SettingsField>
        <SettingsField label="E-posta" hint="E-posta adresi değiştirilemez.">
          <input className={settingsInputClassName(true)} value={user.email} readOnly />
        </SettingsField>
        <div className="pt-2">
          <PrimaryButton type="submit" disabled={isSaving}>
            {isSaving ? <Spinner size="sm" label="Kaydediliyor..." className="text-white" /> : "Değişiklikleri kaydet"}
          </PrimaryButton>
        </div>
      </form>
    </SettingsSection>
  );
}
