import { Outlet, useLocation } from "react-router-dom";
import { PageTitle } from "../components/common";
import { SettingsNavigationCard } from "./SettingsNavigationCard";

export function SettingsLayout() {
  const location = useLocation();

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
      <PageTitle title="Ayarlar" subtitle="Profil, güvenlik, workspace ve faturalandırma tercihlerinizi yönetin." />
      <SettingsNavigationCard />
      <div className="min-w-0">
        <Outlet key={location.key} />
      </div>
    </div>
  );
}
