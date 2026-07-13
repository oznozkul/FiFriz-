import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";
import type { HybridSearchMatchType } from "../../types/hybridSearch";

export type PaletteItem = {
  key: string;
  group: string;
  label: string;
  subtitle?: string;
  snippet?: string;
  extension?: string;
  icon?: LucideIcon;
  kind: "navigate" | "action" | "search";
  path?: string;
  actionId?: string;
  pinEligible?: boolean;
  matchType?: HybridSearchMatchType;
  onSelect: () => void;
};

export function getPaletteGroupOrder(folderGroupLabel: string) {
  return [
    "Sabitlenenler",
    "Son Açılan Belgeler",
    "Son Aramalar",
    "Belgeler",
    "Paylaşılan Belgeler",
    folderGroupLabel,
    "Etiketler",
    "Akıllı Koleksiyonlar",
    "Entegrasyonlar",
    "Kullanıcılar",
    "Aksiyonlar",
    "AI Aksiyonları",
    "Ayarlar",
  ];
}

export function highlightText(text: string, query: string): ReactNode {
  const term = query.trim();
  if (term.length < 2) return text;

  const lowerText = text.toLocaleLowerCase("tr-TR");
  const lowerTerm = term.toLocaleLowerCase("tr-TR");
  const index = lowerText.indexOf(lowerTerm);
  if (index === -1) return text;

  return (
    <>
      {text.slice(0, index)}
      <mark className="rounded bg-violet-soft px-0.5 font-semibold text-violet-brand">{text.slice(index, index + term.length)}</mark>
      {text.slice(index + term.length)}
    </>
  );
}

export function filterCatalogByQuery<T extends { label: string }>(items: T[], query: string) {
  const normalized = query.trim().toLocaleLowerCase("tr-TR");
  if (!normalized) return items;
  return items.filter((item) => item.label.toLocaleLowerCase("tr-TR").includes(normalized));
}
