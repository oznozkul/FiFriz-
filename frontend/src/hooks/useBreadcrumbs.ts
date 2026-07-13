import { useMemo } from "react";
import { useLocation } from "react-router-dom";
import type { BreadcrumbItem } from "../components/common/Breadcrumb";
import { useBreadcrumbContext } from "../context/BreadcrumbContext";
import { routeMetadata, smartCollectionTitles } from "../data/routeMetadata";
import { settingsNavItems } from "../data/settingsNav";
import { useAuth } from "../services/AuthContext";
import { getFolderNavLabel } from "../hooks/useFolderLabels";

function matchRouteMeta(pathname: string, tenantType?: string | null) {
  if (pathname === "/folders") {
    return { label: getFolderNavLabel(tenantType), section: "Organizasyon" };
  }

  if (routeMetadata[pathname]) return routeMetadata[pathname];

  if (pathname.startsWith("/collections/")) {
    const key = pathname.split("/")[2] ?? "";
    return {
      label: smartCollectionTitles[key] ?? "Koleksiyon",
      section: "Belgeler",
      parentPath: "/collections",
    };
  }

  if (pathname.startsWith("/settings/")) {
    const settingsItem = settingsNavItems.find((item) => item.to === pathname);
    if (settingsItem) {
      return { label: settingsItem.label, section: "Ayarlar", parentPath: "/settings/profile" };
    }
  }

  return null;
}

export function useBreadcrumbs() {
  const { pathname } = useLocation();
  const breadcrumbContext = useBreadcrumbContext();
  const { activeTenant } = useAuth();

  return useMemo(() => {
    if (breadcrumbContext?.items.length) {
      return breadcrumbContext.items;
    }

    const items: BreadcrumbItem[] = [];
    const meta = matchRouteMeta(pathname, activeTenant?.type);

    if (pathname.startsWith("/documents/") && pathname !== "/documents") {
      items.push({ label: "Belgeler", href: "/documents" });
      return items;
    }

    if (meta?.section && meta.section !== meta.label) {
      if (meta.section === "Belgeler") {
        items.push({ label: "Belgeler", href: "/documents" });
      } else if (meta.section === "Ayarlar") {
        items.push({ label: "Ayarlar", href: "/settings/profile" });
      } else if (meta.section === "Yapay Zeka") {
        items.push({ label: "Yapay Zeka", href: "/ai-assistant" });
      } else if (meta.section === "Organizasyon") {
        items.push({ label: "Organizasyon", href: "/folders" });
      } else {
        items.push({ label: meta.section });
      }
    }

    if (pathname.startsWith("/collections/")) {
      items.push({ label: "Akıllı Koleksiyonlar", href: "/collections" });
      if (meta) items.push({ label: meta.label, href: pathname });
      return items;
    }

    if (pathname.startsWith("/settings/") && meta) {
      items.push({ label: meta.label, href: pathname });
      return items;
    }

    if (meta) {
      items.push({ label: meta.label, href: pathname });
    }

    return items;
  }, [activeTenant?.type, breadcrumbContext?.items, pathname]);
}
