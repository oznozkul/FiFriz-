import { ShieldCheck } from "lucide-react";
import { useAuth } from "../../services/AuthContext";
import { EmptyState } from "../../components/common";
import { SettingsSection } from "./shared";
import { roleLabels } from "./constants";

export function RolesPermissionsSettingsPage() {
  const { activeTenant } = useAuth();
  const isOrganization = activeTenant?.type === "organization";

  if (!isOrganization) {
    return (
      <SettingsSection title="Roller & Yetkiler" description="Workspace erişim rollerini yönetin.">
        <EmptyState
          title="Kurumsal workspace özelliği"
          description="Rol ve yetki yönetimi kurumsal workspace'lerde kullanılabilir."
        />
      </SettingsSection>
    );
  }

  return (
    <SettingsSection title="Roller & Yetkiler" description="Kurumsal workspace rolleri ve yetki matrisi.">
      <div className="rounded-xl border border-dashed border-line bg-slate-50/60 px-6 py-12 text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-violet-soft text-violet-brand">
          <ShieldCheck className="h-6 w-6" />
        </div>
        <h2 className="text-base font-bold text-ink">Yakında</h2>
        <p className="mx-auto mt-2 max-w-md text-sm text-slate-500">
          Detaylı rol ve yetki yönetimi yakında eklenecek. Şu an aktif rolünüz:{" "}
          <span className="font-semibold text-violet-brand">{roleLabels[activeTenant?.role ?? "member"]}</span>
        </p>
      </div>
    </SettingsSection>
  );
}
