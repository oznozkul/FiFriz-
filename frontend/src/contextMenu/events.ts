import { useEffect } from "react";
import type { ContextMenuActionDetail } from "./types";

export const CONTEXT_MENU_ACTION_EVENT = "context-menu:action";

export function publishContextMenuAction(detail: ContextMenuActionDetail) {
  window.dispatchEvent(new CustomEvent<ContextMenuActionDetail>(CONTEXT_MENU_ACTION_EVENT, { detail }));
}

export function useContextMenuActionHandler(
  handler: (detail: ContextMenuActionDetail) => void,
  enabled = true,
) {
  useEffect(() => {
    if (!enabled) return;

    const listener = (event: Event) => {
      handler((event as CustomEvent<ContextMenuActionDetail>).detail);
    };

    window.addEventListener(CONTEXT_MENU_ACTION_EVENT, listener);
    return () => window.removeEventListener(CONTEXT_MENU_ACTION_EVENT, listener);
  }, [enabled, handler]);
}
