import { ChevronLeft, ChevronRight, X } from "lucide-react";
import type { ComponentType } from "react";
import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { BrandMark, IconButton } from "../components/common";
import {
  aiNavGroup,
  dashboardNavItem,
  integrationsNavItem,
  productNavGroups,
  settingsNavItem,
} from "../data/sidebarNav";
import { useFolderLabels } from "../hooks/useFolderLabels";
import { useAuth } from "../services/AuthContext";
import { SidebarQuickAccess } from "./SidebarQuickAccess";
import { SidebarUserCard } from "./SidebarUserCard";

const collapsedStorageKey = "fifriz_sidebar_collapsed";

type SidebarProps = {
  isOpen?: boolean;
  onClose?: () => void;
};

const navLinkBase =
  "flex h-9 w-full items-center gap-3 rounded-lg border-l-2 px-3 text-left text-sm font-medium transition";
const navLinkActive = "border-violet-brand bg-violet-soft text-violet-brand";
const navLinkInactive = "border-transparent text-slate-600 hover:border-violet-ring/60 hover:bg-violet-soft/50 hover:text-ink";

function NavLinkItem({
  path,
  label,
  icon: IconComponent,
  end,
  roles,
  pathname,
  onClose,
  collapsed,
}: {
  path: string;
  label: string;
  icon: ComponentType<{ className?: string }>;
  end?: boolean;
  roles?: Array<"owner" | "admin" | "member">;
  pathname: string;
  onClose?: () => void;
  collapsed?: boolean;
}) {
  const { activeTenant } = useAuth();
  const folderLabels = useFolderLabels();

  if (roles && roles.length > 0) {
    if (!activeTenant || !roles.includes(activeTenant.role)) return null;
  }

  const displayLabel = path === "/folders" ? folderLabels.navLabel : label;
  const active = end ? pathname === path : pathname === path || pathname.startsWith(`${path}/`);

  return (
    <Link
      to={path}
      onClick={() => onClose?.()}
      title={collapsed ? displayLabel : undefined}
      aria-current={active ? "page" : undefined}
      className={`${navLinkBase} ${active ? navLinkActive : navLinkInactive} ${collapsed ? "justify-center border-l-0 px-0" : ""}`}
    >
      <IconComponent className={`h-4 w-4 shrink-0 ${active ? "text-violet-brand" : "text-slate-500"}`} />
      {!collapsed ? displayLabel : null}
    </Link>
  );
}

function NavGroupSection({
  group,
  pathname,
  onClose,
  groupKey,
  collapsed,
}: {
  group: (typeof productNavGroups)[number];
  pathname: string;
  onClose?: () => void;
  groupKey: string;
  collapsed?: boolean;
}) {
  return (
    <div key={groupKey}>
      {!collapsed && group.label ? (
        <p className="mb-1.5 px-3 text-[11px] font-semibold uppercase tracking-wide text-slate-400">{group.label}</p>
      ) : null}
      <div className="space-y-0.5">
        {group.items.map((item, itemIndex) => (
          <NavLinkItem
            key={`${groupKey}-${item.path}-${itemIndex}`}
            {...item}
            pathname={pathname}
            onClose={onClose}
            collapsed={collapsed}
          />
        ))}
      </div>
    </div>
  );
}

