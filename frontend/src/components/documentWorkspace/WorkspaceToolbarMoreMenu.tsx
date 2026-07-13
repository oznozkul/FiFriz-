import {
  Activity,
  Bot,
  GitBranch,
  HeartPulse,
  Maximize2,
  Minimize2,
  MoreHorizontal,
  PanelRightClose,
  PanelRightOpen,
  Sparkles,
  Workflow,
} from "lucide-react";
import { useEffect, useRef, useState, type ReactNode } from "react";
import type { WorkspacePanelSection } from "./documentWorkspaceUtils";

type WorkspaceToolbarMoreMenuProps = {
  isFullscreen: boolean;
  isPanelOpen: boolean;
  onToggleFullscreen: () => void;
  onTogglePanel: () => void;
  onOpenMobilePanel: () => void;
  onFocusSection: (section: WorkspacePanelSection) => void;
};

type MenuItem = {
  label: string;
  icon: ReactNode;
  onClick: () => void;
  show?: boolean;
};

export function WorkspaceToolbarMoreMenu({
  isFullscreen,
  isPanelOpen,
  onToggleFullscreen,
  onTogglePanel,
  onOpenMobilePanel,
  onFocusSection,
}: WorkspaceToolbarMoreMenuProps) {
  const [open, setOpen] = useState(false);
  const [isDesktop, setIsDesktop] = useState(() =>
    typeof window !== "undefined" ? window.matchMedia("(min-width: 1024px)").matches : true,
  );
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const mediaQuery = window.matchMedia("(min-width: 1024px)");
    const onChange = () => setIsDesktop(mediaQuery.matches);
    onChange();
    mediaQuery.addEventListener("change", onChange);
    return () => mediaQuery.removeEventListener("change", onChange);
  }, []);

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (event: MouseEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    window.addEventListener("mousedown", onPointerDown);
    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("mousedown", onPointerDown);
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  const items: MenuItem[] = [
    { label: "Onay süreci", icon: <Workflow className="h-4 w-4" />, onClick: () => focus("workflow") },
    { label: "Versiyonlar", icon: <GitBranch className="h-4 w-4" />, onClick: () => focus("versions") },
    { label: "Aktivite", icon: <Activity className="h-4 w-4" />, onClick: () => focus("activity") },
    { label: "Belge sağlığı", icon: <HeartPulse className="h-4 w-4" />, onClick: () => focus("health") },
    { label: "AI asistan", icon: <Bot className="h-4 w-4" />, onClick: () => focus("ai") },
    { label: "AI özeti", icon: <Sparkles className="h-4 w-4" />, onClick: () => focus("summary") },
    {
      label: isFullscreen ? "Tam ekrandan çık" : "Tam ekran",
      icon: isFullscreen ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />,
      onClick: () => {
        setOpen(false);
        onToggleFullscreen();
      },
    },
    {
      label: isPanelOpen ? "Paneli gizle" : "Paneli göster",
      icon: isPanelOpen ? <PanelRightClose className="h-4 w-4" /> : <PanelRightOpen className="h-4 w-4" />,
      onClick: () => {
        setOpen(false);
        onTogglePanel();
      },
      show: isDesktop,
    },
    {
      label: "Bilgi paneli",
      icon: <PanelRightOpen className="h-4 w-4" />,
      onClick: () => {
        setOpen(false);
        onOpenMobilePanel();
      },
      show: !isDesktop,
    },
  ].filter((item) => item.show !== false);

  function focus(section: WorkspacePanelSection) {
    setOpen(false);
    onFocusSection(section);
  }

  return (
    <div ref={rootRef} className="relative">
      <button
        type="button"
        aria-label="Daha fazla"
        aria-expanded={open}
        className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-line bg-white text-slate-600 transition hover:border-violet-ring hover:bg-violet-soft/40 hover:text-violet-brand focus:outline-none focus:ring-4 focus:ring-violet-ring/70"
        onClick={() => setOpen((current) => !current)}
      >
        <MoreHorizontal className="h-4 w-4" />
      </button>
      {open ? (
        <div className="absolute right-0 top-[calc(100%+6px)] z-50 min-w-[200px] rounded-xl border border-line bg-white py-1 shadow-panel">
          {items.map((item) => (
            <button
              key={item.label}
              type="button"
              className="flex w-full items-center gap-2.5 px-3 py-2.5 text-left text-sm font-semibold text-slate-700 transition hover:bg-violet-soft/40 hover:text-ink"
              onClick={item.onClick}
            >
              <span className="text-slate-500">{item.icon}</span>
              {item.label}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
