import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listPins, listRecent, type PinItem, type RecentItem } from "../services/navigationService";

type QuickAccessProps = {
  collapsed?: boolean;
};

export function SidebarQuickAccess({ collapsed = false }: QuickAccessProps) {
  const [pins, setPins] = useState<PinItem[]>([]);
  const [recent, setRecent] = useState<RecentItem[]>([]);

  useEffect(() => {
    void listPins()
      .then((items) => setPins(items.slice(0, 5)))
      .catch(() => setPins([]));
    void listRecent()
      .then((response) => {
        const merged = [
          ...response.documents,
          ...response.folders,
          ...response.smartCollections,
          ...response.integrations,
          ...response.settingsPages,
        ]
          .sort((a, b) => new Date(b.openedAt).getTime() - new Date(a.openedAt).getTime())
          .slice(0, 5);
        setRecent(merged);
      })
      .catch(() => setRecent([]));
  }, []);

  if (collapsed) return null;

  return (
    <div className="space-y-3 border-t border-line pt-4">
      <div className="px-3">
        <p className="text-[11px] font-semibold uppercase tracking-wide text-slate-400">Hızlı Erişim</p>
      </div>

      {pins.length > 0 ? (
        <div className="space-y-0.5 px-1">
          <p className="px-2 text-[10px] font-bold uppercase tracking-wide text-slate-400">Sabitlenenler</p>
          {pins.map((pin) => (
            <Link
              key={pin.id}
              to={pin.path}
              className="flex h-8 items-center rounded-lg px-2 text-xs font-semibold text-slate-600 transition hover:bg-violet-soft/50 hover:text-ink"
              title={pin.label}
            >
              <span className="truncate">{pin.label}</span>
            </Link>
          ))}
        </div>
      ) : null}

      {recent.length > 0 ? (
        <div className="space-y-0.5 px-1">
          <p className="px-2 text-[10px] font-bold uppercase tracking-wide text-slate-400">Son Açılanlar</p>
          {recent.map((item) => (
            <Link
              key={`${item.entityType}-${item.entityId}`}
              to={item.path}
              className="flex h-8 items-center rounded-lg px-2 text-xs font-semibold text-slate-600 transition hover:bg-violet-soft/50 hover:text-ink"
              title={item.label}
            >
              <span className="truncate">{item.label}</span>
            </Link>
          ))}
        </div>
      ) : null}

      <Link to="/recent" className="mx-2 block rounded-lg px-2 py-1.5 text-xs font-semibold text-violet-brand transition hover:bg-violet-soft/50">
        Tümünü gör
      </Link>
    </div>
  );
}
