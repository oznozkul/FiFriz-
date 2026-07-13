import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { EmptyState, LoadingState, PageTitle } from "../components/common";
import { listRecent, type RecentItem, type RecentResponse } from "../services/navigationService";

export function RecentPage() {
  const [recent, setRecent] = useState<RecentResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    void listRecent()
      .then(setRecent)
      .catch(() => setRecent({ documents: [], folders: [], smartCollections: [], integrations: [], settingsPages: [] }))
      .finally(() => setIsLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <PageTitle title="Son Açılanlar" subtitle="Son ziyaret ettiğiniz belgeler, klasörler ve sayfalar." />

      {isLoading ? (
        <LoadingState variant="panel" label="Son açılanlar yükleniyor..." />
      ) : !recent ? (
        <EmptyState title="Son açılan öğe yok" description="Belgeler ve sayfalar açtıkça burada listelenecek." />
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          <RecentSection title="Son Açılan Belgeler" items={recent.documents} emptyLabel="Henüz belge açılmadı." />
          <RecentSection title="Son Açılan Klasörler" items={recent.folders} emptyLabel="Henüz klasör açılmadı." />
          <RecentSection title="Son Koleksiyonlar" items={recent.smartCollections} emptyLabel="Henüz koleksiyon açılmadı." />
          <RecentSection title="Son Entegrasyonlar" items={recent.integrations} emptyLabel="Henüz entegrasyon açılmadı." />
          <RecentSection title="Son Ayarlar" items={recent.settingsPages} emptyLabel="Henüz ayar sayfası açılmadı." />
        </div>
      )}
    </div>
  );
}

function RecentSection({ title, items, emptyLabel }: { title: string; items: RecentItem[]; emptyLabel: string }) {
  return (
    <section className="panel overflow-hidden">
      <div className="border-b border-line px-5 py-4">
        <h2 className="text-base font-bold text-ink">{title}</h2>
      </div>
      {items.length === 0 ? (
        <p className="px-5 py-6 text-sm text-slate-500">{emptyLabel}</p>
      ) : (
        <ul className="divide-y divide-line">
          {items.map((item) => (
            <li key={`${item.entityType}-${item.entityId}`}>
              <Link to={item.path} className="flex items-center justify-between gap-3 px-5 py-3 transition hover:bg-violet-soft/35">
                <span className="truncate font-semibold text-ink">{item.label}</span>
                <span className="shrink-0 text-xs text-slate-400">{formatDate(item.openedAt)}</span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("tr-TR", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
