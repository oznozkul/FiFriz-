import { HelpCircle, LogOut, RefreshCw, Settings, User } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../../services/AuthContext";
import { toast } from "../../services/toast";

type UserMenuProps = {
  align?: "left" | "right";
};

export function UserMenu({ align = "right" }: UserMenuProps) {
  const { user, activeTenant, logout } = useAuth();
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

  if (!user) return null;

  const initials = user.fullName
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part: string) => part[0]?.toUpperCase() ?? "")
    .join("");

  return (
    <div ref={wrapperRef} className="relative">
      <button
        type="button"
        className="flex items-center gap-2 rounded-lg border border-line bg-white px-2 py-1.5 transition hover:border-violet-ring hover:bg-violet-soft/30"
        aria-expanded={isOpen}
        aria-haspopup="menu"
        onClick={() => setIsOpen((open) => !open)}
      >
        <span className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-slate-700 to-violet-brand text-xs font-bold text-white">
          {initials || "U"}
        </span>
        <span className="hidden max-w-[8rem] truncate text-left text-sm font-semibold text-ink md:block">{user.fullName}</span>
      </button>

      {isOpen ? (
        <div
          className={`absolute top-[calc(100%+0.5rem)] z-50 w-56 overflow-hidden rounded-xl border border-line bg-white shadow-panel ${align === "right" ? "right-0" : "left-0"}`}
          role="menu"
        >
          <div className="border-b border-line px-4 py-3">
            <p className="truncate text-sm font-semibold text-ink">{user.fullName}</p>
            <p className="truncate text-xs text-slate-500">{user.email}</p>
            {activeTenant ? <p className="mt-1 truncate text-xs font-medium text-violet-brand">{activeTenant.name}</p> : null}
          </div>
          <div className="p-1">
            <MenuLink icon={User} label="Profil" to="/settings/profile" onNavigate={() => setIsOpen(false)} />
            <MenuButton
              icon={RefreshCw}
              label="Workspace Değiştir"
              onClick={() => {
                setIsOpen(false);
                toast.info("Workspace değiştirme yakında eklenecek.");
              }}
            />
            <MenuLink icon={Settings} label="Ayarlar" to="/settings" onNavigate={() => setIsOpen(false)} />
            <MenuButton
              icon={HelpCircle}
              label="Yardım"
              onClick={() => {
                setIsOpen(false);
                toast.info("Yardım merkezi yakında eklenecek.");
              }}
            />
            <button
              type="button"
              className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm font-semibold text-rose-600 transition hover:bg-rose-50"
              role="menuitem"
              onClick={() => {
                setIsOpen(false);
                logout();
              }}
            >
              <LogOut className="h-4 w-4" />
              Çıkış
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function MenuLink({
  icon: Icon,
  label,
  to,
  onNavigate,
}: {
  icon: typeof User;
  label: string;
  to: string;
  onNavigate: () => void;
}) {
  return (
    <Link
      className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-semibold text-slate-700 transition hover:bg-violet-soft/50 hover:text-violet-brand"
      role="menuitem"
      to={to}
      onClick={onNavigate}
    >
      <Icon className="h-4 w-4" />
      {label}
    </Link>
  );
}

function MenuButton({
  icon: Icon,
  label,
  onClick,
}: {
  icon: typeof User;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm font-semibold text-slate-700 transition hover:bg-violet-soft/50 hover:text-violet-brand"
      role="menuitem"
      onClick={onClick}
    >
      <Icon className="h-4 w-4" />
      {label}
    </button>
  );
}
