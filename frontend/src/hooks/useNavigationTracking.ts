import { useEffect } from "react";
import { useLocation } from "react-router-dom";
import { routeMetadata, smartCollectionTitles } from "../data/routeMetadata";
import { settingsNavGroups } from "../data/settingsNav";
import { trackVisit } from "../services/navigationService";

export function useNavigationTracking() {
  const { pathname } = useLocation();

  useEffect(() => {
    if (pathname.startsWith("/settings/")) {
      const item = settingsNavGroups.flatMap((group) => group.items).find((entry) => entry.to === pathname);
      if (!item) return;
      void trackVisit({
        entityType: "settings_page",
        entityId: pathname.replace("/settings/", ""),
        label: item.label,
        path: pathname,
      }).catch(() => undefined);
      return;
    }

    if (pathname.startsWith("/collections/")) {
      const key = pathname.split("/")[2];
      if (!key) return;
      void trackVisit({
        entityType: "smart_collection",
        entityId: key,
        label: smartCollectionTitles[key] ?? "Koleksiyon",
        path: pathname,
      }).catch(() => undefined);
    }
  }, [pathname]);
}

export function usePageVisitLabel(pathname: string) {
  if (routeMetadata[pathname]) return routeMetadata[pathname].label;
  return undefined;
}
