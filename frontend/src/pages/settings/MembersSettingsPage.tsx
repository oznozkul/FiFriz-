import { useCallback, useEffect, useState } from "react";
import { EmptyState, PageState } from "../../components/common";
import { useAuth } from "../../services/AuthContext";
import { listTenantMembers } from "../../services/tenantService";
import type { TenantMember } from "../../services/tenantService";
import { roleLabels } from "./constants";
import { SettingsSection } from "./shared";

export function MembersSettingsPage() {
  const { activeTenant } = useAuth();
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const isOrganization = activeTenant?.type === "organization";

  useEffect(() => {
    if (!isOrganization) {
      setIsLoading(false);
      return;
    }

    let cancelled = false;
    setIsLoading(true);
    setLoadError(null);
    listTenantMembers()
      .then((data) => {
        if (!cancelled) setMembers(data);
      })
      .catch((error) => {
        if (!cancelled) {
          setMembers([]);
          setLoadError(error instanceof Error ? error.message : "Üyeler yüklenemedi.");
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [isOrganization]);

  const retryLoad = useCallback(() => {
    if (!isOrganization) return;
    setIsLoading(true);
    setLoadError(null);
    listTenantMembers()
      .then(setMembers)
      .catch((error) => {
        setMembers([]);
        setLoadError(error instanceof Error ? error.message : "Üyeler yüklenemedi.");
      })
      .finally(() => setIsLoading(false));
  }, [isOrganization]);

  if (!isOrganization) {
    return (
      <SettingsSection title="Üyeler" description="Workspace üyelerini görüntüleyin ve yönetin.">
        <EmptyState
          title="Kurumsal workspace özelliği"
          description="Bu özellik kurumsal workspace için kullanılabilir. Bireysel hesaplarda üye yönetimi bulunmaz."
        />
      </SettingsSection>
    );
  }

  return (
    <SettingsSection title="Üyeler" description="Kurumsal workspace üyeleri ve rolleri.">
      <PageState
        isLoading={isLoading}
        error={loadError}
        isEmpty={members.length === 0}
        variant="inline"
        loadingLabel="Üyeler yükleniyor..."
        errorTitle="Üyeler yüklenemedi"
        onRetry={retryLoad}
        emptyTitle="Üye bulunamadı"
        emptyDescription="Bu workspace'te henüz üye listesi boş."
      >
        <div className="overflow-hidden rounded-lg border border-line">
          <table className="hidden w-full text-left text-sm md:table">
            <thead className="bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
              <tr>
                <th className="px-4 py-3">Ad Soyad</th>
                <th className="px-4 py-3">E-posta</th>
                <th className="px-4 py-3">Rol</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {members.map((member) => (
                <tr key={member.userId} className="bg-white">
                  <td className="px-4 py-3 font-semibold text-ink">{member.fullName}</td>
                  <td className="px-4 py-3 text-slate-600">{member.email}</td>
                  <td className="px-4 py-3">
                    <span className="rounded-md bg-violet-soft px-2 py-1 text-xs font-bold text-violet-brand">
                      {roleLabels[member.role] ?? member.role}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="divide-y divide-line md:hidden">
            {members.map((member) => (
              <article key={member.userId} className="space-y-2 bg-white px-4 py-4">
                <p className="font-semibold text-ink">{member.fullName}</p>
                <p className="text-sm text-slate-600">{member.email}</p>
                <span className="inline-flex rounded-md bg-violet-soft px-2 py-1 text-xs font-bold text-violet-brand">
                  {roleLabels[member.role] ?? member.role}
                </span>
              </article>
            ))}
          </div>
        </div>
      </PageState>
    </SettingsSection>
  );
}
