type DashboardWelcomeHeaderProps = {
  fullName: string;
  tenantName: string;
  isPersonal: boolean;
};

export function DashboardWelcomeHeader({ fullName, tenantName, isPersonal }: DashboardWelcomeHeaderProps) {
  const firstName = fullName.trim().split(/\s+/)[0] || "Kullanıcı";

  return (
    <section className="panel rounded-2xl bg-gradient-to-br from-white via-white to-violet-soft/40 p-6 shadow-panel">
      <p className="text-xs font-semibold uppercase tracking-wide text-violet-brand">Hoş geldin</p>
      <h1 className="mt-2 text-2xl font-bold tracking-tight text-ink md:text-3xl">{firstName}</h1>
      <p className="mt-2 max-w-2xl text-sm leading-relaxed text-slate-600">
        {isPersonal
          ? `${tenantName} çalışma alanınızda belgelerinizi, tarihlerinizi ve favorilerinizi tek ekrandan yönetin.`
          : `${tenantName} için onaylar, aktiviteler ve paylaşımlarınızın özeti burada.`}
      </p>
    </section>
  );
}
