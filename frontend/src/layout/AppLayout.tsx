import { useEffect, useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { CommandPalette } from "../components/common/CommandPalette";
import { UploadDocumentModal } from "../components/documents";
import { BreadcrumbProvider, useBreadcrumbContext } from "../context/BreadcrumbContext";
import { useCommandPalette } from "../hooks/useCommandPalette";
import { useNavigationTracking } from "../hooks/useNavigationTracking";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

function AppLayoutContent() {
  const location = useLocation();
  const breadcrumbContext = useBreadcrumbContext();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const commandPalette = useCommandPalette();

  useNavigationTracking();

  useEffect(() => {
    breadcrumbContext?.resetItems();
  }, [breadcrumbContext, location.pathname]);

  useEffect(() => {
    const syncCollapsed = () => {
      const aside = document.querySelector("aside[data-sidebar-collapsed]");
      setSidebarCollapsed(aside?.getAttribute("data-sidebar-collapsed") === "true");
    };
    syncCollapsed();
    const observer = new MutationObserver(syncCollapsed);
    const aside = document.querySelector("aside[data-sidebar-collapsed]");
    if (aside) observer.observe(aside, { attributes: true, attributeFilter: ["data-sidebar-collapsed"] });
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (!sidebarOpen) return;
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previous;
    };
  }, [sidebarOpen]);

  return (
    <div className="min-h-screen bg-white text-sm text-slate-700">
      <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      <div className={`flex min-h-screen min-w-0 flex-col transition-all duration-200 ${sidebarCollapsed ? "lg:pl-[4.5rem]" : "lg:pl-64"}`}>
        <Topbar onMenuClick={() => setSidebarOpen(true)} onCommandPaletteClick={commandPalette.open} />
        <main className="min-h-0 flex-1 bg-slate-50/70 px-4 pb-5 pt-4 sm:px-5 md:px-7 md:pb-8 md:pt-5">
          <Outlet key={location.key} />
        </main>
      </div>
      <UploadDocumentModal isOpen={uploadModalOpen} onClose={() => setUploadModalOpen(false)} />
      <CommandPalette isOpen={commandPalette.isOpen} onClose={commandPalette.close} onUploadClick={() => setUploadModalOpen(true)} />
    </div>
  );
}

export function AppLayout() {
  return (
    <BreadcrumbProvider>
      <AppLayoutContent />
    </BreadcrumbProvider>
  );
}
