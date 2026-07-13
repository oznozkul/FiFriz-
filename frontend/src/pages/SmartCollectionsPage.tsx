import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { smartCollectionIcon } from "../components/smartCollections/smartCollectionIcons";
import { EmptyState, LoadingState, PageTitle, Button } from "../components/common";
import {
  buildSmartCollectionMenuItems,
  ContextMenuActions,
  getSmartCollectionMenuContext,
  useContextMenuActionHandler,
} from "../contextMenu";
import { useContextMenu } from "../hooks/useContextMenu";
import { useAuth } from "../services/AuthContext";
import { listSmartCollections, type SmartCollection } from "../services/smartCollectionService";
import { createPin } from "../services/navigationService";
import { toast } from "../services/toast";

export function SmartCollectionsPage() {
  const navigate = useNavigate();
  const { activeTenant } = useAuth();
  const role = activeTenant?.role ?? "member";
  const { menu, bindContextMenu } = useContextMenu();
  const [collections, setCollections] = useState<SmartCollection[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const fetchCollections = useCallback(async () => {
    setIsLoading(true);
    setLoadError(null);
    try {
      setCollections(await listSmartCollections());
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Akıllı koleksiyonlar yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchCollections();
  }, [fetchCollections]);

  useContextMenuActionHandler(
    useCallback(
      (detail) => {
        if (detail.context.entityType !== "smart-collection") return;
        const collection = detail.context.entity as SmartCollection;

        switch (detail.action) {
          case ContextMenuActions.smartCollection.refresh:
            void fetchCollections();
            toast.success("Koleksiyonlar yenilendi.");
            return;
          case ContextMenuActions.smartCollection.pin:
            void createPin({
              entityType: "smart_collection",
              entityId: collection.key,
              label: collection.title,
              path: `/collections/${collection.key}`,
            })
              .then(() => toast.success("Koleksiyon sabitlendi."))
              .catch((error) => toast.error(error instanceof Error ? error.message : "Sabitleme başarısız."));
            return;
          default:
            navigate(`/collections/${collection.key}`);
        }
      },
      [fetchCollections, navigate],
    ),
  );

  return (
    <div className="space-y-6">
      <PageTitle
        title="Akıllı Koleksiyonlar"
        subtitle="Belgelerinize özel hazır filtrelenmiş görünümler. Klasör değildir; belgeleri taşımaz."
      />

      {isLoading ? (
        <LoadingState variant="panel" label="Koleksiyonlar yükleniyor..." />
      ) : loadError ? (
        <EmptyState
          title="Koleksiyonlar yüklenemedi"
          description={loadError}
          action={
            <Button variant="secondary" onClick={() => void fetchCollections()}>
              Tekrar dene
            </Button>
          }
        />
      ) : collections.length === 0 ? (
        <EmptyState title="Akıllı koleksiyon bulunamadı" description="Hazır koleksiyonlar yüklendiğinde burada görünecek." />
      ) : (
        <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {collections.map((collection) => {
            const Icon = smartCollectionIcon(collection.icon);
            const menuProps = bindContextMenu({
              items: buildSmartCollectionMenuItems({ collection, role }),
              context: getSmartCollectionMenuContext(collection),
            });

            return (
              <div
                key={collection.key}
                className="panel group flex cursor-pointer flex-col p-5 transition hover:border-violet-ring hover:shadow-md"
                role="link"
                tabIndex={0}
                onClick={() => navigate(`/collections/${collection.key}`)}
                onKeyDown={(event) => {
                  menuProps.onKeyDown?.(event);
                  if (event.key === "Enter") navigate(`/collections/${collection.key}`);
                }}
                onContextMenu={menuProps.onContextMenu}
                onTouchStart={menuProps.onTouchStart}
                onTouchMove={menuProps.onTouchMove}
                onTouchEnd={menuProps.onTouchEnd}
                onTouchCancel={menuProps.onTouchCancel}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-violet-soft text-violet-brand transition group-hover:bg-violet-brand group-hover:text-white">
                    <Icon className="h-5 w-5" />
                  </div>
                  <span className="rounded-full border border-violet-200 bg-violet-soft px-2.5 py-0.5 text-xs font-bold text-violet-brand">
                    {collection.count}
                  </span>
                </div>
                <h3 className="mt-4 text-base font-bold text-ink">{collection.title}</h3>
                <p className="mt-2 flex-1 text-sm text-slate-600">{collection.description}</p>
                <p className="mt-4 text-sm font-semibold text-violet-brand group-hover:underline">Koleksiyonu aç</p>
              </div>
            );
          })}
        </section>
      )}

      {menu}
    </div>
  );
}