export function Sidebar({ isOpen = false, onClose }: SidebarProps) {
  const { pathname } = useLocation();
  const SettingsIcon = settingsNavItem.icon;
  const settingsActive = pathname === "/settings" || pathname.startsWith("/settings/");
  const DashboardIcon = dashboardNavItem.icon;
  const IntegrationsIcon = integrationsNavItem.icon;
  const [collapsed, setCollapsed] = useState(() => {
    if (typeof window === "undefined") return false;
    return window.localStorage.getItem(collapsedStorageKey) === "1";
  });

  useEffect(() => {
    window.localStorage.setItem(collapsedStorageKey, collapsed ? "1" : "0");
  }, [collapsed]);

  const [isLargeScreen, setIsLargeScreen] = useState(() =>
    typeof window !== "undefined" ? window.matchMedia("(min-width: 1024px)").matches : true,
  );

  useEffect(() => {
    const media = window.matchMedia("(min-width: 1024px)");
    const handleChange = () => setIsLargeScreen(media.matches);
    handleChange();
    media.addEventListener("change", handleChange);
    return () => media.removeEventListener("change", handleChange);
  }, []);

  const navCollapsed = collapsed && isLargeScreen;
  const sidebarWidth = navCollapsed ? "w-64 lg:w-[4.5rem]" : "w-64";

  return (
    <>
      <div
        className={`fixed inset-0 z-40 bg-slate-900/30 transition lg:hidden ${isOpen ? "opacity-100" : "pointer-events-none opacity-0"}`}
        onClick={onClose}
      />
      <aside
        className={`fixed inset-y-0 left-0 z-50 flex ${sidebarWidth} shrink-0 flex-col border-r border-line bg-white p-4 transition-all duration-200 lg:translate-x-0 ${
          isOpen ? "translate-x-0" : "max-lg:-translate-x-full"
        }`}
        data-sidebar-collapsed={navCollapsed ? "true" : "false"}
      >
        <div className={`mb-5 flex items-center gap-3 ${navCollapsed ? "justify-center px-0" : "px-2"}`}>
          <BrandMark />
          {!navCollapsed ? <span className="text-base font-bold text-ink">FiFriz</span> : null}
          <IconButton className={`${navCollapsed ? "" : "ml-auto"} h-8 w-8 lg:hidden`} aria-label="Menüyü kapat" onClick={onClose}>
            <X className="h-4 w-4" />
          </IconButton>
          <IconButton
            className={`hidden h-8 w-8 lg:inline-flex ${navCollapsed ? "" : "ml-auto"}`}
            aria-label={navCollapsed ? "Kenar çubuğunu genişlet" : "Kenar çubuğunu daralt"}
            onClick={() => setCollapsed((value) => !value)}
          >
            {navCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </IconButton>
        </div>

        <nav className="min-h-0 flex-1 space-y-5 overflow-y-auto scrollbar-hidden">
          <div className="space-y-0.5">
            <NavLinkItem
              path={dashboardNavItem.path}
              label={dashboardNavItem.label}
              icon={DashboardIcon}
              end={dashboardNavItem.end}
              pathname={pathname}
              onClose={onClose}
              collapsed={navCollapsed}
            />
          </div>

          {productNavGroups.map((group, index) => (
            <NavGroupSection
              key={`product-${group.label ?? index}`}
              group={group}
              groupKey={`product-${index}`}
              pathname={pathname}
              onClose={onClose}
              collapsed={navCollapsed}
            />
          ))}

          <div className="space-y-0.5">
            <NavLinkItem
              path={integrationsNavItem.path}
              label={integrationsNavItem.label}
              icon={IntegrationsIcon}
              end={integrationsNavItem.end}
              pathname={pathname}
              onClose={onClose}
              collapsed={navCollapsed}
            />
          </div>

          <NavGroupSection group={aiNavGroup} groupKey="ai" pathname={pathname} onClose={onClose} collapsed={navCollapsed} />
        </nav>

        <SidebarQuickAccess collapsed={navCollapsed} />

        <div className="mt-4 shrink-0 space-y-2 border-t border-line pt-4">
          <Link
            to={settingsNavItem.path}
            onClick={() => onClose?.()}
            title={navCollapsed ? settingsNavItem.label : undefined}
            aria-current={settingsActive ? "page" : undefined}
            className={`${navLinkBase} ${settingsActive ? navLinkActive : navLinkInactive} ${navCollapsed ? "justify-center border-l-0 px-0" : ""}`}
          >
            <SettingsIcon className={`h-4 w-4 shrink-0 ${settingsActive ? "text-violet-brand" : "text-slate-500"}`} />
            {!navCollapsed ? settingsNavItem.label : null}
          </Link>
          {!navCollapsed ? <SidebarUserCard /> : null}
        </div>
      </aside>
    </>
  );
}
