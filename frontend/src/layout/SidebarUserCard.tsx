import { useAuth } from "../services/AuthContext";
import { roleLabels, tenantTypeLabels } from "../pages/settings/constants";

export function SidebarUserCard() {
  const { user, activeTenant } = useAuth();

  if (!user || !activeTenant) return null;

  const initials = user.fullName
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");

  return (
    <div className="rounded-lg border border-line bg-slate-50/80 p-3">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-slate-700 to-violet-brand text-xs font-bold text-white">
          {initials || "U"}
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-semibold text-ink">{user.fullName}</p>
          <p className="truncate text-xs font-medium text-violet-brand">{activeTenant.name}</p>
        </div>
      </div>
      <p className="mt-2 truncate text-[11px] text-slate-500">
        {tenantTypeLabels[activeTenant.type] ?? activeTenant.type} · {roleLabels[activeTenant.role] ?? activeTenant.role}
      </p>
    </div>
  );
}
