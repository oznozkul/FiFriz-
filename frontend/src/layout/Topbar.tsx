import { Menu, Search } from "lucide-react";
import { NotificationBell } from "../components/notifications/NotificationBell";
import { Breadcrumb, IconButton, UserMenu } from "../components/common";
import { useBreadcrumbs } from "../hooks/useBreadcrumbs";

type TopbarProps = {
  onMenuClick?: () => void;
  onCommandPaletteClick?: () => void;
};

export function Topbar({ onMenuClick, onCommandPaletteClick }: TopbarProps) {
  const breadcrumbs = useBreadcrumbs();

  return (
    <header className="sticky top-0 z-30 flex h-16 shrink-0 items-center gap-2 border-b border-line bg-white px-3 sm:gap-3 sm:px-4 md:px-6 lg:px-8">
      <div className="flex min-w-0 flex-1 items-center gap-2">
        <IconButton className="shrink-0 lg:hidden" aria-label="Menüyü aç" onClick={onMenuClick}>
          <Menu className="h-4 w-4" />
        </IconButton>
        <Breadcrumb items={breadcrumbs} className="hidden min-w-0 sm:block" maxVisible={3} />
      </div>

      <div className="hidden min-w-0 flex-1 md:flex md:max-w-xl">
        <button
          type="button"
          className="flex h-10 w-full items-center gap-3 rounded-xl border border-line bg-slate-50/80 px-4 text-sm font-medium text-slate-500 transition hover:border-violet-ring hover:bg-white focus:outline-none focus:ring-4 focus:ring-violet-ring/70"
          onClick={onCommandPaletteClick}
        >
          <Search className="h-4 w-4 shrink-0" />
          <span className="min-w-0 flex-1 truncate text-left">Ara veya komut çalıştır...</span>
          <kbd className="hidden shrink-0 rounded border border-line bg-white px-2 py-0.5 text-[11px] font-semibold text-slate-400 lg:inline">Ctrl+K</kbd>
        </button>
      </div>

      <div className="flex shrink-0 items-center gap-1 sm:gap-2">
        <IconButton className="md:hidden" aria-label="Komut paletini aç" onClick={onCommandPaletteClick}>
          <Search className="h-4 w-4" />
        </IconButton>

        <NotificationBell />

        <div className="hidden h-8 w-px shrink-0 bg-line sm:block" aria-hidden="true" />

        <UserMenu />
      </div>
    </header>
  );
}
