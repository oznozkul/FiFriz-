import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";

export type BreadcrumbItem = {
  label: string;
  href?: string;
};

type BreadcrumbContextValue = {
  items: BreadcrumbItem[];
  setItems: (items: BreadcrumbItem[]) => void;
  resetItems: () => void;
};

const BreadcrumbContext = createContext<BreadcrumbContextValue | null>(null);

export function BreadcrumbProvider({ children }: { children: ReactNode }) {
  const [items, setItemsState] = useState<BreadcrumbItem[]>([]);

  const setItems = useCallback((next: BreadcrumbItem[]) => {
    setItemsState((current) => {
      if (
        current.length === next.length &&
        current.every((item, index) => item.label === next[index]?.label && item.href === next[index]?.href)
      ) {
        return current;
      }
      return next;
    });
  }, []);

  const resetItems = useCallback(() => {
    setItemsState((current) => (current.length === 0 ? current : []));
  }, []);

  const value = useMemo(
    () => ({
      items,
      setItems,
      resetItems,
    }),
    [items, resetItems, setItems],
  );

  return <BreadcrumbContext.Provider value={value}>{children}</BreadcrumbContext.Provider>;
}

export function useBreadcrumbContext() {
  return useContext(BreadcrumbContext);
}
