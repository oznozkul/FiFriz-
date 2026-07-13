import { ChevronRight } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { publishContextMenuAction } from "../../contextMenu/events";
import type { ContextMenuContext, ContextMenuItem, ContextMenuState } from "../../contextMenu/types";

type ContextMenuProps = {
  state: ContextMenuState;
  onClose: () => void;
};

type FlatItem = {
  item: ContextMenuItem;
  depth: number;
  parentId?: string;
};

function flattenItems(items: ContextMenuItem[], depth = 0, parentId?: string): FlatItem[] {
  const flat: FlatItem[] = [];
  for (const item of items) {
    if (item.divider) {
      flat.push({ item, depth, parentId });
      continue;
    }
    flat.push({ item, depth, parentId });
  }
  return flat;
}

function getSelectableItems(items: ContextMenuItem[]): ContextMenuItem[] {
  return items.filter((item) => !item.divider && !item.disabled);
}

export function ContextMenu({ state, onClose }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const [openSubmenuId, setOpenSubmenuId] = useState<string | null>(null);
  const selectableItems = getSelectableItems(state.items);
  const flatItems = flattenItems(state.items);

  const handleSelect = useCallback(
    (item: ContextMenuItem) => {
      if (item.disabled || item.divider) return;
      if (item.children && item.children.length > 0) {
        setOpenSubmenuId(item.id);
        return;
      }
      if (item.action) {
        publishContextMenuAction({ action: item.action, context: state.context });
      }
      onClose();
    },
    [onClose, state.context],
  );

  useEffect(() => {
    setActiveIndex(0);
    setOpenSubmenuId(null);
  }, [state.x, state.y, state.items]);

  useEffect(() => {
    const handlePointerDown = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        onClose();
      }
    };
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    const handleScroll = () => onClose();

    window.addEventListener("mousedown", handlePointerDown);
    window.addEventListener("keydown", handleEscape);
    window.addEventListener("scroll", handleScroll, true);
    return () => {
      window.removeEventListener("mousedown", handlePointerDown);
      window.removeEventListener("keydown", handleEscape);
      window.removeEventListener("scroll", handleScroll, true);
    };
  }, [onClose]);

  useEffect(() => {
    const menu = menuRef.current;
    if (!menu) return;

    const rect = menu.getBoundingClientRect();
    const padding = 8;
    let left = state.x;
    let top = state.y;

    if (left + rect.width > window.innerWidth - padding) {
      left = window.innerWidth - rect.width - padding;
    }
    if (top + rect.height > window.innerHeight - padding) {
      top = window.innerHeight - rect.height - padding;
    }

    menu.style.left = `${Math.max(padding, left)}px`;
    menu.style.top = `${Math.max(padding, top)}px`;
  }, [state.x, state.y, state.items, openSubmenuId]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (selectableItems.length === 0) return;

      if (event.key === "ArrowDown") {
        event.preventDefault();
        setActiveIndex((index) => (index + 1) % selectableItems.length);
      }
      if (event.key === "ArrowUp") {
        event.preventDefault();
        setActiveIndex((index) => (index - 1 + selectableItems.length) % selectableItems.length);
      }
      if (event.key === "Enter") {
        event.preventDefault();
        const item = selectableItems[activeIndex];
        if (item) handleSelect(item);
      }
      if (event.key === "ArrowRight") {
        const item = selectableItems[activeIndex];
        if (item?.children?.length) {
          event.preventDefault();
          setOpenSubmenuId(item.id);
        }
      }
      if (event.key === "ArrowLeft" && openSubmenuId) {
        event.preventDefault();
        setOpenSubmenuId(null);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [activeIndex, handleSelect, openSubmenuId, selectableItems]);

  return createPortal(
    <div
      ref={menuRef}
      className="fixed z-[110] min-w-[220px] overflow-visible rounded-xl border border-line bg-white py-1.5 shadow-2xl"
      style={{ left: state.x, top: state.y }}
      role="menu"
      aria-label="Bağlam menüsü"
    >
      {flatItems.map(({ item }) => {
        if (item.divider) {
          return <div key={item.id} className="my-1 border-t border-line" role="separator" />;
        }

        const selectableIndex = selectableItems.indexOf(item);
        const isActive = selectableIndex === activeIndex;
        const Icon = item.icon;
        const hasChildren = Boolean(item.children?.length);
        const isSubmenuOpen = openSubmenuId === item.id;

        return (
          <div key={item.id} className="relative">
            <button
              className={`flex w-full items-center gap-2.5 px-3 py-2 text-left text-sm font-semibold transition duration-150 disabled:cursor-not-allowed disabled:opacity-50 ${
                item.danger
                  ? `text-red-600 hover:bg-red-50 ${isActive ? "bg-red-50" : ""}`
                  : `text-slate-700 hover:bg-violet-soft/60 ${isActive ? "bg-violet-soft/70" : ""}`
              }`}
              type="button"
              role="menuitem"
              disabled={item.disabled}
              onMouseEnter={() => {
                if (selectableIndex >= 0) setActiveIndex(selectableIndex);
                if (hasChildren) setOpenSubmenuId(item.id);
                else setOpenSubmenuId(null);
              }}
              onClick={() => handleSelect(item)}
            >
              {Icon ? <Icon className="h-4 w-4 shrink-0 opacity-80" /> : <span className="h-4 w-4 shrink-0" />}
              <span className="min-w-0 flex-1 truncate">{item.label}</span>
              {item.shortcut ? <span className="ml-4 text-[11px] font-medium text-slate-400">{item.shortcut}</span> : null}
              {hasChildren ? <ChevronRight className="h-4 w-4 shrink-0 text-slate-400" /> : null}
            </button>

            {hasChildren && isSubmenuOpen && item.children ? (
              <div className="absolute left-full top-0 z-[120] ml-1 min-w-[200px] overflow-hidden rounded-xl border border-line bg-white py-1.5 shadow-2xl">
                {item.children.map((child) => {
                  if (child.divider) {
                    return <div key={child.id} className="my-1 border-t border-line" role="separator" />;
                  }
                  const ChildIcon = child.icon;
                  return (
                    <button
                      key={child.id}
                      className={`flex w-full items-center gap-2.5 px-3 py-2 text-left text-sm font-semibold transition duration-150 disabled:cursor-not-allowed disabled:opacity-50 ${
                        child.danger ? "text-red-600 hover:bg-red-50" : "text-slate-700 hover:bg-violet-soft/60"
                      }`}
                      type="button"
                      role="menuitem"
                      disabled={child.disabled}
                      onClick={() => {
                        if (child.disabled || !child.action) return;
                        publishContextMenuAction({ action: child.action, context: state.context });
                        onClose();
                      }}
                    >
                      {ChildIcon ? <ChildIcon className="h-4 w-4 shrink-0 opacity-80" /> : <span className="h-4 w-4 shrink-0" />}
                      <span className="min-w-0 flex-1 truncate">{child.label}</span>
                      {child.shortcut ? <span className="ml-4 text-[11px] font-medium text-slate-400">{child.shortcut}</span> : null}
                    </button>
                  );
                })}
              </div>
            ) : null}
          </div>
        );
      })}
    </div>,
    document.body,
  );
}

export type ContextMenuTriggerOptions = {
  items: ContextMenuItem[];
  context: ContextMenuContext;
};
