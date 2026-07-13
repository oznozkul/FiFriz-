import { ChevronLeft, ChevronRight, RefreshCw } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { EmptyState, IconButton, LoadingState, PageTitle, PrimaryButton } from "../components/common";
import {
  formatNotificationTime,
  getNotificationLink,
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
  notificationTypeLabels,
  type NotificationItem,
} from "../services/notificationService";
import { toast } from "../services/toast";

export function NotificationsPage() {
  const [items, setItems] = useState<NotificationItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isMarkingAll, setIsMarkingAll] = useState(false);
  const navigate = useNavigate();

  const totalPages = useMemo(() => Math.max(1, Math.ceil(totalCount / pageSize)), [totalCount, pageSize]);

  const fetchNotifications = async (targetPage = page) => {
    setIsLoading(true);
    try {
      const response = await listNotifications({
        unreadOnly,
        page: targetPage,
        pageSize,
      });
      setItems(response.items);
      setTotalCount(response.totalCount);
      setPage(response.page);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Bildirimler yüklenemedi.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void fetchNotifications(1);
  }, [unreadOnly]);

  const handleMarkAllRead = async () => {
    setIsMarkingAll(true);
    try {
      await markAllNotificationsRead();
      await fetchNotifications(page);
      toast.success("Tüm bildirimler okundu olarak işaretlendi.");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Bildirimler okundu olarak işaretlenemedi.");
    } finally {
      setIsMarkingAll(false);
    }
  };

  const handleNotificationClick = async (notification: NotificationItem) => {
    try {
      if (!notification.isRead) {
        await markNotificationRead(notification.id);
        setItems((current) => current.map((item) => (item.id === notification.id ? { ...item, isRead: true } : item)));
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Bildirim okundu olarak işaretlenemedi.");
      return;
    }

    const link = getNotificationLink(notification);
    if (link) {
      navigate(link);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageTitle title="Bildirimler" subtitle="Sizi ilgilendiren son olayları takip edin." />
        <div className="flex items-center gap-2">
          <IconButton aria-label="Yenile" onClick={() => void fetchNotifications(page)}>
            <RefreshCw className="h-4 w-4" />
          </IconButton>
          <PrimaryButton className="h-10 px-3 text-sm" disabled={isMarkingAll} onClick={() => void handleMarkAllRead()}>
            Tümünü Okundu Yap
          </PrimaryButton>
        </div>
      </div>

      <section className="panel p-4">
        <label className="inline-flex items-center gap-2 text-sm font-semibold text-slate-600">
          <input
            checked={unreadOnly}
            className="h-4 w-4 rounded border-line text-violet-brand focus:ring-violet-brand"
            type="checkbox"
            onChange={(event) => setUnreadOnly(event.target.checked)}
          />
          Sadece okunmamış
        </label>
      </section>

      {isLoading ? (
        <LoadingState variant="panel" label="Bildirimler yükleniyor..." />
      ) : items.length === 0 ? (
        <EmptyState
          title="Henüz bildirimin yok."
          description={unreadOnly ? "Okunmamış bildirim bulunmuyor." : "Yeni paylaşım, yorum veya versiyon olayları burada görünecek."}
        />
      ) : (
        <section className="panel divide-y divide-line overflow-hidden">
          {items.map((notification) => (
            <button
              key={notification.id}
              type="button"
              className={`w-full px-5 py-4 text-left transition hover:bg-violet-soft/30 ${notification.isRead ? "bg-white" : "bg-violet-soft/40"}`}
              onClick={() => void handleNotificationClick(notification)}
            >
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-sm font-bold text-ink">{notification.title}</span>
                <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-500">
                  {notificationTypeLabels[notification.type] ?? notification.type}
                </span>
                {!notification.isRead ? (
                  <span className="rounded-full bg-violet-soft px-2 py-0.5 text-xs font-semibold text-violet-brand">Yeni</span>
                ) : null}
              </div>
              <p className="mt-2 text-sm text-slate-600">{notification.message}</p>
              <p className="mt-2 text-xs font-medium text-slate-400">{formatNotificationTime(notification.createdAt)}</p>
            </button>
          ))}
        </section>
      )}

      {!isLoading && totalCount > pageSize ? (
        <div className="flex items-center justify-between">
          <p className="text-sm font-semibold text-slate-500">
            Sayfa {page} / {totalPages}
          </p>
          <div className="flex items-center gap-2">
            <IconButton aria-label="Önceki sayfa" disabled={page <= 1} onClick={() => void fetchNotifications(page - 1)}>
              <ChevronLeft className="h-4 w-4" />
            </IconButton>
            <IconButton aria-label="Sonraki sayfa" disabled={page >= totalPages} onClick={() => void fetchNotifications(page + 1)}>
              <ChevronRight className="h-4 w-4" />
            </IconButton>
          </div>
        </div>
      ) : null}
    </div>
  );
}
