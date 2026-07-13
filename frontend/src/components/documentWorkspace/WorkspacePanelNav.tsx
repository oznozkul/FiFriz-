import type { WorkspacePanelSection } from "./documentWorkspaceUtils";
import { WORKSPACE_ADVANCED_NAV, WORKSPACE_AI_NAV, WORKSPACE_MAIN_NAV } from "./workspacePanelConfig";

type WorkspacePanelNavProps = {
  activeSection?: WorkspacePanelSection;
  folderLabel: string;
  onSelect: (section: WorkspacePanelSection) => void;
};

export function WorkspacePanelNav({ activeSection, folderLabel, onSelect }: WorkspacePanelNavProps) {
  const mainNav = WORKSPACE_MAIN_NAV.map((item) => (item.id === "folder" ? { ...item, label: folderLabel } : item));

  return (
    <div className="shrink-0 border-b border-line bg-white">
      <div className="space-y-3 px-3 py-3">
        <NavGroup label="Belge" items={mainNav} activeSection={activeSection} onSelect={onSelect} />
        <NavGroup label="Yapay Zeka" items={WORKSPACE_AI_NAV} activeSection={activeSection} onSelect={onSelect} />
        <NavGroup label="Gelişmiş" items={WORKSPACE_ADVANCED_NAV} activeSection={activeSection} onSelect={onSelect} />
      </div>
    </div>
  );
}

function NavGroup({
  label,
  items,
  activeSection,
  onSelect,
}: {
  label: string;
  items: { id: WorkspacePanelSection; label: string }[];
  activeSection?: WorkspacePanelSection;
  onSelect: (section: WorkspacePanelSection) => void;
}) {
  return (
    <div>
      <p className="mb-1.5 px-1 text-[10px] font-semibold uppercase tracking-wide text-slate-400">{label}</p>
      <div className="flex flex-wrap gap-1.5">
        {items.map((item) => {
          const isActive = activeSection === item.id;
          return (
            <button
              key={item.id}
              type="button"
              className={`shrink-0 rounded-lg px-2.5 py-1.5 text-xs font-semibold transition focus:outline-none focus:ring-4 focus:ring-violet-ring/70 ${
                isActive ? "bg-violet-soft text-violet-brand ring-1 ring-violet-ring" : "text-slate-600 hover:bg-white hover:text-ink"
              }`}
              onClick={() => onSelect(item.id)}
            >
              {item.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}
