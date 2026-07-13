import { NavLink } from "react-router-dom";
import { settingsNavItems } from "../data/settingsNav";

export function SettingsNavigationCard() {
  return (
    <nav aria-label="Ayarlar menüsü" className="panel overflow-hidden p-2">
      <div className="flex gap-1 overflow-x-auto scrollbar-hidden">
        {settingsNavItems.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `inline-flex shrink-0 items-center gap-2 rounded-xl px-3 py-2.5 text-sm font-semibold transition focus:outline-none focus:ring-4 focus:ring-violet-ring/70 sm:px-4 ${
                  isActive
                    ? "bg-violet-soft text-violet-brand ring-1 ring-violet-ring"
                    : "text-slate-600 hover:bg-slate-50 hover:text-ink"
                }`
              }
            >
              <Icon className="h-4 w-4 shrink-0 opacity-90" />
              <span className="whitespace-nowrap">{item.label}</span>
            </NavLink>
          );
        })}
      </div>
    </nav>
  );
}
