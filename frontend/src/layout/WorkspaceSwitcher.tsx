import { Check, ChevronDown } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../services/AuthContext";
import { toast } from "../services/toast";
import type { AuthTenant } from "../types/auth";
import { roleLabels, tenantTypeLabels } from "../pages/settings/constants";

type WorkspaceSwitcherProps = {
  onNavigate?: () => void;
};

export function WorkspaceSwitcher({ onNavigate }: WorkspaceSwitcherProps) {
  const { user, activeTenant, tenants } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      if (!wrapperRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    window.addEventListener("pointerdown", handlePointerDown);
    return () => window.removeEventListener("pointerdown", handlePointerDown);
  }, []);

  if (!user || !activeTenant) return null;

  const initials = user.fullName
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");

  const handleSelect = (tenant: AuthTenant) => {
    if (tenant.id === activeTenant.id) {
      setIsOpen(false);
      return;
    }
    setIsOpen(false);
    toast.info("Workspace değiştirme yakında eklenecek.");
    onNavigate?.();
  };

  return (
    <div ref={wrapperRef} className="relative">
      <button
        type="button"
        className="flex w-full items-center gap-3 rounded-lg border border-line bg-white p-3 text-left transition hover:border-violet-ring"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        onClick={() => setIsOpen((open) => !open)}
      >
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-slate-700 to-violet-brand text-xs font-bold text-white">
          {initials || "U"}
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-semibold text-ink">{activeTenant.name}</p>
          <p className="truncate text-xs text-slate-500">{user.fullName}</p>
        </div>
        <ChevronDown className={`h-4 w-4 shrink-0 text-slate-400 transition ${isOpen ? "rotate-180" : ""}`} />
      </button>

      {isOpen ? (
        <div className="absolute bottom-[calc(100%+0.5rem)] left-0 right-0 z-50 overflow-hidden rounded-xl border border-line bg-white shadow-panel" role="listbox">
          <div className="border-b border-line px-3 py-2">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">Workspace</p>
          </div>
          <div className="max-h-56 overflow-y-auto p-1">
            {tenants.map((tenant) => {
              const isActive = tenant.id === activeTenant.id;
              return (
                <button
                  key={tenant.id}
                  type="button"
                  role="option"
                  aria-selected={isActive}
                  className={`flex w-full items-start gap-2 rounded-lg px-3 py-2.5 text-left transition ${
                    isActive ? "bg-violet-soft text-violet-brand" : "text-slate-700 hover:bg-slate-50"
                  }`}
                  onClick={() => handleSelect(tenant)}
                >
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-semibold">{tenant.name}</p>
                    <p className="mt-0.5 text-xs opacity-80">
                      {tenantTypeLabels[tenant.type] ?? tenant.type} · {roleLabels[tenant.role] ?? tenant.role}
                    </p>
                  </div>
                  {isActive ? <Check className="mt-0.5 h-4 w-4 shrink-0" /> : null}
                </button>
              );
            })}
          </div>
          <div className="border-t border-line p-1">
            <Link
              className="block rounded-lg px-3 py-2 text-sm font-semibold text-slate-600 transition hover:bg-slate-50 hover:text-violet-brand"
              to="/settings/workspace"
              onClick={() => {
                setIsOpen(false);
                onNavigate?.();
              }}
            >
              Workspace ayarları
            </Link>
          </div>
        </div>
      ) : null}
    </div>
  );
}
