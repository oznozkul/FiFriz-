import { useCallback, useRef, useState, type KeyboardEvent, type MouseEvent, type TouchEvent } from "react";
import { ContextMenu, type ContextMenuTriggerOptions } from "../components/common/ContextMenu";
import type { ContextMenuState } from "../contextMenu/types";

const LONG_PRESS_MS = 500;
const MOVE_TOLERANCE_PX = 12;

export function useContextMenu() {
  const [menuState, setMenuState] = useState<ContextMenuState | null>(null);
  const longPressTimerRef = useRef<number | null>(null);
  const touchStartRef = useRef<{ x: number; y: number } | null>(null);

  const closeMenu = useCallback(() => {
    setMenuState(null);
  }, []);

  const openMenu = useCallback((x: number, y: number, options: ContextMenuTriggerOptions) => {
    if (options.items.length === 0) return;
    setMenuState({
      x,
      y,
      items: options.items,
      context: options.context,
    });
  }, []);

  const clearLongPress = useCallback(() => {
    if (longPressTimerRef.current !== null) {
      window.clearTimeout(longPressTimerRef.current);
      longPressTimerRef.current = null;
    }
    touchStartRef.current = null;
  }, []);

  const bindContextMenu = useCallback(
    (options: ContextMenuTriggerOptions) => {
      const openAt = (x: number, y: number) => openMenu(x, y, options);

      return {
        onContextMenu: (event: MouseEvent) => {
          event.preventDefault();
          event.stopPropagation();
          openAt(event.clientX, event.clientY);
        },
        onKeyDown: (event: KeyboardEvent) => {
          const isContextMenuKey = event.key === "ContextMenu";
          const isShiftF10 = event.key === "F10" && event.shiftKey;
          if (!isContextMenuKey && !isShiftF10) return;

          event.preventDefault();
          event.stopPropagation();
          const target = event.currentTarget as HTMLElement;
          const rect = target.getBoundingClientRect();
          openAt(rect.left + rect.width / 2, rect.top + rect.height / 2);
        },
        onTouchStart: (event: TouchEvent) => {
          const touch = event.touches[0];
          if (!touch) return;
          touchStartRef.current = { x: touch.clientX, y: touch.clientY };
          clearLongPress();
          longPressTimerRef.current = window.setTimeout(() => {
            openAt(touch.clientX, touch.clientY);
            clearLongPress();
          }, LONG_PRESS_MS);
        },
        onTouchMove: (event: TouchEvent) => {
          const start = touchStartRef.current;
          const touch = event.touches[0];
          if (!start || !touch) return;
          const deltaX = Math.abs(touch.clientX - start.x);
          const deltaY = Math.abs(touch.clientY - start.y);
          if (deltaX > MOVE_TOLERANCE_PX || deltaY > MOVE_TOLERANCE_PX) {
            clearLongPress();
          }
        },
        onTouchEnd: () => {
          clearLongPress();
        },
        onTouchCancel: () => {
          clearLongPress();
        },
      };
    },
    [clearLongPress, openMenu],
  );

  const menu = menuState ? <ContextMenu state={menuState} onClose={closeMenu} /> : null;

  return {
    menu,
    menuState,
    openMenu,
    closeMenu,
    bindContextMenu,
  };
}

export type ContextMenuTriggerBindings = ReturnType<ReturnType<typeof useContextMenu>["bindContextMenu"]>;
